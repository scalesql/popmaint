package app

import (
	"cmp"
	"context"
	"popmaint/pkg/config.go"
	"popmaint/pkg/maint"
	"popmaint/pkg/mssqlz"
	"popmaint/pkg/state"
	"slices"
	"time"

	"github.com/pkg/errors"
)

type Engine struct {
	out   mssqlz.Outputer
	st    *state.State
	start time.Time
}

func NewEngine(out mssqlz.Outputer, st *state.State) Engine {
	return Engine{
		out:   out,
		st:    st,
		start: time.Now(),
	}
}

func (engine *Engine) runPlan(ctx context.Context, plan config.Plan, noexec bool) int {
	// TODO sort the plan entries and run in order
	return engine.runCheckDB(ctx, plan, noexec)
}

func (engine *Engine) runCheckDB(ctx context.Context, plan config.Plan, noexec bool) int {
	var err error
	exitCode := 0
	timeLimit := time.Duration(plan.CheckDB.TimeLimit)
	engine.out.WriteStringf("%s: checkdb: time_limit: %s  noindex: %t  physical_only: %t  max_size_mb: %d  info_messages: %t", plan.Name, timeLimit, plan.CheckDB.NoIndex, plan.CheckDB.PhysicalOnly, plan.CheckDB.MaxSizeMB, plan.CheckDB.InfoMessages)

	// loop through all servers and get databases
	databases := make([]mssqlz.Database, 0)
	for _, fqdn := range plan.Servers {
		srv, err := mssqlz.GetServer(ctx, fqdn)
		if err != nil {
			engine.out.WriteError(errors.Wrap(err, fqdn))
			continue
		}
		dbs, err := mssqlz.OnlineDatabases(ctx, fqdn)
		if err != nil {
			engine.out.WriteError(errors.Wrap(err, fqdn))
			continue
		}
		databases = append(databases, dbs...)
		engine.out.WriteStringf("fqdn: %s  server: %s  databases: %d", fqdn, srv.ServerName, len(dbs))
	}
	// sort the databases, filter, and get state
	for i, db := range databases {
		tm, ok := engine.st.GetLastCheckDBDate(db)
		if ok {
			databases[i].LastDBCC = tm
		}
	}
	sortDatabasesForDBCC(databases)
	// sort.SliceStable(databases, func(i, j int) bool {
	// 	if databases[i].LastDBCC.Before(databases[j].LastDBCC) {
	// 		return true
	// 	}

	// 	return databases[i].DatabaseMB > databases[j].DatabaseMB
	// })
	// if !noexec, run each one
	var totals struct {
		count int
		size  int
	}
	start := time.Now()
	for _, db := range databases {
		if timeLimit.Seconds() > 0 {
			if time.Now().After(start.Add(timeLimit)) {
				engine.out.WriteStringf("%s: time_limit (%s) exceeded", plan.Name, timeLimit)
				break
			}
		}
		if plan.CheckDB.MaxSizeMB > 0 && db.DatabaseMB > plan.CheckDB.MaxSizeMB {
			continue
		}
		engine.out.WriteStringf("%s: database: %s  size_mb: %d  last_dbcc: %s", db.ServerName, db.DatabaseName, db.DatabaseMB, db.LastDBCC.Format("2006-01-02 15:04:05"))

		err = maint.CheckDB(ctx, engine.out, db.FQDN, db, plan, noexec)
		if err != nil {
			// Log the error and keep going
			exitCode = 1
			engine.out.WriteErrorf("%s: %s", db.ServerName, err.Error())
			continue // don't set the state
		}
		if !noexec { // if we really did it, save it
			err = engine.st.SetLastCheckDBDate(db)
			if err != nil {
				engine.out.WriteError(err)
				exitCode = 1
			}
			// TODO Log the run time for this: FQDN, server, database, size, duration (rounded to second)
		}
		totals.count++
		totals.size += db.DatabaseMB
	}
	engine.out.WriteStringf("totals: databases: %d  size_mb: %d  time=%s", totals.count, totals.size, time.Since(start).Round(1*time.Second))
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
