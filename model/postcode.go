package model

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

var isPostcode = regexp.MustCompile(`^[A-Z]{1,2}\d[A-Z\d]?\s\d[A-Z]{2}$`)

type Postcode string

func NewPostcode(s string) (Postcode, error) {
	p := s

	// Remove bad characters.
	p = strings.ReplaceAll(p, " ", "")
	p = strings.ReplaceAll(p, "\"", "")
	p = strings.ReplaceAll(p, "'", "")
	p = strings.ReplaceAll(p, "`", "")
	p = strings.ReplaceAll(p, "[", "")
	p = strings.ReplaceAll(p, "]", "")

	if len(p) < 5 {
		return "", fmt.Errorf("invalid postcode length: '%s' (original '%s)", p, s)
	}
	p = strings.ToUpper(p)

	// Handle case where O is used instead of 0.
	b := []byte(p)
	if b[len(b)-3] == 'O' {
		b[len(b)-3] = '0'
		p = string(b)
	}

	inwardCodeIndex := len(p) - 3
	p = p[:inwardCodeIndex] + " " + p[inwardCodeIndex:]
	if !isPostcode.MatchString(p) {
		return "", fmt.Errorf("invalid postcode format: '%s' (original '%s')", p, s)
	}
	return Postcode(p), nil
}

func (p *Postcode) UnmarshalJSON(data []byte) error {
	postcode, err := NewPostcode(string(data))
	if err != nil {
		return err
	}

	*p = postcode
	return nil
}

type Postcodes []Postcode

func ParsePostcodes(postcodeStrings []string) Postcodes {
	var postcodes = []Postcode{}
	for _, s := range postcodeStrings {
		p, err := NewPostcode(s)
		if err != nil {
			log.Printf("%v", err)
			continue
		}
		postcodes = append(postcodes, p)
	}
	return postcodes
}

func (p *Postcodes) UnmarshalJSON(data []byte) error {
	postcodeStrings := strings.Split(string(data), ",")
	postcodes := ParsePostcodes(postcodeStrings)
	*p = postcodes
	return nil
}
