package clients

import (
	"context"
	"crypto/tls"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Test that the outage count is fetched from the live SP Energy endpoint.
func Test_getSPEnergyOutageCount(t *testing.T) {
	ctx := context.Background()
	var client = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	_, err := MakeSPEnergyClient(client).getOutageCount(ctx)
	assert.NoError(t, err)
}

// Test that the requested number of outages is fetched from SP Energy.
func Test_getSPEnergyOutages(t *testing.T) {
	ctx := context.Background()
	var client = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	outages, err := MakeSPEnergyClient(client).getOutages(ctx, 2)
	assert.NoError(t, err)
	assert.NotNil(t, outages)
	assert.Len(t, outages.Outages, 2)
}

// Test that the count-then-fetch flow returns outages from SP Energy.
func Test_ListSPEnergyOutages(t *testing.T) {
	ctx := context.Background()
	var client = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	_, err := MakeSPEnergyClient(client).ListOutages(ctx)
	assert.NoError(t, err)
}
