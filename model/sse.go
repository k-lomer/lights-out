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
	ID           string   `json:"reference"`
	Start        *SseTime `json:"loggedAt"`
	EstimatedEnd *SseTime `json:"estimatedRestoration"`
	Updated      *SseTime `json:"updated"`
	Resolved     bool     `json:"resolved"`
	Postcodes    []string `json:"affectedAreas"`
}

func (so SseOutage) ToOutage() Outage {
	postcodes, _ := ParsePostcodes(so.Postcodes, false)

	var start *time.Time
	if so.Start != nil {
		start = &so.Start.Time
	}

	var end *time.Time
	if so.EstimatedEnd != nil {
		end = &so.EstimatedEnd.Time
	}

	return Outage{
		DNO:          DnoSse,
		ID:           so.ID,
		Start:        toUTC(start),
		EstimatedEnd: toUTC(end),
		ActualEnd:    toUTC(so.actualEndTime()),
		Postcodes:    postcodes,
		Status:       so.status(),
	}
}

// actualEndTime reports the real restoration time. SSE has no dedicated restored
// field, but a resolved fault's last update is when it was marked restored.
func (so SseOutage) actualEndTime() *time.Time {
	if so.Resolved && so.Updated != nil {
		return &so.Updated.Time
	}
	return nil
}

func (so SseOutage) status() Status {
	if so.actualEndTime() != nil {
		return StatusResolved
	}
	if so.Start != nil && so.Start.After(time.Now()) {
		return StatusFuture
	}
	return StatusActive
}

func (so SseOutages) ToOutages() []Outage {
	outages := make([]Outage, len(so.Outages))
	for i, f := range so.Outages {
		outages[i] = f.ToOutage()
	}
	return outages
}
