Application Notes
=================





Architecture
------------
* /cmd/defrag/main.go
* app.Run(dev, plan, noexec) (/pkg/app/run.go)
    * Setup logging (getLogger) -- returns WithGroup("popmaint")
    * Read Config
    * engine := NewEngine(logger, st)
    * engine.runPlan(ctx, plan, noexec) (/pkg/app/chekdb)
        * engine.runCheckDB(ctx, plan, noexec) (/pkg/app/chekdb)
            * go through servers and get databases
            * process each database

* Build `App` struct that the `Config` structs hang off
* Also that state stuff hangs off
