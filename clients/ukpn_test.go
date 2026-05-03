package clients

import (
	"context"
	"net/http"
	"testing"
	"time"
)

var client *http.Client = &http.Client{
	Timeout: 30 * time.Second,
}

func Test_ListUkpnOutages(t *testing.T) {
	ctx := context.Background()
	_, err := ListUkpnOutages(ctx, client)
	if err != nil {
		t.Error(err)
	}
}
