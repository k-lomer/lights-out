package model

import (
	"encoding/json"
	"strings"
	"time"
)

const energyNorthWestTimeLayout = "2006-01-02T15:04:05"

type EnergyNorthWestOutageType string

const (
	energyNorthWestCurrentFault       EnergyNorthWestOutageType = "CurrentFault"
	energyNorthWestResolvedFault      EnergyNorthWestOutageType = "ResolvedFault"
	energyNorthWestTodaysPlannedWorks EnergyNorthWestOutageType = "TodaysPlannedWorks"
	energyNorthWestFuturePlannedWorks EnergyNorthWestOutageType = "FuturePlannedWorks"
)

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
	ID           string                    `json:"faultNumber"`
	Type         EnergyNorthWestOutageType `json:"Type"`
	Start        *EnergyNorthWestTime      `json:"date"`
	EstimatedEnd *EnergyNorthWestTime      `json:"estimatedTimeOfRestoration"`
	ActualEnd    *EnergyNorthWestTime      `json:"actualTimeOfRestoration"`
	Postcodes    Postcodes                 `json:"AffectedPostcodes"`
}

func (enw EnergyNorthWestOutage) ToOutage() Outage {
	var start *time.Time
	if enw.Start != nil {
		start = &enw.Start.Time
	}

	var estimatedEnd *time.Time
	if enw.EstimatedEnd != nil {
		estimatedEnd = &enw.EstimatedEnd.Time
	}

	var actualEnd *time.Time
	if enw.ActualEnd != nil {
		actualEnd = &enw.ActualEnd.Time
	}

	return Outage{
		DNO:          DnoEnergyNorthWest,
		ID:           enw.ID,
		Start:        toUTC(start),
		EstimatedEnd: toUTC(estimatedEnd),
		ActualEnd:    toUTC(actualEnd),
		Postcodes:    enw.Postcodes,
		Status:       enw.status(actualEnd),
	}
}

// status maps the Energy North West fault type to a canonical status. Cancelled
// or unrecognised types fall back to whether the power is already back on.
func (enw EnergyNorthWestOutage) status(actualEnd *time.Time) Status {
	switch enw.Type {
	case energyNorthWestCurrentFault, energyNorthWestTodaysPlannedWorks:
		return StatusActive
	case energyNorthWestResolvedFault:
		return StatusResolved
	case energyNorthWestFuturePlannedWorks:
		return StatusPlanned
	default:
		if actualEnd != nil {
			return StatusResolved
		}
		return StatusActive
	}
}

func (enwo EnergyNorthWestOutages) ToOutages() []Outage {
	outages := make([]Outage, len(enwo.Outages))
	for i, f := range enwo.Outages {
		outages[i] = f.ToOutage()
	}
	return outages
}
