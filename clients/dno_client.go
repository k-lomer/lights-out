package clients

import (
	"context"
	"io"
	"time"

	"github.com/k-lomer/lights-out/cache"
	"github.com/k-lomer/lights-out/model"
)

// drainAndClose reads any remaining bytes from a response body and closes it,
// so the underlying connection can be returned to the pool and reused.
// A failure leaves nothing actionable, so errors are ignored.
func drainAndClose(body io.ReadCloser) {
	io.Copy(io.Discard, body) //nolint:errcheck
	body.Close()              //nolint:errcheck
}

type DnoClient interface {
	ListOutages(ctx context.Context) ([]model.Outage, error)
	GetDno() model.Dno
	LastUpdate() *time.Time
	SetUpdated()
	UpdateLock()
	UpdateUnlock()
}

func ListOutages(ctx context.Context, client DnoClient, outageCache *cache.OutageCache) ([]model.Outage, error) {
	dno := string(client.GetDno())

	if outages, err := outageCache.Get(dno); err == nil {
		return outages, nil
	}

	client.UpdateLock()
	defer client.UpdateUnlock()

	// Another call may have refreshed the cache while we waited on the
	// lock, so re-check cache before fetching from the client.
	lastUpdate := client.LastUpdate()
	if lastUpdate != nil && time.Since(*lastUpdate) < outageCache.GetTtl() {
		if outages, err := outageCache.Get(dno); err == nil {
			return outages, nil
		}
	}

	outages, err := client.ListOutages(ctx)
	if err != nil {
		return nil, err
	}

	client.SetUpdated()
	outageCache.Set(dno, outages)

	return outages, nil
}
