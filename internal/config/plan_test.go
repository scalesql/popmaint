package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlanRemoveDupes(t *testing.T) {
	assert := assert.New(t)
	plan := Plan{
		Servers: []string{
			"a",
			"a",
			"b",
			"x",
			"b",
			"b",
			"A",
		},
	}
	dupes := plan.RemoveDupes()
	assert.Equal([]string{"a", "b", "x"}, plan.Servers)
	assert.Equal([]string{"a", "b", "b", "A"}, dupes)
}
