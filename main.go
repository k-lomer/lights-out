package main

import (
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

func aggregateOutages(outages *[][]model.Outage) []model.Outage {
	var totalOutages []model.Outage
	for _, r := range *outages {
		totalOutages = append(totalOutages, r...)
	}
	return totalOutages
}

func ListHandler(w http.ResponseWriter, r *http.Request) {
	var clientResults [][]model.Outage
	var wg sync.WaitGroup
	clientProviders := 2
	clientErrors := 0
	wg.Add(clientProviders)

	go func() {
		defer wg.Done()
		outages, err := clients.ListNorthernPowergridOutages(r.Context(), client)
		if err != nil {
			log.Printf("error getting NortherPowergrid outages: %v", err)
			clientErrors += 1
			return
		}
		clientResults = append(clientResults, outages)
	}()

	go func() {
		defer wg.Done()
		outages, err := clients.ListSseOutages(r.Context(), client)
		if err != nil {
			log.Printf("error getting SSE outages: %v", err)
			clientErrors += 1
			return
		}
		clientResults = append(clientResults, outages)
	}()

	wg.Wait()

	if clientErrors == clientProviders {
		http.Error(w, "all DNO clients failed", http.StatusInternalServerError)
		return
	}

	totalOutages := aggregateOutages(&clientResults)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(totalOutages); err != nil {
		log.Printf("error encoding outages: %v", err)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /list", ListHandler)
	s := http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("starting server")
	err := s.ListenAndServe()
	if err != nil {
		if err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}
}
