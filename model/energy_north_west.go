package model

import (
	"strings"
	"time"
)

const energyNorthWestTimeLayout = "2006-01-02T15:04:05"

type EnergyNorthWestTime struct {
	time.Time
}

func (t *EnergyNorthWestTime) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)

	parsed, err := time.ParseInLocation(energyNorthWestTimeLayout, s, ukLocation)
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}

type EnergyNorthWestPostcodes struct {
	postcodes []string
}

func (p *EnergyNorthWestPostcodes) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), ` "`)
	p.postcodes = strings.Split(s, ", ")
	return nil
}

type EnergyNorthWestOutages struct {
	Outages      []EnergyNorthWestOutage `json:"Items"`
	TotalOutages int                     `json:"TotalResults"`
}

type EnergyNorthWestOutage struct {
	ID           string                   `json:"faultNumber"`
	Start        EnergyNorthWestTime      `json:"date"`
	EstimatedEnd *EnergyNorthWestTime     `json:"estimatedTimeOfRestoration"`
	ActualEnd    *EnergyNorthWestTime     `json:"actualTimeOfRestoration"`
	Postcodes    EnergyNorthWestPostcodes `json:"AffectedPostcodes"`
}

func (enw EnergyNorthWestOutage) ToOutage() Outage {
	var end *time.Time
	if enw.ActualEnd != nil {
		end = &enw.ActualEnd.Time
	} else if enw.EstimatedEnd != nil {
		end = &enw.EstimatedEnd.Time
	}

	return Outage{
		DNO:       "EnergyNorthWest",
		ID:        enw.ID,
		Start:     enw.Start.Time,
		End:       end,
		Postcodes: enw.Postcodes.postcodes,
	}
}

func (enwo EnergyNorthWestOutages) ToOutages() []Outage {
	outages := make([]Outage, len(enwo.Outages))
	for i, f := range enwo.Outages {
		outages[i] = f.ToOutage()
	}
	return outages
}
