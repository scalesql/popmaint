package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/scalesql/popmaint/internal/failure"
	"github.com/scalesql/popmaint/pkg/build"
	"github.com/scalesql/popmaint/pkg/config"
	"github.com/scalesql/popmaint/pkg/lx"
	"github.com/scalesql/popmaint/pkg/state"
	"golang.org/x/term"
)

var ErrRunError = errors.New("error running plan")

func Run(cmdLine CommandLine) int {
	defer failure.HandlePanic(build.Commit(), build.Built().Format(time.RFC3339))
	jobid := fmt.Sprintf("%s_%s", time.Now().Format("20060102_150405"), cmdLine.Plan)
	exename, err := os.Executable()
	if err != nil {
		fmt.Println("ERROR", err.Error())
		return 1
	}
	exenameBase := filepath.Base(exename)
	ctx := context.Background()

	usr, err := user.Current()
	if err != nil {
		fmt.Println("ERROR", fmt.Errorf("os.user: %w", err).Error())
		return 1
	}
	hn, err := os.Hostname()
	if err != nil {
		fmt.Println("ERROR", fmt.Errorf("os.hostname: %w", err).Error())
		return 1
	}

	logger, err := lx.New(jobid, cmdLine.Plan, "popmaint")
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
	logger.SetFormatJSON(cmdLine.Dev)

	logger.SetCached("exename()", exenameBase)
	logger.SetCached("commit()", build.Commit())
	logger.SetCached("version()", build.Version())
	logger.SetCached("built()", build.Built().Format(time.RFC3339))
	logger.SetCached("user()", usr.Username)

	logger.AddFields("job_id", jobid)
	if cmdLine.Dev {
		logger.AddFields("app.exec.dev", true)
	}

	appconfig, err := config.ReadConfig()
	if err != nil {
		logger.Error(fmt.Errorf("config.readconfig: %w", err).Error())
		return 1
	}
	err = logger.SetMappings(appconfig.Logging.Fields)
	if err != nil {
		logger.Error(fmt.Errorf("logger.setmappings: %w", err).Error())
		return 1
	}
	logger.AddFields(
		"app.version", build.Version(),
		"app.commit", build.Commit(),
		"app.built", build.Built().Format(time.RFC3339),
		"app.exec.path", exename,
		"app.exec.name", exenameBase,
		"app.exec.no_exec", cmdLine.NoExec,
		"app.exec.user", usr.Username,
		"app.exec.pid", os.Getpid(),
		"app.exec.host", hn,
		"app.exec.is_terminal", term.IsTerminal(int(os.Stdout.Fd())),
	)
	logger.Info(fmt.Sprintf("%s: %s (%s) built %s", strings.ToUpper(exenameBase), build.Version(), build.Commit(), build.Built()))
	msg := strings.ToUpper(exenameBase)
	if cmdLine.Dev {
		msg += fmt.Sprintf("  cmdLine.Dev: %t", cmdLine.Dev)
	}
	if cmdLine.NoExec {
		msg += fmt.Sprintf("  cmdLine.NoExec: %t", cmdLine.NoExec)
	}
	logger.Info(msg, "log_retention_days", appconfig.Logging.LogRetentionDays)

	err = lx.CleanUpLogs(appconfig.Logging.LogRetentionDays, "json", "*.ndjson")
	if err != nil {
		logger.Error(fmt.Errorf("lx.cleanuplogs: %w", err).Error())
		return 1
	}

	plan, err := config.ReadPlan(cmdLine.Plan)
	if err != nil {
		logger.Error(fmt.Errorf("config.readplan: %w", err).Error())
		return 1
	}
	dupes := plan.RemoveDupes()
	for _, str := range dupes {
		logger.Warn(fmt.Sprintf("%s: duplicate server: %s", cmdLine.Plan, str))
	}
	logger.Info(fmt.Sprintf("PLAN: %s  servers: %d  cmdLine.NoExec: %t", cmdLine.Plan, len(plan.Servers), cmdLine.NoExec),
		"servers", len(plan.Servers),
		//"cmdLine.NoExec", cmdLine.NoExec,
		"maxdop_cores", plan.MaxDopCores,
		"maxdop_percent", plan.MaxDopPercent)
	st, err := state.New(cmdLine.Plan)
	if err != nil {
		logger.Error(fmt.Errorf("state.new: %w", err).Error())
		return 1
	}
	defer func(st *state.State) {
		if err := st.Close(); err != nil {
			logger.Error(fmt.Errorf("state.close: %w", err).Error())
		}
	}(st)

	engine := NewEngine(logger, plan, st)
	return engine.runPlan(ctx, cmdLine.NoExec)
}
