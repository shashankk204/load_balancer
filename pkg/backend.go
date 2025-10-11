package core

import (
	"log"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

type Backend struct {
	URL          *url.URL
	Alive        bool
	ReverseProxy *httputil.ReverseProxy
}



func NewBackend(rawURL string) *Backend {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		log.Fatalf("Invalid backend URL: %v", err)
	}
	return &Backend{
		URL:          parsedURL,
		Alive:        true,
		ReverseProxy: httputil.NewSingleHostReverseProxy(parsedURL),
	}
}



type BackendPool struct {
	Backends []*Backend
	Current  int64
}



func (BP *BackendPool) GetNextBackend() *Backend {
	next:=atomic.AddInt64(&BP.Current,1);
	return BP.Backends[int(next)%len(BP.Backends)]
}



