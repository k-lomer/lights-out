package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/k-lomer/lights-out/model"
)

const apiBaseUrlSse = "https://ssen-powertrack-api.opcld.com"
const apiRouteSseLiveFaults = "/gridiview/reporter/info/livefaults"

type SseClient struct {
	httpClient *http.Client
}

func MakeSseClient(client *http.Client) SseClient {
	return SseClient{
		httpClient: client,
	}
}

func (client SseClient) GetDno() model.Dno {
	return model.DnoSse
}

func (client SseClient) ListOutages(ctx context.Context) ([]model.Outage, error) {
	req, err := http.NewRequestWithContext(ctx,
		http.MethodGet, apiBaseUrlSse+apiRouteSseLiveFaults, nil)
	if err != nil {
		return nil, err
	}

	res, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Extract to SSE model
	defer drainAndClose(res.Body)
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
