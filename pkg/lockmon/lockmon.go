package lockmon

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-sql/sqlexp"
	"github.com/pkg/errors"
)

// TODO montior for blocking OR being blocked

// TODO
// This will log any errors in the SQL that runs
// it will return blocking errors, or errors detecting blocking
// it will also set a flag that it encountered an error
type Result struct {
	SQLResult    sql.Result
	Err          error
	Source       string
	Success      bool
	LoggedErrors bool
	// Blocking information?
}

var ErrBlocking = fmt.Errorf("blocking detected")

// var ErrBlocked = fmt.Errorf("being blocked")

type monitor struct {
	log     ExecLogger
	pool    *sql.DB
	spid    int16
	timeout time.Duration
	ch      chan Result
}

// ExecMonitor runs a SQL statement and watches to see if it is blocking anything.  If so, it kills it.
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
	//log.Info(fmt.Sprintf("spid=%d", spid))

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

	//log.Info("waiting for results, ok := <-ch...")

	// get the first result. this is all we really care about
	// this will be the blocking monitor or the actual SQL result
	results := <-ch
	//log.Debug(fmt.Sprintf("results: %+v  ok: %v\n", results, ok))
	//log.Debug("cancelling context...")

	// once we have the first result (either the batch completed or we had blocking)
	// we cancel the context.  this will cancel the other GO routine
	cancel()
	<-ch // wait for the second GO routine to finish
	//log.Info("second value!")
	// wait for all the senders to
	return results
}

func (mon *monitor) runMonitor(ctx context.Context) {
	//mon.log.Infof("runmon...")
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
out:
	for {
		select {
		case <-ctx.Done():
			//mon.log.Info("runmon: was cancelled")
			// ticker.Stop()
			break out
		case <-ticker.C:
			//mon.log.Info("runmon: ticker...")
			blocked, err := getBlockingCount(ctx, mon.pool, mon.spid)
			if err != nil { // if we get an error checking blocking, we are done
				mon.log.Error(fmt.Sprintf("runmon: getblockingcount: %v", err))
				mon.ch <- Result{Err: err, Success: false, Source: "monitor"}
				return
			}
			if blocked > 0 {
				// TODO get more blocking information
				mon.ch <- Result{Err: ErrBlocking, Success: false, Source: "monitor"}
				return
			}
		}
	}
	//println("here?")
	//mon.log.Info("runmon: sending... (how do I get here?)")
	mon.ch <- Result{Err: nil, Success: true, Source: "monitor"}
}

func (mon *monitor) runStmt(ctx context.Context, conn *sql.Conn, stmt string, log ExecLogger) {
	// r, err := conn.ExecContext(ctx, stmt)
	// result := Result{SQLResult: r, Err: err, Source: "exec", Success: err == nil}
	loggedErrors, err := mon.execStmtContext(ctx, conn, stmt, log)
	result := Result{Err: err, Source: "exec", LoggedErrors: loggedErrors, Success: err == nil}
	mon.ch <- result
}

func (mon *monitor) execStmtContext(ctx context.Context, conn *sql.Conn, stmt string, log ExecLogger) (bool, error) {
	errs := make([]error, 0)
	var loggedErrors bool

	retmsg := &sqlexp.ReturnMessage{}
	rows, qe := conn.QueryContext(ctx, stmt, retmsg)
	if qe != nil {
		return loggedErrors, qe
	}
	defer rows.Close()
	results := true
	first := true
	for /*qe == nil && */ results {
		msg := retmsg.Message(ctx)
		switch m := msg.(type) {
		case sqlexp.MsgNotice:
			//println(m.Message.String())
			log.Info(m.Message.String())
		case sqlexp.MsgError:
			//println("ERROR:", m.Error.Error())
			//println(FormatRootError(m.Error))
			//errs = append(errs, m.Error)
			log.Error(FormatRootError(m.Error))
			errs = append(errs, m.Error)
			loggedErrors = true
		case sqlexp.MsgRowsAffected:
			if m.Count == 1 {
				log.Info("(1 row affected)")
			} else {
				log.Info(fmt.Sprintf("(%d rows affected)", m.Count))
			}
		case sqlexp.MsgNextResultSet:
			// TODO: reset the "qe" value
			results = rows.NextResultSet()
			//log.Printf("sqlexp.MsgNextResultSet: results: %v\n", results)
			if err := rows.Err(); err != nil {
				// retcode = -100
				// qe = s.handleError(&retcode, err)
				// s.Format.AddError(err)
				// This is where "context canceled" error shows up
				// which is context.Canceled
				if !errors.Is(err, context.Canceled) {
					log.Error(fmt.Sprintf("MsgNextResultSet: rows.Err(): %s\n", err))
					loggedErrors = true
					errs = append(errs, err)
				}
			}
			if results {
				first = true
			}
		case sqlexp.MsgNext: // next row
			//var val int
			//out.WriteString("sqlexp.MsgNext")
			// TODO: return rows as "row: a=1 b=2 z='test'"
			// Send rows to out.WriteRows(*sql.Rows)
			for rows.Next() {
				if first {
					//headers, _ := rows.Columns()
					//logger.Info(fmt.Sprintf("header: %v", headers), slog.Bool("sql_output", true))
					first = false
					log.Warn("MSSQL result set discarded")
				}
				// if err := rows.Scan(&val); err != nil {
				// 	return err
				// }
				// log.Printf("val=%d\n", val)
				// TODO do something with this row logger.Info("a row", slog.Bool("sql_output", true))
			}
		}
	}
	if len(errs) == 0 {
		return loggedErrors, nil
	}
	return loggedErrors, errs[0]
}

func getSessionID(ctx context.Context, conn *sql.Conn) (int16, error) {
	var spid int16
	err := conn.QueryRowContext(ctx, "SELECT @@SPID;").Scan(&spid)
	if err != nil {
		return 0, errors.Wrap(err, "select @@spid")
	}
	return spid, err
}

// getBlockingCount returns the count of requests blocked by a sessioni
func getBlockingCount(ctx context.Context, pool *sql.DB, spid int16) (int, error) {
	var blocked int
	err := pool.QueryRowContext(ctx, `	SELECT COUNT(*) as blocked_sessions 
										FROM sys.dm_exec_requests 
										WHERE 	(wait_time > 1000 AND blocking_session_id = ?) -- we are blocking
										OR 		(wait_time > 5000 AND session_id = ? and blocking_session_id IS NOT NULL AND blocking_session_id <> 0); -- we are blocked`, spid, spid).Scan(&blocked)
	return blocked, err
}

// func MonitorLocking(ctx context.Context, lw Logger) {
// 	lw.Infof("lockmon: entering...")
// 	ticker := time.NewTicker(1 * time.Second)

// 	go func() {
// 		defer ticker.Stop()
// 		lw.Infof("lockmon: starting...")
// 		for {
// 			select {
// 			case <-ctx.Done(): // The parent will cancel the context which cancels us
// 				lw.Infof("lockmon: ctx.done()")
// 				return
// 			case <-ticker.C:
// 				// poll for locking
// 				lw.Tracef("lockmon: polling")
// 			}
// 		}
// 	}()
// }
