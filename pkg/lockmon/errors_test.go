package lockmon

import (
	"fmt"
	"testing"

	mssql "github.com/microsoft/go-mssqldb"
	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	// basic error
	assert := assert.New(t)
	e1 := mssql.Error{Number: 50000, State: 1, Class: 16, LineNo: 1, Message: "Test"}
	e1str := FormatRootError(e1)
	assert.Equal("Msg 50000, Level 16, State 1, Line 1: Test", e1str)

	// stream error
	e2 := mssql.StreamError{InnerError: e1}
	assert.Equal(e1str, FormatRootError(e2))

	// diagnostics
	e3 := e1
	e3.ServerName = "TEST\\TXN"
	e3.ProcName = "TheProc"
	e3diag := ErrorDiagnostics(e3)
	assert.Equal("Server TEST\\TXN, Msg 50000, Level 16, State 1, Procedure TheProc, Line 1: Test", e3diag)

	// diagnostics with All populated
	e3.All = []mssql.Error{e1, e1}
	want := fmt.Sprintf("%s\n[0] %s\n[1] %s", e3diag, e1str, e1str)
	assert.Equal(want, ErrorDiagnostics(e3))

	// ServerError only has private fields
	// RetryableError only has private fields
}
