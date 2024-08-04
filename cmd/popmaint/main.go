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
	var exitCode int

	exename, err := os.Executable()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	flag.StringVar(&plan, "plan", "", "plan to run")
	flag.BoolVar(&noexec, "noexec", false, "do not execute the DBCC (display only)")
	flag.BoolVar(&dev, "dev", false, "enable DEV settings")
	flag.BoolVar(&version, "version", false, "print the version and exit")
	//flag.IntVar(&exitCode, "exit", 0, "exit immediately with this code")
	flag.IntVar(&exitCode, "exit", 0, "if not zero, exit immediately with this code")

	flag.Parse()
	if exitCode != 0 {
		exit(exitCode)
		return
	}
	if version {
		fmt.Printf("%s: %s (%s) built %s\n", exename, build.Version(), build.Commit(), build.Built())
		return
	}
	if plan == "" {
		log.Fatal("fatal: --plan is required")
		os.Exit(1)
	}
	exitCode = app.Run(dev, plan, noexec)
	exit(exitCode)
}

func exit(code int) {
	if code == 0 {
		os.Exit(0)
	}
	if code > 125 {
		os.Exit(125)
	}
	if code < 0 {
		os.Exit(0)
	}
	os.Exit(code)
}
