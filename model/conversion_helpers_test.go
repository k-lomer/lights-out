package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// assertConverted checks the invariants that hold for every outage.
// The DNO is set to the expected constant.
// Every postcode is in normalised form.
// When startRequired is true it asserts Start is non-nil, for providers that always have a start time.
func assertConverted(t *testing.T, outages []Outage, dno Dno, startRequired bool) {
	for _, o := range outages {
		assert.Equal(t, dno, o.DNO)
		if startRequired {
			assert.NotNil(t, o.Start)
		}
		// Times are standardised to UTC in the canonical model.
		if o.Start != nil {
			assert.Equal(t, time.UTC, o.Start.Location())
		}
		if o.EstimatedEnd != nil {
			assert.Equal(t, time.UTC, o.EstimatedEnd.Location())
		}
		if o.ActualEnd != nil {
			assert.Equal(t, time.UTC, o.ActualEnd.Location())
		}
		for _, p := range o.Postcodes {
			assert.Regexp(t, isPostcode, string(p))
		}
	}
}

// assertTimeEqual asserts got is non-nil and points to the same instant as want.
func assertTimeEqual(t *testing.T, want time.Time, got *time.Time) {
	if assert.NotNil(t, got) {
		assert.True(t, want.Equal(*got), "want %s, got %s", want, got)
	}
}
