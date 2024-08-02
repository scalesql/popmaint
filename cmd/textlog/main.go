package main

import (
	"log/slog"
	"os"

	"gitlab.com/greyxor/slogor"
)

func main() {
	println("-------------------------------------------------------------------------------| <- 80 chars")
	println("-----------------------------------------------------------------------------------------------------------------------| <- 120 chars")
	nestedAttr := []any{slog.Int("size", 37), slog.Group("global", slog.Group("host", slog.String("name", "D40"), slog.String("ip", "10.34.2.45")))}
	kv := []any{"k1", 1, "k2", "two"}
	slog.Info("Test #1")
	logger := slog.Default()
	logger.Info("Test #2")
	logger.Warn("Test #3")
	println("-------------------------------------------------------------------------------|")
	// slogor
	l2 := slog.New(slogor.NewHandler(os.Stdout, slogor.Options{
		ShowSource: true,
		Level:      slog.LevelDebug,
	}))
	l2.Debug("this is a test")
	l2.Info("Test")
	l2.Warn("OH my", kv...)
	l2.Error("oops")
	l2.Info("Junk", slog.Int("size", 37), slog.Group("global", slog.String("name", "D40"), slog.Int("mine", 875)))
	l2.Error("Something bad has happened", nestedAttr...)

	println("-------------------------------------------------------------------------------|")

}
