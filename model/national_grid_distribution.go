package model

import (
	"strings"
	"time"
)

const nationalGridTimeLayout = "2006-01-02 15:04:05"

type OptionalNationalGridTime struct {
	*time.Time
}

func (o OptionalNationalGridTime) String() string {
	if o.Time == nil {
		return "nil"
	}
	return o.Time.String()
}

func (o *OptionalNationalGridTime) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "1900-01-01 00:00:00" {
		o.Time = nil
		return nil
	}

	parsed, err := time.Parse(nationalGridTimeLayout, s)
	if err != nil {
		return err
	}
	o.Time = &parsed
	return nil
}

type NationalGridPowercuts struct {
	Incidents []NationalGridPowercut `json:"incidents"`
}

type NationalGridPowercut struct {
	ID        string                   `json:"id"`
	Start     OptionalNationalGridTime `json:"startTime"`
	End       OptionalNationalGridTime `json:"etr"`
	Postcodes []string                 `json:"postcodes"`
}

func (ngp NationalGridPowercut) ToOutage() Outage {

	return Outage{
		DNO:       "NationalGridDistribution",
		ID:        ngp.ID,
		Start:     *ngp.Start.Time,
		End:       ngp.End.Time,
		Postcodes: ngp.Postcodes,
	}
}

func (ukpni NationalGridPowercuts) ToOutages() []Outage {
	outages := make([]Outage, len(ukpni.Incidents))
	for i, f := range ukpni.Incidents {
		outages[i] = f.ToOutage()
	}
	return outages
}
