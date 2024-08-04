package lockmon

type Logger interface {
	Errorf(string, ...any)
	Infof(string, ...any)
	Debugf(string, ...any)
	Tracef(string, ...any)
}

type ExecLogger interface {
	Error(msg string, args ...any)
	Warn(msg string, args ...any)
	Info(msg string, args ...any)
	Debug(msg string, args ...any)
}

// nilwriter is used if no logger is passed in
type nilwriter struct{}

func (nilwriter) Error(msg string, args ...any) {}
func (nilwriter) Warn(msg string, args ...any)  {}
func (nilwriter) Info(msg string, args ...any)  {}
func (nilwriter) Debug(msg string, args ...any) {}
func (nilwriter) Trace(msg string, args ...any) {}
