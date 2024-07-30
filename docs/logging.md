Logging
=======
* text logs includes only the message

TODO
----
* Create a CheckDB log struct
* And/or maybe a log struct for each type of event
* Each row gets some kind of descriptor {popmaint.type} -- or action?
    * app - start, setttings, summary, etc.
    * checkdb
    * defrag, backup, etc.

### Fields v2
* time
* msg
* global.host.name
* ...collector.version
* ...collector.name="popmaint.exe"
* popmaint{
    * plan, action, time_limit, duration
}

### Fields
* time
* msg
* level
* global
    * host
        * name
    * version
    * collector="popmaint.exe"
* popmaint
    * exec
        * plan
        * action (app, checkdb, backup, defrag, history, etc.)
        * time_limit (3m)
        * duration(_sec)
    * domain
    * server_name
    * computer
    * instance
    * databases (count of databases)
    * tempdb_mb
    * database
        * name
        * size_mb
        * schema
            * object
    * duration (string)
    * duration_sec (int)
    * checkdb - the struct from config
    * defrag
        * fragmentation

* server
    * name 
    * computer
    * instance
    * version
    * edition
    * databases
        * size_mb
        * schemas
            * object
* checkdb|defrag|etc.
    * time_limit

```go
type Attributes struct {
    Plan string
    Action string // enum or constants
    Domain string
    Server string
    Computer string
    Instance string
    Database string 
}
```
