package handlers

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"slices"
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

func GetOutages(ctx context.Context, qp model.QueryParams) ([]model.Outage, error) {
	var dnoOutages [][]model.Outage
	var wg sync.WaitGroup
	var dnoClients = []clients.DnoClient{
		clients.MakeEnergyNorthWestClient(client),
		clients.MakeNationalGridDistributionClient(client),
		clients.MakeNorthernPowergridClient(client),
		clients.MakeSPEnergyClient(insecureClient),
		clients.MakeSseClient(client),
		clients.MakeUKPowerNetworkClient(client),
	}
	clientErrors := 0
	wg.Add(len(dnoClients))

	for _, client := range dnoClients {
		go func() {
			defer wg.Done()
			outages, err := client.ListOutages(ctx)
			if err != nil {
				log.Printf("error getting outages for %s: %v", client.GetDno(), err)
				clientErrors += 1
				return
			}
			dnoOutages = append(dnoOutages, outages)
		}()
	}

	wg.Wait()

	if clientErrors == len(dnoClients) {
		return nil, errors.New("all DNO clients failed")
	}

	totalOutages := model.AggregateOutages(&dnoOutages)

	// sort to ensure determinism
	slices.SortFunc(totalOutages, model.KeyComp)

	// PageSize 0 means return all results
	if qp.PageSize == 0 {
		return totalOutages, nil
	}

	startIndex := min(uint(len(totalOutages)), qp.PageSize*qp.PageIndex)
	endIndex := min(uint(len(totalOutages)), qp.PageSize*(qp.PageIndex+1))
	pageOutages := totalOutages[startIndex:endIndex]
	return pageOutages, nil
}

func ListHandler(w http.ResponseWriter, r *http.Request) {
	qp, err := model.ParseQueryParams(r.URL.Query())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	outages, err := GetOutages(r.Context(), qp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(outages); err != nil {
		log.Printf("error encoding outages: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
