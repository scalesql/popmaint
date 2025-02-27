package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"time"

	"github.com/scalesql/popmaint/internal/lockmon"
)

func main() {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{Level: slog.LevelDebug}),
	)
	log.Info("Starting lockmon.exe...")
	server := os.Args[1] // "sqlserver://host/instance"
	stmt := "RAISERROR('ONE', 0, 1) WITH NOWAIT; PRINT 'TWO'; WAITFOR DELAY '00:00:10'; PRINT 'THREE';"
	if len(os.Args) > 2 {
		stmt = os.Args[2]
	}
	pool, err := sql.Open("sqlserver", server)
	if err != nil {
		log.Error(err.Error())
		return
	}
	defer pool.Close()
	// lockmon.TraceLogging = true // enable tracing
	r := lockmon.ExecMonitor(ctx, log, pool, stmt, time.Duration(0))
	if r.Err != nil {
		log.Error(r.Err.Error())
		return
	}
}
