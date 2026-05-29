package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/k-lomer/lights-out/model"
)

const apiBaseUrlNationalGridDistribution = "https://powercuts.nationalgrid.co.uk"
const apiRouteNationalGridDistribution = "/__powercuts/getTabularView"

type NationalGridDistributionClient struct {
	httpClient *http.Client
}

func MakeNationalGridDistributionClient(client *http.Client) NationalGridDistributionClient {
	return NationalGridDistributionClient{
		httpClient: client,
	}
}

func (client NationalGridDistributionClient) GetDno() model.Dno {
	return model.DnoNationalGridDistribution
}

func (client NationalGridDistributionClient) ListOutages(ctx context.Context) ([]model.Outage, error) {
	req, err := http.NewRequestWithContext(ctx,
		http.MethodGet, apiBaseUrlNationalGridDistribution+apiRouteNationalGridDistribution, nil)
	if err != nil {
		return nil, err
	}

	res, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Extract to National Grid Distribution model
	defer drainAndClose(res.Body)
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected return code from NationalGridDistribution, %d", res.StatusCode)
	}
	var powercuts model.NationalGridPowercuts
	err = json.NewDecoder(res.Body).Decode(&powercuts)
	if err != nil {
		return nil, err
	}

	// Convert to Outage model
	return powercuts.ToOutages(), nil
}
