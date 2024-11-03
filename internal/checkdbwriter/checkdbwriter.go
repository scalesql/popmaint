package checkdbwriter

import (
	"strconv"
	"sync"

	"github.com/oriser/regroup"
)

/*
type ExecLogger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}
*/

var estimateRegex = regroup.MustCompile(`(?m)^Estimated.+\s=\s(?P<kb>\d+)\.`)

type AW struct {
	mu       *sync.Mutex
	messages []string
}

func New() *AW {
	return &AW{
		mu:       &sync.Mutex{},
		messages: make([]string, 0),
	}
}

// Estimate tries to extract the estimate from the messages
func (aw *AW) EstimateKB() int {
	aw.mu.Lock()
	defer aw.mu.Unlock()
	for _, str := range aw.messages {
		match, err := estimateRegex.Groups(str)
		if err != nil {
			continue
		}
		kbstr, ok := match["kb"]
		if !ok {
			continue
		}
		kb, err := strconv.Atoi(kbstr)
		if err != nil {
			continue
		}
		return kb
	}
	return 0
}

func (aw *AW) Messages() []string {
	aw.mu.Lock()
	defer aw.mu.Unlock()
	return aw.messages
}

func (aw *AW) Info(msg string, _ ...any) {
	aw.mu.Lock()
	defer aw.mu.Unlock()
	aw.messages = append(aw.messages, msg)
}

func (aw *AW) Error(msg string, _ ...any) {
	aw.mu.Lock()
	defer aw.mu.Unlock()
	aw.messages = append(aw.messages, msg)
}

func (aw *AW) Debug(msg string, _ ...any) {
	aw.mu.Lock()
	defer aw.mu.Unlock()
	aw.messages = append(aw.messages, msg)
}
