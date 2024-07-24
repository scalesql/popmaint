package maint

import (
	"context"
	"fmt"
	"popmaint/pkg/config.go"
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

func CheckDB(ctx context.Context, out mssqlz.Outputer, host string, db mssqlz.Database, plan config.Plan, noexec bool) error {
	pool, err := mssqlh.Open(host, "master")
	if err != nil {
		return err
	}
	defer pool.Close()
	stmt := makeStatement(db.DatabaseName, plan)
	out.WriteStringf("%s: %s: %s", db.ServerName, db.DatabaseName, stmt)
	if !noexec {
		err = mssqlz.ExecContext(ctx, pool, stmt, out)
	}
	return err
}

func makeStatement(db string, plan config.Plan) string {
	stmt := fmt.Sprintf("DBCC CHECKDB(%s", mssqlh.QuoteName(db))
	if plan.CheckDB.NoIndex {
		stmt += ", NOINDEX"
	}
	stmt += ")"
	clauses := make([]string, 0)
	if !plan.CheckDB.InfoMessages {
		clauses = append(clauses, "NO_INFOMSGS")
	}
	if plan.CheckDB.PhysicalOnly {
		clauses = append(clauses, "PHYSICAL_ONLY")
	}
	if len(clauses) > 0 {
		stmt += " WITH " + strings.Join(clauses, ", ")
	}
	stmt += ";"
	return stmt
}
