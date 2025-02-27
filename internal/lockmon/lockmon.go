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
	pool    *sql.DB
	spid    int16
	timeout time.Duration
	ch      chan Result
}

// ExecMonitor runs a SQL statement and watches to see if it is blocking anything.  If so, it kills it.
// This will log any errors in the SQL that runs.
// It will return blocking errors, or errors detecting blocking.
// It will also set a flag that it encountered an error.
func ExecMonitor(ctx context.Context, log ExecLogger, pool *sql.DB, stmt string, timeout time.Duration) Result {
	if pool == nil {
		return Result{Err: fmt.Errorf("pool is nil")}
	}
	if log == nil {
		log = nilwriter{}
	}

	var cancel context.CancelFunc
	if ctx == nil {
		ctx = context.Background()
	}
	// if we were passed a timeout, add that to the context
	if timeout == time.Duration(0) {
		ctx, cancel = context.WithCancel(ctx)
	} else {
		ctx, cancel = context.WithTimeout(ctx, timeout)
	}
	defer cancel()

	// get one connection we will use to run the statement
	conn, err := pool.Conn(ctx)
	if err != nil {
		return Result{Err: errors.Wrap(err, "pool.conn")}
	}
	defer conn.Close()

	// get the SPID for that session
	spid, err := getSessionID(ctx, conn)
	if err != nil {
		return Result{Err: errors.Wrap(err, "getsession")}
	}
	if TraceLogging {
		log.Debug(fmt.Sprintf("spid=%d", spid))
	}

	// make a channel for the results
	// two GO routines will need to return results
	ch := make(chan Result, 2)
	defer close(ch)

	mon := monitor{
		log:     log,
		pool:    pool,
		spid:    spid,
		timeout: timeout,
		ch:      ch,
	}

	go mon.runStmt(ctx, conn, stmt, log)
	go mon.runMonitor(ctx)

	// get the first result. this is all we really care about
	// this will be the blocking monitor or the actual SQL result
	results := <-ch

	// once we have the first result (either the batch completed or we had blocking)
	// we cancel the context.  this will cancel the other GO routine
	cancel()
	<-ch // wait for the second GO routine to finish
	return results
}

func (mon *monitor) runMonitor(ctx context.Context) {
	defer failure.HandlePanic(build.Commit(), build.Built().Format(time.RFC3339))
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
out:
	for {
		select {
		case <-ctx.Done():
			break out // there is a send down below
		case <-ticker.C:
			sessions, err := getBlocking(ctx, mon.pool, mon.spid)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					// if the context is canceled, send a message on the channel so we don't block
					break out // there is a send down below
				}
				// log the error and keep going
				mon.log.Error(fmt.Errorf("getblocking: %w", err).Error())
				continue
			}

			if len(sessions) <= 1 {
				continue // no blocking, only our session
			}

			// len(sessions) > 1 so there is blocking
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

	retmsg := &sqlexp.ReturnMessage{}
	// passing in retmsg as an arguement actives sqlexp.
	// we can use this to get the messages from the server
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
			log.Info(fmt.Sprintf("msg: %T", msg))
		}
		switch m := msg.(type) {
		case sqlexp.MsgNotice:
			log.Info(m.Message.String()) // this is the actual PRINT/RAISERROR message

			// handle any errors in this message
			// this captures KILL from the server
			switch e := m.Message.(type) {
			case mssql.Error:
				qe = handleError(log, e)
			}
		case sqlexp.MsgError:
			log.Error(FormatRootError(m.Error))
			errs = append(errs, m.Error)
			loggedErrors = true
			qe = handleError(log, m.Error)
		case sqlexp.MsgRowsAffected:
			if m.Count == 1 {
				log.Info("(1 row affected)")
			} else {
				log.Info(fmt.Sprintf("(%d rows affected)", m.Count))
			}
		case sqlexp.MsgNextResultSet:
			// if no more rows, this will fall though because
			// results will be false
			results = rows.NextResultSet()
			if err := rows.Err(); err != nil {
				// This is where "context canceled" error shows up
				// which is context.Canceled
				qe = handleError(log, err)
				if !errors.Is(err, context.Canceled) {
					log.Error(fmt.Sprintf("msgnextresultset: rows.err(): %s\n", err))
					loggedErrors = true
					errs = append(errs, err)
				}
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

func getSessionID(ctx context.Context, conn *sql.Conn) (int16, error) {
	var spid int16
	err := conn.QueryRowContext(ctx, "SELECT @@SPID;").Scan(&spid)
	if err != nil {
		return 0, errors.Wrap(err, "select @@spid")
	}
	return spid, err
}

func getBlocking(ctx context.Context, pool *sql.DB, spid int16) ([]Session, error) {
	sessions := make([]Session, 0)
	// This should always return 1 or more sessions
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
		WHERE 	(wait_time > 1000 AND blocking_session_id = @spid) -- we are blocking
		OR 		(wait_time > 5000 AND session_id = @spid and blocking_session_id IS NOT NULL AND blocking_session_id <> 0) -- we are blocked
		OR		session_id = @spid ; -- and this session
	`, sql.Named("spid", spid))
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
func handleError(log ExecLogger, err error) error {
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
			log.Info(fmt.Sprintf("Msg %d, Level %d, State %d", errNumber, errSeverity, errState))
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
