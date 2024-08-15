PopMaint
========

TODO
----
_ have separate errors for app errors and CHECKDB errors
_ license for non-commercial
_ parse the TOML into mappings
_ functions for the severity/level: severity_name(), severity_level()

* If physical only, then we can't do extended logical checks.  It errors.
* Skip any server less than 2014 and log a warning?
_ Get the functions and constants to the logger
_ color logging for dev?

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

