package mssqlz

import (
	"context"
	"strings"
	"time"

	"github.com/billgraziano/mssqlh"
	"github.com/jmoiron/sqlx"
	_ "github.com/microsoft/go-mssqldb"
)

// Database holds information about a SQL Server database.  It is
// primarily used for CheckDB.
type Database struct {
	FQDN           string
	DatabaseName   string    `db:"database_name"`
	DatabaseMB     int       `db:"database_mb"`
	LastDBCC       time.Time `db:"last_dbcc"`
	ServerName     string    `db:"server_name"`
	Domain         string
	Computer       string
	Instance       string
	EngineEdition  int    `db:"engine_edition"`
	ProductVersion string `db:"product_version"`
	MajorVersion   int    `db:"major_version"`
	MaxDop         int
	CPUCount       int `db:"cpu_count"`
}

// Path returns a string in the format /domain/computer/instance/database.  This is used
// as a key in maps for the database.
func (db Database) Path() string {
	str := "/" + strings.Join([]string{db.Domain, db.Computer, db.Instance, db.DatabaseName}, "/")
	//str = strings.ToLower(str)
	return str
}

// OnlineDatabases returns a list of datbases that are online.  It includes system databases.
func OnlineDatabases(ctx context.Context, fqdn string) ([]Database, error) {
	databases := make([]Database, 0)
	pool, err := mssqlh.Open(fqdn, "master")
	if err != nil {
		return databases, err
	}
	defer pool.Close()
	poolx := sqlx.NewDb(pool, "mssql")
	defer poolx.Close()

	err = poolx.SelectContext(ctx, &databases, dblistQuery)
	for i := range databases {
		databases[i].FQDN = fqdn
	}
	return databases, err
}

// TODO: Add updateability
var dblistQuery = `

;WITH CTE AS (
	SELECT	d.[name] AS [database_name],
			(sum(size)/128) as [database_mb]
	FROM	sys.databases d
	JOIN	sys.master_files mf ON mf.database_id = d.database_id
	WHERE	mf.[type] = 0
	AND		d.[state] = 0
	GROUP BY d.[name]
)
SELECT cte.*
	,@@SERVERNAME AS server_name
	,COALESCE(DEFAULT_DOMAIN(), '') as domain
	,COALESCE(CAST(SERVERPROPERTY('ComputerNamePhysicalNetBIOS') AS NVARCHAR(128)), '') AS computer
	,COALESCE(CAST(SERVERPROPERTY('InstanceName') AS NVARCHAR(128)), '') AS instance
	,COALESCE(CAST(SERVERPROPERTY('EngineEdition') AS INT), 0) AS engine_edition
	,COALESCE(CAST(SERVERPROPERTY('ProductVersion') AS NVARCHAR(128)), '') AS product_version
	,COALESCE(CAST(SERVERPROPERTY('ProductMajorVersion') AS INT), '') AS major_version
	,(SELECT [value] FROM sys.configurations WHERE configuration_id = 1539) AS [maxdop]
	,(SELECT cpu_count FROM sys.dm_os_sys_info) AS cpu_count 
FROM CTE

`
