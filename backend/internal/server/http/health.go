package http

import (
	"encoding/json"
	"net/http"
)

// HealthHandler returns a simple 200 OK with a JSON status for infrastructure checks.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
