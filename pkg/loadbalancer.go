package core

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)



type LoadBalancer struct {
	Routes map[string]*BackendPool
	mux    sync.RWMutex
}

func Initialize_LB() *LoadBalancer {
	return &LoadBalancer{
		Routes:      make(map[string]*BackendPool)}
}

func NewRoute(urls []string) *BackendPool {
	backends := make([]*Backend, 0, len(urls))
	for _, u := range urls {
		backends = append(backends, NewBackend(u))
	}
	return &BackendPool{Backends: backends}
}

func (lb *LoadBalancer) AddRoute(prefix string, urls []string,strategy Strategy) {
	pool := NewRoute(urls)
	pool.Strategy = strategy
	lb.Routes[prefix] = pool
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	start := time.Now()
	for prefix, BP := range lb.Routes {
		if strings.HasPrefix(req.URL.Path, prefix) {
			target := BP.GetNextBackend()
			if target==nil{
				http.Error(w,"No Backend availabe",http.StatusServiceUnavailable)
				return
			}

			target.IncActive()
			proxy := target.ReverseProxy
			proxy.ServeHTTP(w, req)
			duration := time.Since(start)
			target.DecActive()
			target.RecordRequest(duration)
			
			log.Printf("[%s] %s -> %s [strategy=%s] in %v (avg %.2f ms, active %d)",
	req.Method, req.URL.Path, target.URL, BP.Strategy, duration, target.AvgLatency(), target.ActiveRequests())

			return
		}
	}
	http.Error(w, "No backend found for route", http.StatusNotFound)
}



func (lb *LoadBalancer) StartHealthChecks(interval time.Duration, healthPath string) {
	client := &http.Client{Timeout: 2 * time.Second}
	ticker := time.NewTicker(interval)

	go func() {
		for range ticker.C {
			for prefix, pool := range lb.Routes {
				for _, b := range pool.Backends {
					go func(b *Backend, prefix string) {
						healthURL := b.URL.String() + healthPath
						resp, err := client.Get(healthURL)
						isAlive := err == nil && resp != nil && resp.StatusCode == http.StatusOK
						if resp != nil {
							resp.Body.Close()
						}
						b.SetAlive(isAlive)
						status := "DOWN"
						if isAlive {
							status = "UP"
						}
						log.Printf("[%s] %s is %s", prefix, b.URL, status)
					}(b, prefix)
				}
			}
		}
	}()
}







func (lb *LoadBalancer) AddBackendToRoute(prefix string, backendURL string,strategy Strategy ) error {
	lb.mux.Lock()
	defer lb.mux.Unlock()

	pool, exists := lb.Routes[prefix]
	if !exists {
		pool = NewRoute([]string{backendURL})
		pool.Strategy=strategy;
		lb.Routes[prefix] = pool
		log.Printf("Created new route %s with backend %s", prefix, backendURL)
		return nil
	}

	newBackend := NewBackend(backendURL)
	pool.Backends = append(pool.Backends, newBackend)
	log.Printf("Added new backend %s to route %s", backendURL, prefix)
	return nil
}

func (lb *LoadBalancer) RemoveBackendFromRoute(prefix string, backendURL string) {
	lb.mux.Lock()
	defer lb.mux.Unlock()

	pool, exists := lb.Routes[prefix]
	if !exists {
		log.Printf("Route %s does not exist", prefix)
		return
	}

	filtered := []*Backend{}
	for _, b := range pool.Backends {
		if b.URL.String() != backendURL {
			filtered = append(filtered, b)
		}
	}

	pool.Backends = filtered
	log.Printf("Removed backend %s from route %s", backendURL, prefix)
}



func (lb *LoadBalancer) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	type BackendMetrics struct {
		URL         string  `json:"url"`
		Alive       bool    `json:"alive"`
		Requests    int64   `json:"total_requests"`
		AvgLatency  float64 `json:"avg_latency_ms"`
		Active      int64   `json:"active_requests"`
	}

	out := map[string][]BackendMetrics{}

	lb.mux.RLock()
	defer lb.mux.RUnlock()

	for prefix, pool := range lb.Routes {
		var metrics []BackendMetrics
		for _, b := range pool.Backends {
			metrics = append(metrics, BackendMetrics{
				URL:        b.URL.String(),
				Alive:      b.IsAlive(),
				Requests:   atomic.LoadInt64(&b.TotalRequests),
				AvgLatency: b.AvgLatency(),
				Active:     atomic.LoadInt64(&b.Active),
			})
		}
		out[prefix] = metrics
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}





