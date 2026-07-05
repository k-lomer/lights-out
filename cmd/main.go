package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"time"

	"github.com/k-lomer/lights-out/clients"
	"github.com/k-lomer/lights-out/handlers"
	"github.com/k-lomer/lights-out/model"
)

func main() {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// There is a certificate issue for the SP Energy API.
	// x509: certificate signed by unknown authority
	// This can be checked with `openssl s_client -connect powercuts.spenergynetworks.co.uk:443 -showcerts`
	// Use InsecureSkipVerify = true to ignore the incomplete certificate chain (missing intermediate certificates).
	// This could be fixed by manually providing the missing certificates.
	insecureClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	dnoClients := map[model.Dno]clients.DnoClient{
		model.DnoEnergyNorthWest:          clients.MakeEnergyNorthWestClient(client),
		model.DnoNationalGridDistribution: clients.MakeNationalGridDistributionClient(client),
		model.DnoNorthernPowergrid:        clients.MakeNorthernPowergridClient(client),
		model.DnoSPEnergy:                 clients.MakeSPEnergyClient(insecureClient),
		model.DnoSse:                      clients.MakeSseClient(client),
		model.DnoUKPowerNetwork:           clients.MakeUKPowerNetworkClient(client),
	}

	mux := http.NewServeMux()
	lh := handlers.NewListHandler(dnoClients)

	mux.Handle("GET /list", lh)
	s := http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("starting server")
	err := s.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %v", err)
	}
}
