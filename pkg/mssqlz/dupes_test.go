package mssqlz

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDupes(t *testing.T) {
	assert := assert.New(t)
	type test struct {
		srv  Server
		dupe bool
	}
	tests := []test{
		{Server{Domain: "D0", Computer: "C0", Instance: ""}, false},
		{Server{Domain: "D0", Computer: "C0", Instance: ""}, true},
		{Server{Domain: "d0", Computer: "C0", Instance: ""}, true},
		{Server{Domain: "d0", Computer: "C0", Instance: "I"}, false},
		{Server{Domain: "d0", Computer: "C0", Instance: "I"}, true},
	}
	dc := NewDupeCheck()
	for i, tc := range tests {
		dupe := dc.IsDupe(tc.srv)
		assert.Equal(tc.dupe, dupe, "test #%d", i)
	}
}
