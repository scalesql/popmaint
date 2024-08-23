package main

import (
	"github.com/scalesql/popmaint/pkg/lx"
)

func main() {
	println("cx.main...")
	px := lx.NewConsoleLogger()
	px.Info("test", "f1", 1)
	cx := px.WithFields("f2", 2)
	cx.Log(lx.LevelInfo, "another test", "f3", 3)
	c2 := cx.WithFields("f4", 4)
	c2.Log(lx.LevelInfo, "final boss", "f5", 5)
}
