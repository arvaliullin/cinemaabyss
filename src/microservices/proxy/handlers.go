package main

import (
	"encoding/json"
	"net/http"
)

// health сообщает о доступности proxy-сервиса.
func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"status": true})
}
