package model

import (
	"strings"
	"time"
)

var ukLocation, _ = time.LoadLocation("Europe/London")

const ukpnTimeLayout = "2006-01-02T15:04:05"
const ukpnTimeLayoutMs = "2006-01-02T15:04:05.999"

type ukpnTime struct {
	time.Time
}

type ukpnTimeMs struct {
	time.Time
}

func (u *ukpnTime) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	parsed, err := time.ParseInLocation(ukpnTimeLayout, s, ukLocation)
	if err != nil {
		return err
	}
	u.Time = parsed
	return nil
}

func (u *ukpnTimeMs) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	parsed, err := time.ParseInLocation(ukpnTimeLayoutMs, s, ukLocation)
	if err != nil {
		return err
	}
	u.Time = parsed
	return nil
}

type UKPowerNetworkIncident struct {
	ID        string      `json:"IncidentReference"`
	Start     ukpnTime    `json:"CreationDateTime"`
	Restored  *ukpnTimeMs `json:"RestoredDateTime"`
	Estimated *ukpnTime   `json:"EstimatedRestorationDate"`
	Postcodes []string    `json:"FullPostcodeData"`
}

func (ukpni UKPowerNetworkIncident) ToOutage() Outage {
	var endTime *time.Time
	if ukpni.Restored != nil {
		endTime = &ukpni.Restored.Time
	} else if ukpni.Estimated != nil {
		endTime = &ukpni.Estimated.Time
	}
	return Outage{
		DNO:       "UKPowerNetwork",
		ID:        ukpni.ID,
		Start:     ukpni.Start.Time,
		End:       endTime,
		Postcodes: ukpni.Postcodes,
	}
}

func UKPowerNetworkIncidentsToOutages(ukpnis []UKPowerNetworkIncident) []Outage {
	var outages []Outage

	for _, incident := range ukpnis {
		outages = append(outages, incident.ToOutage())
	}

	return outages
}
