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

// Test that the restored (millisecond layout) and estimated times populate the end fields independently.
func Test_UKPowerNetwork_SplitsEndTimes(t *testing.T) {
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
	assertTimeEqual(t, ukpnExpectedTime(t, ukpnTimeLayoutMs, "2026-06-25T18:18:56.06"), got.ActualEnd)
	assertTimeEqual(t, ukpnExpectedTime(t, ukpnTimeLayout, "2026-06-25T20:30:00"), got.EstimatedEnd)
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

	assertTimeEqual(t, ukpnExpectedTime(t, ukpnTimeLayout, "2026-06-25T20:30:00"), got.EstimatedEnd)
	assert.Nil(t, got.ActualEnd)
}

// Test that a missing restored and estimated time leaves both end times nil.
func Test_UKPowerNetwork_NoEnd(t *testing.T) {
	var o UKPowerNetworkOutage
	require.NoError(t, json.Unmarshal([]byte(`{
		"IncidentReference": "UKPN-3",
		"CreationDateTime": "2026-06-25T16:36:34",
		"FullPostcodeData": ["N166RJ"]
	}`), &o))

	got := o.ToOutage()

	assert.Nil(t, got.EstimatedEnd)
	assert.Nil(t, got.ActualEnd)
}

// Test that the start comes from the received or planned date for planned work and the creation date otherwise.
func Test_UKPowerNetwork_StartTime(t *testing.T) {
	cases := []struct {
		name string
		json string
		want string // Expected start in the UKPN layout; empty means nil.
	}{
		{
			"planned work in progress uses the received date, not the weeks-early creation date",
			`{"IncidentReference":"UKPN-1","CreationDateTime":"2026-06-11T12:40:26","ReceivedDate":"2026-07-05T09:01:00","PlannedDate":"2026-07-05T09:00:00","FullPostcodeData":["N166RJ"]}`,
			"2026-07-05T09:01:00",
		},
		{
			"planned work not yet begun has no received date and falls back to the planned date",
			`{"IncidentReference":"UKPN-2","CreationDateTime":"2026-06-11T12:40:26","PlannedDate":"2026-07-05T09:00:00","FullPostcodeData":["N166RJ"]}`,
			"2026-07-05T09:00:00",
		},
		{
			"unplanned work with a received date still uses the creation date",
			`{"IncidentReference":"UKPN-3","CreationDateTime":"2026-07-05T18:28:46","ReceivedDate":"2026-07-05T18:28:42","FullPostcodeData":["N166RJ"]}`,
			"2026-07-05T18:28:46",
		},
		{
			"no times leaves the start nil",
			`{"IncidentReference":"UKPN-4","FullPostcodeData":["N166RJ"]}`,
			"",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var o UKPowerNetworkOutage
			require.NoError(t, json.Unmarshal([]byte(tc.json), &o))

			got := o.ToOutage().Start
			if tc.want == "" {
				assert.Nil(t, got)
				return
			}
			assertTimeEqual(t, ukpnExpectedTime(t, ukpnTimeLayout, tc.want), got)
		})
	}
}

// Test that a restoration, a future start, or an ongoing outage maps to the canonical status.
func Test_UKPowerNetwork_Status(t *testing.T) {
	now := time.Now()
	future := ukpnTime{Time: now.Add(24 * time.Hour)}
	past := ukpnTime{Time: now.Add(-24 * time.Hour)}
	restored := ukpnTimeMs{Time: now}

	cases := []struct {
		name string
		o    UKPowerNetworkOutage
		want Status
	}{
		{"restored is resolved", UKPowerNetworkOutage{Creation: &past, Restored: &restored}, StatusResolved},
		{"restored planned work is resolved", UKPowerNetworkOutage{Planned: &past, Received: &past, Restored: &restored}, StatusResolved},
		{"planned work not yet begun is future", UKPowerNetworkOutage{Planned: &future}, StatusFuture},
		{"planned work in progress is active", UKPowerNetworkOutage{Planned: &past, Received: &past}, StatusActive},
		{"unplanned ongoing work is active", UKPowerNetworkOutage{Creation: &past}, StatusActive},
		{"no times is active", UKPowerNetworkOutage{}, StatusActive},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.o.status())
		})
	}
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
