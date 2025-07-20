package app

import (
	"context"
	"time"

	"github.com/scalesql/popmaint/internal/config"
	"github.com/scalesql/popmaint/internal/lx"
	"github.com/scalesql/popmaint/internal/state"
)

const (
	ActionCheckdb       = "checkdb"
	ActionBackupHistory = "backup_history"
	ActionDBMailHistory = "dbmail_history"
	ActionAgentHistory  = "agent_history"
)

// Engine is the main application engine that runs the maintenance plan
type Engine struct {
	logger lx.Logger
	st     state.Stater
	start  time.Time
	JobID  string
	Plan   config.Plan
}

func NewEngine(logger lx.Logger, plan config.Plan, st state.Stater) Engine {
	engine := Engine{
		logger: logger,
		st:     st,
		start:  time.Now(),
		Plan:   plan,
	}
	return engine
}

func (engine *Engine) runPlan(ctx context.Context, noexec bool) int {
	// TODO sort the plan actions and run in order
	result := engine.runCheckDB(ctx, noexec)

	// only run if configured
	// this also catches negative retentions
	if engine.Plan.BackupHistory.RetainDays > 0 {
		r1 := engine.runBackupHistory(ctx, noexec)
		if r1 > result {
			result = r1
		}
	}

	// only run if configured
	// this also catches negative retentions
	if engine.Plan.AgentHistory.RetainDays > 0 {
		r1 := engine.runAgentHistory(ctx, noexec)
		if r1 > result {
			result = r1
		}
	}

	// only run if configured
	// this also catches negative retentions
	if engine.Plan.DBMailHistory.RetainDays > 0 {
		r1 := engine.runDBMailHistory(ctx, noexec)
		if r1 > result {
			result = r1
		}
	}

	return result
}
