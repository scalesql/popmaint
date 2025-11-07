package lx

import (
	"io"
	"maps"
	"os"
	"sync"
	"sync/atomic"
)

// atomicLogNumber is used for line numbers in the JSON file
// this is used for sorting if the timestamps are identical
var atomicLogNumber atomic.Uint32

// PX is the parent structure for logging
type PX struct {
	mu             *sync.Mutex
	console        io.Writer
	logFolder      string
	logFilePattern string
	jsonFile       io.WriteCloser
	logFileName    string
	level          LogLevel
	jobid          string // yyyymmdd_hhmmss_plan
	payload        string // field name of the payload
	// sequence       uint64

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
	useUTC     bool // log in UTC instead of local time
}

// New returns a new logger.
// jobid is 20240813_055211_plan1
func New(jobid, plan string, opts ...Option) (PX, error) {
	px, err := setup(jobid)
	if err != nil {
		return PX{}, err
	}
	for _, opt := range opts {
		opt(&px)
	}
	if px.payload == "" {
		px.payload = "popmaint"
	}
	err = px.SetLogFile(plan)
	if err != nil {
		return PX{}, err
	}
	return px, nil
}

type Option func(*PX)

func WithPayload(payload string) Option {
	return func(px *PX) {
		px.payload = payload
	}
}

func WithLogFolder(folder string) Option {
	return func(px *PX) {
		px.logFolder = folder
	}
}

func WithFileName(name string) Option {
	return func(px *PX) {
		px.logFilePattern = name
	}
}

// setup defaults for the PX object
func setup(jobid string) (PX, error) {
	px := PX{
		mu:       &sync.Mutex{},
		console:  os.Stdout,
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

// SetUTC sets the flag to use UTC
func (px *PX) SetUTC(utc bool) {
	px.mu.Lock()
	defer px.mu.Unlock()
	px.useUTC = utc
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

// LogFileName returns the name of the file we are writing to
func (px *PX) LogFileName() string {
	px.mu.Lock()
	defer px.mu.Unlock()
	return px.logFileName
}
