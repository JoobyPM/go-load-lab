package cache

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// Item represents a sample data structure for our in-memory cache
type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Global cache variables
var (
	Items    []Item
	Hydrated bool
)

// HydrateCache simulates the process of loading data into memory.
// This marks the cache as hydrated once complete.
func HydrateCache() {
	rand.Seed(time.Now().UnixNano())

	const totalItems = 100
	for i := 0; i < totalItems; i++ {
		name := fmt.Sprintf("Item-%d-%d", i, rand.Intn(9999))
		Items = append(Items, Item{ID: i, Name: name})
	}
	Hydrated = true
}

// BusyWait keeps the CPU busy for approximately ms milliseconds by
// doing floating-point math in a tight loop.
func BusyWait(ms int) {
	if ms <= 0 {
		return
	}
	end := time.Now().Add(time.Duration(ms) * time.Millisecond)
	var result float64
	for time.Now().Before(end) {
		result += math.Sqrt(float64(time.Now().UnixNano() % 999999))
	}
	_ = result
}
