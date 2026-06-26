package model

import (
	"encoding/json"
	"log"
)

// decodeOutages decodes a JSON array of outages element by element
func decodeOutages[T any](raws []json.RawMessage, dno Dno) []T {
	outages := make([]T, 0, len(raws))
	for _, raw := range raws {
		var o T
		if err := json.Unmarshal(raw, &o); err != nil {
			log.Printf("%s: skipping outage that failed to decode: %v", dno, err)
			continue
		}
		outages = append(outages, o)
	}
	return outages
}
