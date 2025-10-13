package core

import (
	"log"
	"net/http/httputil"
	"net/url"
	// "sync"
	"sync/atomic"
)

type Backend struct {
	URL          *url.URL
	Alive        int32
	// mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
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




