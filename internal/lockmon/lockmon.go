package lockmon

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-sql/sqlexp"
	mssql "github.com/microsoft/go-mssqldb"
	"github.com/pkg/errors"
	"github.com/scalesql/popmaint/internal/build"
	"github.com/scalesql/popmaint/internal/failure"
	"github.com/scalesql/popmaint/internal/mssqlz"
)

// Enable extra logging mostly for DEV
var TraceLogging = false

type Result struct {
	SQLResult    sql.Result
	Err          error
	Source       string
	Success      bool
	LoggedErrors bool
	Sessions     []Session
}

type Session struct {
	SessionID  int16  `db:"session_id"`
	BlockingID int16  `db:"blocking_session_id"`
	LastWait   string `db:"last_wait_type"`
	WaitTime   int32  `db:"wait_time"`
	Statement  string `db:"statement_text"`
}

var ErrBlocking = fmt.Errorf("blocking detected")

type monitor struct {
	log     ExecLogger
	server  string
	spid    int16
	timeout time.Duration
	ch      chan Result
}

// ExecMonitor runs a SQL statement and watches to see if it is blocking anything.  If so, it kills it.
// This will log any errors in the SQL that runs.
// It will return blocking errors, or errors detecting blocking.
// It will also set a flag that it encountered an error.
func ExecMonitor(ctx context.Context, log ExecLogger, host string, stmt string, timeout, blocking, blocked time.Duration) Result {

	// setupt the two connection pools
	workpool, err := mssqlz.PoolWithSuffix(host, "master", "worker")
	if err != nil {
		return Result{Err: fmt.Errorf("mssqlz.poolwithsuffix.worker: %w", err)}
	}
	defer workpool.Close()

	// setup the monitor pool
	monpool, err := mssqlz.PoolWithSuffix(host, "master", "monitor")
	if err != nil {
		return Result{Err: fmt.Errorf("mssqlz.poolwithsuffix.monitor: %w", err)}
	}
	defer monpool.Close()

	if log == nil {
		log = nilwriter{}
	}
	var cancel context.CancelFunc
	if ctx == nil {
		ctx = context.Background()
	}

	// this context will cancel in case of blocking or the statement completes
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()

	// get one connection we will use to run the statement
	conn, err := workpool.Conn(ctx)
	if err != nil {
		return Result{Err: errors.Wrap(err, "pool.conn")}
	}
	defer conn.Close()

	// get the SPID for that session
	server, spid, err := getServerAndSession(ctx, conn)
	if err != nil {
		return Result{Err: errors.Wrap(err, "getserverandsession")}
	}
	if TraceLogging {
		log.Debug(fmt.Sprintf("spid=%d", spid))
	}

	// make a channel for the results
	// two GO routines will need to return results
	ch := make(chan Result, 2)
	defer close(ch)

	mon := monitor{
		server:  server,
		log:     log,
		spid:    spid,
		timeout: timeout,
		ch:      ch,
	}

	go mon.runStmt(ctx, conn, stmt, log)
	go mon.runMonitor(ctx, monpool, blocking, blocked, log)

	// get the first result. this is all we really care about
	// this will be the blocking monitor or the actual SQL result
	results := <-ch

	// once we have the first result (either the batch completed or we had blocking)
	// we cancel the context.  this will cancel the other GO routine
	cancel()
	<-ch // wait for the second GO routine to finish
	return results
}

