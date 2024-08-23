package lx

import (
	"fmt"
	"os"
	"sync"
)

// ConsoleWriter just writes the text and JSON to the console
type ConsoleWriter struct{}

func (cw ConsoleWriter) Write(bb []byte) (int, error) {
	fmt.Printf("%s\n", string(bb))
	return len(bb), nil
}

func (cw ConsoleWriter) Close() error {
	return nil
}

func NewConsoleLogger() PX {
	return PX{
		mu:       &sync.Mutex{},
		console:  os.Stdout,
		jsonFile: ConsoleWriter{},
		payload:  "payload",
		jobid:    "jobid",
		level:    LevelInfo,
		cached:   make(map[string]any),
		mappings: []Field{},
	}
}
