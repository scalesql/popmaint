package mssqlz

import (
	"context"
	"strings"

	"github.com/billgraziano/mssqlh/v2"
	"github.com/jmoiron/sqlx"
)

type Server struct {
	FQDN           string
	ServerName     string `db:"server_name"`
	Domain         string
	Computer       string
	Instance       string
	EngineEdition  int    `db:"engine_edition"`
	ProductVersion string `db:"product_version"`
	MajorVersion   int    `db:"major_version"`
	MaxDop         int
	CPUCount       int `db:"cpu_count"`
}

// Path returns a string in the format /domain/computer/instance.  This is used
// as a key in maps for the server.
func (srv Server) Path() string {
	str := "/" + strings.Join([]string{srv.Domain, srv.Computer, srv.Instance}, "/")
	//str = strings.ToLower(str)
	return str
}

// GetServer returns information on SQL Server
func GetServer(ctx context.Context, fqdn string) (Server, error) {
	srv := Server{FQDN: fqdn}
	pool, err := mssqlh.Open(fqdn, "master")
	if err != nil {
		return srv, err
	}
	defer pool.Close()
	poolx := sqlx.NewDb(pool, "mssql")
	defer poolx.Close()

	err = poolx.GetContext(ctx, &srv, dbServerQuery)
	return srv, err
}

var dbServerQuery = `

SELECT @@SERVERNAME AS server_name
	,COALESCE(DEFAULT_DOMAIN(), '') as domain
	,COALESCE(CAST(SERVERPROPERTY('ComputerNamePhysicalNetBIOS') AS NVARCHAR(128)), '') AS computer
	,COALESCE(CAST(SERVERPROPERTY('InstanceName') AS NVARCHAR(128)), '') AS instance
	,COALESCE(CAST(SERVERPROPERTY('EngineEdition') AS INT), 0) AS engine_edition
	,COALESCE(CAST(SERVERPROPERTY('ProductVersion') AS NVARCHAR(128)), '') AS product_version
	,COALESCE(CAST(SERVERPROPERTY('ProductMajorVersion') AS INT), '') AS major_version
	,(SELECT [value] FROM sys.configurations WHERE configuration_id = 1539) AS [maxdop]
	,(SELECT cpu_count FROM sys.dm_os_sys_info) AS cpu_count ; 
`
