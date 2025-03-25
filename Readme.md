# Installation

To use `cache-sev` in your Go project, you can install it using the following:
```bash 
go get github.com/ivansevryukov1995/cache-sev
```
# Evict algorithm
* lru
* lfu

# Commands
* Get
* Put

# Example

Run the server code below `go run .` :

```bash 
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

var cache Cacher[int, string]

const cacheCapacity = 2
const ttl = time.Second * 10

func init() {
	cache, err := NewCache[int, string]("lru", cacheCapacity)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(cache)
}

func computeExpensiveOperation(key int) string {
	// Simulation of an expensive operation
	time.Sleep(time.Millisecond * 200) // 200 мс
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
	cache.Put(key, result, ttl)

	w.Header().Set("X-Cache", "MISS")
	json.NewEncoder(w).Encode(result)
}

func main() {
	http.HandleFunc("/compute", getHandler)
	http.ListenAndServe(":8080", nil)
}


```
Use the `curl` command to save the `key` value to the cache:
```bash 
Example:
curl "http://localhost:8080/compute?key=1"
```