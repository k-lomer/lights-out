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

// Test that the fault type maps to the canonical status, with a timestamp fallback.
func Test_EnergyNorthWest_Status(t *testing.T) {
	cases := []struct {
		name      string
		faultType EnergyNorthWestOutageType
		actualEnd string
		want      Status
	}{
		{"current fault is active", energyNorthWestCurrentFault, "", StatusActive},
		{"resolved fault is resolved", energyNorthWestResolvedFault, "2026-06-25T12:00:00", StatusResolved},
		{"todays planned works is planned", energyNorthWestTodaysPlannedWorks, "", StatusActive},
		{"future planned works is planned", energyNorthWestFuturePlannedWorks, "", StatusPlanned},
		{"unknown type with restoration is resolved", EnergyNorthWestOutageType("CancelledPlanned"), "2026-06-25T12:00:00", StatusResolved},
		{"unknown type without restoration is active", EnergyNorthWestOutageType("CancelledPlanned"), "", StatusActive},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			o := EnergyNorthWestOutage{ID: "ENW-status", Type: tc.faultType}
			if tc.actualEnd != "" {
				o.ActualEnd = &EnergyNorthWestTime{Time: enwTime(t, tc.actualEnd)}
			}

			assert.Equal(t, tc.want, o.ToOutage().Status)
		})
	}
}

// Test that the estimated and actual restoration times are populated independently.
func Test_EnergyNorthWest_SplitsEndTimes(t *testing.T) {
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
	assertTimeEqual(t, enwTime(t, "2026-06-25T14:00:00"), got.EstimatedEnd)
	assertTimeEqual(t, enwTime(t, "2026-06-25T12:00:00"), got.ActualEnd)
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

	assertTimeEqual(t, enwTime(t, "2026-06-25T14:00:00"), got.EstimatedEnd)
	assert.Nil(t, got.ActualEnd)
}

// Test that a missing actual and estimated end leaves both end times nil.
func Test_EnergyNorthWest_NoEnd(t *testing.T) {
	var o EnergyNorthWestOutage
	require.NoError(t, json.Unmarshal([]byte(`{
		"faultNumber": "ENW-3",
		"date": "2026-06-25T10:00:00",
		"AffectedPostcodes": "AB1 2CD"
	}`), &o))

	got := o.ToOutage()

	assert.Nil(t, got.EstimatedEnd)
	assert.Nil(t, got.ActualEnd)
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
