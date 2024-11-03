POPMAINT
========
POPMAINT is a SQL Server maintenance solution similar to SQL Server Maintenance Plans.  It is based on GIT-friendly configuration files.


Command-line Flags
-----
* `-plan plan_name` - Name of the plan to run. Do not include the `toml` extension.
* `-noexec` 
* `-version`
* `-exit #` - Run and exit with this code.  If an error is encountered, the app exits with a non-zero code.  This is for testing how your scheduler responds to this.
* `-log-level level` - "trace", "debug", "verbose", "info", "warn", or "error"


Plan Files (TOML)
-----------------
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
excluded = ["tempdb", "msdb"]
min_interval_days = 0

no_index = true 
info_messages = false 
physical_only = true 
estimate_only = false   
extended_logical_checks = false 
data_purity = false 
```

Application Configuration (TOML)
--------------------------------
```toml
[log]
retention_days = 1

[log.fields]
global.log.host = "hostname()"
global.log.user = "user()"
```

Goals
-----
* Replace SQL Server Maintenance Plans
* Configuration via GIT-friendly files
* JSON logging designed for Elastic Search
* Spread load across servers and time zones

Roadmap - Short Term
--------------------
* Store `state` and logs in a database
* Configure SQL Server credentials from environment variables

Roadmap - Long Term
-------------------
* History cleanup
* Statistics
* Backup
* Reindex
