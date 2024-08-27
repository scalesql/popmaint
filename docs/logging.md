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
    * event = checkdb completed, defrag completed, etc.
    * duration (1m3s)?
    * duration_sec
    * app
        * version, commit, built
        * exec // attributes of this execution
            * pid, user
            * no_exec, dev, debug, trace
            * path 
            * host 
            * OS?  OS vesion?
        * go
            * gomaxprocs, goos (runtime.GOOS), goarch (runtime.GOARCH) -- maybe just one time, goversion?
            * memory? threads?
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

Fields (Object)
---------------
* popmaint
    * server
        * edition, version, version_major, enterprise (bool)
    * object|id|bucket|container
        * fqdn
        * domain
        * computer
        * server (server\instance)
        * instance
        * avail_group
        * is_primary 
        * database
        * schema
        * table
        * index
        * path (/domain/server/etc.)
    
