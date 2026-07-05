package model

import (
	"cmp"
	"net/url"
	"slices"
	"time"
)

type Status string

const (
	StatusActive   Status = "Active"
	StatusFuture   Status = "Future"
	StatusResolved Status = "Resolved"
)

var AllStatusList = [3]Status{
	StatusActive,
	StatusFuture,
	StatusResolved,
}

// ukLocation is the assumed timezone for DNOs that report timestamps without a
// timezone offset (Energy North West, National Grid, SP Energy, UK Power Network).
// We assume those times without zones are UK local and parse them here before
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
	DNO          Dno        `json:"dno"`
	ID           string     `json:"id"`
	Start        *time.Time `json:"start_time"`
	EstimatedEnd *time.Time `json:"estimated_end"`
	ActualEnd    *time.Time `json:"actual_end"`
	Postcodes    Postcodes  `json:"postcodes"`
	LastUpdated  time.Time  `json:"last_updated_time"`
	Status       Status     `json:"status"`
}

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

func SetLastUpdated(outages []Outage, t time.Time) {
	for i := range outages {
		outages[i].LastUpdated = t
	}
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

func FilterByStatus(outages []Outage, status []Status) []Outage {
	return slices.Collect(func(yield func(Outage) bool) {
		for _, o := range outages {
			for _, s := range status {
				if o.Status == s {
					if !yield(o) {
						return
					}
					break
				}
			}
		}
	})
}
