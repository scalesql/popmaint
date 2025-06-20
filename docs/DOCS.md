POPMAINT
========


Security
------------------------------------------------------------------

### Required Permissions
Permissions can be granted at the server level or individually in each database.

* If the service account is a system administrator
    * `sysadmin` role
* If the service account is not a system administrator
    * `VIEW ANY DEFINITION`
    * `VIEW SERVER STATE` 
    * `VIEW ANY DATABASE` 
    * `db_owner` in each database
* Repository
    * `db_owner` in the repository database (`PopMaint`) or higher


Command-line Flags
------------------------------------------------------------------
Usage: `popmaint -flags plan_name`

* `plan_name` - Name of the plan to run. Do not include the `.toml` extension.
* `-noexec` 
* `-version`
* `-exit #` - Run and exit with this code.  If an error is encountered, the app exits with the non-zero code.  This is for testing how your job scheduler responds to this.
* `-log-level` *level* - where *level* is one of `trace`, `debug`, `verbose`, `info`, `warn`, or `error`


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
password = "${POPMAINT_REPO_PASSWORD}"
```

For the `username` and `password`, environment variables can be substituted with `${environment_variable_name}` in the field value.  If the username and password aren't set, it will use a trusted connection.  NOTE: There is a bug that doesn't support passwords with characters that would need to be URL encoded.

Plan Configuration (TOML)
------------------------------------------------------------------
Plan files should be stored in the `plans` folder and have a `toml` extension.

```toml
servers = [
    "SERVER1\\SQL2022x",
    ]

maxdop_cores = 2
maxdop_pct = 50
maxdop_pct_maxdop = 0 

[log]
# level = "debug"

[checkdb]
time_limit = "60m"  # don't start new statements
statement_timeout = "3h" # cancel a running statement
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

### Time Limits
* `time_limit` is the time after which no new CHECKDB statement will start.
*  `statement_timeout` is the duration after which a CHECKDB statement should be cancelled.  This is typically much higher than `time_limt`. It typically handles stuck processes or being blocked.  This should be higher than CHECKDB will ever run normally.

### MAXDOP Settings
There are three settings that control the MAXDOP that CHECKDB uses.

* `maxdop_cores` - cap at an absolute number of cores 
* `maxdop_pct` - cap at a percentage of the cores (rounded down)
* `maxdop_pct_maxdop` - cap at a percentage of server MAXDOP rounded down

All these are optional.  If any of these are set, the MAXDOP is set at the lowest value from any of them providing it is lower than the server MAXDOP and server cores.

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




