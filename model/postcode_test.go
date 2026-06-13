package model

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewPostcode(t *testing.T) {
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

func Test_ParsePostcodes(t *testing.T) {
	var postcodeTestCases = []struct {
		input       []string
		expected    []Postcode
		stopOnError bool
		expectErr   bool
	}{
		{[]string{}, []Postcode{}, true, false},
		{[]string{}, []Postcode{}, false, false},
		{[]string{""}, []Postcode{}, true, true},
		{[]string{""}, []Postcode{}, false, false},
		{[]string{"A33AA"}, []Postcode{Postcode("A3 3AA")}, true, false},
		{[]string{"A33AA"}, []Postcode{Postcode("A3 3AA")}, false, false},
		{[]string{"A33AA", "A44AA"}, []Postcode{Postcode("A3 3AA"), Postcode("A4 4AA")}, true, false},
		{[]string{"A33AA", "A44AA"}, []Postcode{Postcode("A3 3AA"), Postcode("A4 4AA")}, false, false},
		{[]string{"A3"}, []Postcode{Postcode("A3 3AA")}, true, true},
		{[]string{"A3"}, []Postcode{}, false, false},
		{[]string{"A3", "A44AA"}, []Postcode{}, true, true},
		{[]string{"A3", "A44AA"}, []Postcode{Postcode("A4 4AA")}, false, false},
		{[]string{"A33AA", "A4"}, []Postcode{}, true, true},
		{[]string{"A33AA", "A4"}, []Postcode{Postcode("A3 3AA")}, false, false},
		{[]string{"A3", "A4"}, []Postcode{}, true, true},
		{[]string{"A3", "A4"}, []Postcode{}, false, false},
	}

	for _, tc := range postcodeTestCases {
		t.Run(strings.Join(tc.input, ",")+" stopOnError "+strconv.FormatBool(tc.stopOnError), func(t *testing.T) {
			p, err := ParsePostcodes(tc.input, tc.stopOnError)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, Postcodes(tc.expected), p)
			}
		})
	}
}
