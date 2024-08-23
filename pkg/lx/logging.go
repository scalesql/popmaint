package lx

import (
	"encoding/json"
	"fmt"
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

func (px PX) Log(level Level, msg string, args ...any) {
	px.mu.Lock()
	defer px.mu.Unlock()
	now := time.Now()
	px.logConsole(now, level, msg)
	m := anys2map(px.payload, args...)
	m["time"] = now
	m["message"] = msg
	m["level"] = level.String()
	//m["global.host.name"] = "D40"

	m, errs := px.applyFuncs(m)
	for _, err := range errs {
		px.logConsole(now, LevelError, fmt.Errorf("px.logjson: %w", err).Error())
	}
	err := px.logJSON(level, m)
	if err != nil {
		px.logConsole(now, LevelError, fmt.Errorf("px.logjson: %w", err).Error())
	}
}

func (px PX) Console(level Level, msg string) {
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
