Logging
=======
All logs are under the `popmaint` parent.
* time
* level
* msg
* popmaint
    * job_id
    * action = ["checkdb", "defrag", etc.]
    * duration (1m3s)?
    * duration_sec
    * app
        * version
        * commit
        * built
    * settings
        * no_exec
    * host
        * domain
        * name
        * OS?  Version?
        * pid
        * exe_path
    * runtime
        * gomaxprocs
        * goos (runtime.GOOS)
        * goarch (runtime.GOARCH)
    * server
        * domain, computer, name (@@SERVERNAME), instance
        * edition
        * version
        * version_major
        * enterprise (bool)
    * database
        * name
        * avail_group
        * size_mb
    * object
        * schema
        * table
        * index
        * size_mb
    * checkdb
        * settings
            *
    * defrag
        * stuff
