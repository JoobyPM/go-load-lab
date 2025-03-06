package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"github.com/JoobyPM/go-load-lab/internal/cache"
)

// WaitHandler waits for `time` ms before responding
func WaitHandler(w http.ResponseWriter, r *http.Request) {
	waitStr := r.URL.Query().Get("time")
	waitTime, err := strconv.Atoi(waitStr)
	if err != nil {
		waitTime = 100
	}
	time.Sleep(time.Duration(waitTime) * time.Millisecond)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Waited %d ms\n", waitTime)))
}

// HeavyCallHandler tries to consume CPU in a more controlled way
// Example: /havy-call?cpu=100m&duration=5
//   - cpu=100m means ~10% CPU usage
//   - duration=5 means hold it for 5 seconds
func HeavyCallHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	cpuStr := query.Get("cpu")
	if cpuStr == "" {
		cpuStr = "100m"
	}
	cpuStr = strings.ToLower(cpuStr)
	cpuStr = strings.TrimSuffix(cpuStr, "m")

	cpuMillicores, err := strconv.Atoi(cpuStr)
	if err != nil {
		cpuMillicores = 100
	}

	fraction := float64(cpuMillicores) / 1000.0
	if fraction > 1.0 {
		fraction = 1.0
	} else if fraction < 0 {
		fraction = 0.0
	}

	durationSec := 5
	if durStr := query.Get("duration"); durStr != "" {
		if val, err := strconv.Atoi(durStr); err == nil && val > 0 {
			durationSec = val
		}
	}

	timesliceMs := 100
	totalSlices := (durationSec * 1000) / timesliceMs

	startTime := time.Now()
	for i := 0; i < totalSlices; i++ {
		busyMs := int(float64(timesliceMs) * fraction)
		sleepMs := timesliceMs - busyMs

		cache.BusyWait(busyMs)
		time.Sleep(time.Duration(sleepMs) * time.Millisecond)
	}
	totalTime := time.Since(startTime)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(
		fmt.Sprintf("Heavy call: requested %dm -> fraction=%.2f, duration=%ds, actual=%.2fs\n",
			cpuMillicores, fraction, durationSec, totalTime.Seconds()),
	))
}
