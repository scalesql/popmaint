package app

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"popmaint/pkg/config"
	"popmaint/pkg/maint"
	"popmaint/pkg/mssqlz"
	"popmaint/pkg/state"
	"slices"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type Engine struct {
	logger *slog.Logger
	st     *state.State
	start  time.Time
}

func NewEngine(logger *slog.Logger, st *state.State) Engine {
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
	child := engine.logger.With(slog.String("action", ActionCheckdb))
	var err error
	exitCode := 0
	timeLimit := time.Duration(plan.CheckDB.TimeLimit)
	child.Info(
		fmt.Sprintf("checkdb: time: %s  max_size_mb: %d  no_index: %t  physical_only: %t",
			timeLimit, plan.CheckDB.MaxSizeMB, plan.CheckDB.NoIndex, plan.CheckDB.PhysicalOnly),
		"time_limit", timeLimit.String(),
		slog.Group(ActionCheckdb,
			"no_index", plan.CheckDB.NoIndex,
			"physical_only", plan.CheckDB.PhysicalOnly,
			"max_size_mb", plan.CheckDB.MaxSizeMB,
			"info_messages", plan.CheckDB.InfoMessages,
		),
	)

	if plan.CheckDB.EstimateOnly && !plan.CheckDB.InfoMessages {
		plan.CheckDB.InfoMessages = true
		child.Warn("WARN: 'info_messages' set to 'true' to display estimates", slog.String("action", ActionCheckdb))
	}

	// loop through all servers and get databases
	databases := make([]mssqlz.Database, 0)
	for _, fqdn := range plan.Servers {
		srv, err := mssqlz.GetServer(ctx, fqdn)
		if err != nil {
			child.Error(errors.Wrap(err, fqdn).Error(), slog.String("action", ActionCheckdb))
			continue
		}
		dbs, err := mssqlz.OnlineDatabases(ctx, fqdn)
		if err != nil {
			child.Error(errors.Wrap(err, fqdn).Error(), slog.String("action", ActionCheckdb))
			continue
		}
		databases = append(databases, dbs...)
		size := 0
		for _, db := range dbs {
			size += db.DatabaseMB
		}
		child.Info(fmt.Sprintf("%s:  server: %s  databases: %d  size_mb: %d", fqdn, srv.ServerName, len(dbs), size),
			slog.String("server", srv.ServerName),
			slog.Int("databases", len(dbs)))
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

		filteredDatabases = append(filteredDatabases, db)
		totals.count++
		totals.size += db.DatabaseMB
	}
	child.Info(fmt.Sprintf("checkdb (filtered): databases: %d  size_mb: %d", totals.count, totals.size),
		slog.Int("databases", totals.count),
		slog.Int("size_mb", totals.size),
	)
	// sort databases
	sortDatabasesForDBCC(filteredDatabases)

	totals.count = 0
	totals.size = 0

	start := time.Now()
	for _, db := range filteredDatabases {
		if timeLimit.Seconds() > 0 {
			if time.Now().After(start.Add(timeLimit)) {
				child.Warn(fmt.Sprintf("%s: time_limit (%s) exceeded", plan.Name, timeLimit))
				break
			}
		}

		child.Info(fmt.Sprintf("checkdb: %s.%s (%d mb)", db.ServerName, db.DatabaseName, db.DatabaseMB),
			slog.String("server", db.ServerName),
			slog.String("database", db.DatabaseName),
			slog.Int("size_mb", db.DatabaseMB),
			slog.Group(ActionCheckdb,
				slog.String("last_dbcc", db.LastDBCC.Format(time.RFC3339)),
			),
		)

		err = maint.CheckDB(ctx, child, db.FQDN, db, plan, noexec)
		if err != nil {
			// Log the error and keep going
			exitCode = 1
			child.Error(fmt.Sprintf("%s: %s", db.ServerName, err.Error()))
			continue // so we don't set the state
		}
		if !noexec { // if we really did it, save it
			err = engine.st.SetLastCheckDB(db)
			if err != nil {
				child.Error(fmt.Errorf("setlastcheckdb: %w", err).Error())
				exitCode = 1
			}
			// TODO Log the run time for this: FQDN, server, database, size, duration (rounded to second)
		}
		totals.count++
		totals.size += db.DatabaseMB
	}
	child.Info(fmt.Sprintf("checkdb: %d database(s) at %d mb in %s", totals.count, totals.size, time.Since(start).Round(1*time.Second).String()),
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
