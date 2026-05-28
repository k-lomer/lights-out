package clients

import (
	"context"
	"io"

	"github.com/k-lomer/lights-out/model"
)

type DnoClient interface {
	ListOutages(ctx context.Context) ([]model.Outage, error)
}

func drainAndClose(body io.ReadCloser) {
	io.Copy(io.Discard, body) //nolint:errcheck
	body.Close()              //nolint:errcheck
}
