package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/k-lomer/lights-out/model"
)

const apiBaseUrlEnergyNorthWest = "https://www.enwl.co.uk"
const apiRouteEnergyNorthWest = "/api/power-outages/search"

func getEnergyNorthWestOutages(ctx context.Context, client *http.Client, pageSize int) (*model.EnergyNorthWestOutages, error) {
	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost, apiBaseUrlEnergyNorthWest+apiRouteEnergyNorthWest, nil)
	if err != nil {
		return nil, err
	}

	// Add query parameters
	q := req.URL.Query()
	q.Add("pageSize", strconv.Itoa(pageSize))
	q.Add("pageNumber", "1")
	q.Add("includeCurrent", "true")
	q.Add("includeResolved", "true")
	q.Add("includeTodaysPlanned", "true")
	q.Add("includeFuturePlanned", "true")
	q.Add("includeCancelledPlanned", "false")
	req.URL.RawQuery = q.Encode()

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected return code from EnergyNorthWest, %d", res.StatusCode)
	}
	var outages model.EnergyNorthWestOutages
	err = json.NewDecoder(res.Body).Decode(&outages)
	if err != nil {
		return nil, err
	}

	return &outages, nil
}

func ListEnergyNorthWestOutages(ctx context.Context, client *http.Client) ([]model.Outage, error) {
	pageSize := 200
	outages, err := getEnergyNorthWestOutages(ctx, client, pageSize)
	if err != nil {
		return nil, err
	}
	if outages.TotalOutages > pageSize {
		// Larger page size required, get all reported plus a small buffer to be safe.
		pageBufferSize := 10
		outages, err = getEnergyNorthWestOutages(ctx, client, outages.TotalOutages+pageBufferSize)
		if err != nil {
			return nil, err
		}
	}

	return outages.ToOutages(), nil
}
