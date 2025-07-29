Development Notes
=================


## Testing
* Tests should NOT run in parallel since some test blocking
* Test results should not be cached since some use SQL Server
* The SQL Server for testing is set by the `POPMAINT_DBSERVER` environment variable
* The GOOSE environment variables are needed to run `goose` manually

### .env File
```
GOOSE_DRIVER=sqlserver
GOOSE_DBSTRING=sqlserver://D40/SQL2016?database=PopMaint
GOOSE_MIGRATION_DIR=./assets/migrations
GOOSE_TABLE=popmaint_db_version

POPMAINT_USER=abc
POPMAINT_PASS=xyz

POPMAINT_DBSERVER=D40,53796
```

### ./vscode/settings.json
```json
{
    "go.testEnvVars": {
        "POPMAINT_DBSERVER": "D40,53796"
    },
    "go.testEnvFile": "${workspaceFolder}/.env",
    "go.testFlags": ["-count=1"]
}
```

### test.cmd
```cmd
SET POPMAINT_DBSERVER=D40,53796
go test -v -count=1 ./... 
```