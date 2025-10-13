package core

import (
	"log"
	"net/http"
	"strings"
	"time"
)



type LoadBalancer struct {
	Routes map[string]*BackendPool
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

func (lb *LoadBalancer) AddRoute(prefix string, urls []string) {
	pool := NewRoute(urls)
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
			proxy := target.ReverseProxy
			proxy.ServeHTTP(w, req)
			duration := time.Since(start)
			log.Printf("[%s] %s -> %s in %v", req.Method, req.URL.Path, target.URL, duration)
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