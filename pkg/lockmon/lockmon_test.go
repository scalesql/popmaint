package lockmon

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNilSQL(t *testing.T) {
	assert := assert.New(t)
	//pool := sql.DB{}
	r := ExecMonitor(context.Background(), nil, nil, "", time.Duration(0))
	assert.NoError(r.Err)
}
