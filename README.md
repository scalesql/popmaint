PopMaint
========

TODO
----
* Look for plans in root folder first, then look in `plans`
    * Assume it has a `toml` extension -- but check for both
* databases to include and exclude -- case-insentive
* Skip any server less than 2014 and log a warning

Architecture
------------
```
/cmd/defrag/main.go
    app.DefragHost(fqdn) -- keeps state and restart
        mssqz.OnlineDatabases(fqdn)
        maint.Defrag(fqdn, database)
```
* Build `App` struct that the `Config` structs hang off
* Also that state stuff hangs off

Terminology
-----------
* Action - DBCC(CheckDB), Defrag, Backup, Stats, CleanUp

File Structure
--------------
```
logs/
├─ json/
└─ text/
   └─ 240712_130214_plan1_log
state/
└─ plan1.state.json
popmaint.exe
popmaint.toml
plan1.toml
plan2.toml
```

Finding Work
------------
* Get all the databases
* Go get the the last DBCC from state
* Sort by oldest DBCC, largest database
* Maybe read-only gets done less frequently?  Only show up every X days?
* Have a minimum number of days before we retest

Before Test
-----------
* Skip any server before SQL Server 2014
* Command-line options
    * ✅ FQDN
    * ✅ Settings: noindex, physical only, max database size
* Log to file and console
    * Purge old log files
    * `popmaint_yymmdd_hhmmss_plan.(log|json)`

Fun Stuff
---------
* Parse the duration and limit it
* Write sql.Rows to a log file
* Write the log file and rotate it

Next Test #1
------------
* Monitor for blocking or blocked by and abort (EXECMON)
* Use `/pkg/execmon` to defrag

Next Test #2
------------
* More DBCC options
* Get all the databases from all the servers.  Sort oldestest, then largest, and go as afar as we can
* Add a time limit
* MAXDOP as pct/value
* database (optional)
* EXTENDED_LOGICAL_CHECKS, data purity

TOML Plan File - core-us-kc, core-br, uskc-epay, uskc-eft,
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


Appify
------
* Server list in file
* DBCC settings in file
* Save state based on this combo
* Write log file based on this combo
* Write log files to NDJSON format

Messages
--------
* Return 
    * message: text, JSON, stdout
    * error: text, JSON, stdout
    * RowSet: text, stdout
* Writer vs Logger, one vs many
* JSON is only for ELK
* Options
    * Output that has message, error or RowSet
    * Send the results on a channel and something else handles them
    * Pass in a Writer
    * Pass in a "thing"
* "Thing" - Write message, write error, write RowSet

Maybe Future?
-------------
* Split up databases based on hash of the domain, server, instance, database
* Don't run on primary? or synchronous node?  Or a flag for that?
* Panic Handler

Issues
------
* If an error occurs, log it, and keep going.  But fail the EXE after we are done.
    * Rebooting server, permission, etc.