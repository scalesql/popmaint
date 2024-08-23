package lx

import (
	"maps"
)

// CX wraps a logger
// but is mostly a passthrough
type CX struct {
	px     *PX            // wrap the logger
	fields map[string]any // fields for this logger
}

// WithFields returns a child logger with some new default fields
func (px PX) WithFields(args ...any) CX {
	cx := CX{
		px: &px,
	}
	cx.fields = args2map(args...)
	return cx
}

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

func (cx CX) Log(level Level, msg string, args ...any) {
	final := make(map[string]any)

	// copy the fields from the child
	maps.Copy(final, cx.fields)

	// copy the fields we just got
	m := args2map(args...)
	maps.Copy(final, m)
	newargs := map2args(final)
	cx.px.Log(level, msg, newargs...)
}

func (cx CX) Console(level Level, msg string) {
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
