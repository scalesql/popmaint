package app

import (
	"context"
	"fmt"
	"time"

	"github.com/scalesql/popmaint/internal/maint"
	"github.com/scalesql/popmaint/internal/mssqlz"
	"github.com/scalesql/popmaint/internal/zerr"
)

func (engine *Engine) runDBMailHistory(ctx context.Context, noexec bool) int {
	plan := engine.Plan
	child := engine.logger.WithFields("action", ActionDBMailHistory)
	exitCode := 0

	child.Info(
		fmt.Sprintf("DBMAILHISTORY: retain_days: %d", plan.DBMailHistory.RetainDays),
	)
	dupecheck := mssqlz.NewDupeCheck()
	for _, fqdn := range plan.Servers {
		srv, err := mssqlz.GetServer(ctx, fqdn)
		if err != nil {
			child.Error(zerr.Wrap(err, fqdn).Error(), "action", ActionDBMailHistory)
			continue
		}
		dupe := dupecheck.IsDupe(srv)
		if dupe {
			child.Warn(fmt.Sprintf("%s: duplicate: %s => %s", ActionDBMailHistory, fqdn, srv.Path()))
			continue
		}

		start := time.Now()
		err = maint.DBMailHistory(ctx, child, srv, plan, noexec)
		if err != nil {
			child.Error(zerr.Wrap(err, srv.Path()).Error(), "action", ActionDBMailHistory)
			exitCode = 1
			continue
		}
		child.Info(fmt.Sprintf("DBMAILHISTORY: %s (%s)", srv.Path(), time.Since(start).Round(1*time.Second).String()), "action", ActionDBMailHistory)
	}
	return exitCode
}
