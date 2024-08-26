package lx

import (
	"fmt"
)

// ConsoleWriter just writes the text and JSON to the console.
// It is mostly used for testing and POC.
type ConsoleWriter struct{}

func (cw ConsoleWriter) Write(bb []byte) (int, error) {
	fmt.Printf("%s\n", string(bb))
	return len(bb), nil
}

func (cw ConsoleWriter) Close() error {
	return nil
}

func NewConsoleLogger() (PX, error) {
	px, err := setup("jobid", "payload")
	if err != nil {
		return px, err
	}
	px.jsonFile = ConsoleWriter{}
	return px, nil
}
