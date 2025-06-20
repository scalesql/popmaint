POPMAINT
========================================
POPMAINT is a command-line SQL Server maintenance solution similar to SQL Server Maintenance Plans.  It is based on GIT-friendly configuration files.  **This is currently under active development and EVERYTHING is subject to change.**

Goals
----------------------------------------
This project exists for two reasons.  First, I want a way to centralize SQL Server maintenance.  I am tired of deploying jobs to every server and then trying to change their schedules or activities.  And I even wrote a tool to [deploy jobs to every server](https://www.sqldsc.com/)!

Second, I want to handle blocking.  Indexing can block. Other maintenance tasks can block.  POPMAINT polls the server every second and kills the maintenance if it is blocking anything.  I hope to expand this to Availability Group send and redo queue sizes.

Our use case for this is to add a new server by editing a file and committing to GIT.  Everything beyond that should be automated.

Features
----------------------------------------
* Command-line application that supports scheduling through SQL Server Agent or other job scheduling application.  I run it from our central ActiveBatch server.
* Configuration via GIT-friendly files
* Log to the console or JSON
* Cancel maintenance tasks if they block other requests
* Spread load across servers, time, and time zones
* Track state to resume where it left off and limit load
* CHECKDB
    * Support all DBCC CHECKDB options
    * Limit cores and runtime to throttle load



Getting Started
---------------
1. Unzip the files to a local folder
2. Create `./plans/sample.toml`

    ```toml
    servers = ["localhost"]

    [checkdb]
    time_limit = "60m"
    included = ["master", "msdb", "model"]
    no_index = true 
    physical_only = true 
    ```
3. Run the executable and execute the plan: `popmaint sample`.  This will run as the current user.  It requires `sysadmin` or `db_owner` permissions on these databases.
4. See `DOCS.html` for further documentation

This will run DBCC against the system databases.

Releases
--------------------------------------------------------------------
### vNext (June 2025)
* Improved MAXDOP calculation
* Improved logging
* Fixed formatting on the documentation
* Skip database snapshots
* Add `statement_timeout` option

### 0.22 (March 2025)
* CHECKDB skips `tempdb`
* MAXDOP settings now include an option to set a percentage of the server MAXDOP
* Fixed the display format of the last time DBCC ran in the log

### 0.21 (February 2025)
* First public release
