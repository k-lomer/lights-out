package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ListUkpnOutages(t *testing.T) {
	var postcodeTestCases = []struct {
		input     string
		expected  string
		expectErr bool
	}{
		{"", "", true},
		{"       ", "", true},
		{"a", "", true},
		{"a7", "", true},
		{"aaaaaaa", "", true},
		{"aa77aaaaa", "", true},
		{"1AA 1AA", "", true},
		{"111 111", "", true},
		{"A3.3AA", "", true},
		{"$%^$&^][", "", true},
		{"a33aa", "A3 3AA", false},
		{"A33AA", "A3 3AA", false},
		{"a3 3aa", "A3 3AA", false},
		{"a3 3aa", "A3 3AA", false},
		{"A3 3AA", "A3 3AA", false},
		{" A33AA", "A3 3AA", false},
		{"A33AA ", "A3 3AA", false},
		{" A33AA ", "A3 3AA", false},
		{"    A33AA   ", "A3 3AA", false},
		{"AA3A 3AA ", "AA3A 3AA", false},
		{"A3A 3AA ", "A3A 3AA", false},
		{"A33 3AA ", "A33 3AA", false},
		{"AA3 3AA ", "AA3 3AA", false},
		{"AA3     3AA ", "AA3 3AA", false},
		{"AA33 3AA ", "AA33 3AA", false},
		{"A0 0AA", "A0 0AA", false},
		{"X01 0ZY", "X01 0ZY", false},
		{"Le167jF", "LE16 7JF", false},
		{"[\"N00A A\"]", "N0 0AA", false},
		{"\"MK417 PJ\"", "MK41 7PJ", false},
		{"\"WN4 9AF", "WN4 9AF", false},
		{"CA13 OEN", "CA13 0EN", false},
	}

	for _, tc := range postcodeTestCases {
		t.Run(tc.input, func(t *testing.T) {
			p, err := NewPostcode(tc.input)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tc.expected, string(p))
			}
		})
	}
}
