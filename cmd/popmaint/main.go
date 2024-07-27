package main

import (
	"flag"
	"log"
	"os"
	"popmaint/pkg/app"
)

func main() {
	var plan string
	var noexec bool
	var dev bool

	flag.StringVar(&plan, "plan", "", "plan to run")
	flag.BoolVar(&noexec, "noexec", false, "do not execute the DBCC (display only)")
	flag.BoolVar(&dev, "dev", false, "enable DEV settings")
	flag.Parse()
	if plan == "" {
		log.Fatal("fatal: --plan is required")
		return
	}
	exitCode := app.Run(dev, plan, noexec)
	if exitCode > 125 {
		exitCode = 125
	}
	if exitCode < 0 {
		exitCode = 0
	}
	os.Exit(exitCode)
}
