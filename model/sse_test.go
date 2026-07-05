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
		{"UUID": "sse-good", "loggedAt": "2026-06-25T11:00:00.000+0000", "estimatedRestoration": "2026-06-26T00:00:00.000+0000", "affectedAreas": ["HP13 7DZ"]},
		{"UUID": "sse-bad", "loggedAt": "not a time", "estimatedRestoration": "2026-06-26T00:00:00.000+0000", "affectedAreas": ["HP13 7DZ"]}
	]}`), &outages))

	require.Len(t, outages.Outages, 1)
	assert.Equal(t, "sse-good", outages.Outages[0].ID)
}

// Test that the timezone-offset layout is parsed for both times.
func Test_Sse_ParsesTimes(t *testing.T) {
	var o SseOutage
	require.NoError(t, json.Unmarshal([]byte(`{
		"UUID": "sse-1",
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
