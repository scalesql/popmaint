package state

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/scalesql/popmaint/internal/config"
	"github.com/scalesql/popmaint/internal/mssqlz"
)

func (st *DBState) SetLastCheckDB(db mssqlz.Database) error {
	stmt := `
	
	MERGE dbo.checkdb_state AS t
USING (VALUES 
    (@p1, @p2, @p3, @p4)
) AS source (domain_name, server_name, [database_name], last_checkdb)
ON (t.domain_name = source.domain_name 
    AND t.server_name = source.server_name 
    AND t.[database_name] = source.[database_name])
WHEN MATCHED THEN 
    UPDATE SET 
        last_checkdb = source.last_checkdb
WHEN NOT MATCHED THEN 
    INSERT (domain_name, server_name, [database_name], last_checkdb)
    VALUES (source.domain_name, source.server_name, source.[database_name], source.last_checkdb);
	
	`
	_, err := st.pool.Exec(stmt, db.Domain, db.ServerName, db.DatabaseName, time.Now().Round(1*time.Second))
	return err
}

func (st *DBState) GetLastCheckDBDate(db mssqlz.Database) (time.Time, bool, error) {
	stmt := `
		SELECT 	last_checkdb
		FROM	dbo.checkdb_state
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

// checkdblog represents a record in the dbo.checkdb_log table
type checkdblog struct {
	LogID        int64     `db:"log_id"`
	PlanName     string    `db:"plan_name"`
	DomainName   string    `db:"domain_name"`
	ServerName   string    `db:"server_name"`
	DatabaseName string    `db:"database_name"`
	CompletedAt  time.Time `db:"completed_at"`
	DurationSec  int32     `db:"duration_sec"`
	DataMB       int64     `db:"data_mb"`
	Maxdop       int32     `db:"maxdop"`
}

func (st *DBState) LogCheckDB(plan config.Plan, jobid string, db mssqlz.Database, dur time.Duration) error {
	m := make(map[string]any)
	m["plan_name"] = plan.Name
	m["job_id"] = jobid
	m["domain_name"] = db.Domain
	m["server_name"] = db.ServerName
	m["database_name"] = db.DatabaseName
	m["completed_at"] = time.Now()
	m["duration_sec"] = int32(dur.Seconds())
	m["data_mb"] = int64(db.DatabaseMB)

	query := `
        INSERT INTO dbo.checkdb_log (
            		plan_name, job_id, domain_name, server_name, database_name, completed_at, duration_sec, data_mb) 
			VALUES (:plan_name, :job_id, :domain_name, :server_name, :database_name, :completed_at, :duration_sec, :data_mb);
    `
	pool := sqlx.NewDb(st.pool, "sqlserver")
	_, err := pool.NamedExec(query, m)
	return err
}
