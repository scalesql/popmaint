package app

import (
	"context"
	"popmaint/pkg/config.go"
	"popmaint/pkg/maint"
	"popmaint/pkg/mssqlz"
	"popmaint/pkg/state"
	"sort"
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
	start := time.Now()
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
	sort.SliceStable(databases, func(i, j int) bool {
		if databases[i].LastDBCC.Before(databases[j].LastDBCC) {
			return true
		}

		return databases[i].DatabaseMB > databases[j].DatabaseMB
	})
	// if !noexec, run each one
	for _, db := range databases {
		if time.Now().After(start.Add(timeLimit)) {
			engine.out.WriteStringf("%s: time_limit (%s) exceeded", plan.Name, timeLimit)
			break
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
	}
	return exitCode
}

// // TODO: CheckDB only checks one mssqlz.Database row and writes the results
// func (ce *Engine) CheckDB(ctx context.Context, host string, plan config.Plan, noexec bool) error {
// 	pool, err := mssqlh.Open(host, "master")
// 	if err != nil {
// 		return err
// 	}
// 	defer pool.Close()

// 	databases, err := mssqlz.OnlineDatabases(ctx, host)
// 	if err != nil {
// 		return err
// 	}

// 	// Get the last CheckDB date from `state`.
// 	// Secondary replicas always report the value from the primary.
// 	// That means we will use our date.
// 	for i, db := range databases {
// 		tm, ok := ce.st.GetLastCheckDBDate(db)
// 		if ok {
// 			databases[i].LastDBCC = tm
// 		}
// 	}

// 	// sort the oldest first, and then the largest
// 	sort.SliceStable(databases, func(i, j int) bool {
// 		if databases[i].LastDBCC.Before(databases[j].LastDBCC) {
// 			return true
// 		}

// 		return databases[i].DatabaseMB > databases[j].DatabaseMB
// 	})

// 	for _, db := range databases {
// 		if plan.CheckDB.MaxSizeMB > 0 && db.DatabaseMB > plan.CheckDB.MaxSizeMB {
// 			continue
// 		}
// 		err = maint.CheckDB(ctx, ce.out, host, db, plan, noexec)
// 		if err != nil {
// 			// TODO keep going on error?
// 			return err
// 		}
// 		if !noexec { // if we really did it, save it
// 			err = ce.st.SetLastCheckDBDate(db)
// 			if err != nil {
// 				ce.out.WriteError(err)
// 				return err
// 			}
// 		}
// 	}
// 	return nil
// }

// // sort by oldest defrag, then largest size
// func sortDatabases(databases []mssqlz.Database) {
// 	sort.SliceStable(databases, func(i, j int) bool {
// 		if databases[i].LastDBCC.Before(databases[j].LastDBCC) {
// 			return true
// 		}

// 		return databases[i].DatabaseMB > databases[j].DatabaseMB
// 	})
// }
