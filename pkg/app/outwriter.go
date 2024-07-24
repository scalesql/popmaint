package app

import (
	"fmt"
	"log"
	"sync"

	"github.com/golang-sql/sqlexp"
	mssql "github.com/microsoft/go-mssqldb"
)

// TODO - remove the error returns, write them to the console
// if we can't write to files, just write that message to the console
// maybe a flag to indicate we are writing the JSON also?
// maybe a field name in the JSON for execution_id?
// maybe that is defrag_execution_id.log|ndjson for now?
// something timestamp-ish and sequential
// xid_servers_plan.(log|ndjson)

type OutWriter struct {
	mu sync.RWMutex
}

func (ow *OutWriter) init() {
	// if ow.mu == nil {
	// 	ow.mu = sync.RWMutex{}
	// }
}

func New() (OutWriter, error) {
	return OutWriter{
		mu: sync.RWMutex{},
	}, nil
}

func (ow *OutWriter) WriteMessage(msg sqlexp.MsgNotice) error {
	ow.init()
	log.Println(msg.Message.String())
	return nil
}

func (ow *OutWriter) WriteError(err error) error {
	switch e := err.(type) {
	case mssql.Error:
		msg := fmt.Sprintf("Msg %d, Level %d, State %d, Line %d %s", e.Number, e.Class, e.State, e.LineNo, e.Message)
		log.Printf("ERROR: %s\n", msg)
	default:
		log.Println(err)
	}

	return nil
}

func (ow *OutWriter) WriteErrorf(fmt string, args ...any) error {
	log.Printf(fmt, args...)
	return nil
}

func (ow *OutWriter) WriteRowSet() error {
	panic("rowset!")
}

func (ow *OutWriter) WriteString(str string) error {
	log.Println(str)
	return nil
}

func (ow *OutWriter) WriteStringf(fmt string, args ...any) error {
	log.Printf(fmt, args...)
	return nil
}
