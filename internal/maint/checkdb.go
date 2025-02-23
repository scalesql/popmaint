package maint

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/scalesql/popmaint/internal/config"
	"github.com/scalesql/popmaint/internal/lockmon"
	"github.com/scalesql/popmaint/internal/mssqlz"

	"github.com/billgraziano/mssqlh"
)

type CheckDBEstimate struct {
	Plan     config.Plan
	NeededKB int
	Messages []string
}

func CheckDB(ctx context.Context, logger lockmon.ExecLogger, host string, db mssqlz.Database, plan config.Plan, noexec bool) error {
	pool, err := mssqlh.Open(host, "master")
	if err != nil {
		return err
	}
	defer pool.Close()
	maxdop, err := plan.MaxDop(db.Cores, db.Maxdop)
	if err != nil {
		return err
	}
	if plan.CheckDB.DataPurity && plan.CheckDB.PhysicalOnly {
		return fmt.Errorf("can't set data_purity and physical_only")
	}
	stmt := makeCheckDBStatement(db.DatabaseName, plan, maxdop)
	logger.Debug(stmt, "server", db.ServerName, "database", db.DatabaseName)

	if !noexec {
		result := lockmon.ExecMonitor(ctx, logger, pool, stmt, time.Duration(0))
		if result.Sessions != nil {
			if len(result.Sessions) > 0 {
				for _, s := range result.Sessions {
					msg := fmt.Sprintf("id=%d blocker=%d %s", s.SessionID, s.BlockingID, mssqlz.TrimSQL(s.Statement, 60))
					logger.Error(msg,
						"stmt", mssqlz.TrimSQL(s.Statement, 200))
				}
			}
		}
		return result.Err
	}
	return nil
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
