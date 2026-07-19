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
	*UpdateTracker
	httpClient *http.Client
}

func MakeSseClient(client *http.Client) SseClient {
	return SseClient{
		UpdateTracker: &UpdateTracker{},
		httpClient:    client,
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

	// Extract to SSE model.
	defer drainAndClose(res.Body)
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected return code from %s, %d", client.GetDno(), res.StatusCode)
	}
	var liveOutages model.SseOutages
	err = json.NewDecoder(res.Body).Decode(&liveOutages)
	if err != nil {
		return nil, fmt.Errorf("decode %s response: %w", client.GetDno(), err)
	}

	ret := liveOutages.ToOutages()
	model.SetLastUpdated(ret, client.SetUpdated())

	return ret, nil
}
