package app

// CommandLine is used to pass command-line arguments
// to the application
type CommandLine struct {
	Plan             string
	NoExec           bool
	LogRetentionDays int
	LogLevel         string
	Dev              bool
}
