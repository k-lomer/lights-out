package clients

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Test that outages are listed from the live National Grid Distribution endpoint.
func Test_ListNationalGridDistributionOutages(t *testing.T) {
	ctx := context.Background()
	var client = &http.Client{
		Timeout: 30 * time.Second,
	}
	_, err := MakeNationalGridDistributionClient(client).ListOutages(ctx)
	assert.NoError(t, err)
}
