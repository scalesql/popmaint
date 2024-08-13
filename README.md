PopMaint
========

TODO
----
_ license for non-commercial
* If physical only, then we can't do extended logical checks.  It errors.
* Need to incorporate ExecMon and writing the output to a logger (or a writer?)
    * Can I wrap a logger in a writer?  No.  I need INFO and ERR.
* Only log successful completions stuff through the passed in logger
    * All the "I'm working on stuff" goes through stdout to console
* Skip any server less than 2014 and log a warning
_ Get the functions and constants to the logger
_ color logging

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

