package clients

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func Test_ListUkpnOutages(t *testing.T) {
	ctx := context.Background()
	var client = &http.Client{
		Timeout: 30 * time.Second,
	}
	_, err := MakeUKPowerNetworkClient(client).ListOutages(ctx)
	if err != nil {
		t.Error(err)
	}
}
