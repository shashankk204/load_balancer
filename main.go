package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/shashankk204/load_balancer/middleware"
	core "github.com/shashankk204/load_balancer/pkg"

	controller "github.com/shashankk204/load_balancer/controller"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type RouteConfig struct {
	Prefix   string   `json:"prefix"`
	Backends []string `json:"backends"`
	Strategy string   `json:"strategy,omitempty"`
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
	core.InitMetrics();
	lb:=core.Initialize_LB()
	rl := middleware.NewRateLimiter(5,10)
	cfg, err := LoadConfig("routes.json")
	if err != nil {
		log.Fatal(err)
	}


	for _, r := range cfg.Routes {
		strategy := core.ParseStrategy(r.Strategy)
		lb.AddRoute(r.Prefix, r.Backends,strategy)
	}

	go lb.StartHealthChecks(5*time.Second, "/health")
 
	adminHandler := &controller.AdminHandler{LB: lb}

	mux := http.NewServeMux()
	mux.Handle("/", middleware.RateLimitMiddleware(rl, lb))               
	mux.Handle("/admin/", adminHandler)   
	mux.Handle("/metrics", promhttp.Handler())
	// mux.HandleFunc("/metrics2", lb.MetricsHandler)



	fmt.Println("Load Balancer started at :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}