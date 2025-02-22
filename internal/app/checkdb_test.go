package app

import (
	"testing"
	"time"

	"github.com/scalesql/popmaint/internal/mssqlz"

	"github.com/stretchr/testify/assert"
)

func TestDatabaseSort(t *testing.T) {
	assert := assert.New(t)
	dd := []mssqlz.Database{
		{
			DatabaseName: "One",
			LastCheckDB:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			DatabaseMB:   100,
		},
		{
			DatabaseName: "Two",
			LastCheckDB:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			DatabaseMB:   300,
		},
		{
			DatabaseName: "One-Plus",
			LastCheckDB:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			DatabaseMB:   200,
		},
		{
			DatabaseName: "Three",
			LastCheckDB:  time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC),
			DatabaseMB:   100,
		},
		{
			DatabaseName: "Four",
			LastCheckDB:  time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC),
			DatabaseMB:   200,
		},
		{
			DatabaseName: "Five",
			LastCheckDB:  time.Date(1950, 1, 1, 0, 0, 0, 0, time.UTC),
			DatabaseMB:   200,
		},
	}
	sortDatabasesForDBCC(dd)
	assert.Equal("Four", dd[0].DatabaseName, "i=0")
	assert.Equal("Three", dd[1].DatabaseName, "i=1")
	assert.Equal("Two", dd[3].DatabaseName, "i=3")
	assert.Equal("One-Plus", dd[4].DatabaseName, "i=3")
	assert.Equal("One", dd[5].DatabaseName, "i=4")
}

func TestIntervalTooEarly(t *testing.T) {
	assert := assert.New(t)
	now := time.Now()
	db := mssqlz.Database{LastCheckDB: now.Add((-1 * 3 * 24) * time.Hour)}
	assert.True(intervalTooEarly(db, 7))
	assert.False(intervalTooEarly(db, 2))
}
