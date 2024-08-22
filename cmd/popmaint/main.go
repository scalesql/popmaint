package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/scalesql/popmaint/internal/failure"
	"github.com/scalesql/popmaint/pkg/app"
	"github.com/scalesql/popmaint/pkg/build"
)

func main() {
	var plan string
	var noexec bool
	var dev bool
	var version bool
	var exitCode int
	var panicFlag bool

	defer failure.HandlePanic(build.Commit(), build.Built().Format(time.RFC3339))

	exename, err := os.Executable()
	if err != nil {
		fmt.Println(err.Error())
		exit(1)
	}
	flag.StringVar(&plan, "plan", "", "plan to run")
	flag.BoolVar(&noexec, "noexec", false, "do not execute the DBCC (display only)")
	flag.BoolVar(&dev, "dev", false, "enable DEV settings")
	flag.BoolVar(&version, "version", false, "print the version and exit")
	flag.IntVar(&exitCode, "exit", 0, "if not zero, exit immediately with this code")
	flag.BoolVar(&panicFlag, "panic", false, "panic and exit")

	flag.Parse()
	if exitCode != 0 {
		fmt.Println("handling -exit flag")
		exit(exitCode)
		return
	}
	if panicFlag {
		fmt.Println("popmaint.exe: handling -panic flag")
		panic("panic: handling -panic flag")
	}
	if version {
		fmt.Printf("%s: %s (%s) built %s\n", exename, build.Version(), build.Commit(), build.Built())
		return
	}
	if plan == "" {
		fmt.Println("FATAL: --plan is required")
		exit(1)
	}
	exitCode = app.Run(dev, plan, noexec)
	exit(exitCode)
}

func exit(code int) {
	if code == 0 {
		os.Exit(0)
	}

	if code > 125 {
		code = 125
	}
	if code < 0 {
		code = 0
	}

	fmt.Printf("%s EXIT: %d\n", time.Now().Format("15:04:05"), code)
	os.Exit(code)
}
