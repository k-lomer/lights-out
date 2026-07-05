package model

import (
	_ "embed"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/national_grid_distribution.json
var nationalGridFixture []byte

// nationalGridTime parses a time in the National Grid layout for building expectations.
func nationalGridTime(t *testing.T, s string) time.Time {
	t.Helper()
	parsed, err := time.ParseInLocation(nationalGridTimeLayout, s, ukLocation)
	require.NoError(t, err)
	return parsed
}

// Test that the real captured National Grid payload decodes and converts cleanly.
func Test_NationalGrid_RealData(t *testing.T) {
	var outages NationalGridOutages
	require.NoError(t, json.Unmarshal(nationalGridFixture, &outages))

	got := outages.ToOutages()

	require.NotEmpty(t, got)
	assert.Len(t, got, len(outages.Outages))
	assertConverted(t, got, DnoNationalGridDistribution, false)
}

// Test that a future start, a past start, or a missing start maps to the canonical status.
func Test_NationalGrid_Status(t *testing.T) {
	now := time.Now()
	future := now.Add(24 * time.Hour)
	past := now.Add(-24 * time.Hour)

	cases := []struct {
		name string
		o    NationalGridOutage
		want Status
	}{
		{"future start is future", NationalGridOutage{Start: OptionalNationalGridTime{Time: &future}}, StatusFuture},
		{"past start is active", NationalGridOutage{Start: OptionalNationalGridTime{Time: &past}}, StatusActive},
		{"missing start is active", NationalGridOutage{}, StatusActive},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.o.status())
		})
	}
}

// Test that real start and end times are parsed from the space-separated layout.
func Test_NationalGrid_ParsesTimes(t *testing.T) {
	var o NationalGridOutage
	require.NoError(t, json.Unmarshal([]byte(`{
		"id": "NG-1",
		"startTime": "2026-06-25 08:00:00",
		"etr": "2026-06-25 16:00:00",
		"postcodes": ["B24 9FF"]
	}`), &o))

	got := o.ToOutage()

	assert.Equal(t, "NG-1", got.ID)
	assertTimeEqual(t, nationalGridTime(t, "2026-06-25 08:00:00"), got.Start)
	assertTimeEqual(t, nationalGridTime(t, "2026-06-25 16:00:00"), got.EstimatedEnd)
	assert.Equal(t, Postcodes{"B24 9FF"}, got.Postcodes)
}

// Test that the 1900-01-01 sentinel in the etr field maps End to nil.
func Test_NationalGrid_SentinelEnd(t *testing.T) {
	var o NationalGridOutage
	require.NoError(t, json.Unmarshal([]byte(`{
		"id": "NG-2",
		"startTime": "2026-06-25 08:00:00",
		"etr": "1900-01-01 00:00:00",
		"postcodes": ["B24 9FF"]
	}`), &o))

	got := o.ToOutage()

	assertTimeEqual(t, nationalGridTime(t, "2026-06-25 08:00:00"), got.Start)
	assert.Nil(t, got.EstimatedEnd)
}

// Test that the 1900-01-01 sentinel in either time field maps both to nil.
func Test_NationalGrid_BothSentinel(t *testing.T) {
	var o NationalGridOutage
	require.NoError(t, json.Unmarshal([]byte(`{
		"id": "NG-3",
		"startTime": "1900-01-01 00:00:00",
		"etr": "1900-01-01 00:00:00",
		"postcodes": ["B24 9FF"]
	}`), &o))

	got := o.ToOutage()

	assert.Nil(t, got.Start)
	assert.Nil(t, got.EstimatedEnd)
}
