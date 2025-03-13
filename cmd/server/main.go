package main

import (
    "io"
    "log"
    "net/http"
    "os"
    "github.com/JoobyPM/go-load-lab/internal/cache"
    "github.com/JoobyPM/go-load-lab/internal/handlers"
)

func main() {
    // ----------------------------------------------------------------
    // 1) Check LOG_FILE env to decide if we log to a file
    // ----------------------------------------------------------------
    logFilePath := os.Getenv("LOG_FILE")
    if logFilePath == "" {
        // LOG_FILE not set => log only to stdout
        log.SetOutput(os.Stdout)
        log.SetFlags(log.LstdFlags | log.Lshortfile)
        log.Println("INFO: LOG_FILE is not set, logging ONLY to stdout")
    } else {
        // LOG_FILE is set => try opening/creating the file + multiwriter
        file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
        if err != nil {
            // If we can’t open the file, fail fast or just fallback to stdout
            log.Fatalf("Failed to open log file %s: %v", logFilePath, err)
        }
		// Ensure file is closed on exit
        defer file.Close()

		// Create a MultiWriter so logs go to both stdout and file
        multiWriter := io.MultiWriter(os.Stdout, file)
        log.SetOutput(multiWriter)

		// Include date/time in each log line
        log.SetFlags(log.LstdFlags | log.Lshortfile)

        log.Printf("INFO: Logging to BOTH stdout and file %s\n", logFilePath)
    }

    // ----------------------------------------------------------------
    // 2) Hydrate in-memory cache
    // ----------------------------------------------------------------
    log.Println("INFO: Hydrating cache ...")
    cache.HydrateCache()
    log.Printf("INFO: Cache hydrated, total items = %d\n", len(cache.Items))

    // ----------------------------------------------------------------
    // 3) Register handlers
    // ----------------------------------------------------------------
	// Serve static files from ./static at the root path
    fs := http.FileServer(http.Dir("./static"))
    http.Handle("/", loggingMiddleware(fs))

    http.HandleFunc("/info", loggingMiddlewareFunc(handlers.InfoHandler))
    http.HandleFunc("/livez", loggingMiddlewareFunc(handlers.LivezHandler))
    http.HandleFunc("/readyz", loggingMiddlewareFunc(handlers.ReadyzHandler))
    http.HandleFunc("/wait", loggingMiddlewareFunc(handlers.WaitHandler))
    http.HandleFunc("/havy-call", loggingMiddlewareFunc(handlers.HeavyCallHandler))
    http.HandleFunc("/items", loggingMiddlewareFunc(handlers.ItemsHandler))

    // ----------------------------------------------------------------
    // 4) Start the server
    // ----------------------------------------------------------------
    port := ":8080"
    log.Printf("INFO: Starting server on port %s ...\n", port)
    err := http.ListenAndServe(port, nil)
    if err != nil {
        log.Fatalf("ERROR: ListenAndServe failed: %v", err)
    }
}

// loggingMiddleware logs each request
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("INFO: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
        next.ServeHTTP(w, r)
    })
}

// loggingMiddlewareFunc is like loggingMiddleware but for http.HandlerFunc
func loggingMiddlewareFunc(fn http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        log.Printf("INFO: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
        fn(w, r)
    }
}
