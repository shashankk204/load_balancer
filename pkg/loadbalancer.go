package core

import (
	"log"
	"net/http"
	"sync/atomic"
)

type LoadBalancer struct {
	Backends []*Backend
	Current  uint64
}

func Initialize_LB(backends []*Backend) *LoadBalancer{
	return &LoadBalancer{Backends: backends,Current: 0};
}


func (lb *LoadBalancer) getNextBackend() *Backend {
	next := atomic.AddUint64(&lb.Current, 1)
	return lb.Backends[int(next)%len(lb.Backends)]
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backend := lb.getNextBackend()
	log.Printf("Forwarding request to %s", backend.URL.String())
	backend.ReverseProxy.ServeHTTP(w, r)
}