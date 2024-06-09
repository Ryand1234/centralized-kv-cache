package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

const (
	baseURL     = "http://localhost:5000"
	totalReqs   = 10000 // Total number of requests
	concurrency = 500    // Number of concurrent requests
	setRatio    = 0.5   // Ratio of set requests (e.g., 0.5 means 50% set requests and 50% get requests)
)

// Function to make a single HTTP GET request
func makeRequest(client *http.Client, wg *sync.WaitGroup, key string) {
	defer wg.Done()
	resp, err := client.Get(fmt.Sprintf("%s/get?key=%s", baseURL, key))
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: Status code %d for GET %s\n", resp.StatusCode, key)
	}
}

// Function to make a single HTTP SET request
func setRequest(client *http.Client, wg *sync.WaitGroup, key, value string) {
	defer wg.Done()
	resp, err := client.Get(fmt.Sprintf("%s/set?key=%s&value=%s&duration=1h", baseURL, key, value))
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	resp.Body.Close()
}

func main() {
	startTime := time.Now()

	var wg sync.WaitGroup
	client := &http.Client{}

	// Channel to limit concurrency
	sem := make(chan struct{}, concurrency)

	// Slice to keep track of set keys
	setKeys := make([]string, 0)

	for i := 0; i < totalReqs; i++ {
		wg.Add(1)
		sem <- struct{}{}
		key := fmt.Sprintf("test%d", i)
		value := fmt.Sprintf("value%d", i)
		if float64(i) < setRatio*float64(totalReqs) {
			// Perform set request
			go func() {
				defer func() { <-sem }()
				setRequest(client, &wg, key, value)
				setKeys = append(setKeys, key)
			}()
		} else {
			// Perform get request on a randomly selected key from setKeys
			go func() {
				defer func() { <-sem }()
				rand.Seed(time.Now().UnixNano()) // Seed the random number generator
				randomIndex := rand.Intn(len(setKeys))
				makeRequest(client, &wg, setKeys[randomIndex])
			}()
		}
	}

	// Wait for all requests to complete
	wg.Wait()
	close(sem)

	duration := time.Since(startTime)
	fmt.Printf("Completed %d requests in %v\n", totalReqs, duration)
	fmt.Printf("Requests per second: %.2f\n", float64(totalReqs)/duration.Seconds())
}

