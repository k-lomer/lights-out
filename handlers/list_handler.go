package handlers

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/k-lomer/lights-out/clients"
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

func GetOutages(ctx context.Context) ([]model.Outage, error) {
	var clientResults [][]model.Outage
	var wg sync.WaitGroup
	var clients = []clients.DnoClient{
		clients.MakeEnergyNorthWestClient(client),
		clients.MakeNationalGridDistributionClient(client),
		clients.MakeNorthernPowergridClient(client),
		clients.MakeSPEnergyClient(insecureClient),
		clients.MakeSseClient(client),
		clients.MakeUKPowerNetworkClient(client),
	}
	clientErrors := 0
	wg.Add(len(clients))

	for _, client := range clients {
		go func() {
			defer wg.Done()
			outages, err := client.ListOutages(ctx)
			if err != nil {
				log.Printf("error getting outages for %s: %v", client.GetDno(), err)
				clientErrors += 1
				return
			}
			clientResults = append(clientResults, outages)
		}()
	}

	wg.Wait()

	if clientErrors == len(clients) {
		return nil, errors.New("all DNO clients failed")
	}

	totalOutages := model.AggregateOutages(&clientResults)
	return totalOutages, nil
}

func ListHandler(w http.ResponseWriter, r *http.Request) {
	outages, err := GetOutages(r.Context())

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(outages); err != nil {
		log.Printf("error encoding outages: %v", err)
	}
}
