package model

import (
	"encoding/json"
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

type EnergyNorthWestOutages struct {
	Outages      []EnergyNorthWestOutage `json:"Items"`
	TotalOutages int                     `json:"TotalResults"`
}

func (enwo *EnergyNorthWestOutages) UnmarshalJSON(data []byte) error {
	var raw struct {
		Items        []json.RawMessage `json:"Items"`
		TotalResults int               `json:"TotalResults"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	enwo.TotalOutages = raw.TotalResults
	enwo.Outages = decodeOutages[EnergyNorthWestOutage](raw.Items, DnoEnergyNorthWest)
	return nil
}

type EnergyNorthWestOutage struct {
	ID           string               `json:"faultNumber"`
	Start        *EnergyNorthWestTime `json:"date"`
	EstimatedEnd *EnergyNorthWestTime `json:"estimatedTimeOfRestoration"`
	ActualEnd    *EnergyNorthWestTime `json:"actualTimeOfRestoration"`
	Postcodes    Postcodes            `json:"AffectedPostcodes"`
}

func (enw EnergyNorthWestOutage) ToOutage() Outage {
	var start *time.Time
	if enw.Start != nil {
		start = &enw.Start.Time
	}

	var end *time.Time
	if enw.ActualEnd != nil {
		end = &enw.ActualEnd.Time
	} else if enw.EstimatedEnd != nil {
		end = &enw.EstimatedEnd.Time
	}

	return Outage{
		DNO:       DnoEnergyNorthWest,
		ID:        enw.ID,
		Start:     toUTC(start),
		End:       toUTC(end),
		Postcodes: enw.Postcodes,
	}
}

func (enwo EnergyNorthWestOutages) ToOutages() []Outage {
	outages := make([]Outage, len(enwo.Outages))
	for i, f := range enwo.Outages {
		outages[i] = f.ToOutage()
	}
	return outages
}
