package model

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test that defaults are applied when no query params are supplied.
func Test_ParseQueryParams_Defaults(t *testing.T) {
	qp, err := ParseQueryParams(url.Values{})
	require.NoError(t, err)

	assert.Equal(t, uint(10), qp.PageSize)
	assert.Equal(t, uint(0), qp.PageIndex)
	assert.Equal(t, Postcodes{}, qp.Postcodes)
	assert.Equal(t, AllDnoList, qp.Dnos)
	assert.Equal(t, []Status{StatusActive}, qp.Status)
}

// Test parsing of the pageSize param and its error branch.
func Test_ParseQueryParams_PageSize(t *testing.T) {
	testCases := []struct {
		name      string
		value     string
		expect    uint
		expectErr string
	}{
		{name: "explicit value", value: "25", expect: 25},
		{name: "zero", value: "0", expect: 0},
		{name: "non-numeric is rejected", value: "lots", expectErr: "failed to parse pageSize 'lots'"},
		{name: "negative is rejected", value: "-1", expectErr: "failed to parse pageSize '-1'"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			qp, err := ParseQueryParams(url.Values{"pageSize": {tc.value}})
			if tc.expectErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expect, qp.PageSize)
		})
	}
}

// Test parsing of the pageIndex param and its error branch.
func Test_ParseQueryParams_PageIndex(t *testing.T) {
	testCases := []struct {
		name      string
		value     string
		expect    uint
		expectErr string
	}{
		{name: "explicit value", value: "3", expect: 3},
		{name: "zero", value: "0", expect: 0},
		{name: "non-numeric is rejected", value: "second", expectErr: "failed to parse pageIndex 'second'"},
		{name: "negative is rejected", value: "-2", expectErr: "failed to parse pageIndex '-2'"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			qp, err := ParseQueryParams(url.Values{"pageIndex": {tc.value}})
			if tc.expectErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expect, qp.PageIndex)
		})
	}
}

// Test that postcodes are parsed, normalised, and invalid ones are rejected.
func Test_ParseQueryParams_Postcodes(t *testing.T) {
	testCases := []struct {
		name      string
		value     string
		expect    Postcodes
		expectErr string
	}{
		{name: "single postcode is normalised", value: "sw1a1aa", expect: Postcodes{"SW1A 1AA"}},
		{name: "multiple comma-separated postcodes", value: "SW1A 1AA,EC1A 1BB", expect: Postcodes{"SW1A 1AA", "EC1A 1BB"}},
		{name: "invalid postcode is rejected", value: "not-a-postcode", expectErr: "failed to parse postcode"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			qp, err := ParseQueryParams(url.Values{"postcodes": {tc.value}})
			if tc.expectErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expect, qp.Postcodes)
		})
	}
}

// Test that ParseQueryParams handles DNO targeting and its error branches.
func Test_ParseQueryParams_DnoTargeting(t *testing.T) {
	allDnosFalse := url.Values{}
	for _, dno := range AllDnoList {
		allDnosFalse.Set(string(dno), "false")
	}

	testCases := []struct {
		name       string
		values     url.Values
		expectErr  string
		expectDnos []Dno
	}{
		{
			name:       "default targets all DNOs",
			values:     url.Values{},
			expectDnos: AllDnoList,
		},
		{
			name:       "explicit true targets the DNO",
			values:     url.Values{string(DnoEnergyNorthWest): {"true"}},
			expectDnos: AllDnoList,
		},
		{
			name:       "value is case-insensitive",
			values:     url.Values{string(DnoEnergyNorthWest): {"FALSE"}},
			expectDnos: AllDnoList[1:],
		},
		{
			name:       "disabling one DNO excludes only it",
			values:     url.Values{string(DnoEnergyNorthWest): {"false"}},
			expectDnos: AllDnoList[1:],
		},
		{
			name:      "disabling all DNOs is rejected",
			values:    allDnosFalse,
			expectErr: "no DNOs targeted",
		},
		{
			name:      "non-boolean DNO value is rejected",
			values:    url.Values{string(DnoEnergyNorthWest): {"maybe"}},
			expectErr: "unexpected non-boolean value for DNO EnergyNorthWest: maybe",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			qp, err := ParseQueryParams(tc.values)
			if tc.expectErr != "" {
				require.Error(t, err)
				assert.EqualError(t, err, tc.expectErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectDnos, qp.Dnos)
		})
	}
}

// Test that ParseQueryParams handles status targeting and its error branches.
func Test_ParseQueryParams_StatusTargeting(t *testing.T) {
	testCases := []struct {
		name         string
		values       url.Values
		expectErr    string
		expectStatus []Status
	}{
		{
			name:         "default targets only Active",
			values:       url.Values{},
			expectStatus: []Status{StatusActive},
		},
		{
			name: "all statuses can be targeted",
			values: url.Values{
				string(StatusFuture):   {"true"},
				string(StatusResolved): {"true"},
			},
			expectStatus: []Status{StatusActive, StatusFuture, StatusResolved},
		},
		{
			name:         "Future can be targeted alongside Active",
			values:       url.Values{string(StatusFuture): {"true"}},
			expectStatus: []Status{StatusActive, StatusFuture},
		},
		{
			name:         "value is case-insensitive",
			values:       url.Values{string(StatusResolved): {"TRUE"}},
			expectStatus: []Status{StatusActive, StatusResolved},
		},
		{
			name: "disabling all statuses is rejected",
			values: url.Values{
				string(StatusActive):   {"false"},
				string(StatusFuture):   {"false"},
				string(StatusResolved): {"false"},
			},
			expectErr: "no Status targeted",
		},
		{
			name:      "non-boolean Active value is rejected",
			values:    url.Values{string(StatusActive): {"maybe"}},
			expectErr: "unexpected non-boolean value for status Active: maybe",
		},
		{
			name:      "non-boolean Future value is rejected",
			values:    url.Values{string(StatusFuture): {"maybe"}},
			expectErr: "unexpected non-boolean value for status Future: maybe",
		},
		{
			name:      "non-boolean Resolved value is rejected",
			values:    url.Values{string(StatusResolved): {"maybe"}},
			expectErr: "unexpected non-boolean value for status Resolved: maybe",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			qp, err := ParseQueryParams(tc.values)
			if tc.expectErr != "" {
				require.Error(t, err)
				assert.EqualError(t, err, tc.expectErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectStatus, qp.Status)
		})
	}
}
