package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/k-lomer/lights-out/model"
)

func ListHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("called list handler")

	outages := []model.Outage{
		{
			DNO:       "SSE",
			ID:        "572d85a1-ceb1-4b95-b0fc-0ec1b56fb6ce",
			Start:     time.Now(),
			End:       time.Now(),
			Postcodes: []string{"DT2 0HS", "DT2 9PW"},
		},
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
