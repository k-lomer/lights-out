package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"slices"
	"sync"

	"github.com/k-lomer/lights-out/clients"
	"github.com/k-lomer/lights-out/model"
)

type ListHandler struct {
	dnoClients map[model.Dno]clients.DnoClient
}

func NewListHandler(dnoClients map[model.Dno]clients.DnoClient) ListHandler {
	return ListHandler{
		dnoClients,
	}
}

func (lh ListHandler) getOutages(ctx context.Context, qp model.QueryParams) ([]model.Outage, error) {
	dnoClients := []clients.DnoClient{}
	for _, dno := range qp.Dnos {
		dnoClients = append(dnoClients, lh.dnoClients[dno])
	}

	dnoOutages := make([][]model.Outage, len(dnoClients))
	dnoErrs := make([]error, len(dnoClients))
	var wg sync.WaitGroup
	for i, client := range dnoClients {
		wg.Add(1)
		go func() {
			defer wg.Done()
			outages, err := client.ListOutages(ctx)
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
	if clientErrors == len(qp.Dnos) {
		return nil, errors.New("all DNO clients failed")
	}

	totalOutages := model.AggregateOutages(&dnoOutages)

	// sort to ensure determinism
	slices.SortFunc(totalOutages, model.KeyComp)

	// filter by postcode
	if len(qp.Postcodes) > 0 {
		hash := qp.Postcodes.GetHashMap()

		totalOutages = slices.Collect(func(yield func(model.Outage) bool) {
			for _, o := range totalOutages {
				for _, p := range o.Postcodes {
					if hash[p] {
						if !yield(o) {
							return
						}
						continue
					}
				}
			}
		})
	}

	// page size 0 means return all results
	if qp.PageSize == 0 {
		return totalOutages, nil
	}

	// return outages based on page size and page index
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
