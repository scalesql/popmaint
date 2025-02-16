package state

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/scalesql/popmaint/internal/mssqlz"
)

// TODO maybe logging the event will need the plan?

type DBState struct {
	pool *sql.DB
}

// NewDBState returns a new database state
func NewDBState(server, database, user, password string) (*DBState, error) {
	if database == "" {
		database = "PopMaint"
	}
	host, instance, port := parseFQDN(server)
	if host == "" {
		return nil, fmt.Errorf("invalid server: %s", server)
	}
	//println(host, instance, port)
	query := url.Values{}
	query.Add("app name", "popmaint.exe")
	query.Add("database", database)
	// query.Add("encrypt", "optional")

	u := &url.URL{
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
	if user != "" || password != "" {
		u.User = url.UserPassword(user, password)
	}
	pool, err := sql.Open("sqlserver", u.String())
	if err != nil {
		return nil, err
	}

	err = pool.Ping()
	if err != nil {
		return nil, err
	}

	// TODO: check for db_owner role
	// TODO: deploy schema
	st := &DBState{
		pool: pool,
	}
	return st, nil
}

func (st *DBState) SetLastCheckDB(db mssqlz.Database) error {
	stmt := `
	
	MERGE dbo.dbcc_state AS t
USING (VALUES 
    (@p1, @p2, @p3, @p4)
) AS source (domain_name, server_name, [database_name], last_dbcc)
ON (t.domain_name = source.domain_name 
    AND t.server_name = source.server_name 
    AND t.[database_name] = source.[database_name])
WHEN MATCHED THEN 
    UPDATE SET 
        last_dbcc = source.last_dbcc
WHEN NOT MATCHED THEN 
    INSERT (domain_name, server_name, [database_name], last_dbcc)
    VALUES (source.domain_name, source.server_name, source.[database_name], source.last_dbcc);
	
	`
	_, err := st.pool.Exec(stmt, db.Domain, db.ServerName, db.DatabaseName, time.Now().Round(1*time.Second))
	return err
}

func (st *DBState) GetLastCheckDBDate(db mssqlz.Database) (time.Time, bool, error) {
	stmt := `
		SELECT 	last_dbcc
		FROM	dbo.dbcc_state
		WHERE	[domain_name] = @p1
		AND		[server_name] = @p2
		AND		[database_name] = @p3;
	`
	var tm time.Time
	err := st.pool.QueryRow(stmt, db.Domain, db.ServerName, db.DatabaseName).Scan(&tm)
	if err == nil { // we found it
		return tm, true, nil
	}
	if errors.Is(err, sql.ErrNoRows) { // we didn't find it
		return time.Time{}, false, nil
	}
	return tm, false, err // there was an actual error
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
