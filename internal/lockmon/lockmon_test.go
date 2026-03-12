package lockmon

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/billgraziano/mssqlh/v2"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNilSQL(t *testing.T) {
	assert := assert.New(t)
	//pool := sql.DB{}
	r := ExecMonitor(context.Background(), nil, "", "", time.Duration(0), time.Duration(0), time.Duration(0))
	assert.Error(r.Err)
}

// TODO TestBlocking -  just like blocked but wait  a few seconds inside the GO routine to start the SQL,
//
//	then we should be blocking that SQL

func setupBlockingTable() (*sql.DB, string, error) {
	server := os.Getenv("POPMAINT_DBSERVER")
	if server == "" {
		return nil, "", fmt.Errorf("POPMAINT_DBSERVER must be set")
	}
	pool, err := mssqlh.Open(server, "master")
	if err != nil {
		return nil, "", err
	}
	_, err = pool.Exec(`
		SET XACT_ABORT ON;
		IF OBJECT_ID('tempdb.dbo.popmaint_lockmon_test') IS NULL
			CREATE TABLE tempdb.dbo.popmaint_lockmon_test (Val VARCHAR(100));
		
		INSERT tempdb.dbo.popmaint_lockmon_test VALUES ('A');
		INSERT tempdb.dbo.popmaint_lockmon_test VALUES ('B');
	`)
	if err != nil {
		return nil, "", err
	}
	return pool, server, nil
}

func teardownBlockingTable(pool *sql.DB) {
	_, _ = pool.Exec(`
		SET XACT_ABORT ON;
		IF OBJECT_ID('tempdb.dbo.popmaint_lockmon_test') IS NOT NULL
			DROP TABLE tempdb.dbo.popmaint_lockmon_test;
	`)

}

func TestBlocking(t *testing.T) {

	t.Run("TestBlocked", func(t *testing.T) { // we are blocked by a session
		require := require.New(t)
		assert := assert.New(t)
		pool, fqdn, err := setupBlockingTable()
		require.NoError(err)
		defer pool.Close()
		defer teardownBlockingTable(pool)

		var wg sync.WaitGroup
		wg.Add(1)
		// start a transaction and hold it
		go func(pool *sql.DB, wg *sync.WaitGroup) {
			_, _ = pool.Exec(`
			WHILE @@TRANCOUNT > 0 ROLLBACK TRAN ;
			BEGIN TRAN 
				SELECT * FROM tempdb.dbo.popmaint_lockmon_test WITH(UPDLOCK, TABLOCK);
				WAITFOR DELAY '00:00:04'
			ROLLBACK TRAN `)
			wg.Done()
		}(pool, &wg)
		time.Sleep(500 * time.Millisecond)
		result := ExecMonitor(context.Background(), consoleWriter{}, fqdn, `
		WHILE @@TRANCOUNT > 0 ROLLBACK TRAN ;
		BEGIN TRAN 
			SELECT * FROM tempdb.dbo.popmaint_lockmon_test WITH(UPDLOCK, TABLOCK);
		ROLLBACK TRAN `,
			time.Duration(0), time.Duration(10*time.Second), time.Duration(2*time.Second))
		assert.Error(result.Err)
		assert.True(errors.Is(result.Err, ErrBlocking))
		assert.Equal(1, len(result.Sessions))
		wg.Wait()
	})

	t.Run("TestBlockingAnother", func(t *testing.T) {
		require := require.New(t)
		assert := assert.New(t)
		pool, fqdn, err := setupBlockingTable()
		require.NoError(err)
		defer pool.Close()
		defer teardownBlockingTable(pool)

		var wg sync.WaitGroup
		wg.Add(1)
		// start a session that will start a transaction after the main session
		go func(pool *sql.DB, wg *sync.WaitGroup) {
			defer wg.Done()
			_, _ = pool.Exec(`
			WAITFOR DELAY '00:00:02' -- wait for the other session
			WHILE @@TRANCOUNT > 0 ROLLBACK TRAN ;
			BEGIN TRAN 
				SELECT TOP 1 * FROM tempdb.dbo.popmaint_lockmon_test WITH(UPDLOCK, TABLOCK);
			ROLLBACK TRAN `)

		}(pool, &wg)
		time.Sleep(500 * time.Millisecond)
		result := ExecMonitor(context.Background(), consoleWriter{}, fqdn, `
		-- This TRAN should run first because of the WAITFOR above
		SET XACT_ABORT ON; -- This is required to make this work
		WHILE @@TRANCOUNT > 0 ROLLBACK TRAN ;
		BEGIN TRAN 
			SELECT TOP 2 * FROM tempdb.dbo.popmaint_lockmon_test WITH(UPDLOCK, TABLOCK);
			WAITFOR DELAY '00:00:10'; -- simulate running for a long time
		ROLLBACK TRAN `,
			time.Duration(0), time.Duration(4*time.Second), time.Duration(10*time.Second))
		assert.Error(result.Err)
		assert.True(errors.Is(result.Err, ErrBlocking))
		assert.Equal(1, len(result.Sessions))
		t.Log("done")
		wg.Wait()
	})

	t.Run("TestNotBlocked", func(t *testing.T) {
		// This test doesn't block because we wait 10s for a 3s transaction
		require := require.New(t)
		assert := assert.New(t)

		pool, fqdn, err := setupBlockingTable()
		require.NoError(err)
		defer pool.Close()
		defer teardownBlockingTable(pool)

		// start a transaction and hold it for 3 seconds
		go func(pool *sql.DB) {
			_, _ = pool.Exec(`
			WHILE @@TRANCOUNT > 0 ROLLBACK TRAN ;
			BEGIN TRAN 
				SELECT 1 FROM tempdb.dbo.popmaint_lockmon_test WITH(UPDLOCK, TABLOCK);
				WAITFOR DELAY '00:00:03'
			ROLLBACK TRAN 
`)
		}(pool)
		time.Sleep(500 * time.Millisecond) // sometimes the second statement runs first
		// run against the transaction with a 10 second blocked timeout
		result := ExecMonitor(context.Background(), consoleWriter{}, fqdn, `
		WHILE @@TRANCOUNT > 0 ROLLBACK TRAN ;
		BEGIN TRAN 
			SELECT 2 FROM tempdb.dbo.popmaint_lockmon_test WITH(UPDLOCK, TABLOCK);
		ROLLBACK TRAN `,
			time.Duration(0), time.Duration(10*time.Second), time.Duration(10*time.Second))
		assert.NoError(result.Err)
		assert.False(errors.Is(result.Err, ErrBlocking))
		assert.Equal(0, len(result.Sessions))
	})
}
