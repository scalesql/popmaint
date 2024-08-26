package lx

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"time"

	"github.com/fatih/color"
	"golang.org/x/term"
)

type Level int

const (
	LevelDebug = iota
	LevelVerbose
	LevelInfo
	LevelWarn
	LevelError
)

func (d Level) String() string {
	return [...]string{"DEBUG", "INFO", "INFO", "WARN", "ERROR"}[d]
}

func (px PX) Debug(msg string, args ...any) {
	px.Log(LevelDebug, msg, args...)
}
func (px PX) Verbose(msg string, args ...any) {
	px.Log(LevelVerbose, msg, args...)
}
func (px PX) Info(msg string, args ...any) {
	px.Log(LevelInfo, msg, args...)
}
func (px PX) Warn(msg string, args ...any) {
	px.Log(LevelWarn, msg, args...)
}
func (px PX) Error(msg string, args ...any) {
	px.Log(LevelError, msg, args...)
}

// Log an event.  Args are passed as "k", value pairs in an array
func (px PX) Log(level Level, msg string, args ...any) {
	px.mu.Lock()
	defer px.mu.Unlock()
	argmap := anys2map("", args...)
	px.log(level, msg, argmap)
}

// Log an event.  Args are passed as map["a.b.c"]any.
// Child loggers build a map and use this.
func (px PX) LogMap(level Level, msg string, m map[string]any) {
	px.mu.Lock()
	defer px.mu.Unlock()
	px.log(level, msg, m)
}

func (px PX) log(level Level, msg string, fields map[string]any) {
	now := time.Now()
	px.logConsole(now, level, msg)

	// start with the fields from the logger
	parms := maps.Clone(px.fields)

	// add in the fields passed as parameters
	maps.Copy(parms, fields)

	// put those in the payload
	parent := make(map[string]any)
	nested, err := dotted2nested(parms)
	if err != nil {
		px.logConsole(now, LevelError, fmt.Errorf("dotted2nested: %w", err).Error())
		// continue and keep trying with what we got
	}
	if len(parms) > 0 {
		if px.payload != "" {
			parent[px.payload] = nested
		} else { // else just put them all at the top level
			maps.Copy(parent, nested)
		}
	}
	// overwrite the top level values I need
	// TODO these should eventually be parameters
	parent["time"] = now
	parent["message"] = msg
	parent["level"] = level.String()

	// apply the functions
	// always done after the payload and the nesting
	p2, errs := px.applyFuncs(parent)
	for _, err := range errs {
		px.logConsole(now, LevelError, fmt.Errorf("px.logjson: %w", err).Error())
		// continue and keep trying with what we got
	}
	// log the JSON
	err = px.logJSON(level, p2)
	if err != nil {
		px.logConsole(now, LevelError, fmt.Errorf("px.logjson: %w", err).Error())
		// continue and keep trying with what we got
	}
}

// Console just writes to the console
func (px PX) Console(level Level, msg string) {
	px.mu.Lock()
	defer px.mu.Unlock()
	px.logConsole(time.Now(), level, msg)
}

func (px *PX) logConsole(now time.Time, level Level, msg string) {
	if level < px.level {
		return
	}
	out := ""

	if level >= LevelWarn {
		out += level.String() + " "
	}
	out += msg
	if term.IsTerminal(int(os.Stdout.Fd())) {
		switch level {
		case LevelError:
			red := color.New(color.FgRed).SprintFunc()
			out = red(out)
		case LevelWarn:
			yellow := color.New(color.FgYellow).SprintFunc()
			out = yellow(out)
		default:
		}
	}
	out += "\n"
	line := fmt.Sprintf("%s %s", now.Format("15:04:05"), out)
	px.console.Write([]byte(line))
}

func (px *PX) logJSON(level Level, m map[string]any) error {
	if level < px.level {
		return nil
	}
	// make a nested map
	nested, err := dotted2nested(m)
	if err != nil {
		return err
	}
	var bb []byte
	if px.formatJSON {
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
