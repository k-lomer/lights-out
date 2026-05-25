package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/k-lomer/lights-out/model"
)

const apiBaseUrlSPEnergy = "https://powercuts.spenergynetworks.co.uk"
const apiRouteSPEnergyExecute = "/webruntime/api/apex/execute?language=en-US&asGuest=true&htmlEncode=false"

func getSPEnergyIncidentCount(ctx context.Context, client *http.Client) (int, error) {
	body := `{"namespace":"","classname":"@udd/01pSr000002yGTp","method":"getImpactDataCount","isContinuation":false,"params":{"postcode":"","statuses":[]},"cacheable":false}`
	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost, apiBaseUrlSPEnergy+apiRouteSPEnergyExecute, strings.NewReader(body))
	if err != nil {
		return 0, err
	}

	res, err := client.Do(req)
	if err != nil {
		return 0, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected return code from SSE, %d", res.StatusCode)
	}

	var result struct {
		Count int `json:"returnValue"`
	}
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return 0, err
	}
	return result.Count, nil
}

func getSPEnergyIncidents(ctx context.Context, client *http.Client, count int) (*model.SPEnergyIncidents, error) {
	body := `{"namespace":"","classname":"@udd/01pSr000002yGTp","method":"getImpactData","isContinuation":false,"params":{"paramsJson":"{\"postcode\":\"\",\"pageNumber\":1,\"pageSize\":` + strconv.Itoa(count) + `,\"statuses\":[]}"},"cacheable":false}`
	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost, apiBaseUrlSPEnergy+apiRouteSPEnergyExecute, strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected return code from SSE, %d", res.StatusCode)
	}
	var incidents model.SPEnergyIncidents
	err = json.NewDecoder(res.Body).Decode(&incidents)
	if err != nil {
		return nil, err
	}

	return &incidents, nil
}

func ListSPEnergyOutages(ctx context.Context, client *http.Client) ([]model.Outage, error) {

	count, err := getSPEnergyIncidentCount(ctx, client)
	if err != nil {
		return nil, err
	}

	incidents, err := getSPEnergyIncidents(ctx, client, count)
	if err != nil {
		return nil, err
	}

	return incidents.ToOutages(), nil
}
