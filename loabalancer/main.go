package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

type Backend struct {
	URL   *url.URL
	Alive bool
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
	i := atomic.AddUint64(&lb.current, 1)
	backend := lb.Backends[i%uint64(len(lb.Backends))]
	return backend
}

func (lb *LoadBalancer) ServeHTTPCustom(w http.ResponseWriter, r *http.Request) {
	backend := lb.getNextServer()
	if backend == nil {
		http.Error(w, "No backend available", http.StatusServiceUnavailable)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(backend.URL)
	proxy.ServeHTTP(w, r)
}

func main() {
	backends := []*Backend{
		{URL: mustParse("http://localhost:8081"), Alive: true},
		{URL: mustParse("http://localhost:8082"), Alive: false},
		{URL: mustParse("http://localhost:8083"), Alive: true},
	}

	lb := &LoadBalancer{Backends: backends}
	fmt.Print(lb)
	http.HandleFunc("/", lb.ServeHTTPCustom)

	log.Println("Load balancer running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
