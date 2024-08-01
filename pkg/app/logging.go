package app

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	slogmulti "github.com/samber/slog-multi"
)

func getLogger(name string, dev bool) (*slog.Logger, []*os.File, error) {
	// noattrs removes all attributes except the big three
	// noattrs := func(groups []string, a slog.Attr) slog.Attr {
	// 	if a.Key == slog.TimeKey || a.Key == slog.LevelKey || a.Key == slog.MessageKey {
	// 		return a
	// 	}
	// 	return slog.Attr{}
	// }

	//global := slog.Group("global", slog.Group("host", slog.String("name", "xxx")))
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// setup the log files
	files := make([]*os.File, 0)
	txtFile, err := getLogFile(name, "log")
	if err != nil {
		return nil, nil, err
	}
	files = append(files, txtFile)
	jsonFile, err := getLogFile(name, "ndjson")
	if err != nil {
		return nil, nil, err
	}
	files = append(files, jsonFile)

	// build the loggers
	// TODO pass in all the "global fields"
	jlog1 := slog.New(slog.NewJSONHandler(jsonFile, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	jlog2 := jlog1.With(
		slog.Group("global",
			slog.Group("host",
				slog.String("name", hostname))),
	)
	jlog3 := jlog2.WithGroup("popmaint").With(slog.String("plan", name))

	//var consoleLogger slog.Handler
	// if dev {
	// 	consoleLogger = tint.NewHandler(colorable.NewColorable(os.Stdout), &tint.Options{
	// 		Level:       slog.LevelDebug,
	// 		TimeFormat:  time.TimeOnly,
	// 		ReplaceAttr: noattrs,
	// 	})
	// } else {
	// 	consoleLogger = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	// 		Level:       slog.LevelDebug,
	// 		ReplaceAttr: noattrs,
	// 	})
	// }
	logger := slog.New(
		slogmulti.Fanout(
			slog.Default().Handler(),
			slog.NewTextHandler(txtFile, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}),
			jlog3.Handler(),
		),
	)
	return logger, files, nil
}

func getLogFile(name string, ext string) (*os.File, error) {
	err := os.MkdirAll(filepath.Join(".", "logs", "text"), os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("os.mkdirall: %w", err)
	}
	err = os.MkdirAll(filepath.Join(".", "logs", "json"), os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("os.mkdirall: %w", err)
	}
	var fileName, file string
	switch ext {
	case "log":
		fileName = fmt.Sprintf("%s_%s.%s", time.Now().Format("20060102_150405"), name, ext)
		file = filepath.Join(".", "logs", "text", fileName)
	case "ndjson":
		fileName = fmt.Sprintf("%s_%s.%s", time.Now().Format("20060102_150405"), name, ext)
		file = filepath.Join(".", "logs", "json", fileName)
	default:
		return nil, fmt.Errorf("invalid extension: %s", ext)
	}

	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return nil, fmt.Errorf("os.openfile: %w", err)
	}
	return f, nil
}

func cleanUpLogs(days int, folder, wildcard string) error {
	cutoff := time.Duration(days*24) * time.Hour
	files, err := filepath.Glob(filepath.Join(".", "logs", folder, wildcard))
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
