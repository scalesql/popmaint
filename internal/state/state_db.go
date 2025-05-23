package state

import (
	"database/sql"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/pressly/goose/v3"
	"github.com/scalesql/popmaint/assets"
)

// TODO maybe logging the event will need the plan?

type Logger interface {
	Fatalf(format string, v ...interface{})
	Printf(format string, v ...interface{})
	Debug(msg string, args ...any)
}

type DBState struct {
	pool *sql.DB
}

// NewDBState returns a new database state
func NewDBState(server, database, user, password string, logger Logger) (*DBState, error) {
	if server == "" {
		return nil, fmt.Errorf("server is required")
	}
	if database == "" {
		return nil, fmt.Errorf("repository database is required")
	}

	u, err := connstr(server, database, user, password)
	if err != nil {
		return nil, fmt.Errorf("connstr: %w", err)
	}
	logger.Debug(fmt.Sprintf("REPOSITORY: %s", redacted(u)))
	pool, err := sql.Open("sqlserver", u.String())
	if err != nil {
		return nil, err
	}

	err = pool.Ping()
	if err != nil {
		return nil, err
	}

	// check for db_owner role
	user, isdbo, err := checkDBOwner(pool)
	if err != nil {
		return nil, fmt.Errorf("checkdbowner: %w", err)
	}
	if !isdbo {
		return nil, fmt.Errorf("user '%s' is not a member of the db_owner role", user)
	}
	goose.SetBaseFS(assets.DBMigrationsFS)
	err = goose.SetDialect("mssql")
	if err != nil {
		return nil, fmt.Errorf("goose.setdialect: %w", err)
	}
	goose.SetTableName("popmaint_db_version")
	goose.SetLogger(logger)
	err = goose.Up(pool, "migrations")
	if err != nil {
		return nil, fmt.Errorf("goose.up: %w", err)
	}

	st := &DBState{
		pool: pool,
	}
	return st, nil
}

func connstr(server, database, user, password string) (url.URL, error) {
	host, instance, port := parseFQDN(server)
	if host == "" {
		return url.URL{}, fmt.Errorf("invalid server: %s", server)
	}

	query := url.Values{}
	query.Add("app name", "popmaint.exe")
	query.Add("database", database)
	// query.Add("encrypt", "optional")

	u := url.URL{
		Scheme:   "sqlserver",
		Host:     host,
		RawQuery: query.Encode(),
	}
	if instance != "" {
		u.Path = instance
	}
	if port != 0 {
		u.Host = fmt.Sprintf("%s:%d", host, port)
	}
	if user != "" && password != "" {
		u.User = url.UserPassword(user, password)
	}
	return u, nil
}

func redacted(u url.URL) string {
	if u.User == nil {
		return u.String()
	}
	usr := u.User.Username()
	if usr == "" {
		usr = "_notset_"
	}
	pwd, set := u.User.Password()
	if !set {
		pwd = "_notset_"
	} else {
		// redact all but the first character
		var newpwd string
		for i := range pwd {
			if i == 0 {
				newpwd += string(pwd[i])
			} else {
				newpwd += "_"
			}
		}
		pwd = newpwd
	}
	u.User = url.UserPassword(usr, pwd)
	return u.String()
}

// checkDBOwner checks if the current user is a member of the db_owner role and returns the user name and the flag
func checkDBOwner(db *sql.DB) (string, bool, error) {
	var userName string
	var isDBOwner int
	err := db.QueryRow("SELECT SUSER_NAME(), IS_ROLEMEMBER('db_owner')").Scan(&userName, &isDBOwner)
	if err != nil {
		return "", false, err
	}
	return userName, isDBOwner == 1, nil
}

// Close the state repository
func (st *DBState) Close() error {
	return st.pool.Close()
}

// parse FQDN splits a host\instance with an optional port
func parseFQDN(s string) (host, instance string, port int) {
	var err error
	parts := strings.FieldsFunc(s, hostSplitter)
	host = parts[0]
	if len(parts) == 1 {
		return host, "", 0
	}
	if len(parts) == 2 {
		port, err = strconv.Atoi(parts[1])
		if err == nil {
			return host, "", port
		}
		instance = parts[1]
		return host, instance, 0
	}
	if len(parts) == 3 {
		instance = parts[1]
		port, _ = strconv.Atoi(parts[2])
		return host, instance, port
	}

	return host, instance, port
}

// hostSplitter splits a string on :,\ and is used to split FQDN names
func hostSplitter(r rune) bool {
	return r == ':' || r == ',' || r == '\\'
}
