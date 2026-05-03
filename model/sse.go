package model

import (
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

type SseFaults struct {
	Faults []SseFault `json:"Faults"`
}

type SseFault struct {
	ID        string   `json:"UUID"`
	Start     SseTime  `json:"loggedAt"`
	End       SseTime  `json:"estimatedRestoration"`
	Postcodes []string `json:"affectedAreas"`
}

func (sf SseFault) ToOutage() Outage {
	return Outage{
		DNO:       "SSE",
		ID:        sf.ID,
		Start:     sf.Start.Time,
		End:       &sf.End.Time,
		Postcodes: sf.Postcodes,
	}
}

func (sf SseFaults) ToOutages() []Outage {
	outages := make([]Outage, len(sf.Faults))
	for i, f := range sf.Faults {
		outages[i] = f.ToOutage()
	}
	return outages
}
