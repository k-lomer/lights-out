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

func ListUkpnOutages(ctx context.Context, client *http.Client) ([]model.Outage, error) {
	req, err := http.NewRequestWithContext(ctx,
		http.MethodGet, apiBaseUrlUkpn+apiRouteUkpnIncidents, nil)
	if err != nil {
		return nil, err
	}

	// Spoof browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:150.0) Gecko/20100101 Firefox/150.0")
	req.Header.Set("Accept", "text/plain")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// Extract to UKPN model
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected return code from UKPN, %d", res.StatusCode)
	}
	var incidents []model.UKPowerNetworkIncident
	err = json.NewDecoder(res.Body).Decode(&incidents)
	if err != nil {
		return nil, err
	}

	// Convert to Outage model
	return model.UKPowerNetworkIncidentsToOutages(incidents), nil
}
