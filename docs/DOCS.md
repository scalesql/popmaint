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
* `--noexec` 
* `--version`
* `--exit #` - Run and exit with this code.  If an error is encountered, the app will exit with a non-zero code.  This is for testing your job scheduler.
* `--log-level` *level* - where *level* is one of `trace`, `debug`, `verbose`, `info`, `warn`, or `error`
* `--help` - Display the command-line options

Application Configuration (popmaint.toml)
------------------------------------------------------------------
```toml
[log]
level = "debug" #info, error, etc.

[log.fields] # see Logging Fields below
global.log.host = "hostname()"
global.log.user = "user()"
my.field = 37

[repository]
server = "server1"
database = "PopMaint"
username = "popmaint"
password = "${POPMAINT_REPO_PASSWORD}"
```

For the `username` and `password` fields, environment variables can be substituted with `${environment_variable_name}` in the field value.  If the username and password aren't set, it will use a trusted connection.  NOTE: There is a bug that doesn't support passwords with characters that would need to be URL encoded.

Plan Configuration (TOML)
------------------------------------------------------------------
Plan files are stored in the `plans` folder and have a `toml` extension.

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
time_limit = "60m"       # don't start new statements after this time
statement_timeout = "3h" # cancel a running statement
blocking_timeout = "1s"  # cancel after 1s if we are blocking
blocked_timeout = "1m"   # wait this long if we are being blocked

included = ["master", "msdb", "corrupt_db"]
excluded = ["otherdb"]
min_interval_days = 0

no_index = true 
info_messages = false 
physical_only = true 
estimate_only = false   
extended_logical_checks = false 
data_purity = false 

[backup_history]
retain_days = 90
statement_timeout = "5m"
blocking_timeout = "5s"
blocked_timeout = "5m"

[dbmail_history]
retain_days = 365
statement_timeout = "3m"
blocking_timeout = "5s"
blocked_timeout = "5m"

[agent_history]
retain_days = 90
statement_timeout = "5m"
blocking_timeout = "5s"
blocked_timeout = "5m"

```

* At least one server is required
* The history cleanups only require the number of days to retain

### Time Limits
* `time_limit` is the time after which no new CHECKDB statement will start.  This is a soft limit.
*  `statement_timeout` is the duration after which a statement should be cancelled.  This is typically much higher than `time_limt`. It typically handles stuck processes or disconnected sessions.  This should be higher than CHECKDB will ever run normally.
* `blocking_timeout` is the amount of time to wait if we are blocking something else.  This is polled every second.  __This defaults to 1 second.__
* `blocked_timeout` is the amount of time to wait if we are being blocked.  This is polled every second. __This defaults to 5 seconds.__

### MAXDOP Settings
There are three settings that control the MAXDOP that CHECKDB uses.

* `maxdop_cores` - cap at an absolute number of cores 
* `maxdop_pct` - cap at a percentage of the cores (rounded down)
* `maxdop_pct_maxdop` - cap at a percentage of server MAXDOP rounded down

All these are optional.  If any of these are set, the MAXDOP is set at the lowest value from any of them providing it is lower than the server MAXDOP and server cores.

Basic Logging (JSON)
------------------------------------------------------------------
* PopMaint has two JSON file logging modes.  The default is Basic Logging.
* The JSON logs are written to `./logs/json` with an extension of `.ndjson`.  The file names are `yyyymmdd_hhmmss_plan.ndjson` in UTC time.
* There are four top-level fields:
    * `time`
    * `level` - DEBUG, INFO, etc.
    * `message`
    * `popmaint` - nested object
* All application specific fields fall under the `popmaint` nested object.
* Log files are purged after 30 days.

Advanced Logging (JSON)
------------------------------------------------------------------
* Advanced logging provides control over file location, log file name, purge process, retention, etc.  It is enabled by setting `advanced=true`.  The suggsted settings for Advanced Logging are: 

    ```toml
    [log]
    advanced=true
    retain_days = 30
    folder = "./logs/json"
    file_name_template = "{{.job_id}}.ndjson"
    purge_glob = "*.ndjson"
    ```

### Advanced Logging Notes
* All fields are required for advanced logging
* `folder` should be entered as `D:\\Logs\\PopMaint` to correctly handle escaping in TOML.
* Setting `retain_days` = 0, turns off purging log files
* `purge_glob` is only required if `retain_days` > 0.  This is the wildcard that will be searched in the log folder.  It filters based on this wildcard and the last write date.
* `use_utc` is not listed above.  It allows `true` or `false`.  If set to `true`, the time reported in `time` will be in UTC instead of the local time with offset.
    * `true`:   2025-10-23T20:01:54.0277865Z
    * `false`:  2025-10-23T15:01:54.0277865-05:00
* `file_name_template` supports the following replacements.  The field is a GO text template so values need a "." field name prefix inside {{ }}.
    * `date` and `date_utc` in YYYYYMMDD format
    * `time` and `time_utc` in HHMMSS format
    * `plan` is the name of the plan
    * `job_id` which is just `{{.date_utc}}_{{.time_utc}}_{{.plan}}`

Logging Fields and Functions
------------------------------------------------------------------
The following functions can be used to manipulate the JSON logging results.  Current development efforts are focused on text logging to support a job scheduling tool such as SQL Server Agent.  

Functions can be used to manipulate the format and value of logging.  The format is

```toml
    a.b.c = "hostname()"
    a.b.d = 37
```
Since this is TOML, the "value" (the function) must be in quotes.  The dotted field names indicate nesting in the JSON.  A sample log might look like this:

```json
{
    "a": {
        "b": {
            "c": "D40",
            "d": 37
        }
    },
    "level": "INFO",
    "message": "sample message...",
    "popmaint": {
        "app": {
            "built": "2020-04-01T00:00:00Z"
        }
    },
    "time": "2025-10-23T20:01:54.0277865-05:00"
}
```

The following functions return values:

* `commit()` - the result of GIT DESCRIBE on the repository when the executable was built
* `built()` the build time in RFC3339 format
* `exename()` - the executable name
* `hostname()` - hostname where the program is running
* `pid()` - process ID of the program
* `version()` - the version of the executable
* `jobid()` - the job ID which is `yyyymmdd_hhmmss_plan`
* `plan()` - the name of the plan

The following functions manipulate the results:

```toml
    new.field = "old.field.move()"
    extra.field = "old.field.copy()"
    removed.field = "delete()"
```




