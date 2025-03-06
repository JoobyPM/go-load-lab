package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"github.com/JoobyPM/go-load-lab/internal/cache"
)

// ItemsHandler returns paginated data from our cache
func ItemsHandler(w http.ResponseWriter, r *http.Request) {
	if !cache.Hydrated {
		http.Error(w, "Cache not hydrated yet", http.StatusServiceUnavailable)
		return
	}

	query := r.URL.Query()
	offsetStr := query.Get("offset")
	limitStr := query.Get("limit")

	offset := 0
	if offsetStr != "" {
		if val, err := strconv.Atoi(offsetStr); err == nil && val >= 0 {
			offset = val
		}
	}
	limit := 10
	if limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 {
			limit = val
		}
	}

	start := offset
	if start > len(cache.Items) {
		start = len(cache.Items)
	}
	end := start + limit
	if end > len(cache.Items) {
		end = len(cache.Items)
	}

	subset := cache.Items[start:end]
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(subset)
}