func (mon *monitor) runMonitor(ctx context.Context, pool *sql.DB, blocking, blocked time.Duration, log ExecLogger) {
	var count976 = 0
	if pool == nil {
		log.Error("runmonitor: pool is nil")
		return
	}
	defer failure.HandlePanic(build.Commit(), build.Built().Format(time.RFC3339))
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
out:
	for {
		select {
		case <-ctx.Done():
			break out // there is a send down below
		case <-ticker.C:
			sessions, err := getBlocking(ctx, pool, mon.spid, blocking, blocked)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					// if the context is canceled, send a message on the channel so we don't block
					break out // there is a send down below
				}

				// check for error 976, if there are less than 10, keep going
				// https://github.com/scalesql/popmaint/issues/7
				var sqlErr mssql.Error
				if errors.As(err, &sqlErr) { // if this is a SQL Server error
					if sqlErr.Number == 976 {
						count976++
						if count976 < 10 { // if 976 and < 10, just keep polling
							continue
						}
					}
				}

				// if there is an error getting blocking, kill the whole thing and return
				// likely the server restarted or disconnected or something
				// just exit and live to fight another day
				mon.ch <- Result{Err: fmt.Errorf("getblocking: %w", err), Success: false, Source: "monitor", Sessions: sessions}
				return
			}

			if len(sessions) == 0 {
				continue // no blocking, only our session
			}

			// len(sessions) > 0 so there is blocking
			mon.ch <- Result{Err: ErrBlocking, Success: false, Source: "monitor", Sessions: sessions}
			return
		}
	}
	mon.ch <- Result{Err: nil, Success: true, Source: "monitor"}
}

func (mon *monitor) runStmt(ctx context.Context, conn *sql.Conn, stmt string, log ExecLogger) {
	defer failure.HandlePanic(build.Commit(), build.Built().Format(time.RFC3339))
	loggedErrors, err := mon.execStmtContext(ctx, conn, stmt, log)
	result := Result{Err: err, Source: "exec", LoggedErrors: loggedErrors, Success: err == nil}
	mon.ch <- result
}

func (mon *monitor) execStmtContext(ctx context.Context, conn *sql.Conn, stmt string, log ExecLogger) (bool, error) {
	errs := make([]error, 0)
	var loggedErrors bool

	// passing in retmsg as an arguement actives sqlexp.
	// we can use this to get the messages from the server
	retmsg := &sqlexp.ReturnMessage{}

	// make sure we have a context
	if ctx == nil {
		ctx = context.Background()
	}

	// make sure we have a cancel function
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)

	// if we have a timeout, let's use that
	if mon.timeout > time.Duration(0) {
		ctx, cancel = context.WithTimeout(ctx, mon.timeout)
	}
	defer cancel()

	rows, qe := conn.QueryContext(ctx, stmt, retmsg)
	if qe != nil {
		return loggedErrors, qe
	}
	defer rows.Close()

	results := true
	first := true
	for qe == nil && results {

		// get the message from the server
		msg := retmsg.Message(ctx)
		if TraceLogging {
			log.Info(fmt.Sprintf("[%s] msg: %T", mon.server, msg))
		}
		switch m := msg.(type) {
		case sqlexp.MsgNotice:
			log.Info(fmt.Sprintf("[%s] %s", mon.server, m.Message.String())) // this is the actual PRINT/RAISERROR message

			// handle any errors in this message
			// this captures KILL from the server
			switch e := m.Message.(type) {
			case mssql.Error:
				qe = handleError(mon.server, log, e)
			}
		case sqlexp.MsgError:
			log.Error(fmt.Sprintf("[%s] %s", mon.server, FormatRootError(m.Error)))
			errs = append(errs, m.Error)
			loggedErrors = true
			qe = handleError(mon.server, log, m.Error)
		case sqlexp.MsgRowsAffected:
			if m.Count == 1 {
				log.Info(fmt.Sprintf("[%s] (1 row affected)", mon.server))
			} else {
				log.Info(fmt.Sprintf("[%s] (%d rows affected)", mon.server, m.Count))
			}
		case sqlexp.MsgNextResultSet:
			// if no more rows, this will fall though because
			// results will be false
			results = rows.NextResultSet()
			if err := rows.Err(); err != nil {
				// This is where "context canceled" errors shows up
				// which is context.Canceled
				if !errors.Is(err, context.Canceled) {
					return true, fmt.Errorf("[%s] statement_timeout exceeded: %s", mon.server, mon.timeout.String())
				}
				// otherwise, the error is likely a SQL Server
				qe = handleError(mon.server, log, err)
				loggedErrors = true
				errs = append(errs, err)

			}
			if results {
				first = true
			}
		case sqlexp.MsgNext: // next row
			for rows.Next() {
				if first { // handle headers
					first = false
					log.Warn("MSSQL result set discarded")
				}
				// scan for values, if we cared about the rows
			}
		default:
			log.Debug("unknown message type")
		}
	}
	if len(errs) == 0 {
		return loggedErrors, nil
	}
	return loggedErrors, errors.New("errors occurred")
}

