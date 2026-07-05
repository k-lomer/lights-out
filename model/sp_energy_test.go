package model

import (
	_ "embed"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/sp_energy.json
var spEnergyFixture []byte

// spTime parses a time in the given SP Energy layout for building expectations.
func spTime(t *testing.T, layout, s string) time.Time {
	t.Helper()
	parsed, err := time.ParseInLocation(layout, s, ukLocation)
	require.NoError(t, err)
	return parsed
}

// Test that the real captured SP Energy payload decodes and converts cleanly.
func Test_SPEnergy_RealData(t *testing.T) {
	var outages SPEnergyOutages
	require.NoError(t, json.Unmarshal(spEnergyFixture, &outages))

	got := outages.ToOutages()

	require.NotEmpty(t, got)
	assert.Len(t, got, len(outages.Outages))
	assertConverted(t, got, DnoSPEnergy, true)
	// SP Energy supplies an estimated fix, an actual restoration, or both, so every outage has an end time.
	for _, o := range got {
		assert.True(t, o.EstimatedEnd != nil || o.ActualEnd != nil)
	}
}

// Test that a future creation date, an actual restoration, or neither maps to the canonical status.
func Test_SPEnergy_Status(t *testing.T) {
	now := time.Now()
	future := SPEnergyStartTime{Time: now.Add(24 * time.Hour)}
	past := SPEnergyStartTime{Time: now.Add(-24 * time.Hour)}
	restored := SPEnergyEndTime{Time: now}

	cases := []struct {
		name string
		o    SPEnergyOutage
		want Status
	}{
		{"future creation date is future", SPEnergyOutage{Start: &future}, StatusFuture},
		{"actual restoration is resolved", SPEnergyOutage{Start: &past, ActualEnd: &restored}, StatusResolved},
		{"past creation with no restoration is active", SPEnergyOutage{Start: &past}, StatusActive},
		{"missing start with no restoration is active", SPEnergyOutage{}, StatusActive},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.o.status())
		})
	}
}

// Test the ISO start layout and the 12-hour AM/PM end layout (the live formats).
func Test_SPEnergy_PrimaryLayouts(t *testing.T) {
	var o SPEnergyOutage
	require.NoError(t, json.Unmarshal([]byte(`{
		"incidentReference": "SP-1",
		"createdDate": "2026-06-25 20:44:45",
		"estimatedFix": "6/26/2026, 4:00 AM",
		"actualRestorationTime": "6/25/2026, 8:53 PM",
		"postcodeList": "WA16 9LP, CW10 9LN"
	}`), &o))

	got := o.ToOutage()

	assert.Equal(t, "SP-1", got.ID)
	assertTimeEqual(t, spTime(t, sPEnergyStartTimeLayout1, "2026-06-25 20:44:45"), got.Start)
	assertTimeEqual(t, spTime(t, sPEnergyEndTimeLayout1, "6/25/2026, 8:53 PM"), got.ActualEnd)
	assert.Equal(t, Postcodes{"WA16 9LP", "CW10 9LN"}, got.Postcodes)
}

// Test the fallback start layout (D/M/YYYY, 24-hour).
func Test_SPEnergy_StartFallbackLayout(t *testing.T) {
	var o SPEnergyOutage
	require.NoError(t, json.Unmarshal([]byte(`{
		"incidentReference": "SP-2",
		"createdDate": "25/6/2026, 14:30",
		"estimatedFix": "6/26/2026, 4:00 AM",
		"postcodeList": "WA16 9LP"
	}`), &o))

	got := o.ToOutage()

	assertTimeEqual(t, spTime(t, sPEnergyStartTimeLayout2, "25/6/2026, 14:30"), got.Start)
}

// Test the fallback end layout (D/M/YYYY, 24-hour).
func Test_SPEnergy_EndFallbackLayout(t *testing.T) {
	var o SPEnergyOutage
	require.NoError(t, json.Unmarshal([]byte(`{
		"incidentReference": "SP-3",
		"createdDate": "2026-06-25 20:44:45",
		"estimatedFix": "25/6/2026, 22:30",
		"postcodeList": "WA16 9LP"
	}`), &o))

	got := o.ToOutage()

	assertTimeEqual(t, spTime(t, sPEnergyEndTimeLayout2, "25/6/2026, 22:30"), got.EstimatedEnd)
}

// Test that the estimated fix and actual restoration times are populated independently.
func Test_SPEnergy_SplitsEndTimes(t *testing.T) {
	var o SPEnergyOutage
	require.NoError(t, json.Unmarshal([]byte(`{
		"incidentReference": "SP-4",
		"createdDate": "2026-06-25 20:44:45",
		"estimatedFix": "6/26/2026, 4:00 AM",
		"actualRestorationTime": "6/25/2026, 8:53 PM",
		"postcodeList": "WA16 9LP"
	}`), &o))

	got := o.ToOutage()

	assertTimeEqual(t, spTime(t, sPEnergyEndTimeLayout1, "6/26/2026, 4:00 AM"), got.EstimatedEnd)
	assertTimeEqual(t, spTime(t, sPEnergyEndTimeLayout1, "6/25/2026, 8:53 PM"), got.ActualEnd)
}

// Test that the estimated fix is used when no actual restoration time is present.
func Test_SPEnergy_FallsBackToEstimatedEnd(t *testing.T) {
	var o SPEnergyOutage
	require.NoError(t, json.Unmarshal([]byte(`{
		"incidentReference": "SP-5",
		"createdDate": "2026-06-25 20:44:45",
		"estimatedFix": "6/26/2026, 4:00 AM",
		"postcodeList": "WA16 9LP"
	}`), &o))

	got := o.ToOutage()

	assertTimeEqual(t, spTime(t, sPEnergyEndTimeLayout1, "6/26/2026, 4:00 AM"), got.EstimatedEnd)
}
