package app

import (
	"context"
	"errors"
	"fmt"
	"popmaint/pkg/config.go"
	"popmaint/pkg/state"
)

var ErrRunError = errors.New("error running plan")

func Run(planName string, noexec bool) int {
	ctx := context.Background()
	out := OutWriter{}
	plan, err := config.ReadPlan(planName)
	if err != nil {
		out.WriteError(err)
		return 1
	}
	out.WriteStringf("%s: servers: %d  noexec: %t", planName, len(plan.Servers), noexec)
	st, err := state.NewState(planName)
	if err != nil {
		out.WriteError(err)
		return 1
	}
	defer func(st *state.State) {
		if err := st.Close(); err != nil {
			out.WriteError(fmt.Errorf("state.close: %w", err))
		}
	}(st)

	engine := NewEngine(&out, st)
	return engine.runPlan(ctx, plan, noexec)
}
