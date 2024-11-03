package lockmon

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/billgraziano/mssqlh"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNilSQL(t *testing.T) {
	assert := assert.New(t)
	//pool := sql.DB{}
	r := ExecMonitor(context.Background(), nil, nil, "", time.Duration(0))
	assert.Error(r.Err)
}

func TestBlocking(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)
	//server := os.Getenv("POPMAINT_DBSERVER")
	server := "D40,53796"
	require.NotEmpty(server, "please set POPMAINT_DBSERVER environment for testing")
	pool, err := mssqlh.Open(server, "master")
	require.NoError(err)
	defer pool.Close()
	// start blocking for 10 seconds
	go func(pool *sql.DB) {
		_, _ = pool.Exec(`
			WHILE @@TRANCOUNT > 0 ROLLBACK TRAN ;
			BEGIN TRAN 
				SELECT TOP 10 *     FROM	AdventureWorks2016.Person.Person WITH(UPDLOCK, TABLOCK);
				WAITFOR DELAY '00:00:10'
			ROLLBACK TRAN 
`)
	}(pool)
	time.Sleep(1 * time.Second)
	go func(pool *sql.DB) {
		time.Sleep(1 * time.Second)
		_, _ = pool.Exec(`
			WHILE @@TRANCOUNT > 0 ROLLBACK TRAN ;
			BEGIN TRAN 
				SELECT TOP 10 *     FROM	AdventureWorks2016.Person.Person WITH(UPDLOCK, TABLOCK);
				WAITFOR DELAY '00:00:10'
			ROLLBACK TRAN 
`)
	}(pool)
	result := ExecMonitor(context.Background(), consoleWriter{}, pool, `WHILE @@TRANCOUNT > 0 ROLLBACK TRAN ;
		BEGIN TRAN 
			SELECT TOP 10 * FROM AdventureWorks2016.Person.Person WITH(UPDLOCK, TABLOCK);
		ROLLBACK TRAN `,
		time.Duration(0))
	assert.Error(result.Err)
	assert.True(errors.Is(result.Err, ErrBlocking))
	assert.Equal(2, len(result.Sessions))
}
