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

type SPEnergyClient struct {
	httpClient *http.Client
}

func MakeSPEnergyClient(client *http.Client) SPEnergyClient {
	return SPEnergyClient{
		httpClient: client,
	}
}

func (client SPEnergyClient) GetDno() model.Dno {
	return model.DnoSPEnergy
}

func (client SPEnergyClient) getIncidentCount(ctx context.Context) (int, error) {
	body := `{"namespace":"","classname":"@udd/01pSr000002yGTp","method":"getImpactDataCount","isContinuation":false,"params":{"postcode":"","statuses":[]},"cacheable":false}`
	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost, apiBaseUrlSPEnergy+apiRouteSPEnergyExecute, strings.NewReader(body))
	if err != nil {
		return 0, err
	}

	res, err := client.httpClient.Do(req)
	if err != nil {
		return 0, err
	}

	// Extract to SP Energy model
	defer drainAndClose(res.Body)
	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected return code from SP Energy, %d", res.StatusCode)
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

func (client SPEnergyClient) getIncidents(ctx context.Context, count int) (*model.SPEnergyIncidents, error) {
	body := `{"namespace":"","classname":"@udd/01pSr000002yGTp","method":"getImpactData","isContinuation":false,"params":{"paramsJson":"{\"postcode\":\"\",\"pageNumber\":1,\"pageSize\":` + strconv.Itoa(count) + `,\"statuses\":[]}"},"cacheable":false}`
	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost, apiBaseUrlSPEnergy+apiRouteSPEnergyExecute, strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	res, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer drainAndClose(res.Body)
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected return code from SPEnergy, %d", res.StatusCode)
	}
	var incidents model.SPEnergyIncidents
	err = json.NewDecoder(res.Body).Decode(&incidents)
	if err != nil {
		return nil, err
	}

	return &incidents, nil
}

func (client SPEnergyClient) ListOutages(ctx context.Context) ([]model.Outage, error) {

	count, err := client.getIncidentCount(ctx)
	if err != nil {
		return nil, err
	}

	incidents, err := client.getIncidents(ctx, count)
	if err != nil {
		return nil, err
	}

	return incidents.ToOutages(), nil
}
