package model

import (
	_ "embed"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/energy_north_west.json
var energyNorthWestFixture []byte

// enwTime parses a time in the Energy North West layout for building expectations.
func enwTime(t *testing.T, s string) time.Time {
	t.Helper()
	parsed, err := time.ParseInLocation(energyNorthWestTimeLayout, s, ukLocation)
	require.NoError(t, err)
	return parsed
}

// Test that the real captured Energy North West payload decodes and converts cleanly.
func Test_EnergyNorthWest_RealData(t *testing.T) {
	var outages EnergyNorthWestOutages
	require.NoError(t, json.Unmarshal(energyNorthWestFixture, &outages))

	got := outages.ToOutages()

	require.NotEmpty(t, got)
	assert.Len(t, got, len(outages.Outages))
	assertConverted(t, got, DnoEnergyNorthWest, true)
}

// Test that the actual restoration time is preferred over the estimated one.
func Test_EnergyNorthWest_PrefersActualEnd(t *testing.T) {
	var o EnergyNorthWestOutage
	require.NoError(t, json.Unmarshal([]byte(`{
		"faultNumber": "ENW-1",
		"date": "2026-06-25T10:00:00",
		"estimatedTimeOfRestoration": "2026-06-25T14:00:00",
		"actualTimeOfRestoration": "2026-06-25T12:00:00",
		"AffectedPostcodes": "AB1 2CD, EF3 4GH"
	}`), &o))

	got := o.ToOutage()

	assert.Equal(t, "ENW-1", got.ID)
	assertTimeEqual(t, enwTime(t, "2026-06-25T10:00:00"), got.Start)
	assertTimeEqual(t, enwTime(t, "2026-06-25T12:00:00"), got.End)
	assert.Equal(t, Postcodes{"AB1 2CD", "EF3 4GH"}, got.Postcodes)
}

// Test that the estimated end is used when no actual restoration time is present.
func Test_EnergyNorthWest_FallsBackToEstimatedEnd(t *testing.T) {
	var o EnergyNorthWestOutage
	require.NoError(t, json.Unmarshal([]byte(`{
		"faultNumber": "ENW-2",
		"date": "2026-06-25T10:00:00",
		"estimatedTimeOfRestoration": "2026-06-25T14:00:00",
		"AffectedPostcodes": "AB1 2CD"
	}`), &o))

	got := o.ToOutage()

	assertTimeEqual(t, enwTime(t, "2026-06-25T14:00:00"), got.End)
}

// Test that a missing actual and estimated end leaves End nil.
func Test_EnergyNorthWest_NoEnd(t *testing.T) {
	var o EnergyNorthWestOutage
	require.NoError(t, json.Unmarshal([]byte(`{
		"faultNumber": "ENW-3",
		"date": "2026-06-25T10:00:00",
		"AffectedPostcodes": "AB1 2CD"
	}`), &o))

	got := o.ToOutage()

	assert.Nil(t, got.End)
}

// Test that an empty affected postcodes string yields no postcodes.
func Test_EnergyNorthWest_EmptyPostcodes(t *testing.T) {
	var o EnergyNorthWestOutage
	require.NoError(t, json.Unmarshal([]byte(`{
		"faultNumber": "ENW-4",
		"date": "2026-06-25T10:00:00",
		"AffectedPostcodes": ""
	}`), &o))

	got := o.ToOutage()

	assert.Empty(t, got.Postcodes)
}
