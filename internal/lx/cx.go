package lx

import (
	"maps"
)

// CX is a "child" logger that wraps a PX.
// It is mostly a pass-through
type CX struct {
	px     *PX            // wrap the logger
	fields map[string]any // fields for this logger
}

// WithFields returns a child logger from a PX with default fields
func (px PX) WithFields(args ...any) CX {
	cx := CX{
		px: &px,
	}
	cx.fields = args2map(args...)
	return cx
}

// WithFields returns a child logger from a CX with additional default fields
func (cx CX) WithFields(args ...any) CX {
	child := CX{
		px:     cx.px,
		fields: make(map[string]any),
	}
	for k, v := range cx.fields {
		child.fields[k] = v
	}
	m := args2map(args...)
	for k, v := range m {
		child.fields[k] = v
	}
	return child
}

// Log an event
func (cx CX) Log(level LogLevel, msg string, args ...any) {
	// start with the default fields
	final := maps.Clone(cx.fields)
	// copy the fields we just got
	m := args2map(args...)
	maps.Copy(final, m)
	cx.px.LogMap(level, msg, final)
}

// Console writes to the console
func (cx CX) Console(level LogLevel, msg string) {
	cx.px.Console(level, msg)
}

func (cx CX) Debug(msg string, args ...any) {
	cx.Log(LevelDebug, msg, args...)
}

func (cx CX) Verbose(msg string, args ...any) {
	cx.Log(LevelVerbose, msg, args...)
}

func (cx CX) Info(msg string, args ...any) {
	cx.Log(LevelInfo, msg, args...)
}

func (cx CX) Warn(msg string, args ...any) {
	cx.Log(LevelWarn, msg, args...)
}

func (cx CX) Error(msg string, args ...any) {
	cx.Log(LevelError, msg, args...)
}
