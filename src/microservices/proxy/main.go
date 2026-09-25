package main

import (
	"log"
	"net/http"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health)
	mux.Handle("/", newProxyHandler(cfg))

	address := ":" + cfg.Port
	log.Printf("Proxy service listening on %s", address)
	log.Fatal(http.ListenAndServe(address, mux))
}
