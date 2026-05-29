package handlers

import (
	"crypto/tls"
	"encoding/json"
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

func ListHandler(w http.ResponseWriter, r *http.Request) {
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
			outages, err := client.ListOutages(r.Context())
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
		http.Error(w, "all DNO clients failed", http.StatusInternalServerError)
		return
	}

	totalOutages := model.AggregateOutages(&clientResults)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(totalOutages); err != nil {
		log.Printf("error encoding outages: %v", err)
	}
}
