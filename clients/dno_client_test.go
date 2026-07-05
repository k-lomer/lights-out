package clients

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/k-lomer/lights-out/cache"
	"github.com/k-lomer/lights-out/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeDnoClient struct {
	*UpdateTracker
	dno   model.Dno
	calls int
}

func newFakeDnoClient(dno model.Dno) *fakeDnoClient {
	return &fakeDnoClient{
		UpdateTracker: &UpdateTracker{},
		dno:           dno,
	}
}

func (f *fakeDnoClient) GetDno() model.Dno {
	return f.dno
}

func (f *fakeDnoClient) ListOutages(_ context.Context) ([]model.Outage, error) {
	f.calls += 1
	return []model.Outage{{DNO: f.dno, ID: strconv.Itoa(f.calls)}}, nil
}

// Test that ListOutages uses the cache.
func Test_ListOutages_Caching(t *testing.T) {
	client := newFakeDnoClient(model.DnoUKPowerNetwork)
	outageCache := cache.MakeOutageCache(time.Minute)

	outages1, err := ListOutages(context.Background(), client, outageCache)
	require.NoError(t, err)
	assert.NotEmpty(t, outages1)

	outages2, err := ListOutages(context.Background(), client, outageCache)
	require.NoError(t, err)

	assert.Equal(t, outages1, outages2)
	assert.Equal(t, 1, client.calls)
}

// Test that ListOutages with a nil cache fetches from the client on every call.
func Test_ListOutages_NoCaching(t *testing.T) {
	client := newFakeDnoClient(model.DnoUKPowerNetwork)

	outages1, err := ListOutages(context.Background(), client, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, outages1)

	outages2, err := ListOutages(context.Background(), client, nil)
	require.NoError(t, err)

	assert.NotEqual(t, outages1, outages2)
	assert.Equal(t, 2, client.calls)
}

// Test that ListOutages re-fetches from the client once the cached value has expired.
func Test_ListOutages_CacheExpired(t *testing.T) {
	client := newFakeDnoClient(model.DnoUKPowerNetwork)
	outageCache := cache.MakeOutageCache(5 * time.Millisecond)

	outages1, err := ListOutages(context.Background(), client, outageCache)
	require.NoError(t, err)

	assert.EventuallyWithT(t, func(ct *assert.CollectT) {
		outages2, err := ListOutages(context.Background(), client, outageCache)
		require.NoError(ct, err)
		assert.NotEqual(ct, outages1, outages2)
	}, 100*time.Millisecond, 1*time.Millisecond)
}
