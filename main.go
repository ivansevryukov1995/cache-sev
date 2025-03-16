package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/ivansevryukov1995/cache-sev/pkg/lru"
)

// var cache *lfu.Cache[int, string]

var cache *lru.Cache[int, string]

func init() {
	const cacheCapacity = 100
	const ttl = time.Second * 10
	const cleanupInterval = time.Second * 5

	// cache = lfu.NewCache[int, string](cacheCapacity, ttl, cleanupInterval)
	cache = lru.NewCache[int, string](cacheCapacity, ttl, cleanupInterval)
}

func computeExpensiveOperation(key int) string {
	// Simulation of an expensive operation
	time.Sleep(time.Second * 2) // 200 мс
	return fmt.Sprintf("Result for key %d is %d", key, rand.Intn(1000))
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	keyStr := r.URL.Query().Get("key")
	key, err := strconv.Atoi(keyStr)
	if err != nil {
		http.Error(w, "Invalid key", http.StatusBadRequest)
		return
	}

	// Checking if the key is in the cache
	if value, found := cache.Get(key); found {
		w.Header().Set("X-Cache", "HIT")
		json.NewEncoder(w).Encode(value)
		return
	}

	// If the key is not in the cache, we perform an expensive calculation
	result := computeExpensiveOperation(key)

	// Saving the result in the cache
	cache.Put(key, result)

	w.Header().Set("X-Cache", "MISS")
	json.NewEncoder(w).Encode(result)
}

func main() {
	http.HandleFunc("/compute", getHandler)
	http.ListenAndServe(":8080", nil)
}
