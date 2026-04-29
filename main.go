package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/k-lomer/lights-out/clients"
)

var client *http.Client = &http.Client{
	Timeout: 30 * time.Second,
}

func ListHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("called list handler")

	outages, err := clients.ListSseOutages(r.Context(), client)
	if err != nil {
		log.Printf("error getting SSE outages: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(outages); err != nil {
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
