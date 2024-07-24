package main

import (
	"flag"
	"log"
	"os"
	"popmaint/pkg/app"
)

func main() {

	//ctx := context.Background()
	//do := maint.CheckDBOptions{}
	var plan string
	var noexec bool
	// flag.BoolVar(&do.NoIndex, "noindex", false, "Set the NOINDEX option")
	// flag.BoolVar(&do.InfoMessage, "messages", false, "Display Info Messages (disable NO_INFOMSGS)")
	// flag.BoolVar(&do.PhysicalOnly, "physical_only", false, "Set the PHYSICAL_ONLY option")
	// flag.IntVar(&do.MaxSizeMB, "max_size", 0, "max database size to check (MB)")
	flag.StringVar(&plan, "plan", "", "plan to run")
	flag.BoolVar(&noexec, "noexec", false, "do not execute the DBCC (display only)")
	flag.Parse()
	if plan == "" {
		log.Fatal("fatal: --plan is required")
		return
	}
	exitCode := app.Run(plan, noexec)
	if exitCode > 125 {
		exitCode = 125
	}
	if exitCode < 0 {
		exitCode = 0
	}
	os.Exit(exitCode)
}
