package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/k-lomer/lights-out/cache"
	"github.com/k-lomer/lights-out/clients"
	"github.com/k-lomer/lights-out/model"
)

type ListHandler struct {
	dnoClients map[model.Dno]clients.DnoClient
	cache      *cache.OutageCache
}

func NewListHandler(dnoClients map[model.Dno]clients.DnoClient, cache *cache.OutageCache) ListHandler {
	return ListHandler{
		dnoClients,
		cache,
	}
}

func (lh ListHandler) getOutages(ctx context.Context, qp model.QueryParams) ([]model.Outage, error) {
	dnoClients := []clients.DnoClient{}
	for _, dno := range qp.Dnos {
		client := lh.dnoClients[dno]
		if client == nil {
			log.Printf("error getting client for %s", dno)
			continue
		}
		dnoClients = append(dnoClients, client)
	}

	dnoOutages := make([][]model.Outage, len(dnoClients))
	dnoErrs := make([]error, len(dnoClients))
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	var wg sync.WaitGroup

	for i, client := range dnoClients {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if err := recover(); err != nil {
					log.Printf("panic occurred getting outages for %s: %v", client.GetDno(), err)
					dnoErrs[i] = fmt.Errorf("%v", err)
				}
			}()

			outages, err := clients.ListOutages(ctx, client, lh.cache)
			if err != nil {
				log.Printf("error getting outages for %s: %v", client.GetDno(), err)
				dnoErrs[i] = err
				return
			}

			dnoOutages[i] = outages
		}()
	}

	wg.Wait()

	clientErrors := 0
	for _, err := range dnoErrs {
		if err != nil {
			clientErrors += 1
		}
	}
	if clientErrors == len(dnoClients) {
		return nil, errors.New("all DNO clients failed")
	}

	totalOutages := model.AggregateOutages(dnoOutages)

	// Sort to ensure determinism.
	slices.SortFunc(totalOutages, model.KeyComp)

	totalOutages = model.FilterByStatus(totalOutages, qp.Status)
	totalOutages = model.FilterByPostcodes(totalOutages, qp.Postcodes)

	// Page size 0 means return all results.
	if qp.PageSize == 0 {
		return totalOutages, nil
	}

	// Return outages based on page size and page index.
	startIndex := min(uint(len(totalOutages)), qp.PageSize*qp.PageIndex)
	endIndex := min(uint(len(totalOutages)), qp.PageSize*(qp.PageIndex+1))
	pageOutages := totalOutages[startIndex:endIndex]
	return pageOutages, nil
}

func (lh ListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	qp, err := model.ParseQueryParams(r.URL.Query())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	outages, err := lh.getOutages(r.Context(), qp)
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
