package main

import (
	"fmt"
	"go-loadbalancer/common"
	"log"
	"net/http"
)

func server1() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello World from :8081")
	})
	mux.HandleFunc("/health", common.HealthCheckerHandler)
	log.Println("Backend 1 on :8081")
	log.Fatal(http.ListenAndServe(":8081", mux))
}

func server2() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello World from :8082")
	})
	mux.HandleFunc("/health", common.HealthCheckerHandler)
	log.Println("Backend 2 on :8082")
	log.Fatal(http.ListenAndServe(":8082", mux))
}

func main() {
	// Run two servers concurrently
	go server1()
	go server2()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello World from :8083")
	})
	mux.HandleFunc("/health", common.HealthCheckerHandler)
	log.Println("Backend 3 on :8083")
	log.Fatal(http.ListenAndServe(":8083", mux))
}
