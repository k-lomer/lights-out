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

func ListNationalGridDistributionOutages(ctx context.Context, client *http.Client) ([]model.Outage, error) {
	req, err := http.NewRequestWithContext(ctx,
		http.MethodGet, apiBaseUrlNationalGridDistribution+apiRouteNationalGridDistribution, nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// Extract to National Grid Distribution model
	defer res.Body.Close()
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
