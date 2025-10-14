package core

import (
	"log"
	"net/http/httputil"
	"net/url"
	"time"

	// "sync"
	"sync/atomic"
)

type Backend struct {
	URL          *url.URL
	Alive        int32
	ReverseProxy *httputil.ReverseProxy
	TotalRequests int64          
	TotalLatency  int64          
	Active        int64 //current number of active requests
}



func NewBackend(rawURL string) *Backend {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		log.Fatalf("Invalid backend URL: %v", err)
	}
	return &Backend{
		URL:          parsedURL,
		Alive:        1,
		
		ReverseProxy: httputil.NewSingleHostReverseProxy(parsedURL),
	}
}


func (b *Backend) SetAlive(alive bool) {
	var val int32
	if alive {
		val = 1
	}
	atomic.StoreInt32(&b.Alive, val)
}

func (b *Backend) IsAlive() bool {
	return atomic.LoadInt32(&b.Alive) == 1
}






type BackendPool struct {
	Backends []*Backend
	Current  int64
}



func (BP *BackendPool) GetNextBackend() *Backend {
	
	for i:=0;i<len(BP.Backends);i++{
		next:=atomic.AddInt64(&BP.Current,1)
		idx:=int(next)%len(BP.Backends)
		b:=BP.Backends[idx]
		if(b.IsAlive()){
			return b
		}
	}
	return nil;
}



func (p *BackendPool) SetBackendAlive(url *url.URL, alive bool) {
	for _, b := range p.Backends {
		if b.URL.String() == url.String() {
			b.SetAlive(alive)
		}
	}
}




// Record a request for this backend
func (b *Backend) RecordRequest(duration time.Duration) {
	atomic.AddInt64(&b.TotalRequests, 1)
	atomic.AddInt64(&b.TotalLatency, duration.Nanoseconds())
}

// Increment active requests
func (b *Backend) IncActive() {
	atomic.AddInt64(&b.Active, 1)
}

// Decrement active requests
func (b *Backend) DecActive() {
	atomic.AddInt64(&b.Active, -1)
}

// Get average latency in milliseconds
func (b *Backend) AvgLatency() float64 {
	reqs := atomic.LoadInt64(&b.TotalRequests)
	if reqs == 0 {
		return 0
	}
	totalNs := atomic.LoadInt64(&b.TotalLatency)
	return float64(totalNs)/1e6/float64(reqs) // convert ns → ms
}

// Get current number of active requests
func (b *Backend) ActiveRequests() int64 {
	return atomic.LoadInt64(&b.Active)
}