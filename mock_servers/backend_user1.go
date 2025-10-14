package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2000 * time.Millisecond)
		fmt.Fprintf(w, "Hello from USERS instance 1!")
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {

		w.WriteHeader(http.StatusOK)
	})

	fmt.Println("Users backend instance 1 running on :8081")
	http.ListenAndServe(":8081", nil)
}
