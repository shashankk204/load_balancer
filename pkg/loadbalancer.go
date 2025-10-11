package core

import (
	"net/http"
	"strings"
)

// type LoadBalancer struct {
// 	Backends []*Backend
// 	// Current  uint64
// 	BackendPool []*BackendPool
// }

// func (lb *LoadBalancer) getNextBackend() *Backend {
// 	next := atomic.AddUint64(&lb.Current, 1)
// 	return lb.Backends[int(next)%len(lb.Backends)]
// }

// func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	fmt.Printf("%s\n",r.URL.Path)
// 	backend := lb.getNextBackend()
// 	log.Printf("Forwarding request to %s", backend.URL.String())
// 	backend.ReverseProxy.ServeHTTP(w, r)
// }

type LoadBalancer struct {
	BackendList []*Backend
	Routes      map[string]*BackendPool
}

func Initialize_LB() *LoadBalancer {
	return &LoadBalancer{BackendList: []*Backend{}, Routes: make(map[string]*BackendPool)}
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
	lb.BackendList = append(lb.BackendList, pool.Backends...)
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for prefix, BP := range lb.Routes {
		if strings.HasPrefix(req.URL.Path, prefix) {
			target := BP.GetNextBackend()
			proxy := target.ReverseProxy
			proxy.ServeHTTP(w, req)
			return
		}
	}
	http.Error(w, "No backend found for route", http.StatusNotFound)
}
