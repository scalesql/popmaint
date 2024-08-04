package mssqlz

import (
	_ "github.com/microsoft/go-mssqldb"
)

// type ExecLogger interface {
// 	Info(msg string, args ...any)
// 	Error(msg string, args ...any)
// 	Debug(msg string, args ...any)
// }

// func ExecContext(ctx context.Context, pool *sql.DB, stmt string, logger ExecLogger) error {
// 	// TODO Handle multiple errors
// 	var err error
// 	retmsg := &sqlexp.ReturnMessage{}
// 	rows, qe := pool.QueryContext(ctx, stmt, retmsg)
// 	if qe != nil {
// 		return qe
// 	}
// 	defer rows.Close()
// 	results := true
// 	first := true
// 	for qe == nil && results {
// 		msg := retmsg.Message(ctx)
// 		switch m := msg.(type) {
// 		case sqlexp.MsgNotice:
// 			//log.Println(m.Message.String())
// 			//out.WriteMessage(m)
// 			logger.Info(m.Message.String(), slog.Bool("sql_output", true))
// 		case sqlexp.MsgError:
// 			//out.WriteError(m.Error)
// 			logger.Error(m.Error.Error(), slog.Bool("sql_output", true))
// 			//logrus.Debug("    MsgError")
// 			//logrus.Errorf("Error: %s\n", m.Error)
// 			// log.Println("Error:", m.Error)
// 			// switch e := m.Error.(type) {
// 			// case mssql.Error:
// 			// 	println(e.State)
// 			// 	println(e.Number)
// 			// 	println(e.Class)
// 			// 	println(e.Message)
// 			// 	//log.Printf("Error: %s\n", e)
// 			// default:
// 			// 	//log.Printf()
// 			// }
// 		case sqlexp.MsgRowsAffected:
// 			if m.Count == 1 {
// 				logger.Info("(1 row affected)", slog.Bool("sql_output", true))
// 			} else {
// 				logger.Info(fmt.Sprintf("(%d rows affected)\n", m.Count), slog.Bool("sql_output", true))
// 			}
// 		case sqlexp.MsgNextResultSet:
// 			// TODO: reset the "qe" value
// 			results = rows.NextResultSet()
// 			//log.Printf("sqlexp.MsgNextResultSet: results: %v\n", results)
// 			if err = rows.Err(); err != nil {
// 				// retcode = -100
// 				// qe = s.handleError(&retcode, err)
// 				// s.Format.AddError(err)
// 				logger.Error(fmt.Sprintf("MsgNextResultSet: rows.Err(): %s\n", err), slog.Bool("sql_output", true))
// 			}
// 			if results {
// 				first = true
// 			}
// 		case sqlexp.MsgNext: // next row
// 			//var val int
// 			//out.WriteString("sqlexp.MsgNext")
// 			// TODO: return rows as "row: a=1 b=2 z='test'"
// 			// Send rows to out.WriteRows(*sql.Rows)
// 			for rows.Next() {
// 				if first {
// 					headers, _ := rows.Columns()
// 					logger.Info(fmt.Sprintf("header: %v", headers), slog.Bool("sql_output", true))
// 					first = false
// 				}
// 				// if err := rows.Scan(&val); err != nil {
// 				// 	return err
// 				// }
// 				// log.Printf("val=%d\n", val)
// 				logger.Info("a row", slog.Bool("sql_output", true))
// 			}
// 		}
// 	}
// 	return nil
// }
