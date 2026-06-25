package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func outage(id string, postcodes ...Postcode) Outage {
	return Outage{ID: id, Postcodes: postcodes}
}

// Test that an empty postcode list returns the outages unchanged.
func Test_FilterByPostcodes_NoPostcodes(t *testing.T) {
	outages := []Outage{
		outage("a", "AB1 2CD"),
		outage("b", "EF3 4GH"),
	}

	got := FilterByPostcodes(outages, Postcodes{})

	assert.Equal(t, outages, got)
}

// Test that only outages affecting a targeted postcode are returned.
func Test_FilterByPostcodes_Matching(t *testing.T) {
	outages := []Outage{
		outage("a", "AB1 2CD"),
		outage("b", "EF3 4GH"),
		outage("c", "IJ5 6KL"),
	}

	got := FilterByPostcodes(outages, Postcodes{"EF3 4GH"})

	assert.Equal(t, []Outage{outage("b", "EF3 4GH")}, got)
}

// Test that an outage covering multiple postcodes is returned once when any match.
func Test_FilterByPostcodes_MultiplePostcodesPerOutage(t *testing.T) {
	outages := []Outage{
		outage("a", "AB1 2CD", "EF3 4GH"),
		outage("b", "IJ5 6KL"),
	}

	got := FilterByPostcodes(outages, Postcodes{"EF3 4GH"})

	assert.Equal(t, []Outage{outage("a", "AB1 2CD", "EF3 4GH")}, got)
}

// Test that all matching outages are returned across multiple targeted postcodes.
func Test_FilterByPostcodes_MultipleMatches(t *testing.T) {
	outages := []Outage{
		outage("a", "AB1 2CD"),
		outage("b", "EF3 4GH"),
		outage("c", "IJ5 6KL"),
	}

	got := FilterByPostcodes(outages, Postcodes{"AB1 2CD", "IJ5 6KL"})

	assert.Equal(t, []Outage{outage("a", "AB1 2CD"), outage("c", "IJ5 6KL")}, got)
}

// Test that no matches returns an empty result.
func Test_FilterByPostcodes_NoMatches(t *testing.T) {
	outages := []Outage{
		outage("a", "AB1 2CD"),
		outage("b", "EF3 4GH"),
	}

	got := FilterByPostcodes(outages, Postcodes{"IJ5 6KL"})

	assert.Empty(t, got)
}
