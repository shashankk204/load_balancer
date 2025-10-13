package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/posts", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from POSTS instance 1!")
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	fmt.Println("Posts backend instance 1 running on :8091")
	http.ListenAndServe(":8091", nil)
}
