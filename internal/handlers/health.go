package handlers

import (
	"net/http"
	"github.com/JoobyPM/go-load-lab/internal/cache"
)

// LivezHandler is for liveness; always returns OK.
func LivezHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// ReadyzHandler is for readiness; returns OK only if cache has been hydrated.
func ReadyzHandler(w http.ResponseWriter, r *http.Request) {
	if cache.Hydrated {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	} else {
		http.Error(w, "Cache not hydrated yet", http.StatusServiceUnavailable)
	}
}
