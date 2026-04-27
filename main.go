package main

import (
	"log"
	"net/http"
	"time"
)

func ListHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("called list handler")
	w.Write([]byte("list"))
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
