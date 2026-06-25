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

	parsed, err := time.ParseInLocation(nationalGridTimeLayout, s, ukLocation)
	if err != nil {
		return err
	}
	o.Time = &parsed
	return nil
}

type NationalGridOutages struct {
	Outages []NationalGridOutage `json:"incidents"`
}

type NationalGridOutage struct {
	ID        string                   `json:"id"`
	Start     OptionalNationalGridTime `json:"startTime"`
	End       OptionalNationalGridTime `json:"etr"`
	Postcodes Postcodes                `json:"postcodes"`
}

func (ngo NationalGridOutage) ToOutage() Outage {
	return Outage{
		DNO:       DnoNationalGridDistribution,
		ID:        ngo.ID,
		Start:     toUTC(ngo.Start.Time),
		End:       toUTC(ngo.End.Time),
		Postcodes: ngo.Postcodes,
	}
}

func (ngo NationalGridOutages) ToOutages() []Outage {
	outages := make([]Outage, len(ngo.Outages))
	for i, f := range ngo.Outages {
		outages[i] = f.ToOutage()
	}
	return outages
}
