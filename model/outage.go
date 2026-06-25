package model

import (
	"cmp"
	"net/url"
	"slices"
	"time"
)

// ukLocation is the assumed timezone for DNOs that report timestamps without a
// timezone offset (Energy North West, National Grid, SP Energy, UK Power Network).
// We assume those naked times are UK local wall-clock and parse them here before
// normalising to UTC in ToOutage. This is an assumption: if a provider actually
// emits UTC, its summer (BST) times will be an hour out. Northern Powergrid and
// SSE include an explicit offset, so they do not rely on this.
var ukLocation, _ = time.LoadLocation("Europe/London")

// Outage is the canonical, provider-agnostic representation of a power cut.
//
// Start and End are always in UTC. The DNOs report times in a mix of zones —
// some as UK local wall-clock, others as UTC — so each client's ToOutage
// normalises them through toUTC, giving consumers one consistent timezone.
type Outage struct {
	DNO       Dno        `json:"dno"`
	ID        string     `json:"id"`
	Start     *time.Time `json:"start_time"`
	End       *time.Time `json:"end_time"`
	Postcodes Postcodes  `json:"postcodes"`
}

// toUTC normalises a time pointer to UTC, preserving nil. Every Outage time is
// stored in UTC, so each ToOutage routes its start and end through this.
func toUTC(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	u := t.UTC()
	return &u
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
