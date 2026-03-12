package maint

import (
	"context"
	"fmt"
	"time"

	"github.com/scalesql/popmaint/internal/config"
	"github.com/scalesql/popmaint/internal/lockmon"
	"github.com/scalesql/popmaint/internal/mssqlz"
)

// BackupHistory deletes old backup history records from the msdb database
func BackupHistory(ctx context.Context, logger lockmon.ExecLogger, server mssqlz.Server, plan config.Plan, noexec bool) error {

	// build the command
	stmt := fmt.Sprintf("DECLARE @limit DATE = DATEADD(dd, -%d, GETDATE()); EXEC msdb.dbo.sp_delete_backuphistory @oldest_date = @limit;", plan.BackupHistory.RetainDays)
	msg := fmt.Sprintf("%s: %s", server.ServerName, stmt)
	logger.Debug(msg, "server", server.ServerName)

	// run the command
	if !noexec {
		result := lockmon.ExecMonitor(ctx, logger, server.FQDN, stmt, time.Duration(plan.BackupHistory.StatementTimeout),
			time.Duration(plan.BackupHistory.BlockingTimeout), time.Duration(plan.BackupHistory.BlockedTimeout))
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
