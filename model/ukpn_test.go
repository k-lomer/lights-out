package model

import (
	_ "embed"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/ukpn.json
var ukpnFixture []byte

// ukpnExpectedTime parses a time in the no-millisecond UKPN layout for building expectations.
func ukpnExpectedTime(t *testing.T, layout, s string) time.Time {
	t.Helper()
	parsed, err := time.ParseInLocation(layout, s, ukLocation)
	require.NoError(t, err)
	return parsed
}

// Test that the real captured UKPN payload decodes and converts cleanly.
func Test_UKPowerNetwork_RealData(t *testing.T) {
	var outages UKPowerNetworkOutages
	require.NoError(t, json.Unmarshal(ukpnFixture, &outages))

	got := UKPowerNetworkToOutages(outages)

	require.NotEmpty(t, got)
	assert.Len(t, got, len(outages))
	assertConverted(t, got, DnoUKPowerNetwork, true)
}

// Test that a single outage with an unparseable time is skipped, not the whole batch.
func Test_UKPowerNetwork_SkipsUndecodableOutage(t *testing.T) {
	var outages UKPowerNetworkOutages
	require.NoError(t, json.Unmarshal([]byte(`[
		{"IncidentReference": "UKPN-good", "CreationDateTime": "2026-06-25T16:36:34", "FullPostcodeData": ["N166RJ"]},
		{"IncidentReference": "UKPN-bad", "CreationDateTime": "not a time", "FullPostcodeData": ["N166RJ"]}
	]`), &outages))

	require.Len(t, outages, 1)
	assert.Equal(t, "UKPN-good", outages[0].ID)
}

// Test that the restored time (millisecond layout) is preferred for the end time.
func Test_UKPowerNetwork_PrefersRestored(t *testing.T) {
	var o UKPowerNetworkOutage
	require.NoError(t, json.Unmarshal([]byte(`{
		"IncidentReference": "UKPN-1",
		"CreationDateTime": "2026-06-25T16:36:34",
		"RestoredDateTime": "2026-06-25T18:18:56.06",
		"EstimatedRestorationDate": "2026-06-25T20:30:00",
		"FullPostcodeData": ["N166RJ", "E59AR"]
	}`), &o))

	got := o.ToOutage()

	assert.Equal(t, "UKPN-1", got.ID)
	assertTimeEqual(t, ukpnExpectedTime(t, ukpnTimeLayout, "2026-06-25T16:36:34"), got.Start)
	assertTimeEqual(t, ukpnExpectedTime(t, ukpnTimeLayoutMs, "2026-06-25T18:18:56.06"), got.End)
	// Postcodes arrive without spaces and are normalised during conversion.
	assert.Equal(t, Postcodes{"N16 6RJ", "E5 9AR"}, got.Postcodes)
}

// Test that the estimated time is used as the end when no restored time is present.
func Test_UKPowerNetwork_FallsBackToEstimated(t *testing.T) {
	var o UKPowerNetworkOutage
	require.NoError(t, json.Unmarshal([]byte(`{
		"IncidentReference": "UKPN-2",
		"CreationDateTime": "2026-06-25T16:36:34",
		"EstimatedRestorationDate": "2026-06-25T20:30:00",
		"FullPostcodeData": ["N166RJ"]
	}`), &o))

	got := o.ToOutage()

	assertTimeEqual(t, ukpnExpectedTime(t, ukpnTimeLayout, "2026-06-25T20:30:00"), got.End)
}

// Test that a missing restored and estimated time leaves End nil.
func Test_UKPowerNetwork_NoEnd(t *testing.T) {
	var o UKPowerNetworkOutage
	require.NoError(t, json.Unmarshal([]byte(`{
		"IncidentReference": "UKPN-3",
		"CreationDateTime": "2026-06-25T16:36:34",
		"FullPostcodeData": ["N166RJ"]
	}`), &o))

	got := o.ToOutage()

	assert.Nil(t, got.End)
}

// Test that invalid postcodes in the full postcode data are silently skipped.
func Test_UKPowerNetwork_SkipsInvalidPostcodes(t *testing.T) {
	var o UKPowerNetworkOutage
	require.NoError(t, json.Unmarshal([]byte(`{
		"IncidentReference": "UKPN-4",
		"CreationDateTime": "2026-06-25T16:36:34",
		"FullPostcodeData": ["N166RJ", "BAD", "E59AR"]
	}`), &o))

	got := o.ToOutage()

	assert.Equal(t, Postcodes{"N16 6RJ", "E5 9AR"}, got.Postcodes)
}
