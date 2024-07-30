PopMaint
========

TODO
----
* build stuff for version, etc. w/PowerShell builder. Like IsItSQL maybe?
* Need to write better JSON field stuff
* logx?
* Just log to stdout until the logger is setup
* Look for plans in root folder first, then look in `plans`
    * Assume it has a `toml` extension -- but check for both
* Skip any server less than 2014 and log a warning
* `min_frequency` -- don't CheckDB unless this many days have passed
* Maybe I should log to a struct?  That has all the fields I want.  And maybe some subfields?  Like CheckDB.


Terminology
-----------
* Action - DBCC(CheckDB), Defrag, Backup, Stats, CleanUp

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

