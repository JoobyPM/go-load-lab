package main

import (
	"fmt"
	"log"
	"net/http"
	"github.com/JoobyPM/go-load-lab/internal/cache"
	"github.com/JoobyPM/go-load-lab/internal/handlers"
)

func main() {
	// Hydrate our in-memory cache
	cache.HydrateCache()

	// Serve static files from ./static at the root path "/"
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// Register dynamic routes
	http.HandleFunc("/info", handlers.InfoHandler)
	http.HandleFunc("/livez", handlers.LivezHandler)
	http.HandleFunc("/readyz", handlers.ReadyzHandler)
	http.HandleFunc("/wait", handlers.WaitHandler)
	http.HandleFunc("/havy-call", handlers.HeavyCallHandler)
	http.HandleFunc("/items", handlers.ItemsHandler)

	port := ":8080"
	fmt.Printf("Serving on port %s...\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
