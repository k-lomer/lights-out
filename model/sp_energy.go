package model

import (
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

type SPEnergyIncidents struct {
	Incidents []SPEnergyIncident `json:"returnValue"`
}

type SPEnergyIncident struct {
	ID           string            `json:"incidentReference"`
	Start        SPEnergyStartTime `json:"CreatedDate"`
	EstimatedEnd SPEnergyEndTime   `json:"estimatedFix"`
	ActualEnd    *SPEnergyEndTime  `json:"actualRestorationTime"`
	Postcodes    Postcodes         `json:"postcodeList"`
}

func (spei SPEnergyIncident) ToOutage() Outage {
	var end *time.Time
	if spei.ActualEnd != nil {
		end = &spei.ActualEnd.Time
	} else {
		end = &spei.EstimatedEnd.Time
	}
	return Outage{
		DNO:       DnoSPEnergy,
		ID:        spei.ID,
		Start:     &spei.Start.Time,
		End:       end,
		Postcodes: spei.Postcodes,
	}
}

func (spei SPEnergyIncidents) ToOutages() []Outage {
	outages := make([]Outage, len(spei.Incidents))
	for i, f := range spei.Incidents {
		outages[i] = f.ToOutage()
	}
	return outages
}
