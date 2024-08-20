package px

import (
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

// PX is the parent structure for logging
type PX struct {
	mu       *sync.Mutex
	console  io.Writer
	jsonFile io.WriteCloser
	level    Level
	jobid    string // yyyymmdd_hhmmss_plan
	payload  string // field name of the payload

	// mappings are the default things we read from a config
	// file and apply to all things we log
	mappings []Field

	// cached holds function results that are pre-calculated
	// and cached.  They should be in the form of "version()": "1.2".
	// It does a simple lookup on the function name.
	cached map[string]any

	formatJSON bool // writes formatted JSON for DEV
}

// jobid is 20240813_055211_plan1
func New(plan, payload string) (PX, error) {
	now := time.Now()
	lx, err := setup(now, plan, payload)
	if err != nil {
		return PX{}, err
	}
	// get the log file
	jsonFile, err := getLogFile(now, plan, "ndjson")
	if err != nil {
		return PX{}, err
	}
	lx.jsonFile = jsonFile

	return lx, nil
}

// setup defaults for the PX object
func setup(now time.Time, plan, payload string) (PX, error) {
	jobid := fmt.Sprintf("%s_%s", now.Format("20060102_150405"), plan)
	lx := PX{
		mu:       &sync.Mutex{},
		console:  os.Stdout,
		payload:  payload,
		jobid:    jobid,
		level:    LevelInfo,
		cached:   make(map[string]any),
		mappings: []Field{},
	}
	// set cached function results
	hn, err := os.Hostname()
	if err != nil {
		return PX{}, err
	}
	lx.cached["hostname()"] = hn
	lx.cached["pid()"] = os.Getpid()

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

func (px *PX) SetMappings(m map[string]any) error {
	dotted, err := nested2dotted(m)
	if err != nil {
		return err
	}
	px.mu.Lock()
	defer px.mu.Unlock()
	mappings := make([]Field, 0, len(dotted))
	for k, v := range dotted {
		kv := Field{K: k, V: v}
		mappings = append(mappings, kv)
	}
	px.mappings = mappings
	return nil
}

// SetFormatJSON determines whether the JSON logs are written
// to one line for formatted for humans.
func (px *PX) SetFormatJSON(format bool) {
	px.mu.Lock()
	defer px.mu.Unlock()
	px.formatJSON = format
}
