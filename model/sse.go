package model

import (
	"encoding/json"
	"strings"
	"time"
)

const sseTimeLayout = "2006-01-02T15:04:05.999-0700"

type SseTime struct {
	time.Time
}

func (st *SseTime) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	parsed, err := time.Parse(sseTimeLayout, s)
	if err != nil {
		return err
	}
	st.Time = parsed
	return nil
}

type SseOutages struct {
	Outages []SseOutage `json:"Faults"`
}

func (so *SseOutages) UnmarshalJSON(data []byte) error {
	var raw struct {
		Faults []json.RawMessage `json:"Faults"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	so.Outages = decodeOutages[SseOutage](raw.Faults, DnoSse)
	return nil
}

type SseOutage struct {
	ID        string   `json:"UUID"`
	Start     *SseTime `json:"loggedAt"`
	End       *SseTime `json:"estimatedRestoration"`
	Postcodes []string `json:"affectedAreas"`
}

func (so SseOutage) ToOutage() Outage {
	postcodes, _ := ParsePostcodes(so.Postcodes, false)

	var start *time.Time
	if so.Start != nil {
		start = &so.Start.Time
	}

	var end *time.Time
	if so.End != nil {
		end = &so.End.Time
	}

	return Outage{
		DNO:       DnoSse,
		ID:        so.ID,
		Start:     toUTC(start),
		End:       toUTC(end),
		Postcodes: postcodes,
	}
}

func (so SseOutages) ToOutages() []Outage {
	outages := make([]Outage, len(so.Outages))
	for i, f := range so.Outages {
		outages[i] = f.ToOutage()
	}
	return outages
}
