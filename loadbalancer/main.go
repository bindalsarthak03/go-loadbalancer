package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
	"time"
)

type Backend struct {
	URL     *url.URL
	Healthy atomic.Bool
}

type LoadBalancer struct {
	Backends []*Backend
	current  uint64
}

func mustParse(rawURL string) *url.URL {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		log.Fatalf("Invalid URL %s: %v", rawURL, err)
	}
	return parsed
}

func (lb *LoadBalancer) getNextServer() *Backend {
	for i := 0; i < len(lb.Backends); i++ {
		index := atomic.AddUint64(&lb.current, 1) % uint64(len(lb.Backends))
		backend := lb.Backends[index]
		if backend.Healthy.Load() {
			return backend
		}
	}
	return nil
}

func (lb *LoadBalancer) ServeHTTPCustom(w http.ResponseWriter, r *http.Request) {
	backend := lb.getNextServer()
	if backend == nil {
		http.Error(w, "No backend available", http.StatusServiceUnavailable)
		return
	}
	log.Printf("Routing request to backend: %s\n", backend.URL)
	proxy := httputil.NewSingleHostReverseProxy(backend.URL)
	proxy.ServeHTTP(w, r)
}

func healthCheck(lb *LoadBalancer) {
	for {
		fmt.Println("Checking backends health...")
		for _, backend := range lb.Backends {
			res, err := http.Get(backend.URL.String() + "/health")
			if err != nil || res.StatusCode != http.StatusOK {
				backend.Healthy.Store(false)
				log.Printf("Backend %s is not healthy: %v", backend.URL.String(), err)
			} else {
				backend.Healthy.Store(true)
				log.Printf("Backend %s is healthy", backend.URL.String())
			}
			if res != nil {
				res.Body.Close()
			}
		}
		time.Sleep(10 * time.Second)
	}
}

func main() {
	backends := []*Backend{
		{URL: mustParse("http://localhost:8081")},
		{URL: mustParse("http://localhost:8082")},
		{URL: mustParse("http://localhost:8083")},
	}
	lb := &LoadBalancer{Backends: backends}

	// initially store true for all the servers
	for _, backend := range lb.Backends {
		backend.Healthy.Store(true)
	}
	// health check of servers every 10sec
	go healthCheck(lb)

	//start loadbalancer
	http.HandleFunc("/", lb.ServeHTTPCustom)

	log.Println("Load balancer running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
