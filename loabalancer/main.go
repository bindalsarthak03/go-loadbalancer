package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

var backendPool = []string{"http://localhost:8081", "http://localhost:8082", "http://localhost:8083"}

var counter uint32

func getNextServer() string {
	i := atomic.AddUint32(&counter, 1)
	return backendPool[(int(i)-1)%(len(backendPool))]
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		target := getNextServer()
		backendUrl, _ := url.Parse(target)
		log.Println("Backend URL:", backendUrl)
		proxy := httputil.NewSingleHostReverseProxy(backendUrl)
		proxy.ServeHTTP(w, r)
	})

	log.Println("Load balancer on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
