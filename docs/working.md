Working Notes
=============


Estimate Only
-------------
* Final structure
    * server, database, db (mb), tempdb (mb), needed (mb), missing (mb)
    * Plan, needed, []string 
* Regular Expression: `^Estimated.+\s=\s(?P<kb>\d+)\.`
* Call `maint.CheckDB` with a writer that accumulates into an array
* Parse that array with the tempdb


