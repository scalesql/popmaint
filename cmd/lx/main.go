package main

import (
	"log"

	"github.com/scalesql/popmaint/internal/lx"
)

func main() {
	println("lx.main...")
	px, err := lx.NewConsoleLogger()
	if err != nil {
		log.Fatal(err)
	}
	px.Info("top 3 only")
	px.Info("field 1a", "1a", 111)
	px.AddFields("f1", 1)
	px.Info("defaults: f1")
	px.AddFields("f3", 3, "app.f4", 4)
	px.Info("defaults: f1, f3, app.f4")
	px.Info("defaults plus app.f3a", "app.f3a", 9)

	cx := px.WithFields("c1", 1)
	cx.Info("child with c1")
	cx = cx.WithFields("cx.f10", 19, "cx.f11", "test")
	cx.Warn("child with c1, cx.f10, fx.f11")
}
