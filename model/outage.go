package model

import (
	"cmp"
	"net/url"
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
