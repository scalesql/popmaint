package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"popmaint/pkg/config"
	"popmaint/pkg/state"
)

var ErrRunError = errors.New("error running plan")

func Run(dev bool, planName string, noexec bool) int {
	ctx := context.Background()
	logger, logFiles, err := getLogger(planName, dev)
	if err != nil {
		log.Printf("getlogger: %s\n", err.Error())
		return 1
	}

	// defer closing each log file
	for i := range logFiles {
		defer func(f *os.File) {
			if err := f.Close(); err != nil {
				log.Printf("file.close: %s\n", err)
			}
		}(logFiles[i])
	}

	appconfig, err := config.ReadConfig()
	if err != nil {
		logger.Error(fmt.Errorf("config.readconfig: %w", err).Error())
		return 1
	}
	logger.Info("running app", slog.Int("log_retention_days", appconfig.LogRetentionDays))
	err = cleanUpLogs(appconfig.LogRetentionDays, "text", "*.log")
	if err != nil {
		logger.Error(fmt.Errorf("cleanuplogs: %w", err).Error())
		return 1
	}
	err = cleanUpLogs(appconfig.LogRetentionDays, "json", "*.nsjson")
	if err != nil {
		logger.Error(fmt.Errorf("cleanuplogs: %w", err).Error())
		return 1
	}

	plan, err := config.ReadPlan(planName)
	if err != nil {
		logger.Error(err.Error())
		return 1
	}
	logger.Info(fmt.Sprintf("running plan %s", planName), "servers", len(plan.Servers), "noexec", noexec, "maxdop_cores", plan.MaxDopCores, "maxdop_percent", plan.MaxDopPercent)
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
