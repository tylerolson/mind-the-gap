package main

import (
	"encoding/json"
	"gtfs-ingestor/gtfs"
	"net/http"
)

func NewAPIServer(port string, store *gtfs.Store) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /arrivals/{stop_id}", func(w http.ResponseWriter, r *http.Request) {
		arrivialUpdate, ok := store.GetArrival(r.PathValue("stop_id"))
		w.Header().Set("Content-Type", "application/json")
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		err := json.NewEncoder(w).Encode(arrivialUpdate)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	return server
}
