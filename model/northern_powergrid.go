package model

import (
	"maps"
	"slices"
	"strings"
	"time"
)

type OptionalNorthernPowergridTime struct {
	*time.Time
}

func (o OptionalNorthernPowergridTime) String() string {
	if o.Time == nil {
		return "nil"
	}
	return o.Time.String()
}

func (o *OptionalNorthernPowergridTime) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "1900-01-01T00:00:00" {
		o.Time = nil
		return nil
	}

	parsed, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return err
	}
	o.Time = &parsed
	return nil
}

type NorthernPowergridPowercut struct {
	ID       string                        `json:"Reference"`
	Start    time.Time                     `json:"LoggedTime"`
	End      OptionalNorthernPowergridTime `json:"EstimatedTimeTillResolution"`
	Postcode string                        `json:"Postcode"`
}

func (npp NorthernPowergridPowercut) ToOutage() Outage {
	return Outage{
		DNO:       "NorthernPowergrid",
		ID:        npp.ID,
		Start:     npp.Start,
		End:       npp.End.Time,
		Postcodes: []string{npp.Postcode},
	}
}

func (npp NorthernPowergridPowercut) getKey() string {
	return npp.ID + npp.Start.String() + npp.End.String()
}

func NorthernPowergridPowercutsToOutages(npps []NorthernPowergridPowercut) []Outage {
	outages := map[string]Outage{}
	for _, npp := range npps {
		k := npp.getKey()
		v, ok := outages[k]
		if ok {
			v.Postcodes = append(v.Postcodes, npp.Postcode)
			outages[k] = v
		} else {
			outages[k] = npp.ToOutage()
		}
	}
	return slices.Collect(maps.Values(outages))
}
