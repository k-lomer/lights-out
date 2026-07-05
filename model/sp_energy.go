package model

import (
	"encoding/json"
	"strings"
	"time"
)

const sPEnergyStartTimeLayout1 = "2006-01-02 15:04:05"
const sPEnergyStartTimeLayout2 = "2/1/2006, 15:04"
const sPEnergyEndTimeLayout1 = "1/2/2006, 3:04 PM"
const sPEnergyEndTimeLayout2 = "2/1/2006, 15:04"

type SPEnergyStartTime struct {
	time.Time
}

type SPEnergyEndTime struct {
	time.Time
}

func (t *SPEnergyStartTime) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	parsed, err := time.ParseInLocation(sPEnergyStartTimeLayout1, s, ukLocation)
	if err != nil {
		parsed, err = time.ParseInLocation(sPEnergyStartTimeLayout2, s, ukLocation)
	}
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}

func (t *SPEnergyEndTime) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	parsed, err := time.ParseInLocation(sPEnergyEndTimeLayout1, s, ukLocation)
	if err != nil {
		parsed, err = time.ParseInLocation(sPEnergyEndTimeLayout2, s, ukLocation)
	}
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}

type SPEnergyOutages struct {
	Outages []SPEnergyOutage `json:"returnValue"`
}

func (speo *SPEnergyOutages) UnmarshalJSON(data []byte) error {
	var raw struct {
		ReturnValue []json.RawMessage `json:"returnValue"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	speo.Outages = decodeOutages[SPEnergyOutage](raw.ReturnValue, DnoSPEnergy)
	return nil
}

type SPEnergyOutage struct {
	ID           string             `json:"incidentReference"`
	Start        *SPEnergyStartTime `json:"createdDate"`
	EstimatedEnd *SPEnergyEndTime   `json:"estimatedFix"`
	ActualEnd    *SPEnergyEndTime   `json:"actualRestorationTime"`
	Postcodes    Postcodes          `json:"postcodeList"`
}

func (speo SPEnergyOutage) ToOutage() Outage {
	var start *time.Time
	if speo.Start != nil {
		start = &speo.Start.Time
	}

	var estimatedEnd *time.Time
	if speo.EstimatedEnd != nil {
		estimatedEnd = &speo.EstimatedEnd.Time
	}

	var actualEnd *time.Time
	if speo.ActualEnd != nil {
		actualEnd = &speo.ActualEnd.Time
	}

	return Outage{
		DNO:          DnoSPEnergy,
		ID:           speo.ID,
		Start:        toUTC(start),
		EstimatedEnd: toUTC(estimatedEnd),
		ActualEnd:    toUTC(actualEnd),
		Postcodes:    speo.Postcodes,
	}
}

func (speo SPEnergyOutages) ToOutages() []Outage {
	outages := make([]Outage, len(speo.Outages))
	for i, f := range speo.Outages {
		outages[i] = f.ToOutage()
	}
	return outages
}
