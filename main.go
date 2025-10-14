package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	core "github.com/shashankk204/load_balancer/pkg"

	controller "github.com/shashankk204/load_balancer/controller"
)

type RouteConfig struct {
	Prefix   string   `json:"prefix"`
	Backends []string `json:"backends"`
}

type Config struct {
	Routes []RouteConfig `json:"routes"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = json.Unmarshal(data, &cfg)
	return &cfg, err
}






func main() {
	cfg, err := LoadConfig("routes.json")
	if err != nil {
		log.Fatal(err)
	}

	lb:=core.Initialize_LB()

	for _, r := range cfg.Routes {
		lb.AddRoute(r.Prefix, r.Backends)
	}

	lb.StartHealthChecks(5*time.Second, "/health")
 
	adminHandler := &controller.AdminHandler{LB: lb}

	mux := http.NewServeMux()
	mux.Handle("/", lb)                   
	mux.Handle("/admin/", adminHandler)   
	mux.HandleFunc("/metrics", lb.MetricsHandler)

	fmt.Println("Load Balancer started at :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}