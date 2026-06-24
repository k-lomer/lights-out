package clients

import (
	"context"
	"crypto/tls"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Note that there is a certificate issue for this API
// x509: certificate signed by unknown authority
// This can be checked with `openssl s_client -connect powercuts.spenergynetworks.co.uk:443 -showcerts`
// Use InsecureSkipVerify = true to ignore the incomplete certificate chain (missing intermediate certificates)

func Test_getSPEnergyIncidentCount(t *testing.T) {
	ctx := context.Background()
	var client = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	count, err := MakeSPEnergyClient(client).getIncidentCount(ctx)
	assert.NoError(t, err)
	assert.NotZero(t, count)
}

func Test_getSPEnergyIncidents(t *testing.T) {
	ctx := context.Background()
	var client = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	incidents, err := MakeSPEnergyClient(client).getIncidents(ctx, 2)
	assert.NoError(t, err)
	assert.NotNil(t, incidents)
	assert.Len(t, incidents.Incidents, 2)
}

func Test_ListSPEnergyOutages(t *testing.T) {
	ctx := context.Background()
	var client = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	res, err := MakeSPEnergyClient(client).ListOutages(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}
