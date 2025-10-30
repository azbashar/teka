package api

import (
	"encoding/json"
	"net/http"

	"github.com/azbashar/teka/internal/config"
)

func getConfig(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	configJson := config.Cfg

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(configJson)
}