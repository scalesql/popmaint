package main

import (
	"context"
	"flag"
	"log"
	"popmaint/pkg/app"
	"popmaint/pkg/maint"
)

func main() {
	out := app.OutWriter{}
	ctx := context.Background()
	do := maint.DefragOptions{}
	var fqdn string

	flag.BoolVar(&do.NoIndex, "noindex", false, "Set the NOINDEX option")
	flag.BoolVar(&do.InfoMessage, "messages", false, "Display Info Messages (disable NO_INFOMSGS)")
	flag.BoolVar(&do.PhysicalOnly, "physical_only", false, "Set the PHYSICAL_ONLY option")
	flag.IntVar(&do.MaxSizeMB, "max_size", 0, "max database size to check (MB)")
	flag.StringVar(&fqdn, "fqdn", "", "server to run against")
	flag.BoolVar(&do.NoExec, "noexec", false, "do not execute the DBCC")
	flag.Parse()
	if fqdn == "" {
		log.Fatal("fatal: --fqdn is required")
		return
	}
	// log settings:
	out.WriteStringf("host: %s  noindex: %t  messages: %t  physical_only: %t  max_size: %d", fqdn, do.NoIndex, do.InfoMessage, do.PhysicalOnly, do.MaxSizeMB)
	err := app.CheckDB(ctx, &out, fqdn, do)
	if err != nil {
		out.WriteError(err)
	}
}
