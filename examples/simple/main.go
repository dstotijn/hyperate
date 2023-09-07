package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/dstotijn/hyperate"
	"golang.org/x/time/rate"
)

var url = flag.String("url", "https://example.com", "The URL to fetch.")

func main() {
	flag.Parse()

	// Create an http.RoundTripper that rate limits outgoing HTTP requests to 10
	// per second, allowing bursts of 5 requests.
	rt := hyperate.New(http.DefaultTransport, rate.NewLimiter(10, 5))
	client := &http.Client{Transport: rt}

	wg := &sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			resp, err := client.Get(fmt.Sprintf("%v?n=%v", *url, i+1))
			if err != nil {
				log.Printf("HTTP request failed: %v", err)
			}
			defer resp.Body.Close()
		}(i)
	}

	wg.Wait()
}
