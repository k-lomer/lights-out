package model

import (
	_ "embed"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/northern_powergrid.json
var northernPowergridFixture []byte

// Test that the real captured Northern Powergrid payload decodes and converts cleanly.
func Test_NorthernPowergrid_RealData(t *testing.T) {
	var outages NorthernPowergridOutages
	require.NoError(t, json.Unmarshal(northernPowergridFixture, &outages))

	got := NorthernPowergridToOutages(outages)

	require.NotEmpty(t, got)
	// Aggregation merges rows sharing a key, so the result is no larger than the input.
	assert.LessOrEqual(t, len(got), len(outages))
	assertConverted(t, got, DnoNorthernPowergrid, true)
}

// Test that the real duplicate-reference rows merge into one outage with every postcode.
func Test_NorthernPowergrid_RealDataMerge(t *testing.T) {
	var outages NorthernPowergridOutages
	require.NoError(t, json.Unmarshal(northernPowergridFixture, &outages))

	got := NorthernPowergridToOutages(outages)

	var merged *Outage
	for i := range got {
		if got[i].ID == "INCD-790173-A" {
			merged = &got[i]
		}
	}
	require.NotNil(t, merged)
	assert.ElementsMatch(t, Postcodes{"HG4 3HJ", "HG4 3HZ", "HG4 3JU", "HG4 3HF", "HG4 3HN"}, merged.Postcodes)
}

// Test that an outage with an unparseable time is skipped, not the whole batch.
func Test_NorthernPowergrid_SkipsUndecodableOutage(t *testing.T) {
	var outages NorthernPowergridOutages
	require.NoError(t, json.Unmarshal([]byte(`[
		{"Reference": "NPG-good", "LoggedTime": "2026-06-25T13:00:00Z", "EstimatedTimeTillResolution": "2026-06-25T18:00:00Z", "Postcode": "NE34 0JA"},
		{"Reference": "NPG-bad", "LoggedTime": "2026-06-25T13:00:00Z", "EstimatedTimeTillResolution": "not a time", "Postcode": "NE34 0HX"}
	]`), &outages))

	require.Len(t, outages, 1)
	assert.Equal(t, "NPG-good", outages[0].ID)
}

// Test that a single outage wraps its lone postcode and parses both times.
func Test_NorthernPowergrid_SingleOutage(t *testing.T) {
	var o NorthernPowergridOutage
	require.NoError(t, json.Unmarshal([]byte(`{
		"Reference": "NPG-1",
		"LoggedTime": "2026-06-25T13:00:00Z",
		"EstimatedTimeTillResolution": "2026-06-25T18:00:00Z",
		"Postcode": "NE34 0JA"
	}`), &o))

	got := o.ToOutage()

	assert.Equal(t, "NPG-1", got.ID)
	assertTimeEqual(t, time.Date(2026, 6, 25, 13, 0, 0, 0, time.UTC), got.Start)
	assertTimeEqual(t, time.Date(2026, 6, 25, 18, 0, 0, 0, time.UTC), got.EstimatedEnd)
	assert.Equal(t, Postcodes{"NE34 0JA"}, got.Postcodes)
}

// Test that the 1900-01-01 sentinel resolution time maps End to nil.
func Test_NorthernPowergrid_SentinelEnd(t *testing.T) {
	var o NorthernPowergridOutage
	require.NoError(t, json.Unmarshal([]byte(`{
		"Reference": "NPG-2",
		"LoggedTime": "2026-06-25T13:00:00Z",
		"EstimatedTimeTillResolution": "1900-01-01T00:00:00",
		"Postcode": "NE34 0JA"
	}`), &o))

	got := o.ToOutage()

	assert.Nil(t, got.EstimatedEnd)
}

// Test that a completed stage message maps the update time to the actual end while a non-completed row leaves it nil.
func Test_NorthernPowergrid_ActualEndFromCompletedMessage(t *testing.T) {
	var completed NorthernPowergridOutage
	require.NoError(t, json.Unmarshal([]byte(`{
		"Reference": "NPG-done",
		"LoggedTime": "2026-06-25T13:00:00Z",
		"EstimatedTimeTillResolution": "2026-06-25T18:00:00Z",
		"UpdateDate": "2026-06-25T17:30:00Z",
		"CustomerStageSequenceMessage": "The scheduled work has now been completed",
		"Postcode": "NE34 0JA"
	}`), &completed))

	got := completed.ToOutage()

	assertTimeEqual(t, time.Date(2026, 6, 25, 18, 0, 0, 0, time.UTC), got.EstimatedEnd)
	assertTimeEqual(t, time.Date(2026, 6, 25, 17, 30, 0, 0, time.UTC), got.ActualEnd)

	var ongoing NorthernPowergridOutage
	require.NoError(t, json.Unmarshal([]byte(`{
		"Reference": "NPG-ongoing",
		"LoggedTime": "2026-06-25T13:00:00Z",
		"EstimatedTimeTillResolution": "2026-06-25T18:00:00Z",
		"UpdateDate": "2026-06-25T17:30:00Z",
		"CustomerStageSequenceMessage": "Our team has arrived in the area affected by the power cut.",
		"Postcode": "NE34 0JA"
	}`), &ongoing))

	got = ongoing.ToOutage()

	assertTimeEqual(t, time.Date(2026, 6, 25, 18, 0, 0, 0, time.UTC), got.EstimatedEnd)
	assert.Nil(t, got.ActualEnd)
}

// Test that a completion message, a future start, or an ongoing outage maps to the canonical status.
func Test_NorthernPowergrid_Status(t *testing.T) {
	now := time.Now()
	future := now.Add(24 * time.Hour)
	past := now.Add(-24 * time.Hour)

	cases := []struct {
		name string
		o    NorthernPowergridOutage
		want Status
	}{
		{"completed message is resolved", NorthernPowergridOutage{Start: &past, Updated: &now, Message: northernPowergridCompletedMessage}, StatusResolved},
		{"future start is future", NorthernPowergridOutage{Start: &future}, StatusFuture},
		{"past start is active", NorthernPowergridOutage{Start: &past}, StatusActive},
		{"missing start is active", NorthernPowergridOutage{}, StatusActive},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.o.status())
		})
	}
}

