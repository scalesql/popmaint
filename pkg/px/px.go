package px

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type Field struct {
	K string
	V any
}

type Level int

const (
	LevelDebug = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (d Level) String() string {
	return [...]string{"DEBUG", "INFO", "WARN", "ERROR"}[d]
}

type PX struct {
	mu         *sync.Mutex
	console    io.Writer
	jsonFile   io.WriteCloser
	FormatJSON bool
	JobID      string
	Payload    string
	Functions  []Field // stuff with functions and moves
	Fields     []Field // fields for child loggers, etc.
	Constants  map[string]any
	//Statics   map[string]any // result of static functions
}

// jobid is 20240813_055211_plan1
func New(name, payload string) (*PX, error) {
	now := time.Now()
	jobid := fmt.Sprintf("%s_%s", now.Format("20060102_150405"), name)
	// get the log file
	jsonFile, err := getLogFile(now, name, "ndjson")
	if err != nil {
		return nil, err
	}
	lx := &PX{
		mu:       &sync.Mutex{},
		console:  os.Stdout,
		jsonFile: jsonFile,
		Payload:  payload,
		JobID:    jobid,
	}
	return lx, nil
}

// Close the JSON log file
func (px *PX) Close() error {
	px.mu.Lock()
	defer px.mu.Unlock()
	if px.jsonFile != nil {
		return px.jsonFile.Close()
	}
	return nil
}

func (px *PX) Debug(msg string, args ...any) {
	px.Log(LevelDebug, msg, args...)
}

func (px *PX) Info(msg string, args ...any) {
	px.Log(LevelInfo, msg, args...)
}
func (px *PX) Warn(msg string, args ...any) {
	px.Log(LevelWarn, msg, args...)
}
func (px *PX) Error(msg string, args ...any) {
	px.Log(LevelError, msg, args...)
}

func (px *PX) Log(level Level, msg string, args ...any) {
	px.mu.Lock()
	defer px.mu.Unlock()
	now := time.Now()
	px.logConsole(now, level, msg)
	m := anys2map(px.Payload, args...)
	m["time"] = now
	m["message"] = msg
	m["level"] = level.String()
	//m["global.host.name"] = "D40"

	m, errs := px.applyFuncs(m)
	for _, err := range errs {
		px.logConsole(now, LevelError, fmt.Errorf("px.logjson: %w", err).Error())
	}
	err := px.logJSON(m)
	if err != nil {
		px.logConsole(now, LevelError, fmt.Errorf("px.logjson: %w", err).Error())
	}
}

func (px *PX) Console(level Level, msg string) {
	px.logConsole(time.Now(), level, msg)
}

func (px *PX) logConsole(now time.Time, level Level, msg string) {
	out := now.Format("15:04:05")
	if level >= LevelWarn {
		out += fmt.Sprintf(" %s", level)
	}
	out += " " + msg
	out += "\n"
	px.console.Write([]byte(out))
}

func (px *PX) logJSON(m map[string]any) error {
	// make a nested map
	nested, err := dotted2nested(m)
	if err != nil {
		return err
	}
	var bb []byte
	if px.FormatJSON {
		bb, err = json.MarshalIndent(nested, "", "    ")
		if err != nil {
			return err
		}
	} else {
		bb, err = json.Marshal(nested)
		if err != nil {
			return err
		}
	}
	bb = append(bb, "\r\n"...)
	_, err = px.jsonFile.Write(bb)
	if err != nil {
		return err
	}
	return nil
}
