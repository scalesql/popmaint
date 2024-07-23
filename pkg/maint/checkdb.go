package maint

import (
	"context"
	"fmt"
	"popmaint/pkg/mssqlz"
	"strings"

	"github.com/billgraziano/mssqlh"
)

type CheckDBOptions struct {
	NoExec       bool
	NoIndex      bool
	InfoMessage  bool
	PhysicalOnly bool
	MaxSizeMB    int
	// ExtendedLogicalChecks bool
	// DataPurity            bool
}

func CheckDB(ctx context.Context, out mssqlz.Outputer, host string, db mssqlz.Database, do CheckDBOptions) error {
	pool, err := mssqlh.Open(host, "master")
	if err != nil {
		return err
	}
	defer pool.Close()
	stmt := makeStatement(db.DatabaseName, do)
	out.WriteStringf("%s: %s (size_mb=%d  last_dbcc=%s)", db.DatabaseName, stmt, db.DatabaseMB, db.LastDBCC.Format("2006-01-02 15:04:05"))
	if !do.NoExec {
		err = mssqlz.ExecContext(ctx, pool, stmt, out)
	}
	return err
}

func makeStatement(db string, do CheckDBOptions) string {
	stmt := fmt.Sprintf("DBCC CHECKDB(%s", mssqlh.QuoteName(db))
	if do.NoIndex {
		stmt += ", NOINDEX"
	}
	stmt += ")"
	clauses := make([]string, 0)
	if !do.InfoMessage {
		clauses = append(clauses, "NO_INFOMSGS")
	}
	if do.PhysicalOnly {
		clauses = append(clauses, "PHYSICAL_ONLY")
	}
	if len(clauses) > 0 {
		stmt += " WITH " + strings.Join(clauses, ", ")
	}
	stmt += ";"
	return stmt
}
