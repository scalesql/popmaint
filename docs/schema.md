State Repository Schema
=======================

Tables
------
* Schema deploy tables
* dbo.dbcc_state
* dbo.dbcc_log
* dbo.stats_state
* dbo.stats_log
* dbo.error_log (?)

dbo.state_json (?)
------------------
* plan_name
* state_json
* last_updated

dbo.dbcc_state
--------------
* domain_name 
* server_name
* database_name
* last_dbcc datetimeoffset(0)
* ag_name
* ag_role 

dbo.app_log
-----------
* start_at
* host, user, plan, version, 
* completed_at
* errors???

dbo.dbcc_log
------------
* domain_name
* server_name
* database_name
* completed_at datetimeoffset(0)
* plan_name
* version
* host, user, pid, etc.
* duration
* result, errors, log, etc. -- All NVARCHAR(MAX)
* settings...


State
-----
* DBCC - domain, server, database
* Stats - just look in each database
* defrag - domain, server, database, object
* history - nope
* backup - nope (or domain, server, database)

