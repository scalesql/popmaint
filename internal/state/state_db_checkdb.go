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
	m["no_index"] = plan.CheckDB.NoIndex
	m["physical_only"] = plan.CheckDB.PhysicalOnly
	m["extended_logical_checks"] = plan.CheckDB.ExtendedLogicalChecks
	m["data_purity"] = plan.CheckDB.DataPurity

	// Recalculate the maxdop we used
	// It may have been better to pass the actual value around
	// But this seemed cleaner
	maxdop, err := plan.MaxDop(db.Cores, db.Maxdop)
	if err != nil {
		return err
	}

	m["server_cores"] = db.Cores
	m["server_maxdop"] = db.Maxdop
	m["stmt_maxdop"] = maxdop

	query := `
        INSERT INTO dbo.checkdb_log (
            		plan_name, job_id, domain_name, server_name, database_name, completed_at, duration_sec, data_mb
					, no_index, physical_only, extended_logical_checks, data_purity
					, server_cores, server_maxdop, stmt_maxdop) 
			VALUES (:plan_name, :job_id, :domain_name, :server_name, :database_name, :completed_at, :duration_sec, :data_mb
				, :no_index, :physical_only, :extended_logical_checks, :data_purity
				, :server_cores, :server_maxdop, :stmt_maxdop);
    `
	pool := sqlx.NewDb(st.pool, "sqlserver")
	_, err = pool.NamedExec(query, m)
	return err
}
