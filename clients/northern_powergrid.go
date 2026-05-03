package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/k-lomer/lights-out/model"
)

const apiBaseUrlNorthernPowergrid = "https://power.northernpowergrid.com"
const apiRoutePowercutsNorthernPowergrid = "/Powercut_API/rest/powercuts/getall"

func ListNorthernPowergridOutages(ctx context.Context, client *http.Client) ([]model.Outage, error) {
	req, err := http.NewRequestWithContext(ctx,
		http.MethodGet, apiBaseUrlNorthernPowergrid+apiRoutePowercutsNorthernPowergrid, nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// Extract to Northern Powergrid model
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected return code from Northern Powergrid, %d", res.StatusCode)
	}
	var powercuts []model.NorthernPowergridPowercut
	err = json.NewDecoder(res.Body).Decode(&powercuts)
	if err != nil {
		return nil, err
	}

	// Convert to Outage model
	return model.NorthernPowergridPowercutsToOutages(powercuts), nil
}
