package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaxDop(t *testing.T) {
	assert := assert.New(t)
	type test struct {
		cores    int
		maxdop   int
		maxCores int
		maxPct   int
		want     int
		err      bool
	}
	tests := []test{
		{8, 8, 4, 0, 4, false},    // 4 of 8
		{8, 8, 0, 50, 4, false},   // 50% of 8
		{8, 8, 6, 50, 4, false},   // 50% of 8 or 6
		{10, 10, 0, 79, 7, false}, // 79% of 8
		{10, 10, 2, 79, 2, false}, // 79% of 8 or 2
		{10, 10, 24, 0, 0, false}, // 24 of 10 which is 0
		{2, 2, 0, 10, 1, false},   // 10% of 2 which is 1
		{2, 2, 2, 10, 1, false},   // 10% of 2 or 2 which is 1
		{16, 2, 0, 100, 0, false}, // 100% of 16 which should be zero
		{16, 2, 0, 101, 0, true},  // 100% of 16 which should be zero
		{16, 2, 0, 101, 0, true},  // 100% of 16 which should be zero
		{16, 4, 6, 100, 0, false}, // limited by maxdop
	}
	for i, tc := range tests {
		plan := Plan{MaxDopCores: tc.maxCores, MaxDopPercent: tc.maxPct}
		got, err := plan.MaxDop(tc.cores, tc.maxdop)
		if tc.err {
			assert.Error(err, "test #%d", i)
		} else {
			assert.NoError(err, "test #%d", i)
		}
		assert.Equal(tc.want, got, "test #%d", i)
	}
}
