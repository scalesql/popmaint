package lockmon

import (
	"fmt"

	mssql "github.com/microsoft/go-mssqldb"
)

// ErrorDiagnostics returns the formatted mssql.Error and any
// errors included in the `All` field.
func ErrorDiagnostics(err mssql.Error) string {
	out := fmterror(err)
	if err.ServerName != "" {
		out = fmt.Sprintf("Server %s, ", err.ServerName) + out
	}
	if len(err.All) == 0 {
		return out
	}
	for i, v := range err.All {
		out += fmt.Sprintf("\n[%d] %s", i, fmterror(v))
	}
	return out
}

// FormatRootError formats the root error from one of the different
// types of errors returned by the "mssql" package.
func FormatRootError(err error) string {
	return fmterror(RootError(err))
}

// func FormatSQLError()
// func SQLErrorDiagnostics()

// RootError attempts to get the root error from the different types of errors
// the "mssql" package can return.
func RootError(err error) error {
	if err == nil {
		return err
	}

	switch v := err.(type) {
	case mssql.Error: // no unwrap, just format, maybe handle All
		return v
	case mssql.StreamError: // check InnerError, unwrap if needed
		if v.InnerError == nil {
			return v
		} else {
			return cause(v.InnerError)
		}
	case mssql.ServerError: // unwrap
		return cause(v)
	case mssql.RetryableError: // unwrap
		return cause(v)
	default:
		return err
	}
}

func cause(err error) error {
	return unwrap(err, 0)
}

func unwrap(err error, depth int) error {
	if depth > 100 {
		return err
	}
	u, ok := err.(interface {
		Unwrap() error
	})
	if !ok {
		return err
	}
	new := u.Unwrap()
	return unwrap(new, depth+1)
}

func fmterror(err error) string {
	if err == nil {
		return ""
	}
	switch v := err.(type) {
	case mssql.Error:
		out := fmt.Sprintf("Msg %d, Level %d, State %d, ", v.Number, v.Class, v.State)
		if v.ProcName != "" {
			out += fmt.Sprintf("Procedure %s, ", v.ProcName)
		}
		out += fmt.Sprintf("Line %d: %s", v.LineNo, v.Message)
		return out
	default:
		return v.Error()
	}
}
