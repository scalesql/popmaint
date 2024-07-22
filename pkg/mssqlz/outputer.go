package mssqlz

import (
	"github.com/golang-sql/sqlexp"
)

// Outputer is used so that messages (RAISERROR and PRINT) from the SQL statement can be written.
type Outputer interface {
	WriteMessage(sqlexp.MsgNotice) error
	WriteError(error) error
	WriteRowSet() error
	WriteString(string) error
	WriteStringf(string, ...any) error
}
