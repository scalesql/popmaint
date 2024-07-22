package mssqlz

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/golang-sql/sqlexp"
	_ "github.com/microsoft/go-mssqldb"
)

func ExecContext(ctx context.Context, pool *sql.DB, stmt string, out Outputer) error {
	// TODO Handle multiple errors
	var err error
	retmsg := &sqlexp.ReturnMessage{}
	rows, qe := pool.QueryContext(ctx, stmt, retmsg)
	if qe != nil {
		return qe
	}
	defer rows.Close()
	results := true
	first := true
	for qe == nil && results {
		msg := retmsg.Message(ctx)
		switch m := msg.(type) {
		case sqlexp.MsgNotice:
			//log.Println(m.Message.String())
			out.WriteMessage(m)
		case sqlexp.MsgError:
			out.WriteError(m.Error)
			//logrus.Debug("    MsgError")
			//logrus.Errorf("Error: %s\n", m.Error)
			// log.Println("Error:", m.Error)
			// switch e := m.Error.(type) {
			// case mssql.Error:
			// 	println(e.State)
			// 	println(e.Number)
			// 	println(e.Class)
			// 	println(e.Message)
			// 	//log.Printf("Error: %s\n", e)
			// default:
			// 	//log.Printf()
			// }
		case sqlexp.MsgRowsAffected:
			if m.Count == 1 {
				out.WriteString("(1 row affected)")
			} else {
				out.WriteString(fmt.Sprintf("(%d rows affected)\n", m.Count))
			}
		case sqlexp.MsgNextResultSet:
			// TODO: reset the "qe" value
			results = rows.NextResultSet()
			//log.Printf("sqlexp.MsgNextResultSet: results: %v\n", results)
			if err = rows.Err(); err != nil {
				// retcode = -100
				// qe = s.handleError(&retcode, err)
				// s.Format.AddError(err)
				out.WriteString(fmt.Sprintf("MsgNextResultSet: rows.Err(): %s\n", err))
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
					headers, _ := rows.Columns()
					out.WriteString(fmt.Sprintf("header: %v", headers))
					first = false
				}
				// if err := rows.Scan(&val); err != nil {
				// 	return err
				// }
				// log.Printf("val=%d\n", val)
				out.WriteString("a row")
			}
		}
	}
	return nil
}
