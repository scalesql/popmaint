package mssqlz

import (
	"database/sql"
	"os"
	"path/filepath"

	"github.com/billgraziano/mssqlh/v2"
	"github.com/scalesql/popmaint/internal/zerr"
	"golang.org/x/text/unicode/norm"
)

// TrimSQL returns the left most characters of a string
// it should handle unicode/utf8 without splitting characters
// It is designed for big strings of SQL statements
// it uses a sketchy test to see if the string is already short enough
// https://stackoverflow.com/questions/61353016/why-doesnt-golang-have-substring
// This seems better: https://stackoverflow.com/questions/46415894/golang-truncate-strings-with-special-characters-without-corrupting-data
func TrimSQL(str string, n int) string {
	if n == 0 {
		return ""
	}
	if n < 1 { // -1 is the whole string
		return str
	}
	if len(str) <= n { // fewer bytes than we want characters
		return str
	}
	str = norm.NFC.String(str)
	result := str
	chars := 0
	dots := false
	// https://go.dev/doc/effective_go#for
	// for over a string ranges over the runes.
	// i is bumped to the first byte to each rune,
	// then we just count runes/characters
	for i := range str {
		if chars >= n {
			result = str[:i]
			dots = true
			break
		}
		chars++
	}
	if dots {
		result += "..."
	}
	return result
}

// PoolWithSuffix generates a *sql.DB for the given host and database.
// The application name is the executable name.  If a suffix is provided,
// it is appended so it looks like appname.exe-suffix.
// The database defaults to master.  It also runs a Ping() to verify
// the connection is working.
func PoolWithSuffix(host, database, suffix string) (*sql.DB, error) {
	// Get the EXE name
	exe, err := os.Executable()
	if err != nil {
		return nil, zerr.Wrap(err, "os.executable")
	}
	appname := filepath.Base(exe)

	if database == "" {
		database = "master"
	}
	if suffix != "" {
		appname += "-" + suffix
	}

	conn := mssqlh.NewConnection(host, "", "", database, appname)
	pool, err := conn.Open()
	if err != nil {
		return nil, zerr.Wrap(err, "conn.open")
	}
	err = pool.Ping()
	if err != nil {
		return nil, zerr.Wrap(err, "pool.ping")
	}
	return pool, nil
}
