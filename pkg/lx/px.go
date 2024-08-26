package lx

import (
	"io"
	"maps"
	"os"
	"sync"
)

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
	// these should be applied last
	mappings []KV

	// cached holds function results that are pre-calculated
	// and cached.  They should be in the form of "version()": "1.2".
	// It does a simple lookup on the function name.
	cached map[string]any

	// fields are any default fields assigned to this logger.
	// it is in the form "a.b.c":7
	fields map[string]any

	formatJSON bool // writes formatted JSON for DEV
}

// New returns a new logger.
// jobid is 20240813_055211_plan1
func New(jobid, plan, payload string) (PX, error) {
	//now := time.Now()
	px, err := setup(jobid, payload)
	if err != nil {
		return PX{}, err
	}
	// get the log file
	jsonFile, err := getLogFile(jobid, "ndjson")
	if err != nil {
		return PX{}, err
	}
	px.jsonFile = jsonFile
	return px, nil
}

// setup defaults for the PX object
func setup(jobid, payload string) (PX, error) {
	//jobid := fmt.Sprintf("%s_%s", now.Format("20060102_150405"), plan)
	px := PX{
		mu:       &sync.Mutex{},
		console:  os.Stdout,
		payload:  payload,
		jobid:    jobid,
		level:    LevelInfo,
		cached:   make(map[string]any),
		mappings: []KV{},
		fields:   make(map[string]any),
	}
	// set cached function results
	hn, err := os.Hostname()
	if err != nil {
		return PX{}, err
	}
	px.cached["hostname()"] = hn
	px.cached["pid()"] = os.Getpid()

	return px, nil
}

// AddFields adds default fields to the logger.
// Keys should be dotted ("a.b.c").
func (px *PX) AddFields(args ...any) {
	px.mu.Lock()
	defer px.mu.Unlock()
	parms := args2map(args...)
	maps.Copy(px.fields, parms)
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

// SetMappings accepts mapping functions and adds them
// to the logger
func (px *PX) SetMappings(m map[string]any) error {
	dotted, err := nested2dotted(m)
	if err != nil {
		return err
	}
	px.mu.Lock()
	defer px.mu.Unlock()
	mappings := make([]KV, 0, len(dotted))
	for k, v := range dotted {
		kv := KV{K: k, V: v}
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
