package main

import (
	"log"
	"net/http"

	"github.com/amila-t-kumarasekara/web-analyzer/internal/config"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	router := http.NewServeMux()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Web Analyzer is running on port " + config.Port))
	})

	log.Printf("Starting server on port %s", config.Port)
	log.Fatal(http.ListenAndServe(":"+config.Port, router))
}
