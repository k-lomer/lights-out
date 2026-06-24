package clients

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_ListNationalGridDistributionOutages(t *testing.T) {
	ctx := context.Background()
	var client = &http.Client{
		Timeout: 30 * time.Second,
	}
	res, err := MakeNationalGridDistributionClient(client).ListOutages(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}
