POPMAINT
========


Security
------------------------------------------------------------------

### Required Permissions
Permissions can be granted at the server level or individually in each database.
* Server-level
    * `sysadmin` role
* Database-level
    * `VIEW SERVER STATE` server permission
    * `db_owner` in each database
* Repository
    * `db_owner` in the repository database (`PopMaint`) or higher


Command-line Flags
------------------------------------------------------------------
Usage: `popmaint -flags plan_name`
* `plan_name` - Name of the plan to run. Do not include the `.toml` extension.
* `-noexec` 
* `-version`
* `-exit #` - Run and exit with this code.  If an error is encountered, the app exits with a non-zero code.  This is for testing how your job scheduler responds to this.
* `-log-level level` - "trace", "debug", "verbose", "info", "warn", or "error"


Application Configuration (TOML)
------------------------------------------------------------------
```toml
[log]
retention_days = 1
level = "debug" #info, error, etc.

[log.fields]
global.log.host = "hostname()"
global.log.user = "user()"

[repository]
server = "server1"
database = "PopMaint"
username = "popmaint"
password = "${POPMAINT_PASSWORD}"
```

For the `username` and `password`, environment variables can be substituted with `${environment_variable_name}` in the field value.  If the username and password aren't set, it will use a trusted connection.

Plan Configuration (TOML)
------------------------------------------------------------------
Plan files should be stored in the `plans` folder and have a `toml` extension.

```toml
servers = [
    "SERVER1\\SQL2022x",
    ]

maxdop_cores = 2
maxdop_percent = 50

[log]
# level = "debug"

[checkdb]
time_limit = "60m"
included = ["master", "msdb", "corrupt_db"]
excluded = ["otherdb"]
min_interval_days = 0

no_index = true 
info_messages = false 
physical_only = true 
estimate_only = false   
extended_logical_checks = false 
data_purity = false 
```

Logging Functions
------------------------------------------------------------------
The following functions can be used to manipulate the JSON logging results.  Current development efforts are focused on text logging to support a job scheduling tool such as SQL Server Agent.  

Functions can be used to manipulate the format and value of logging.  The format is

```toml
    a.b.c = "commit()"
```
Since this is TOML, the "value" (the function) must be in quotes.

The following functions return values:

* `commit()` - the result of GIT DESCRIBE on the repository when the executable was built
* `built()` the build time in RFC3339 format
* `exename()` - the executable name
* `hostname()` - hostname where the program is running
* `pid()` - process ID of the program
* `version()` - the version of the executable
* `jobid()` - the job ID which is `yyyymmdd_hhmmss_plan_name`.  This should match the JSON log file name.

The following functions manipulate the results:

```toml
    new.field = "old.field.move()"
    extra.field = "old.field.copy()"
    removed.field = "delete()"
```




