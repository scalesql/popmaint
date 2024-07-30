package main

import (
	"log/slog"
	"os"
)

func main() {
	slog.Info("startup...")
	host, _ := os.Hostname()
	l0 := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With(
		slog.Group("global", slog.Group("host", slog.String("name", host))))
	l0.Info("line 2")
	l1 := l0.WithGroup("popmaint")
	l1.Info("line 3", slog.Int("test", 7))
	l1.Info("line 4", slog.Int("a.b.c", 99))
}
