package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/scalesql/popmaint/pkg/mssqlz"
)

var ErrClosed = errors.New("state is closed")

// State captures the state so that we don't need to recreate it.
// It currently only supports the last CheckDB date.
// Defrag will likely just be a map of pending work.
// Update stats may keep a last date.  But we also have that in
// the database.
type State struct {
	mu       *sync.RWMutex
	closed   bool
	fileName string
	Plan     string `json:"plan"`
	CheckDB  struct {
		M map[string]time.Time `json:"last_checkdb"`
	} `json:"checkdb"`
}

func New(plan string) (*State, error) {
	st := State{
		mu:       &sync.RWMutex{},
		fileName: filepath.Join(".", "state", fmt.Sprintf("%s.state.json", strings.ToLower(plan))),
		Plan:     plan,
	}
	st.CheckDB.M = make(map[string]time.Time)
	err := os.MkdirAll(filepath.Join(".", "state"), os.ModePerm)
	if err != nil {
		return &st, err
	}

	err = st.read()
	if err != nil {
		return &st, err
	}
	err = st.write()
	if err != nil {
		return &st, err
	}
	return &st, nil
}

func (st *State) Flush() error {
	st.mu.Lock()
	defer st.mu.Unlock()
	return st.write()
}

func (st *State) Close() error {
	st.mu.Lock()
	defer st.mu.Unlock()
	err := st.write()
	st.closed = true
	return err
}

func (st *State) SetLastCheckDB(db mssqlz.Database) error {
	st.mu.Lock()
	defer st.mu.Unlock()
	k := db.Path()
	st.CheckDB.M[k] = time.Now().Round(1 * time.Second)
	return st.write() // these are big operations so we will write with each completion
}

func (st *State) GetLastCheckDBDate(db mssqlz.Database) (time.Time, bool) {
	st.mu.RLock()
	defer st.mu.RUnlock()
	k := db.Path()
	// If we don't find it, just return
	tm, ok := st.CheckDB.M[k]
	return tm, ok
}

func (st *State) write() error {
	if st.closed {
		return ErrClosed
	}
	bb, err := json.MarshalIndent(st, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(st.fileName, bb, os.ModePerm)
}

func (st *State) read() error {
	if st.closed {
		return ErrClosed
	}
	// if the file doesn't exist, just go on
	if _, err := os.Stat(st.fileName); errors.Is(err, os.ErrNotExist) {
		return nil
	}
	bb, err := os.ReadFile(st.fileName)
	if err != nil {
		return err
	}
	return json.Unmarshal(bb, st)
}
