package lx

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func getLogFile(jobid, ext string) (*os.File, error) {
	// err := os.MkdirAll(filepath.Join(".", "logs", "text"), os.ModePerm)
	// if err != nil {
	// 	return nil, fmt.Errorf("os.mkdirall: %w", err)
	// }
	err := os.MkdirAll(filepath.Join(".", "logs", "json"), os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("os.mkdirall: %w", err)
	}
	var fileName, file string
	switch ext {
	case "log":
		fileName = fmt.Sprintf("%s.%s", jobid, ext)
		file = filepath.Join(".", "logs", "text", fileName)
	case "ndjson":
		fileName = fmt.Sprintf("%s.%s", jobid, ext)
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

// CleanUpLogs removes older log files in the specified folder
func CleanUpLogs(days int, folder, wildcard string) error {
	if days == 0 {
		days = 30
	}
	cutoff := time.Now().Add(time.Duration(days*-24) * time.Hour)
	files, err := filepath.Glob(filepath.Join(".", "logs", folder, wildcard))
	if err != nil {
		return err
	}
	for _, name := range files {
		fi, err := os.Stat(name)
		if err != nil {
			return err
		}
		if fi.ModTime().Before(cutoff) {
			err = os.Remove(name)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