// Test that rows sharing reference, start and end merge into one outage with all postcodes.
func Test_NorthernPowergrid_MergesDuplicates(t *testing.T) {
	var outages NorthernPowergridOutages
	require.NoError(t, json.Unmarshal([]byte(`[
		{"Reference": "NPG-3", "LoggedTime": "2026-06-25T13:00:00Z", "EstimatedTimeTillResolution": "2026-06-25T18:00:00Z", "Postcode": "NE34 0JA"},
		{"Reference": "NPG-3", "LoggedTime": "2026-06-25T13:00:00Z", "EstimatedTimeTillResolution": "2026-06-25T18:00:00Z", "Postcode": "NE34 0HX"}
	]`), &outages))

	got := NorthernPowergridToOutages(outages)

	require.Len(t, got, 1)
	assert.Equal(t, "NPG-3", got[0].ID)
	assert.Equal(t, Postcodes{"NE34 0JA", "NE34 0HX"}, got[0].Postcodes)
}

// Test that rows differing only in resolution time are kept as separate outages.
func Test_NorthernPowergrid_DistinctKeysNotMerged(t *testing.T) {
	var outages NorthernPowergridOutages
	require.NoError(t, json.Unmarshal([]byte(`[
		{"Reference": "NPG-4", "LoggedTime": "2026-06-25T13:00:00Z", "EstimatedTimeTillResolution": "2026-06-25T18:00:00Z", "Postcode": "NE34 0JA"},
		{"Reference": "NPG-4", "LoggedTime": "2026-06-25T13:00:00Z", "EstimatedTimeTillResolution": "2026-06-25T19:00:00Z", "Postcode": "NE34 0HX"}
	]`), &outages))

	got := NorthernPowergridToOutages(outages)

	assert.Len(t, got, 2)
}
