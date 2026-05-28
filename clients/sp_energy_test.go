package clients

import (
	"context"
	"crypto/tls"
	"net/http"
	"testing"
	"time"
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
	if err != nil {
		t.Error(err)
	}
	if count == 0 {
		t.Error("got 0 incident count")
	}
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
	if err != nil {
		t.Error(err)
	}
	if incidents == nil {
		t.Error("failed to get incidents")
	} else if len(incidents.Incidents) != 2 {
		t.Errorf("expected 2 incidents, got %d", len(incidents.Incidents))
	}
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
	if err != nil {
		t.Error(err)
	}
	if len(res) == 0 {
		t.Error("got 0 outages")
	}
}
