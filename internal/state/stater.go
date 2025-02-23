package state

import (
	"time"

	"github.com/scalesql/popmaint/internal/mssqlz"
)

// Stater allows us to store application state in JSON files or a database
type Stater interface {
	Close() error
	SetLastCheckDB(mssqlz.Database) error
	GetLastCheckDBDate(mssqlz.Database) (time.Time, bool, error)
	LogCheckDB(string, string, mssqlz.Database, time.Duration) error
}
