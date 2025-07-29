package maint

import (
	"context"
	"fmt"
	"time"

	"github.com/billgraziano/mssqlh/v2"
	"github.com/scalesql/popmaint/internal/config"
	"github.com/scalesql/popmaint/internal/lockmon"
	"github.com/scalesql/popmaint/internal/mssqlz"
)

// AgentHistory deletes old backup history records from the msdb database
func AgentHistory(ctx context.Context, logger lockmon.ExecLogger, server mssqlz.Server, plan config.Plan, noexec bool) error {
	// open connection
	pool, err := mssqlh.Open(server.FQDN, "msdb")
	if err != nil {
		return err
	}
	defer pool.Close()

	// build the command
	stmt := fmt.Sprintf("DECLARE @limit DATE = DATEADD(dd, -%d, GETDATE()); EXEC msdb.dbo.sp_purge_jobhistory @oldest_date = @limit;", plan.AgentHistory.RetainDays)
	msg := fmt.Sprintf("%s: %s", server.ServerName, stmt)
	logger.Debug(msg, "server", server.ServerName)

	// run the command
	if !noexec {
		result := lockmon.ExecMonitor(ctx, logger, pool, stmt, time.Duration(plan.AgentHistory.StatementTimeout),
			time.Duration(plan.AgentHistory.BlockingTimeout), time.Duration(plan.AgentHistory.BlockedTimeout))
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
