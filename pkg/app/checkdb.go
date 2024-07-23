package app

import (
	"context"
	"popmaint/pkg/maint"
	"popmaint/pkg/mssqlz"
	"popmaint/pkg/state"
	"sort"

	"github.com/billgraziano/mssqlh"
)

type CheckDBEngine struct {
	out mssqlz.Outputer
	do  maint.CheckDBOptions
	st  *state.State
}

func NewCheckDBEngine(out mssqlz.Outputer, st *state.State, do maint.CheckDBOptions) CheckDBEngine {
	return CheckDBEngine{
		out: out,
		do:  do,
		st:  st,
	}
}

func (ce *CheckDBEngine) CheckDB(ctx context.Context, host string) error {
	pool, err := mssqlh.Open(host, "master")
	if err != nil {
		return err
	}
	defer pool.Close()

	databases, err := mssqlz.OnlineDatabases(ctx, host)
	if err != nil {
		return err
	}

	// Get the last CheckDB date from `state`.
	// Secondary replicas always report the value from the primary.
	// That means we will use our date.
	for i, db := range databases {
		tm, ok := ce.st.GetCheckDB(db)
		if ok {
			databases[i].LastDBCC = tm
		}
	}

	// sort the oldest first, and then the largest
	sort.SliceStable(databases, func(i, j int) bool {
		if databases[i].LastDBCC.Before(databases[j].LastDBCC) {
			return true
		}

		return databases[i].DatabaseMB > databases[j].DatabaseMB
	})

	for _, db := range databases {
		if ce.do.MaxSizeMB > 0 && db.DatabaseMB > ce.do.MaxSizeMB {
			continue
		}
		err = maint.CheckDB(ctx, ce.out, host, db, ce.do)
		if err != nil {
			// TODO keep going on error?
			return err
		}
		if !ce.do.NoExec { // if we really did it, save it
			err = ce.st.SaveCheckDB(db)
			if err != nil {
				ce.out.WriteError(err)
				return err
			}
		}
	}
	return nil
}

// // sort by oldest defrag, then largest size
// func sortDatabases(databases []mssqlz.Database) {
// 	sort.SliceStable(databases, func(i, j int) bool {
// 		if databases[i].LastDBCC.Before(databases[j].LastDBCC) {
// 			return true
// 		}

// 		return databases[i].DatabaseMB > databases[j].DatabaseMB
// 	})
// }
