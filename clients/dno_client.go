package clients

import (
	"context"
	"io"

	"github.com/k-lomer/lights-out/model"
)

type DnoClient interface {
	ListOutages(ctx context.Context) ([]model.Outage, error)
	GetDno() model.Dno
}

// drainAndClose reads any remaining bytes from a response body and closes it,
// so the underlying connection can be returned to the pool and reused.
// Both calls are best-effort: a failure leaves nothing actionable, so the errors
// are deliberately ignored.
func drainAndClose(body io.ReadCloser) {
	io.Copy(io.Discard, body) //nolint:errcheck
	body.Close()              //nolint:errcheck
}