func getServerAndSession(ctx context.Context, conn *sql.Conn) (string, int16, error) {
	var server string
	var spid int16
	err := conn.QueryRowContext(ctx, "SELECT @@SERVERNAME, @@SPID;").Scan(&server, &spid)
	if err != nil {
		return "", 0, errors.Wrap(err, "select @@servername, @@spid")
	}
	return server, spid, err
}

func getBlocking(ctx context.Context, pool *sql.DB, spid int16, blocking, blocked time.Duration) ([]Session, error) {
	sessions := make([]Session, 0)

	// set default values for blocking and blocked
	if blocking == 0 {
		blocking = time.Duration(1 * time.Second)
	}
	if blocked == 0 {
		blocked = time.Duration(5 * time.Second)
	}

	rows, err := pool.QueryContext(ctx, `
		SELECT
			r.session_id,
			COALESCE(r.blocking_session_id, 0) AS blocking_session_id,
			COALESCE(r.last_wait_type, '') AS last_wait_type,
			r.wait_time,
			COALESCE(SUBSTRING(h.[text], (r.statement_start_offset/2)+1,   
				((CASE r.statement_end_offset  
				WHEN -1 THEN DATALENGTH(h.text)  
				ELSE r.statement_end_offset  
				END - r.statement_start_offset)/2) + 1), '') AS statement_text   
		FROM sys.dm_exec_requests r
		CROSS APPLY sys.dm_exec_sql_text(r.sql_handle) h
		WHERE 	(wait_time > @blocking AND blocking_session_id = @spid) -- we are blocking
		OR 		(wait_time > @blocked AND session_id = @spid and blocking_session_id IS NOT NULL AND blocking_session_id <> 0) -- we are blocked
	`, sql.Named("spid", spid), sql.Named("blocking", blocking.Milliseconds()), sql.Named("blocked", blocked.Milliseconds()))
	if err != nil {
		return sessions, err
	}
	defer rows.Close()
	for rows.Next() {
		s := Session{}
		err = rows.Scan(&s.SessionID, &s.BlockingID, &s.LastWait, &s.WaitTime, &s.Statement)
		if err != nil {
			return sessions, err
		}
		sessions = append(sessions, s)
	}
	err = rows.Err()
	return sessions, err
}

// handleError handles any errors returned in a message.  This code is mostly copied from
// github.com/microsoft.com/sqlcmd/pgk/sqlcmd/sqlcmd.go.
func handleError(server string, log ExecLogger, err error) error {
	if err == nil {
		return nil
	}
	// we really only return on 127 or if the server tells us to
	// var minSeverityToExit uint8 = 11

	var errNumber int32
	var errSeverity uint8
	var errState uint8

	switch sqlError := err.(type) {
	case mssql.Error:
		errNumber = sqlError.Number
		errSeverity = sqlError.Class
		errState = sqlError.State
		if TraceLogging {
			log.Info(fmt.Sprintf("[%s] Msg %d, Level %d, State %d", server, errNumber, errSeverity, errState))
		}
	case mssql.StreamError:
		return sqlError
	case mssql.ServerError:
		return sqlError
	case mssql.RetryableError:
		return sqlError.Unwrap() // this should already be retried
	}

	// 127 is the magic exit code
	if errState == 127 {
		return ErrExitRequested
	}

	// I want all the errors until I find a case where I don't
	// if errSeverity >= minSeverityToExit {
	// 	return ErrExitRequested
	// }
	return nil
}
