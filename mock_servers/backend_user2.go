package main

import (
	"fmt"
	"net/http"
	 
)

func main() {
	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		
		fmt.Fprintf(w, "Hello from USERS instance 2!")
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	fmt.Println("Users backend instance 2 running on :8082")
	http.ListenAndServe(":8082", nil)
}
