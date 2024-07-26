package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaxDop(t *testing.T) {
	assert := assert.New(t)
	type test struct {
		cores    int
		maxCores int
		maxPct   int
		want     int
	}
	tests := []test{
		{8, 4, 0, 4},
		{8, 0, 50, 4},
		{8, 6, 50, 4},
		{10, 0, 79, 7},
		{10, 2, 79, 2},
		{10, 24, 0, 0},
	}
	for _, tc := range tests {
		plan := Plan{MaxDopCores: tc.maxCores, MaxDopPercent: tc.maxPct}
		got, _ := plan.MaxDop(tc.cores)
		assert.Equal(tc.want, got)
	}
}
