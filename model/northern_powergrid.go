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

type NorthernPowergridOutage struct {
	ID       string                        `json:"Reference"`
	Start    *time.Time                    `json:"LoggedTime"`
	End      OptionalNorthernPowergridTime `json:"EstimatedTimeTillResolution"`
	Postcode Postcode                      `json:"Postcode"`
}

func (npo NorthernPowergridOutage) ToOutage() Outage {
	return Outage{
		DNO:       DnoNorthernPowergrid,
		ID:        npo.ID,
		Start:     toUTC(npo.Start),
		End:       toUTC(npo.End.Time),
		Postcodes: []Postcode{npo.Postcode},
	}
}

func (npo NorthernPowergridOutage) getKey() string {
	start := "nil"
	if npo.Start != nil {
		start = npo.Start.String()
	}
	return npo.ID + start + npo.End.String()
}

func NorthernPowergridToOutages(npos []NorthernPowergridOutage) []Outage {
	outages := map[string]Outage{}
	for _, npo := range npos {
		k := npo.getKey()
		v, ok := outages[k]
		if ok {
			v.Postcodes = append(v.Postcodes, npo.Postcode)
			outages[k] = v
		} else {
			outages[k] = npo.ToOutage()
		}
	}
	return slices.Collect(maps.Values(outages))
}
