package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaxDop(t *testing.T) {
	assert := assert.New(t)
	type test struct {
		cores     int
		maxdop    int
		planCores int
		planPct   int
		want      int
		err       bool
	}
	tests := []test{
		{8, 8, 0, 0, 0, false},    // nothing is set
		{8, 8, -1, -1, 0, true},   // negative plan cores
		{8, 8, 4, 0, 4, false},    // 4 of 8
		{8, 8, 0, 50, 4, false},   // 50% of 8
		{8, 8, 6, 50, 4, false},   // 50% of 8 or 6
		{10, 10, 0, 79, 7, false}, // 79% of 8
		{10, 10, 2, 79, 2, false}, // 79% of 8 or 2
		{10, 10, 24, 0, 0, false}, // 24 of 10 which is 0
		{2, 2, 0, 10, 1, false},   // 10% of 2 which is 1
		{2, 0, 0, 10, 1, false},   // 10% of 2 which is 1
		{2, 2, 2, 10, 1, false},   // 10% of 2 or 2 which is 1
		{16, 2, 0, 100, 0, false}, // 100% of 16 which should be zero
		{16, 2, 0, 101, 0, true},  // 100% of 16 which should be zero
		{16, 2, 0, 101, 0, true},  // 100% of 16 which should be zero
		{16, 4, 6, 100, 0, false}, // limited by maxdop
		{4, 0, 1, 0, 1, false},    // #12, MAXDOP is not set
	}
	for i, tc := range tests {
		plan := Plan{MaxDopCores: tc.planCores, MaxDopPercent: tc.planPct}
		got, err := plan.MaxDop(tc.cores, tc.maxdop)
		if tc.err {
			assert.Error(err, "test #%d", i)
		} else {
			assert.NoError(err, "test #%d", i)
		}
		assert.Equal(tc.want, got, "test #%d", i)
	}
}

func TestMaxDopPctMaxDop(t *testing.T) {
	assert := assert.New(t)
	type test struct {
		cores         int
		maxdop        int
		planCores     int
		planPct       int
		planPctMaxdop int
		want          int
		err           bool
	}
	tests := []test{
		{8, 8, 0, 0, -1, 0, true},     // negative maxdop cores
		{8, 8, 4, 0, 0, 4, false},     // 4 of 8
		{8, 8, 0, 50, 0, 4, false},    // 50% of 8
		{8, 8, 6, 50, 0, 4, false},    // 50% of 8 or 6
		{10, 10, 0, 79, 0, 7, false},  // 79% of 8
		{10, 10, 2, 79, 0, 2, false},  // 79% of 8 or 2
		{10, 10, 24, 0, 0, 0, false},  // 24 of 10 which is 0
		{2, 2, 0, 10, 0, 1, false},    // 10% of 2 which is 1
		{2, 0, 0, 10, 0, 1, false},    // 10% of 2 which is 1
		{2, 2, 2, 10, 0, 1, false},    // 10% of 2 or 2 which is 1
		{16, 2, 0, 100, 50, 1, false}, // 50% of 2 which is 1
		{16, 4, 6, 100, 0, 0, false},  // limited by maxdop
		{4, 0, 1, 0, 0, 1, false},     // #12, MAXDOP is not set
		{8, 0, 16, 0, 0, 0, false},    // no value set, want 0
		{8, 4, 0, 0, 1, 1, false},     // 1% of maxdop is floored at 1
	}
	for i, tc := range tests {
		plan := Plan{MaxDopCores: tc.planCores, MaxDopPercent: tc.planPct, MaxDopPercentMaxDop: tc.planPctMaxdop}
		got, err := plan.MaxDop(tc.cores, tc.maxdop)
		if tc.err {
			assert.Error(err, "test #%d", i)
		} else {
			assert.NoError(err, "test #%d", i)
		}
		assert.Equal(tc.want, got, "test #%d", i)
	}
}
