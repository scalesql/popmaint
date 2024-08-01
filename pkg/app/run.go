package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/scalesql/popmaint/pkg/build"
	"github.com/scalesql/popmaint/pkg/config"
	"github.com/scalesql/popmaint/pkg/state"
)

var ErrRunError = errors.New("error running plan")

func Run(dev bool, planName string, noexec bool) int {
	exename, err := os.Executable()
	if err != nil {
		slog.Error(err.Error())
		return 1
	}
	exenameBase := filepath.Base(exename)
	ctx := context.Background()
	logger, logFiles, err := getLogger(planName, dev)
	if err != nil {
		slog.Error(err.Error())
		return 1
	}

	// defer closing each log file
	for i := range logFiles {
		defer func(f *os.File) {
			if err := f.Close(); err != nil {
				slog.Error(err.Error())
			}
		}(logFiles[i])
	}
	if dev {
		fmt.Println(strings.Repeat("-", 80))
	}

	appconfig, err := config.ReadConfig()
	if err != nil {
		logger.Error(fmt.Errorf("config.readconfig: %w", err).Error())
		return 1
	}
	logger.Info(fmt.Sprintf("%s: %s (%s) built %s", exenameBase, build.Version(), build.Commit(), build.Built()))
	// TODO add build information to global, and full exepath
	msg := exenameBase
	if dev {
		msg += fmt.Sprintf("  dev: %t", dev)
	}
	if noexec {
		msg += fmt.Sprintf("  noexec: %t", noexec)
	}
	logger.Info(msg, slog.Int("log_retention_days", appconfig.LogRetentionDays))

	err = cleanUpLogs(appconfig.LogRetentionDays, "text", "*.log")
	if err != nil {
		logger.Error(fmt.Errorf("cleanuplogs: %w", err).Error())
		return 1
	}
	err = cleanUpLogs(appconfig.LogRetentionDays, "json", "*.ndjson")
	if err != nil {
		logger.Error(fmt.Errorf("cleanuplogs: %w", err).Error())
		return 1
	}

	plan, err := config.ReadPlan(planName)
	if err != nil {
		logger.Error(err.Error())
		return 1
	}
	logger.Info(fmt.Sprintf("plan: %s  servers: %d  noexec: %t", planName, len(plan.Servers), noexec),
		"servers", len(plan.Servers),
		"noexec", noexec,
		"maxdop_cores", plan.MaxDopCores,
		"maxdop_percent", plan.MaxDopPercent)
	st, err := state.New(planName)
	if err != nil {
		logger.Error(fmt.Errorf("state.new: %w", err).Error())
		return 1
	}
	defer func(st *state.State) {
		if err := st.Close(); err != nil {
			logger.Error(fmt.Errorf("state.close: %w", err).Error())
		}
	}(st)

	engine := NewEngine(logger, st)
	return engine.runPlan(ctx, plan, noexec)
}
