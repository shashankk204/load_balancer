package main

import (
	"fmt"
	"log"
	"net/http"

	core "github.com/shashankk204/load_balancer/pkg"
)

func main() {

	lb:=core.Initialize_LB()



	lb.AddRoute("/api/users", []string{
		"http://www.google.com",
		"http://www.facebook.com",
	})
	lb.AddRoute("/api/posts", []string{
		"http://www.amazon.com",
		"http://www.bing.com",
	})

	


	fmt.Println("Load Balancer started at :8080")
	if err := http.ListenAndServe(":8080", lb); err != nil {
		log.Fatal(err)
	}
}
