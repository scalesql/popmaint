package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/scalesql/popmaint/pkg/app"
	"github.com/scalesql/popmaint/pkg/build"
)

func main() {
	var plan string
	var noexec bool
	var dev bool
	var version bool

	exename, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	flag.StringVar(&plan, "plan", "", "plan to run")
	flag.BoolVar(&noexec, "noexec", false, "do not execute the DBCC (display only)")
	flag.BoolVar(&dev, "dev", false, "enable DEV settings")
	flag.BoolVar(&version, "version", false, "print the version and exit")
	flag.Parse()

	if version {
		fmt.Printf("%s: %s (%s) built %s\n", exename, build.Version(), build.Commit(), build.Built())
		return
	}
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
