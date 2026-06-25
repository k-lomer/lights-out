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

var client *http.Client = &http.Client{
	Timeout: 30 * time.Second,
}

var insecureClient *http.Client = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}

func NewDnoClients() map[model.Dno]clients.DnoClient {
	return map[model.Dno]clients.DnoClient{
		model.DnoEnergyNorthWest:          clients.MakeEnergyNorthWestClient(client),
		model.DnoNationalGridDistribution: clients.MakeNationalGridDistributionClient(client),
		model.DnoNorthernPowergrid:        clients.MakeNorthernPowergridClient(client),
		model.DnoSPEnergy:                 clients.MakeSPEnergyClient(insecureClient),
		model.DnoSse:                      clients.MakeSseClient(client),
		model.DnoUKPowerNetwork:           clients.MakeUKPowerNetworkClient(client),
	}
}

func main() {
	mux := http.NewServeMux()
	lh := handlers.NewListHandler(NewDnoClients())

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
