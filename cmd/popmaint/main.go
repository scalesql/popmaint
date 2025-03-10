package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"time"

	"github.com/scalesql/popmaint/internal/app"
	"github.com/scalesql/popmaint/internal/build"
	"github.com/scalesql/popmaint/internal/failure"
)

func main() {
	var version bool
	var exitCode int
	var panicFlag bool
	var printEnv bool
	var cmdLine app.CommandLine
	var help bool

	defer failure.HandlePanic(build.Commit(), build.Built().Format(time.RFC3339))

	exename, err := os.Executable()
	if err != nil {
		fmt.Println(err.Error())
		exit(1)
	}
	flag.BoolVar(&cmdLine.NoExec, "noexec", false, "do not execute the DBCC (display only)")
	flag.BoolVar(&cmdLine.Dev, "dev", false, "enable DEV settings")
	flag.BoolVar(&version, "version", false, "print the version and exit")
	flag.IntVar(&exitCode, "exit", 0, "if not zero, exit immediately with this code")
	flag.BoolVar(&panicFlag, "panic", false, "panic and exit")
	flag.StringVar(&cmdLine.LogLevel, "log-level", "", "log level (trace, debug, verbose, info, warn, error)")
	flag.BoolVar(&printEnv, "env", false, "print related environment variables")
	flag.BoolVar(&help, "help", false, "print this help message")
	flag.BoolVar(&help, "?", false, "print this help message")

	flag.Parse()
	if exitCode != 0 {
		fmt.Println("handling -exit flag")
		exit(exitCode)
		return
	}

	if help {
		fmt.Println("usage: popmaint.exe [flags] <plan>")
		flag.PrintDefaults()
		return
	}

	if panicFlag {
		fmt.Println("popmaint.exe: handling -panic flag")
		panic("panic: handling -panic flag")
	}
	if version {
		fmt.Printf("EXE:  %s: %s (%s) built %s\n", exename, build.Version(), build.Commit(), build.Built())
		printextraversion()
		return
	}
	if printEnv {
		printenv()
		return
	}

	if len(flag.Args()) == 0 {
		fmt.Println("usage: popmaint.exe [flags] <plan>")
		fmt.Println("FATAL: plan is required")
		exit(1)
	}
	cmdLine.Plan = flag.Arg(0)

	exitCode = app.Run(cmdLine, os.Getenv)
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

func printenv() {
	vars := os.Environ()
	for _, v := range vars {
		fmt.Println(v)
	}
}

// print the hostname and username
func printextraversion() {
	hn, err := os.Hostname()
	if err != nil {
		fmt.Println("ERROR (HOST):", err.Error())
	} else {
		fmt.Println("HOST:", hn)
	}

	u, err := user.Current()
	if err != nil {
		fmt.Println("ERROR (USER):", err.Error())
	} else {
		fmt.Printf("USER: %s (%s)\n", u.Username, u.Name)
	}
}
