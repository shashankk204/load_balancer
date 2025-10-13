package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/posts", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from POSTS instance 2!")
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	fmt.Println("Posts backend instance 2 running on :8092")
	http.ListenAndServe(":8092", nil)
}
