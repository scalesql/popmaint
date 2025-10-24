package lx

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"time"
)

// SetLogFile sets the log file location and name based on the
// settings in PX
func (px *PX) SetLogFile(plan string) error {
	if px.logFolder == "" {
		px.logFolder = filepath.Join(".", "logs", "json")
	}
	err := os.MkdirAll(px.logFolder, os.ModePerm)
	if err != nil {
		return fmt.Errorf("os.mkdirall: %w", err)
	}
	var fileName, file string
	fileName = fmt.Sprintf("%s.%s", px.jobid, "ndjson")
	if px.logFilePattern != "" {
		fileName, err = getLogFileName(px.logFilePattern, plan, px.jobid)
		if err != nil {
			return err
		}
	}
	file = filepath.Join(px.logFolder, fileName)
	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("os.openfile: %w", err)
	}
	px.jsonFile = f
	px.logFileName = f.Name()
	return nil
}

func getLogFileName(tmpl, plan, jobid string) (string, error) {
	vals := map[string]string{
		"date":     time.Now().Format("20060102"),
		"time":     time.Now().Format("150405"),
		"date_utc": time.Now().UTC().Format("20060102"),
		"time_utc": time.Now().UTC().Format("150405"),
		"plan":     plan,
		"job_id":   jobid,
	}

	t, err := template.New("file").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("parse: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, vals); err != nil {
		return "", fmt.Errorf("template: %w", err)
	}

	return buf.String(), nil
}

// CleanUpLogs removes older log files in the specified folder
func CleanUpLogs(days int, folder, wildcard string) error {
	// the setlogsettings() function will already have set this to 30 by default
	// so zero really does mean we don't want to purge
	if days == 0 {
		return nil
	}
	cutoff := time.Now().Add(time.Duration(days*-24) * time.Hour)
	files, err := filepath.Glob(filepath.Join(folder, wildcard))
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
