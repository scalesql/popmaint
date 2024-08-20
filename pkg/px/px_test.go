package px

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSetup(t *testing.T) {
	require := require.New(t)
	lx, err := setup(time.Now(), "plan", "payload")
	require.NotNil(lx)
	require.NoError(err)
}
