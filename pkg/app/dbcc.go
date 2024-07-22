package app

import (
	"context"
	"popmaint/pkg/maint"
	"popmaint/pkg/mssqlz"
	"sort"

	"github.com/billgraziano/mssqlh"
)

func CheckDB(ctx context.Context, out mssqlz.Outputer, host string, do maint.DefragOptions) error {
	pool, err := mssqlh.Open(host, "master")
	if err != nil {
		return err
	}
	defer pool.Close()

	databases, err := mssqlz.OnlineDatabases(ctx, host)
	if err != nil {
		return err
	}

	// sort the oldest first, and then the largest
	sortDatabases(databases)
	for _, db := range databases {
		if do.MaxSizeMB > 0 && db.DatabaseMB > do.MaxSizeMB {
			continue
		}
		err = maint.CheckDB(ctx, out, host, db, do)
		if err != nil {
			return err
		}
	}
	return nil
}

// sort by oldest defrag, then largest size
func sortDatabases(databases []mssqlz.Database) {
	sort.SliceStable(databases, func(i, j int) bool {
		if databases[i].LastDBCC.Before(databases[j].LastDBCC) {
			return true
		}

		return databases[i].DatabaseMB > databases[j].DatabaseMB
	})
}
