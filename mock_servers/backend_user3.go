package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		fmt.Fprintf(w, "Hello from USERS instance 3!")
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	fmt.Println("Users backend instance 3 running on :8083")
	http.ListenAndServe(":8083", nil)
}
