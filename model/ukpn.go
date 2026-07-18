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
	Creation  *ukpnTime   `json:"CreationDateTime"`
	Received  *ukpnTime   `json:"ReceivedDate"`
	Planned   *ukpnTime   `json:"PlannedDate"`
	Restored  *ukpnTimeMs `json:"RestoredDateTime"`
	Estimated *ukpnTime   `json:"EstimatedRestorationDate"`
	Postcodes []string    `json:"FullPostcodeData"`
}

func (ukpno UKPowerNetworkOutage) ToOutage() Outage {
	postcodes, _ := ParsePostcodes(ukpno.Postcodes, false)

	return Outage{
		DNO:          DnoUKPowerNetwork,
		ID:           ukpno.ID,
		Start:        toUTC(ukpno.startTime()),
		EstimatedEnd: toUTC(ukpno.estimatedEndTime()),
		ActualEnd:    toUTC(ukpno.actualEndTime()),
		Postcodes:    postcodes,
		Status:       ukpno.status(),
	}
}

// startTime picks the outage start. For planned work the creation date is when
// the job was scheduled, often weeks before the outage, so the received date
// (when the outage actually began) or the scheduled planned date is used
// instead.
func (ukpno UKPowerNetworkOutage) startTime() *time.Time {
	if ukpno.Planned != nil {
		if ukpno.Received != nil {
			return &ukpno.Received.Time
		}
		return &ukpno.Planned.Time
	}

	if ukpno.Creation != nil {
		return &ukpno.Creation.Time
	}
	return nil
}

func (ukpno UKPowerNetworkOutage) estimatedEndTime() *time.Time {
	if ukpno.Estimated != nil {
		return &ukpno.Estimated.Time
	}
	return nil
}

func (ukpno UKPowerNetworkOutage) actualEndTime() *time.Time {
	if ukpno.Restored != nil {
		return &ukpno.Restored.Time
	}
	return nil
}

func (ukpno UKPowerNetworkOutage) status() Status {
	if ukpno.actualEndTime() != nil {
		return StatusResolved
	}
	start := ukpno.startTime()
	if start != nil && start.After(time.Now()) {
		return StatusFuture
	}
	return StatusActive
}

func UKPowerNetworkToOutages(ukpnos UKPowerNetworkOutages) []Outage {
	outages := make([]Outage, len(ukpnos))
	for i, outage := range ukpnos {
		outages[i] = outage.ToOutage()
	}

	return outages
}
