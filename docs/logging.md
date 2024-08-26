Logging
=======
* All logs are under the `popmaint` parent.
* Consider an "id" bucket that is the server, database, object, etc.

Fields
------
* time
* level
* msg
* popmaint
    * job_id
    * action = ["checkdb", "defrag", etc.]
    * duration (1m3s)?
    * duration_sec
    * app
        * version, commit, built
        * settings
            * no_exec
        * host
            * domain, name, pid, executable (full path) 
            * OS?  Version?
        * runtime
            * gomaxprocs, goos (runtime.GOOS), goarch (runtime.GOARCH)
    * server
        * domain, computer, name (@@SERVERNAME), instance
        * edition
        * version
        * version_major
        * enterprise (bool)
        * fqdn
    * database
        * name
        * avail_group
        * size_mb
    * object
        * schema, table, index, size_mb
    * path
        * /domain/computer/instance/database/schema/object/index
        * /domain/ag/database/schema/object/index
        * /domain/ag/database
    * checkdb
        * settings
            * ...
    * defrag
        * stuff
