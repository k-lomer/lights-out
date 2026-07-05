package model

import (
	"encoding/json"
	"strings"
	"time"
)

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

type UKPowerNetworkOutages []UKPowerNetworkOutage

func (u *UKPowerNetworkOutages) UnmarshalJSON(data []byte) error {
	var raws []json.RawMessage
	if err := json.Unmarshal(data, &raws); err != nil {
		return err
	}
	*u = decodeOutages[UKPowerNetworkOutage](raws, DnoUKPowerNetwork)
	return nil
}

type UKPowerNetworkOutage struct {
	ID        string      `json:"IncidentReference"`
	Start     *ukpnTime   `json:"CreationDateTime"`
	Restored  *ukpnTimeMs `json:"RestoredDateTime"`
	Estimated *ukpnTime   `json:"EstimatedRestorationDate"`
	Postcodes []string    `json:"FullPostcodeData"`
}

func (ukpno UKPowerNetworkOutage) ToOutage() Outage {
	var startTime *time.Time
	if ukpno.Start != nil {
		startTime = &ukpno.Start.Time
	}

	var estimatedEnd *time.Time
	if ukpno.Estimated != nil {
		estimatedEnd = &ukpno.Estimated.Time
	}

	var actualEnd *time.Time
	if ukpno.Restored != nil {
		actualEnd = &ukpno.Restored.Time
	}
	postcodes, _ := ParsePostcodes(ukpno.Postcodes, false)

	return Outage{
		DNO:          DnoUKPowerNetwork,
		ID:           ukpno.ID,
		Start:        toUTC(startTime),
		EstimatedEnd: toUTC(estimatedEnd),
		ActualEnd:    toUTC(actualEnd),
		Postcodes:    postcodes,
	}
}

func UKPowerNetworkToOutages(ukpnos UKPowerNetworkOutages) []Outage {
	var outages []Outage

	for _, outage := range ukpnos {
		outages = append(outages, outage.ToOutage())
	}

	return outages
}
