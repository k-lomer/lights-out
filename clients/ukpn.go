package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/k-lomer/lights-out/model"
)

const apiBaseUrlUkpn = "https://www.ukpowernetworks.co.uk"
const apiRouteUkpnIncidents = "/api/power-cut/all-incidents-light"

type UKPowerNetworkClient struct {
	*UpdateTracker
	httpClient *http.Client
}

func MakeUKPowerNetworkClient(client *http.Client) UKPowerNetworkClient {
	return UKPowerNetworkClient{
		UpdateTracker: &UpdateTracker{},
		httpClient:    client,
	}
}

func (client UKPowerNetworkClient) GetDno() model.Dno {
	return model.DnoUKPowerNetwork
}

func (client UKPowerNetworkClient) ListOutages(ctx context.Context) ([]model.Outage, error) {
	req, err := http.NewRequestWithContext(ctx,
		http.MethodGet, apiBaseUrlUkpn+apiRouteUkpnIncidents, nil)
	if err != nil {
		return nil, err
	}

	// Spoof browser.
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:150.0) Gecko/20100101 Firefox/150.0")
	req.Header.Set("Accept", "text/plain")

	res, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Extract to UKPN model.
	defer drainAndClose(res.Body)
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected return code from %s, %d", client.GetDno(), res.StatusCode)
	}
	var outages model.UKPowerNetworkOutages
	err = json.NewDecoder(res.Body).Decode(&outages)
	if err != nil {
		return nil, err
	}

	return model.UKPowerNetworkToOutages(outages), nil
}
