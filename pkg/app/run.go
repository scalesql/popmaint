package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/scalesql/popmaint/internal/failure"
	"github.com/scalesql/popmaint/pkg/build"
	"github.com/scalesql/popmaint/pkg/config"
	"github.com/scalesql/popmaint/pkg/lx"
	"github.com/scalesql/popmaint/pkg/state"
)

var ErrRunError = errors.New("error running plan")

func Run(dev bool, planName string, noexec bool) int {
	defer failure.HandlePanic(build.Commit(), build.Built().Format(time.RFC3339))
	jobid := fmt.Sprintf("%s_%s", time.Now().Format("20060102_150405"), planName)
	exename, err := os.Executable()
	if err != nil {
		fmt.Println("ERROR", err.Error())
		return 1
	}
	exenameBase := filepath.Base(exename)
	ctx := context.Background()

	err = lx.CleanUpLogs(30, "json", "*.ndjson")
	if err != nil {
		fmt.Println("ERROR cleanuplogs: ", err.Error())
		return 1
	}

	logger, err := lx.New(jobid, planName, "popmaint")
	if err != nil {
		fmt.Println("ERROR", err.Error())
		return 1
	}
	defer func(lx lx.PX) {
		err := lx.Close()
		if err != nil {
			fmt.Println("ERROR", err.Error())
		}
	}(logger)
	logger.SetFormatJSON(dev)

	logger.SetCached("exename()", exenameBase)
	logger.SetCached("commit()", build.Commit())
	logger.SetCached("version()", build.Version())
	logger.SetCached("built()", build.Built().Format(time.RFC3339))

	cx := logger.WithFields("job_id", jobid)

	appconfig, err := config.ReadConfig()
	if err != nil {
		cx.Error(fmt.Errorf("config.readconfig: %w", err).Error())
		return 1
	}
	err = logger.SetMappings(appconfig.Logging.Fields)
	if err != nil {
		cx.Error(fmt.Errorf("logger.setmappings: %w", err).Error())
		return 1
	}
	cx = cx.WithFields(
		"app.version", build.Version(),
		"app.commit", build.Commit(),
		"app.built", build.Built().Format(time.RFC3339),
		"app.name", exenameBase,
		"settings.no_exec", noexec,
	)
	cx.Info(fmt.Sprintf("%s: %s (%s) built %s", strings.ToUpper(exenameBase), build.Version(), build.Commit(), build.Built()))
	msg := strings.ToUpper(exenameBase)
	if dev {
		msg += fmt.Sprintf("  dev: %t", dev)
	}
	if noexec {
		msg += fmt.Sprintf("  noexec: %t", noexec)
	}
	cx.Info(msg, "log_retention_days", appconfig.LogRetentionDays)

	plan, err := config.ReadPlan(planName)
	if err != nil {
		logger.Error(fmt.Errorf("config.readplan: %w", err).Error())
		return 1
	}
	dupes := plan.RemoveDupes()
	for _, str := range dupes {
		cx.Warn(fmt.Sprintf("%s: duplicate server: %s", planName, str))
	}
	cx.Info(fmt.Sprintf("PLAN: %s  servers: %d  noexec: %t", planName, len(plan.Servers), noexec),
		"servers", len(plan.Servers),
		"noexec", noexec,
		"maxdop_cores", plan.MaxDopCores,
		"maxdop_percent", plan.MaxDopPercent)
	st, err := state.New(planName)
	if err != nil {
		cx.Error(fmt.Errorf("state.new: %w", err).Error())
		return 1
	}
	defer func(st *state.State) {
		if err := st.Close(); err != nil {
			cx.Error(fmt.Errorf("state.close: %w", err).Error())
		}
	}(st)

	engine := NewEngine(cx, plan, st)
	return engine.runPlan(ctx, noexec)
}
