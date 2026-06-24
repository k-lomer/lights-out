package clients

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_ListSseOutages(t *testing.T) {
	ctx := context.Background()
	var client = &http.Client{
		Timeout: 30 * time.Second,
	}
	res, err := MakeSseClient(client).ListOutages(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}
