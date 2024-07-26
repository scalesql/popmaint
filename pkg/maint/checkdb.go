package maint

import (
	"context"
	"fmt"
	"log/slog"
	"popmaint/pkg/config"
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

func CheckDB(ctx context.Context, logger *slog.Logger, host string, db mssqlz.Database, plan config.Plan, noexec bool) error {
	pool, err := mssqlh.Open(host, "master")
	if err != nil {
		return err
	}
	defer pool.Close()
	maxdop, err := plan.MaxDop(db.Cores)
	if err != nil {
		return err
	}
	if plan.CheckDB.DataPurity && plan.CheckDB.PhysicalOnly {
		return fmt.Errorf("can't set data_purity and physical_only")
	}
	stmt := makeCheckDBStatement(db.DatabaseName, plan, maxdop)
	logger.Debug(stmt, slog.String("server", db.ServerName), slog.String("database", db.DatabaseName))

	if !noexec {
		err = mssqlz.ExecContext(ctx, pool, stmt, logger)
	}
	return err
}

func makeCheckDBStatement(db string, plan config.Plan, maxdop int) string {
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
	if maxdop > 0 {
		clauses = append(clauses, fmt.Sprintf("MAXDOP=%d", maxdop))
	}
	if plan.CheckDB.ExtendedLogicalChecks {
		clauses = append(clauses, "EXTENDED_LOGICAL_CHECKS ")
	}
	if plan.CheckDB.DataPurity {
		clauses = append(clauses, "DATA_PURITY")
	}
	if plan.CheckDB.EstimateOnly {
		clauses = append(clauses, "ESTIMATEONLY")
	}

	if len(clauses) > 0 {
		stmt += " WITH " + strings.Join(clauses, ", ")
	}
	stmt += ";"
	return stmt
}
