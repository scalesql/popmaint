package lx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetup(t *testing.T) {
	require := require.New(t)
	lx, err := setup("xxx_jobid", "payload")
	require.NotNil(lx)
	require.NoError(err)
}
