package app

import (
	"context"
	"fmt"
	"time"

	"github.com/scalesql/popmaint/internal/maint"
	"github.com/scalesql/popmaint/internal/mssqlz"
	"github.com/scalesql/popmaint/internal/zerr"
)

func (engine *Engine) runBackupHistory(ctx context.Context, noexec bool) int {
	plan := engine.Plan
	child := engine.logger.WithFields("action", ActionBackupHistory)
	exitCode := 0

	child.Info(
		fmt.Sprintf("BACKUPHISTORY: retain_days: %d", plan.BackupHistory.RetainDays),
	)
	dupecheck := mssqlz.NewDupeCheck()
	for _, fqdn := range plan.Servers {
		srv, err := mssqlz.GetServer(ctx, fqdn)
		if err != nil {
			child.Error(zerr.Wrap(err, fqdn).Error(), "action", ActionBackupHistory)
			continue
		}
		dupe := dupecheck.IsDupe(srv)
		if dupe {
			child.Warn(fmt.Sprintf("%s: duplicate: %s => %s", ActionBackupHistory, fqdn, srv.Path()))
			continue
		}

		start := time.Now()
		err = maint.BackupHistory(ctx, child, srv, plan, noexec)
		if err != nil {
			child.Error(zerr.Wrap(err, srv.Path()).Error(), "action", ActionBackupHistory)
			exitCode = 1
			continue
		}
		child.Info(fmt.Sprintf("BACKUPHISTORY: %s (%s)", srv.Path(), time.Since(start).Round(1*time.Second).String()), "action", ActionBackupHistory)
	}
	return exitCode
}
