package model

import "time"

var ukLocation, _ = time.LoadLocation("Europe/London")

type Outage struct {
	DNO       string     `json:"dno"`
	ID        string     `json:"id"`
	Start     time.Time  `json:"start_time"`
	End       *time.Time `json:"end_time"`
	Postcodes []string   `json:"postcodes"`
}
