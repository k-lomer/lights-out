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

type NorthernPowergridClient struct {
	httpClient *http.Client
}

func MakeNorthernPowergridClient(client *http.Client) NorthernPowergridClient {
	return NorthernPowergridClient{
		httpClient: client,
	}
}

func (client NorthernPowergridClient) GetDno() model.Dno {
	return model.DnoNorthernPowergrid
}

func (client NorthernPowergridClient) ListOutages(ctx context.Context) ([]model.Outage, error) {
	req, err := http.NewRequestWithContext(ctx,
		http.MethodGet, apiBaseUrlNorthernPowergrid+apiRoutePowercutsNorthernPowergrid, nil)
	if err != nil {
		return nil, err
	}

	res, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Extract to Northern Powergrid model
	defer drainAndClose(res.Body)
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected return code from Northern Powergrid, %d", res.StatusCode)
	}
	var outages []model.NorthernPowergridOutage
	err = json.NewDecoder(res.Body).Decode(&outages)
	if err != nil {
		return nil, err
	}

	// Convert to Outage model
	return model.NorthernPowergridToOutages(outages), nil
}
