package model

import (
	"encoding/json"
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

type NorthernPowergridOutages []NorthernPowergridOutage

func (n *NorthernPowergridOutages) UnmarshalJSON(data []byte) error {
	var raws []json.RawMessage
	if err := json.Unmarshal(data, &raws); err != nil {
		return err
	}
	*n = decodeOutages[NorthernPowergridOutage](raws, DnoNorthernPowergrid)
	return nil
}

const northernPowergridCompletedMessage = "The scheduled work has now been completed"

type NorthernPowergridOutage struct {
	ID           string                        `json:"Reference"`
	Start        *time.Time                    `json:"LoggedTime"`
	EstimatedEnd OptionalNorthernPowergridTime `json:"EstimatedTimeTillResolution"`
	Updated      *time.Time                    `json:"UpdateDate"`
	Message      string                        `json:"CustomerStageSequenceMessage"`
	Postcode     Postcode                      `json:"Postcode"`
}

func (npo NorthernPowergridOutage) ToOutage() Outage {
	return Outage{
		DNO:          DnoNorthernPowergrid,
		ID:           npo.ID,
		Start:        toUTC(npo.Start),
		EstimatedEnd: toUTC(npo.EstimatedEnd.Time),
		ActualEnd:    toUTC(npo.actualEndTime()),
		Postcodes:    []Postcode{npo.Postcode},
		Status:       npo.status(),
	}
}

// actualEndTime reports the real restoration time. Northern Powergrid has no
// dedicated restored field, but a completed stage message means the incident is
// over, so its last update time is used as the actual end.
func (npo NorthernPowergridOutage) actualEndTime() *time.Time {
	if npo.Message == northernPowergridCompletedMessage {
		return npo.Updated
	}
	return nil
}

func (npo NorthernPowergridOutage) status() Status {
	if npo.actualEndTime() != nil {
		return StatusResolved
	}
	if npo.Start != nil && npo.Start.After(time.Now()) {
		return StatusFuture
	}
	return StatusActive
}

func (npo NorthernPowergridOutage) getKey() string {
	start := "nil"
	if npo.Start != nil {
		start = npo.Start.String()
	}
	actualEndTimeString := ""
	actualEndTime := npo.actualEndTime()
	if actualEndTime != nil {
		actualEndTimeString = actualEndTime.String()
	}
	return npo.ID + start + npo.EstimatedEnd.String() + actualEndTimeString
}

func NorthernPowergridToOutages(npos NorthernPowergridOutages) []Outage {
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
	return slices.AppendSeq(make([]Outage, 0, len(outages)), maps.Values(outages))
}
