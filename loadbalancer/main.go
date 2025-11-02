package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
	"time"
)

type Backend struct {
	URL          *url.URL
	Healthy      atomic.Bool
	RequestCount uint64
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

	atomic.AddUint64(&backend.RequestCount, 1)
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

func (lb *LoadBalancer) StatusHandler(w http.ResponseWriter, r *http.Request) {
	status := make([]map[string]interface{}, 0)

	for _, backend := range lb.Backends {
		status = append(status, map[string]interface{}{
			"url":      backend.URL.String(),
			"healthy":  backend.Healthy.Load(),
			"requests": atomic.LoadUint64(&backend.RequestCount),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

const dashboardHTML = `
<!DOCTYPE html>
<html>
<head>
	<title>Load Balancer Dashboard</title>
	<meta http-equiv="refresh" content="5">
	<style>
		body { font-family: Arial, sans-serif; background: #f5f5f5; }
		table { border-collapse: collapse; margin: 20px auto; width: 80%; background: white; }
		th, td { padding: 10px 15px; border: 1px solid #ccc; text-align: center; }
		th { background: #222; color: white; }
		tr:nth-child(even) { background: #eee; }
		h2 { text-align: center; margin-top: 20px; }
	</style>
</head>
<body>
	<h2>Load Balancer Status</h2>
	<table>
		<tr><th>Backend URL</th><th>Healthy</th><th>Requests</th></tr>
		{{range .}}
			<tr>
				<td>{{.URL}}</td>
				<td>{{if .Healthy}}✅{{else}}❌{{end}}</td>
				<td>{{.Requests}}</td>
			</tr>
		{{end}}
	</table>
</body>
</html>
`

func (lb *LoadBalancer) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	data := []map[string]interface{}{}
	for _, backend := range lb.Backends {
		data = append(data, map[string]interface{}{
			"URL":      backend.URL.String(),
			"Healthy":  backend.Healthy.Load(),
			"Requests": atomic.LoadUint64(&backend.RequestCount),
		})
	}

	tmpl := template.Must(template.New("dashboard").Parse(dashboardHTML))
	tmpl.Execute(w, data)
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
	http.HandleFunc("/status", lb.StatusHandler)
	http.HandleFunc("/dashboard", lb.DashboardHandler)

	log.Println("Load balancer running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
