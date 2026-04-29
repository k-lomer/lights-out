package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/k-lomer/lights-out/model"
)

const apiBaseUrl = "https://ssen-powertrack-api.opcld.com"
const apiRouteLiveFaults = "/gridiview/reporter/info/livefaults"

func ListSseOutages(ctx context.Context, client *http.Client) ([]model.Outage, error) {
	req, err := http.NewRequestWithContext(ctx,
		http.MethodGet, apiBaseUrl+apiRouteLiveFaults, nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// Extract to SSE model
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected return code from SSE, %d", res.StatusCode)
	}
	var liveFaults model.SseFaults
	err = json.NewDecoder(res.Body).Decode(&liveFaults)
	if err != nil {
		return nil, err
	}

	// Convert to Outage model
	return liveFaults.ToOutages(), nil
}
