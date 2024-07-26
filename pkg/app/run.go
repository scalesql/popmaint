package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"popmaint/pkg/config"
	"popmaint/pkg/state"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-colorable"
	slogmulti "github.com/samber/slog-multi"
)

var ErrRunError = errors.New("error running plan")

func getLogger(txtfile *os.File) (*slog.Logger, error) {
	//tintLogger := slog.New(tint.NewHandler(os.Stderr, nil))
	l0 := slog.New(
		slogmulti.Fanout(
			slog.NewTextHandler(txtfile, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}),
			tint.NewHandler(colorable.NewColorable(os.Stderr), &tint.Options{
				Level:      slog.LevelDebug,
				TimeFormat: time.TimeOnly,
			}),
		),
	)
	// TODO add this for the JSON handler: l1 := l0.With(slog.String("plan", planName))
	return l0, nil
}

func getLogFile(name string, ext string) (*os.File, error) {
	err := os.MkdirAll(filepath.Join(".", "logs", "text"), os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("os.mkdirall: %w", err)
	}
	fileName := fmt.Sprintf("%s_%s.%s", time.Now().Format("20060102_150405"), name, ext)
	file := filepath.Join(".", "logs", "text", fileName)
	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return nil, fmt.Errorf("os.openfile: %w", err)
	}
	return f, nil
}

func cleanUpLogs(n int) error {
	cutoff := time.Duration(n*24) * time.Hour
	files, err := filepath.Glob(filepath.Join(".", "logs", "text", "*.log"))
	if err != nil {
		return err
	}
	for _, name := range files {
		fi, err := os.Stat(name)
		if err != nil {
			return err
		}
		if diff := time.Since(fi.ModTime()); diff > cutoff {
			//fmt.Printf("Deleting %s which is %s old\n", name, diff)
			err = os.Remove(name)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func Run(planName string, noexec bool) int {
	ctx := context.Background()
	// get a log file and pass it in
	// then defer the closing
	txtFile, err := getLogFile(planName, "log")
	if err != nil {
		log.Printf("getlogfile: %s\n", err.Error())
		return 1
	}
	defer func(f *os.File) {
		if err := f.Close(); err != nil {
			log.Printf("textlogfile.close: %s\n", err)
		}
	}(txtFile)

	logger, err := getLogger(txtFile)
	if err != nil {
		log.Printf("getlogger: %s\n", err.Error())
		return 1
	}

	appconfig, err := config.ReadConfig()
	if err != nil {
		logger.Error(fmt.Errorf("config.readconfig: %w", err).Error())
		return 1
	}
	logger.Info("running app", slog.Int("log_retention_days", appconfig.LogRetentionDays))
	err = cleanUpLogs(appconfig.LogRetentionDays)
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
