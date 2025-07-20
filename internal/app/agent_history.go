package app

import (
	"context"
	"fmt"
	"time"

	"github.com/scalesql/popmaint/internal/maint"
	"github.com/scalesql/popmaint/internal/mssqlz"
	"github.com/scalesql/popmaint/internal/zerr"
)

func (engine *Engine) runAgentHistory(ctx context.Context, noexec bool) int {
	plan := engine.Plan
	child := engine.logger.WithFields("action", ActionAgentHistory)
	exitCode := 0

	child.Info(
		fmt.Sprintf("AGENTHISTORY: retain_days: %d", plan.AgentHistory.RetainDays),
	)
	dupecheck := mssqlz.NewDupeCheck()
	for _, fqdn := range plan.Servers {
		srv, err := mssqlz.GetServer(ctx, fqdn)
		if err != nil {
			child.Error(zerr.Wrap(err, fqdn).Error(), "action", ActionAgentHistory)
			continue
		}
		dupe := dupecheck.IsDupe(srv)
		if dupe {
			child.Warn(fmt.Sprintf("%s: duplicate: %s => %s", ActionAgentHistory, fqdn, srv.Path()))
			continue
		}

		start := time.Now()
		err = maint.AgentHistory(ctx, child, srv, plan, noexec)
		if err != nil {
			child.Error(zerr.Wrap(err, srv.Path()).Error(), "action", ActionAgentHistory)
			exitCode = 1
			continue
		}
		child.Info(fmt.Sprintf("AGENTHISTORY: %s (%s)", srv.Path(), time.Since(start).Round(1*time.Second).String()), "action", ActionAgentHistory)
	}
	return exitCode
}
