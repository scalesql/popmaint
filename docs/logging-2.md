Logging v2
==========
* TOML file that holds configuration (popmaint.toml)
* All the `global` and `logstah` fields should be set at the start
* All app specific fields go in the `popmaint` bucket
    * `action` field that holds app, checkdb, defrag, backup, history, etc.
    * Common fields are domain, server, computer, instance, database, schema, object


TOML file
---------
```toml
[logging]
retain_days = 30

[logging.json]
payload_field = "popmaint"

[logging.fields]
global.host.name = "d40"
global.vesion = "version()"
```

Functions
---------
* hostname()
* version()
* git()
* built() -- date time
* toupper(x)
* tolower(x)
* now()


GO Functions (logx package?)
------------
```go
func dotsToSlog(key string, value any) slog.Attr {
    // split key on "."
    // last one is a Value
    // everyone above is a group
}

type KV is key, value

func dotsToAttrs([]KV) []slogAttr {
    // how do I get an array of stuff?
}
```


Fields
------
```toml
logstash.logtype="exe_name()"
global.host.name="hostname()"
"global.log.vendor:Microsoft"
"global.log.application:MSSQL"
"global.log.type:Application"
global.log.version = "version()"
global.log.hash = "git()"
global.log.vendor = "scaleSQL" # vendor()?
global.log.application = "popmaint"
global.log.version = "version()"

global.log.type="Application"
# global.log.collector.application="exe_name()",
# global.log.collector.version="version()" 
# global.log.collector.hash="git()"
# global.log.collector.host="hostname()"

global.log.lvl.value=6
global.log.lvl.keyword="info"
global.log.lvl.severity="6.Informational"

"mssql.name:sqleventtype"
"mssql.mssql_computer:global.host.name"
"mssql.mssql_fqdn:global.host.fqdn"
"mssql.mssql_version:global.log.version"
#"mssql.name:global.log.name"
"mssql.xe_severity_value:global.log.lvl.value"
"mssql.xe_severity_keyword:global.log.lvl.keyword"
```


SQLXEWriter
-----------
```toml
adds =  [   
			"logstash_type:corpsql",
            "logstash.pipeline: corpsql",
            "logstash.logtype: xelogstash.exe",
            "global.log.vendor:Microsoft",
            "global.log.application:MSSQL",
            "global.log.type:Application",
            "global.log.collector.application:xelogstash.exe", 
            
            "global.log.collector.version:'$(VERSION)'" ,
			"global.log.collector.hash:'$(GITHASH)'",
			"global.log.collector.host:$(HOST)",
			"mssql.xe_ingest_time:$(NOW)"
        ] 

copies = [  
			"mssql.name:sqleventtype",
			"mssql.mssql_computer:global.host.name",
            "mssql.mssql_fqdn:global.host.fqdn",
            "mssql.mssql_version:global.log.version",
			#"mssql.name:global.log.name",
			"mssql.xe_severity_value:global.log.lvl.value",
			"mssql.xe_severity_keyword:global.log.lvl.keyword"
        ]
```        
