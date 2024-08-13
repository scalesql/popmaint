package app

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/scalesql/popmaint/pkg/checkdbwriter"
	"github.com/scalesql/popmaint/pkg/config"
	"github.com/scalesql/popmaint/pkg/maint"
	"github.com/scalesql/popmaint/pkg/mssqlz"
	"github.com/scalesql/popmaint/pkg/px"
	"github.com/scalesql/popmaint/pkg/state"

	"github.com/pkg/errors"
)

type Engine struct {
	logger px.PX
	st     *state.State
	start  time.Time
}

func NewEngine(logger px.PX, st *state.State) Engine {
	return Engine{
		logger: logger,
		st:     st,
		start:  time.Now(),
	}
}

func (engine *Engine) runPlan(ctx context.Context, plan config.Plan, noexec bool) int {
	// TODO sort the plan entries and run in order
	// use participle to get all the sections in order
	return engine.runCheckDB(ctx, plan, noexec)
}

const (
	ActionCheckdb = "checkdb"
	ActionBackup  = "backup"
)

func (engine *Engine) runCheckDB(ctx context.Context, plan config.Plan, noexec bool) int {
	//child := engine.logger.With(slogxx.String("action", ActionCheckdb))
	child := engine.logger
	var err error
	exitCode := 0
	timeLimit := time.Duration(plan.CheckDB.TimeLimit)
	child.Info(
		fmt.Sprintf("CHECKDB: time: %s  max_size_mb: %d  no_index: %t  physical_only: %t",
			timeLimit, plan.CheckDB.MaxSizeMB, plan.CheckDB.NoIndex, plan.CheckDB.PhysicalOnly),
		"time_limit", timeLimit.String(),
		"checkdb.no_index", plan.CheckDB.NoIndex,
		"checkdb.physical_only", plan.CheckDB.PhysicalOnly,
		"checkdb.max_size_mb", plan.CheckDB.MaxSizeMB,
		"checkdb.info_messages", plan.CheckDB.InfoMessages,
	)

	if plan.CheckDB.EstimateOnly && !plan.CheckDB.InfoMessages {
		plan.CheckDB.InfoMessages = true
		child.Warn("'info_messages' set to 'true' to display estimates", "action", ActionCheckdb)
	}

	// loop through all servers and get databases
	databases := make([]mssqlz.Database, 0)
	dupecheck := mssqlz.NewDupeCheck()
	for _, fqdn := range plan.Servers {
		srv, err := mssqlz.GetServer(ctx, fqdn)
		if err != nil {
			child.Error(errors.Wrap(err, fqdn).Error(), "action", ActionCheckdb)
			continue
		}
		dupe := dupecheck.IsDupe(srv)
		if dupe {
			child.Warn(fmt.Sprintf("CHECKDB: duplicate: %s", srv.Path()))
			continue
		}
		dbs, err := mssqlz.OnlineDatabases(ctx, fqdn)
		if err != nil {
			child.Error(errors.Wrap(err, fqdn).Error(), "action", ActionCheckdb)
			continue
		}
		databases = append(databases, dbs...)
		size := 0
		for _, db := range dbs {
			size += db.DatabaseMB
		}
		child.Info(fmt.Sprintf("CHECKDB: %s:  server: %s  databases: %d  size_mb: %d", fqdn, srv.ServerName, len(dbs), size),
			"server", srv.ServerName,
			"databases", len(dbs))
	}

	// sort the databases, filter, and get state
	for i, db := range databases {
		tm, ok := engine.st.GetLastCheckDBDate(db)
		if ok {
			databases[i].LastDBCC = tm
		}
	}

	// filter databases
	var totals struct {
		count int
		size  int
	}

	filteredDatabases := make([]mssqlz.Database, 0, len(databases))
	for _, db := range databases {
		if plan.CheckDB.MaxSizeMB > 0 && db.DatabaseMB > plan.CheckDB.MaxSizeMB {
			continue
		}

		// if this database is in []excluded, just keep going
		if len(plan.CheckDB.Excluded) > 0 {
			if contains(plan.CheckDB.Excluded, db.DatabaseName) {
				continue
			}
		}
		// if this database isn't in []included, just keep going
		if len(plan.CheckDB.Included) > 0 {
			if !contains(plan.CheckDB.Included, db.DatabaseName) {
				continue
			}
		}

		// check the minimum interval
		if intervalTooEarly(db, plan.CheckDB.MinIntervalDays) {
			continue
		}

		filteredDatabases = append(filteredDatabases, db)
		totals.count++
		totals.size += db.DatabaseMB
	}
	child.Info(fmt.Sprintf("CHECKDB: (filtered): databases: %d  size_mb: %d", totals.count, totals.size),
		"databases", totals.count,
		"size_mb", totals.size,
	)
	// sort databases
	sortDatabasesForDBCC(filteredDatabases)

	totals.count = 0
	totals.size = 0

	start := time.Now()
	for _, db := range filteredDatabases {
		if timeLimit.Seconds() > 0 {
			if time.Now().After(start.Add(timeLimit)) {
				child.Warn(fmt.Sprintf("CHECKDB: %s: time_limit (%s) exceeded", plan.Name, timeLimit))
				break
			}
		}

		child.Info(fmt.Sprintf("CHECKDB: %s.%s (%d mb)  last_dbcc: %s",
			db.ServerName, db.DatabaseName, db.DatabaseMB, db.LastDBCC),
			"server", db.ServerName,
			"database", db.DatabaseName,
			"size_mb", db.DatabaseMB,
			"checkdb.last_dbcc", db.LastDBCC.Format(time.RFC3339),
		)

		// get the estimated tempdb space
		estimatePlan := plan
		estimatePlan.CheckDB.EstimateOnly = true
		aw := checkdbwriter.New()
		err = maint.CheckDB(ctx, child, db.FQDN, db, estimatePlan, false)
		if err != nil {
			child.Error(fmt.Errorf("CHECKDB estimate: %w", err).Error())
		} else {
			estimatedKB := aw.EstimateKB()
			if estimatedKB != 0 {
				db.EstimatedTempdb = estimatedKB / 1024
			}
		}
		//fmt.Printf("estimate: rows: %d  (%d KB)\n", len(aw.Messages()), aw.EstimateKB())
		t0 := time.Now()
		err = maint.CheckDB(ctx, child, db.FQDN, db, plan, noexec)
		if err != nil {
			// Log the error and keep going
			exitCode = 1
			child.Error(fmt.Sprintf("CHECKDB: %s: %s", db.ServerName, err.Error()))
			continue // so we don't set the state
		}
		if !noexec { // if we really did it, save it
			err = engine.st.SetLastCheckDB(db)
			if err != nil {
				child.Error(fmt.Errorf("CHECKDB: setlastcheckdb: %w", err).Error())
				exitCode = 1
			}
			// TODO Log the run time for this: FQDN, server, database, size, duration (rounded to second)
			if !plan.CheckDB.EstimateOnly {
				child.Info(fmt.Sprintf("CHECKDB: %s.%s size_mb=%d  duration=%s", db.ServerName, db.DatabaseName, db.DatabaseMB, time.Since(t0).Round(1*time.Second)),
					"size_mb", db.DatabaseMB,
					"duration", time.Since(t0).Round(1*time.Second).String(),
					"duration_sec", int(time.Since(t0).Round(1*time.Second).Seconds()),
					"checkdb", plan.CheckDB,
				)
			}
		}
		totals.count++
		totals.size += db.DatabaseMB
	}
	child.Info(fmt.Sprintf("CHECKDB: %d database(s) at %d mb in %s", totals.count, totals.size, time.Since(start).Round(1*time.Second).String()),
		"databases", totals.count,
		"size_mb", totals.size,
		"duration", time.Since(start).Round(1*time.Second).String(),
		"duration_sec", time.Since(start).Round(1*time.Second).Seconds(),
	)
	return exitCode
}

// sort by oldest defrag, then largest size
func sortDatabasesForDBCC(databases []mssqlz.Database) {
	slices.SortStableFunc(databases, func(a, b mssqlz.Database) int {
		return coalesce(
			cmp.Compare(a.LastDBCC.Unix(), b.LastDBCC.Unix()),     // ascending
			cmp.Compare(int64(b.DatabaseMB), int64(a.DatabaseMB)), // descending
		)
	})
}

func coalesce[T comparable](vals ...T) T {
	var zero T
	for _, val := range vals {
		if val != zero {
			return val
		}
	}
	return zero
}

func contains(list []string, value string) bool {
	for _, str := range list {
		if strings.EqualFold(str, value) {
			return true
		}
	}
	return false
}

// intervalTooEarly returns true if we are trying to run
// CheckDB too early
func intervalTooEarly(db mssqlz.Database, days int) bool {
	if days == 0 {
		return false
	}
	nextTime := db.LastDBCC.Add((time.Duration(days)*24 - 1) * time.Hour)
	return nextTime.After(time.Now())
}
