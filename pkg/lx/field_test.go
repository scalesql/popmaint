package lx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArgs2Map(t *testing.T) {
	assert := assert.New(t)
	type test struct {
		args   []any
		result map[string]any
	}
	// 0 1 2 3 4 5
	// a			(len=1)
	// a 1 			(len=2)
	// a 1 b       	(len=3)
	// a 1 b 2     	(len=4)
	// a 1 b 2 c   	(len=5)
	// a 1 b 2 c 3 	(len=6)

	tests := []test{
		{[]any{}, make(map[string]any)},
		{[]any{"a"}, map[string]any{
			"!BADKEY-0": "a",
		}},
		{[]any{7}, map[string]any{
			"!BADKEY-0": 7,
		}},
		{[]any{"a", 1}, map[string]any{
			"a": 1,
		}},
		{[]any{1, "a"}, map[string]any{
			"!BADKEY-0": 1,
			"!BADKEY-1": "a",
		}},
		{[]any{"a", 1, "b"}, map[string]any{
			"a":         1,
			"!BADKEY-2": "b",
		}},
		{[]any{"a", 1, "b", 2}, map[string]any{
			"a": 1,
			"b": 2,
		}},
		{[]any{"a", 1, "b", 2, 7}, map[string]any{
			"a":         1,
			"b":         2,
			"!BADKEY-4": 7,
		}},
		{[]any{"a", 1, "b", 2, "c", "x"}, map[string]any{
			"a": 1,
			"b": 2,
			"c": "x",
		}},
	}
	for _, tc := range tests {
		m := args2map(tc.args...)
		assert.Equal(len(tc.result), len(m))
		assert.Equal(tc.result, m)
	}
}
