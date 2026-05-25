package clients

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func Test_ListSseOutages(t *testing.T) {
	ctx := context.Background()
	var client *http.Client = &http.Client{
		Timeout: 30 * time.Second,
	}
	res, err := ListSseOutages(ctx, client)
	if err != nil {
		t.Error(err)
	}
	if len(res) == 0 {
		t.Error("didn't get any outages")
	}
}
