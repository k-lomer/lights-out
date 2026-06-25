package model

import (
	"cmp"
	"net/url"
	"slices"
	"time"
)

var ukLocation, _ = time.LoadLocation("Europe/London")

type Outage struct {
	DNO       Dno        `json:"dno"`
	ID        string     `json:"id"`
	Start     *time.Time `json:"start_time"`
	End       *time.Time `json:"end_time"`
	Postcodes Postcodes  `json:"postcodes"`
}

func (o Outage) GetKey() string {
	return string(o.DNO) + "_" + url.QueryEscape(o.ID)
}

func AggregateOutages(outages [][]Outage) []Outage {
	var totalOutages []Outage
	for _, r := range outages {
		totalOutages = append(totalOutages, r...)
	}
	return totalOutages
}

func KeyComp(o1, o2 Outage) int {
	return cmp.Compare(o1.GetKey(), o2.GetKey())
}

func FilterByPostcodes(outages []Outage, postcodes Postcodes) []Outage {
	// An empty postcode list leaves the outages unchanged.
	if len(postcodes) == 0 {
		return outages
	}

	hash := postcodes.GetHashMap()

	return slices.Collect(func(yield func(Outage) bool) {
		for _, o := range outages {
			for _, p := range o.Postcodes {
				if hash[p] {
					if !yield(o) {
						return
					}
					break
				}
			}
		}
	})
}
