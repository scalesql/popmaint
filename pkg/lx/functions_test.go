package lx

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNestedMap1(t *testing.T) {
	assert := assert.New(t)
	src := map[string]any{
		"a": 1,
		"b": 2,
	}
	after, err := dotted2nested(src)
	assert.NoError(err)
	assert.Equal(src, after)
}
func TestNestedMap2(t *testing.T) {
	assert := assert.New(t)
	src := map[string]any{
		"a.c":     1,
		"b":       2,
		"a.d":     3,
		"m.n.o.p": 4,
	}
	after, err := dotted2nested(src)
	assert.NoError(err)
	assert.Equal(3, len(after))
}

func TestNestedDuplicate(t *testing.T) {
	assert := assert.New(t)
	src := map[string]any{
		"a.b.c": 1,
		"a.b.d": 2,
		"a.b":   3,
	}
	_, err := dotted2nested(src)
	assert.Error(err)
	//assert.Equal(src, after)
}

func TestNestedDuplicateValFirst(t *testing.T) {
	assert := assert.New(t)
	src := map[string]any{
		"a.b":   1,
		"a.b.d": 2,
	}
	_, err := dotted2nested(src)
	// t.Log(err)
	assert.Error(err)
	//assert.Equal(src, after)
}

func TestAny2MapOne(t *testing.T) {
	assert := assert.New(t)
	m := anys2map("", "a")
	assert.Equal(1, len(m))
	assert.Equal("!BADKEY", m["a"])
}

func TestAny2MapFour(t *testing.T) {
	assert := assert.New(t)
	m := anys2map("", "a", 1, "b", 2)
	assert.Equal(2, len(m))
	assert.Equal(1, m["a"])
	assert.Equal(2, m["b"])
}

func TestAny2MapFive(t *testing.T) {
	assert := assert.New(t)
	m := anys2map("", "a", 1, "b", 2, "c")
	assert.Equal(3, len(m))
	assert.Equal(1, m["a"])
	assert.Equal(2, m["b"])
	assert.Equal("!BADKEY", m["c"])
}

func TestParseFunction(t *testing.T) {
	assert := assert.New(t)
	type test struct {
		got any
		val any
		fld string
		fn  string
		err bool
	}
	now := time.Now()
	tests := []test{
		{"", nil, "", "", false},
		{"a", "a", "", "", false},
		{"a.b", "a.b", "", "", false},
		{"f()", nil, "", "f()", false},
		{"a.f()", nil, "a", "f()", false},
		{"a.b.f()", nil, "a.b", "f()", false},
		{"a().b()", nil, "", "", true},
		{"a.b().c", nil, "", "", true},
		{37, 37, "", "", false},
		{now, now, "", "", false},
	}
	for _, tc := range tests {
		val, field, function, err := parseFunc(tc.got)
		//println(len(fields))
		assert.Equal(tc.fld, field)
		assert.Equal(tc.fn, function)
		if tc.val != nil {
			assert.Equal(tc.val, val)
		} else {
			assert.Nil(tc.val)
		}
		if tc.err {
			assert.Error(err)
		} else {
			assert.NoError(err)
		}
	}
}

func TestApplyFuncs(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	hostname, err := os.Hostname()
	require.NoError(err)
	//now := time.Now()
	px, err := setup("xxxx_jobid", "payload")
	require.NoError(err)
	px.mappings = []KV{
		{"ab", 37},
		{"x.y", "new"},
		{"ghn", "hostname()"},
		{"gone", "delete()"},
		{"copied", "a.copy()"},
		{"moved", "b.move()"},
	}
	src := map[string]any{
		"a":    1,
		"gone": 19,
		"b":    11,
	}
	m, errs := px.applyFuncs(src)
	assert.Equal(0, len(errs), "too many errs")
	assert.Equal(1, m["a"])
	assert.Equal(37, m["ab"])
	assert.Equal("new", m["x.y"])
	assert.Equal(hostname, m["ghn"])
	// copy
	v, ok := m["gone"]
	assert.Equal(nil, v)
	assert.False(ok)
	assert.Equal(1, m["copied"])
	// move
	v, ok = m["b"]
	assert.Nil(v)
	assert.False(ok)
	v, ok = m["moved"]
	assert.Equal(11, v)
	assert.True(ok)

}
