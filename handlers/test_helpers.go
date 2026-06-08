package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/k-lomer/lights-out/model"
)

func assertStatus(t *testing.T, code int, expectedCode int) {
	if code != expectedCode {
		t.Fatalf("unexpected status code: expected %d, got %d", expectedCode, code)
	}
}

func decodeOutages(t *testing.T, responseBody *bytes.Buffer) []model.Outage {
	var outages []model.Outage
	err := json.NewDecoder(responseBody).Decode(&outages)
	if err != nil {
		t.Fatal(err)
	}

	return outages
}

func checkDnoOutages(t *testing.T, outages []model.Outage) {
	dnoOutageCount := map[model.Dno]int{}
	for _, o := range outages {
		dnoOutageCount[o.DNO] += 1
	}

	for _, dno := range model.AllDnoList {
		if dnoOutageCount[dno] == 0 {
			t.Errorf("Got no outages for %s", dno)
		}
	}
}

func addQueryParams(req *http.Request, k string, v string) {
	q := req.URL.Query()
	q.Add(k, v)
	req.URL.RawQuery = q.Encode()
}
