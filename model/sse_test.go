package model

import (
	_ "embed"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/sse.json
var sseFixture []byte

// Test that the real captured SSE payload decodes and converts cleanly.
func Test_Sse_RealData(t *testing.T) {
	var outages SseOutages
	require.NoError(t, json.Unmarshal(sseFixture, &outages))

	got := outages.ToOutages()

	require.NotEmpty(t, got)
	assert.Len(t, got, len(outages.Outages))
	assertConverted(t, got, DnoSse, true)
	// SSE always supplies both a logged and an estimated restoration time.
	for _, o := range got {
		assert.NotNil(t, o.Start)
		assert.NotNil(t, o.EstimatedEnd)
	}
}

// Test that an outage with an unparseable time is skipped, not the whole batch.
func Test_Sse_SkipsUndecodableOutage(t *testing.T) {
	var outages SseOutages
	require.NoError(t, json.Unmarshal([]byte(`{"Faults": [
		{"reference": "sse-good", "loggedAt": "2026-06-25T11:00:00.000+0000", "estimatedRestoration": "2026-06-26T00:00:00.000+0000", "affectedAreas": ["HP13 7DZ"]},
		{"reference": "sse-bad", "loggedAt": "not a time", "estimatedRestoration": "2026-06-26T00:00:00.000+0000", "affectedAreas": ["HP13 7DZ"]}
	]}`), &outages))

	require.Len(t, outages.Outages, 1)
	assert.Equal(t, "sse-good", outages.Outages[0].ID)
}

// Test that the timezone-offset layout is parsed for both times.
func Test_Sse_ParsesTimes(t *testing.T) {
	var o SseOutage
	require.NoError(t, json.Unmarshal([]byte(`{
		"reference": "sse-1",
		"loggedAt": "2026-06-25T11:00:00.000+0000",
		"estimatedRestoration": "2026-06-26T00:00:00.000+0100",
		"affectedAreas": ["HP13 7DZ", "HP13 7EA"]
	}`), &o))

	got := o.ToOutage()

	assert.Equal(t, "sse-1", got.ID)
	assertTimeEqual(t, time.Date(2026, 6, 25, 11, 0, 0, 0, time.UTC), got.Start)
	assertTimeEqual(t, time.Date(2026, 6, 25, 23, 0, 0, 0, time.UTC), got.EstimatedEnd)
	assert.Equal(t, Postcodes{"HP13 7DZ", "HP13 7EA"}, got.Postcodes)
}

// Test that a resolved fault maps the update time to the actual end while an unresolved one leaves it nil.
func Test_Sse_ActualEndFromResolved(t *testing.T) {
	var resolved SseOutage
	require.NoError(t, json.Unmarshal([]byte(`{
		"reference": "sse-done",
		"loggedAt": "2026-06-25T11:00:00.000+0000",
		"estimatedRestoration": "2026-06-26T00:00:00.000+0000",
		"updated": "2026-06-25T23:30:00.000+0000",
		"resolved": true,
		"affectedAreas": ["HP13 7DZ"]
	}`), &resolved))

	got := resolved.ToOutage()

	assertTimeEqual(t, time.Date(2026, 6, 26, 0, 0, 0, 0, time.UTC), got.EstimatedEnd)
	assertTimeEqual(t, time.Date(2026, 6, 25, 23, 30, 0, 0, time.UTC), got.ActualEnd)

	var ongoing SseOutage
	require.NoError(t, json.Unmarshal([]byte(`{
		"reference": "sse-ongoing",
		"loggedAt": "2026-06-25T11:00:00.000+0000",
		"estimatedRestoration": "2026-06-26T00:00:00.000+0000",
		"updated": "2026-06-25T23:30:00.000+0000",
		"resolved": false,
		"affectedAreas": ["HP13 7DZ"]
	}`), &ongoing))

	got = ongoing.ToOutage()

	assertTimeEqual(t, time.Date(2026, 6, 26, 0, 0, 0, 0, time.UTC), got.EstimatedEnd)
	assert.Nil(t, got.ActualEnd)
}

// Test that a resolved fault, a future start, or an ongoing outage maps to the canonical status.
func Test_Sse_Status(t *testing.T) {
	now := time.Now()
	future := &SseTime{Time: now.Add(24 * time.Hour)}
	past := &SseTime{Time: now.Add(-24 * time.Hour)}
	updated := &SseTime{Time: now}

	cases := []struct {
		name string
		o    SseOutage
		want Status
	}{
		{"resolved is resolved", SseOutage{Start: past, Updated: updated, Resolved: true}, StatusResolved},
		{"future start is future", SseOutage{Start: future}, StatusFuture},
		{"past start is active", SseOutage{Start: past}, StatusActive},
		{"missing start is active", SseOutage{}, StatusActive},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.o.status())
		})
	}
}

// Test that invalid postcodes in the affected areas are silently skipped.
func Test_Sse_SkipsInvalidPostcodes(t *testing.T) {
	var o SseOutage
	require.NoError(t, json.Unmarshal([]byte(`{
		"UUID": "sse-2",
		"loggedAt": "2026-06-25T11:00:00.000+0000",
		"estimatedRestoration": "2026-06-26T00:00:00.000+0000",
		"affectedAreas": ["HP13 7DZ", "NOT A POSTCODE", "HP13 7EA"]
	}`), &o))

	got := o.ToOutage()

	assert.Equal(t, Postcodes{"HP13 7DZ", "HP13 7EA"}, got.Postcodes)
}

// Test that an empty affected areas list yields no postcodes.
func Test_Sse_EmptyPostcodes(t *testing.T) {
	var o SseOutage
	require.NoError(t, json.Unmarshal([]byte(`{
		"UUID": "sse-3",
		"loggedAt": "2026-06-25T11:00:00.000+0000",
		"estimatedRestoration": "2026-06-26T00:00:00.000+0000",
		"affectedAreas": []
	}`), &o))

	got := o.ToOutage()

	assert.Empty(t, got.Postcodes)
}
