package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/scalesql/popmaint/internal/build"
	"github.com/scalesql/popmaint/internal/config"
	"github.com/scalesql/popmaint/internal/failure"
	"github.com/scalesql/popmaint/internal/lx"
	"github.com/scalesql/popmaint/internal/state"
	"golang.org/x/term"
)

var ErrRunError = errors.New("error running plan")

func Run(cmdLine CommandLine, getenv func(string) string) int {
	defer failure.HandlePanic(build.Commit(), build.Built().Format(time.RFC3339))
	jobid := fmt.Sprintf("%s_%s", time.Now().Format("20060102_150405"), cmdLine.Plan)
	exename, err := os.Executable()
	if err != nil {
		fmt.Println("ERROR", err.Error())
		return 1
	}
	exenameBase := filepath.Base(exename)
	ctx := context.Background()

	userName, err := currentUserName()
	if err != nil {
		fmt.Println("ERROR", fmt.Errorf("currentUserName: %w", err).Error())
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
	logger.SetCached("user()", userName)

	logger.AddFields("job_id", jobid)
	if cmdLine.Dev {
		logger.AddFields("app.exec.dev", true)
	}

	// Read the app.toml config file
	appconfig, err := config.ReadConfig(getenv)
	if err != nil {
		logger.Error(fmt.Errorf("config.readconfig: %w", err).Error())
		return 1
	}
	err = logger.SetMappings(appconfig.Log.Fields)
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
		"app.exec.user", userName,
		"app.exec.pid", os.Getpid(),
		"app.exec.host", hn,
		"app.exec.is_terminal", term.IsTerminal(int(os.Stdout.Fd())),
	)
	logger.Info(fmt.Sprintf("%s %s (%s) built %s", strings.ToUpper(exenameBase), build.Version(), build.Commit(), build.Built()))
	msg := fmt.Sprintf("%s on %s as %s", strings.ToUpper(exenameBase), hn, userName)
	if cmdLine.Dev {
		msg += fmt.Sprintf("  cmdLine.Dev: %t", cmdLine.Dev)
	}
	if cmdLine.NoExec {
		msg += fmt.Sprintf("  cmdLine.NoExec: %t", cmdLine.NoExec)
	}
	logger.Info(msg, "log_retention_days", appconfig.Log.LogRetentionDays)

	err = lx.CleanUpLogs(appconfig.Log.LogRetentionDays, "json", "*.ndjson")
	if err != nil {
		logger.Error(fmt.Errorf("lx.cleanuplogs: %w", err).Error())
		return 1
	}

	// Read the plan.toml file
	plan, err := config.ReadPlan(cmdLine.Plan)
	if err != nil {
		logger.Error(fmt.Errorf("config.readplan: %w", err).Error())
		return 1
	}

	// Set the logging level
	// command-line overrides plan.toml overrides app.toml
	logSource := ""
	if appconfig.Log.Level != "" {
		err := logger.SetLevelString(appconfig.Log.Level)
		if err != nil {
			logger.Error(fmt.Errorf("popmaint.toml: %w", err).Error())
			return 1
		}
		logSource = "popmaint.toml"
	}

	if plan.Log.Level != "" {
		err := logger.SetLevelString(plan.Log.Level)
		if err != nil {
			logger.Error(fmt.Errorf("%s.toml: %w", cmdLine.Plan, err).Error())
			return 1
		}
		logSource = fmt.Sprintf("%s.toml", cmdLine.Plan)
	}

	if cmdLine.LogLevel != "" {
		err := logger.SetLevelString(cmdLine.LogLevel)
		if err != nil {
			logger.Error(fmt.Errorf("-log-level: %w", err).Error())
			return 1
		}
		logSource = "command-line -log-level"
	}
	if logger.Level() != lx.LevelInfo {
		logger.Info(fmt.Sprintf("logging at %s from %s", logger.Level(), logSource))
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

	// Set the proper state provider
	var st state.Stater
	if appconfig.Repository.Server == "" {
		st, err = state.NewFileState(cmdLine.Plan)
		if err != nil {
			logger.Error(fmt.Errorf("state.new: %w", err).Error())
			return 1
		}
		logger.Info("State: file-based")
	} else {
		st, err = state.NewDBState(
			appconfig.Repository.Server,
			appconfig.Repository.Database,
			appconfig.Repository.UserName,
			appconfig.Repository.Password,
			logger)
		if err != nil {
			logger.Error(fmt.Errorf("state.new: %w", err).Error())
			return 1
		}
		logger.Info(fmt.Sprintf("State: server: %s  database: %s", appconfig.Repository.Server, appconfig.Repository.Database))
	}
	defer func(st state.Stater) {
		if err := st.Close(); err != nil {
			logger.Error(fmt.Errorf("state.close: %w", err).Error())
		}
	}(st)

	engine := NewEngine(logger, plan, st)
	engine.JobID = jobid
	return engine.runPlan(ctx, cmdLine.NoExec)
}
