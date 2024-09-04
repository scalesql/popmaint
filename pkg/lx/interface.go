package lx

var _ Logger = (*PX)(nil)
var _ Logger = (*CX)(nil)

type Logger interface {
	Debug(msg string, args ...any)
	Verbose(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Log(level LogLevel, msg string, args ...any)
	Console(level LogLevel, msg string)
	WithFields(args ...any) CX
}
