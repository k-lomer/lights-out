package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func outage(id string, postcodes ...Postcode) Outage {
	return Outage{ID: id, Postcodes: postcodes}
}

// validOutage returns a minimal outage that passes Validate.
func validOutage() Outage {
	return Outage{
		DNO:    DnoSse,
		ID:     "abc",
		Status: StatusActive,
	}
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

func statusOutage(id string, status Status) Outage {
	return Outage{ID: id, Status: status}
}

// Test that only outages with a targeted status are returned.
func Test_FilterByStatus_Matching(t *testing.T) {
	outages := []Outage{
		statusOutage("a", StatusActive),
		statusOutage("b", StatusFuture),
		statusOutage("c", StatusResolved),
	}

	got := FilterByStatus(outages, []Status{StatusFuture})

	assert.Equal(t, []Outage{statusOutage("b", StatusFuture)}, got)
}

// Test that all matching outages are returned across multiple targeted statuses.
func Test_FilterByStatus_MultipleStatuses(t *testing.T) {
	outages := []Outage{
		statusOutage("a", StatusActive),
		statusOutage("b", StatusFuture),
		statusOutage("c", StatusResolved),
	}

	got := FilterByStatus(outages, []Status{StatusActive, StatusResolved})

	assert.Equal(t, []Outage{statusOutage("a", StatusActive), statusOutage("c", StatusResolved)}, got)
}

// Test that no matching status returns an empty result.
func Test_FilterByStatus_NoMatches(t *testing.T) {
	outages := []Outage{
		statusOutage("a", StatusActive),
		statusOutage("b", StatusActive),
	}

	got := FilterByStatus(outages, []Status{StatusResolved})

	assert.Empty(t, got)
}

// Test that an empty status list matches nothing.
func Test_FilterByStatus_NoStatuses(t *testing.T) {
	outages := []Outage{
		statusOutage("a", StatusActive),
		statusOutage("b", StatusFuture),
	}

	got := FilterByStatus(outages, []Status{})

	assert.Empty(t, got)
}

// Test that a well-formed outage with nil times validates without error.
func Test_Validate_Valid(t *testing.T) {
	assert.NoError(t, validOutage().Validate())
}

// Test that non-nil UTC times and a valid postcode pass validation.
func Test_Validate_ValidWithTimes(t *testing.T) {
	now := time.Now().UTC()
	o := validOutage()
	o.Start = &now
	o.EstimatedEnd = &now
	o.ActualEnd = &now
	o.LastUpdated = now
	o.Postcodes = Postcodes{"AB1 2CD"}

	assert.NoError(t, o.Validate())
}

// Test that each malformed field is rejected by Validate.
func Test_Validate_Invalid(t *testing.T) {
	nonUTC := time.Date(2026, 7, 18, 12, 0, 0, 0, time.FixedZone("BST", 3600))

	cases := map[string]func(*Outage){
		"invalid DNO":          func(o *Outage) { o.DNO = "NotADno" },
		"missing ID":           func(o *Outage) { o.ID = "" },
		"non-UTC Start":        func(o *Outage) { o.Start = &nonUTC },
		"non-UTC EstimatedEnd": func(o *Outage) { o.EstimatedEnd = &nonUTC },
		"non-UTC ActualEnd":    func(o *Outage) { o.ActualEnd = &nonUTC },
		"invalid postcode":     func(o *Outage) { o.Postcodes = Postcodes{"not a postcode"} },
		"non-UTC LastUpdated":  func(o *Outage) { o.LastUpdated = nonUTC },
		"invalid status":       func(o *Outage) { o.Status = "Bogus" },
	}

	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			o := validOutage()
			mutate(&o)

			assert.Error(t, o.Validate())
		})
	}
}

// Test that a list of only valid outages is returned unchanged with no errors.
func Test_FilterValidOnly_AllValid(t *testing.T) {
	a := validOutage()
	a.ID = "a"
	b := validOutage()
	b.ID = "b"

	got, errs := FilterValidOnly([]Outage{a, b})

	assert.Equal(t, []Outage{a, b}, got)
	assert.Empty(t, errs)
}

// Test that invalid outages are dropped and reported while valid ones are kept in order.
func Test_FilterValidOnly_MixedValidity(t *testing.T) {
	valid1 := validOutage()
	valid1.ID = "keep1"
	invalid := validOutage()
	invalid.ID = "" // Missing ID fails validation.
	valid2 := validOutage()
	valid2.ID = "keep2"

	got, errs := FilterValidOnly([]Outage{valid1, invalid, valid2})

	assert.Equal(t, []Outage{valid1, valid2}, got)
	assert.Len(t, errs, 1)
}

// Test that an all-invalid list yields no outages and one error per outage.
func Test_FilterValidOnly_AllInvalid(t *testing.T) {
	invalid := validOutage()
	invalid.DNO = "Nope"

	got, errs := FilterValidOnly([]Outage{invalid, invalid})

	assert.Empty(t, got)
	assert.Len(t, errs, 2)
}

// Test that an empty input returns no outages and no errors.
func Test_FilterValidOnly_Empty(t *testing.T) {
	got, errs := FilterValidOnly([]Outage{})

	assert.Empty(t, got)
	assert.Empty(t, errs)
}
