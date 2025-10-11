package main

import (
	"fmt"
	"log"
	"net/http"

	core "github.com/shashankk204/load_balancer/pkg"
)

func main() {
	mux:=http.NewServeMux()
	backends := []*core.Backend{
		core.NewBackend("https://localhost:9001") ,
		core.NewBackend("https://localhost:9002"),
		core.NewBackend("https://localhost:9002"),
	}

	lb := core.Initialize_LB(backends)

	mux.HandleFunc("/hmm",lb.ServeHTTP)


	fmt.Println("Load Balancer started at :8080")
	if err := http.ListenAndServe(":8080", lb); err != nil {
		log.Fatal(err)
	}
}
