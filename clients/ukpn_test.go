package clients

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Test that outages are listed from the live UK Power Network endpoint.
func Test_ListUkpnOutages(t *testing.T) {
	ctx := context.Background()
	var client = &http.Client{
		Timeout: 30 * time.Second,
	}
	outages, err := MakeUKPowerNetworkClient(client).ListOutages(ctx)
	assert.NoError(t, err)
	for _, o := range outages {
		assert.NoError(t, o.Validate())
	}
}
