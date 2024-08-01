package main

import (
	"log/slog"
	"strings"
)

func main() {
	slog.Info("Test #1")
	logger := slog.Default()
	logger.Info("Test #2")
	logger.Warn("Test #3")
	chars := strings.Repeat("X", 80)
	println(chars)
}
