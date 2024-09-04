package lx

import (
	"fmt"
	"strings"
)

// Level is the logging level
type LogLevel int

const (
	LevelTrace LogLevel = iota
	LevelDebug
	LevelVerbose
	LevelInfo
	LevelWarn
	LevelError
)

// String returns a string representation of a log level
func (level LogLevel) String() string {
	return [...]string{"TRACE", "DEBUG", "VERBOSE", "INFO", "WARN", "ERROR"}[level]
}

// ParseLevel parses a string value to a LogLevel.
// It returns an error with the valid levels
func ParseLevel(str string) (LogLevel, error) {
	m := make(map[string]LogLevel)
	for i := LevelTrace; i <= LevelError; i++ {
		m[strings.ToLower(i.String())] = i
	}
	level, ok := m[strings.ToLower(str)]
	if !ok {
		return level, fmt.Errorf("invalid log level '%s': use trace, debug, verbose, info, warn, or error", str)
	}
	return level, nil
}

// Level returns the log level for a logger
func (px *PX) Level() LogLevel {
	px.mu.Lock()
	defer px.mu.Unlock()
	return px.level
}

// SetLevel sets the logging level for the logger
func (px *PX) SetLevel(level LogLevel) {
	px.mu.Lock()
	defer px.mu.Unlock()
	px.level = level
}

// SetLevelString sets a log level from a string
func (px *PX) SetLevelString(str string) error {
	px.mu.Lock()
	defer px.mu.Unlock()
	level, err := ParseLevel(str)
	if err != nil {
		return err
	}
	px.level = level
	return nil
}
