PopMaint
========

TODO
----
_ don't DBCC database snapshots, just skip them
_ failure to delete a log file is a logged error, not an abort
_ if one server and it can't be reached, it doesn't return an error
_ `-info` flag that prints the version, host, user, pid, etc.
_ use the functional source license or the PolyForm internal use license 
_ add command-line and TOML settings for trace, debug, verbose
_ check for required permissions

_ priority: history, stats, local backups, reindex, 

_ have separate errors for app errors and CHECKDB errors
_ add trace level
_ state should be per action: `shared-dev.checkdb.state.json`?
_ email certain kinds of alerts?
_ license for non-commercial
_ functions for the severity/level: severity_name(), severity_level()
_ switch "popmaint" to a global value in the `global` package?
_ maybe a skip primary setting so we only check the secondaries?
_ maybe the version can report in 0.15 instead of 0.15.0
* If physical only, then we can't do extended logical checks.  It errors.
_ mail purge - `sysmail_delete_mailitems_sp` and `sysmail_delete_log_sp`

Terminology
-----------
* Action - checkdb, defrag, backup, stats, cleanUp, history

File Structure
--------------
```
logs/
├─ json/
└─ text/
   └─ 240712_130214_plan1.log
plans/
└─ plan1.toml
state/
└─ plan1.state.json
popmaint.exe
popmaint.toml
```

Fun Stuff
---------
* Write sql.Rows to a log file

Next Test #1
------------
* Monitor for blocking or blocked by and abort (EXECMON)
* Use `/pkg/execmon` to defrag

---------
```toml
servers = ["ab", "c"]
time_limit = "4h"
min_repeat_days = 7 # wait at least this long to defrag

[checkdb]
time_limit = "2h"
no_index = true
messages = true
max_size_mb = 100

[defrag]
[backup]
[server.abc]
checkdb.exclude = ["a", "b", "c"]
```

TOML App File - `popmaint.toml`
-------------------------------
[logging]
retain_days = 30

[logging.text]
[logging.json]
[logging.console]

