package core

import (
	"context"
	"fmt"
	"log"
	"maps"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	// core "github.com/shashankk204/load_balancer/pkg"
	"github.com/shashankk204/load_balancer/pkg/logger"
	
)

//TODO:
// Add a timeout or context propagation to cancel stuck proxy calls.
// Handle panics in backends gracefully:

type LoadBalancer struct {
	Routes map[string]*BackendPool
	mux    sync.RWMutex
}

func Initialize_LB() *LoadBalancer {
	return &LoadBalancer{
		Routes:      make(map[string]*BackendPool),
	}
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


// This is a Response Writer Wrapper pattern used to intercept and capture HTTP response details that are normally not accessible after the response is sent.
type responseWriterWrapper struct {
    http.ResponseWriter
    statusCode   int
    responseSize int64
}

func (w *responseWriterWrapper) WriteHeader(statusCode int) {
    w.statusCode = statusCode
    w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriterWrapper) Write(data []byte) (int, error) {
    n, err := w.ResponseWriter.Write(data)
    w.responseSize += int64(n)
    return n, err
}


func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	start := time.Now()
	ctx := logger.WithRequestID(req.Context())

	lb.mux.RLock()  
    defer lb.mux.RUnlock()
	
	for prefix, BP := range lb.Routes {
		if strings.HasPrefix(req.URL.Path, prefix) {  //TODO:need to implement Trie Data structre for Prefix matching


			RouteActiveRequests.WithLabelValues(prefix).Inc()
            defer RouteActiveRequests.WithLabelValues(prefix).Dec()
			


			if req.ContentLength > 0 {
                RouteRequestSize.WithLabelValues(prefix).Observe(float64(req.ContentLength))
            }


			target := BP.GetNextBackend()
			if target==nil{
				RouteErrorsTotal.WithLabelValues(prefix, "no_backend_available").Inc()
				logger.Error(ctx, "No backend available", map[string]string{
					"method": req.Method,
					"path":   req.URL.Path,
				})
				http.Error(w,"No Backend availabe",http.StatusServiceUnavailable)
				return
			}

			target.IncActive()
			BackendActiveConnections.WithLabelValues(prefix, target.URL.String(), target.URL.Host).Inc()
			defer BackendActiveConnections.WithLabelValues(prefix, target.URL.String(), target.URL.Host).Dec()
			defer target.DecActive()

			BackendSelectionTotal.WithLabelValues(prefix, target.URL.String(), target.URL.Host, string(BP.Strategy)).Inc()


			 responseWrapper := &responseWriterWrapper{
                ResponseWriter: w,
                statusCode:     http.StatusOK,
                responseSize:   0,
            }
            

			proxy := target.ReverseProxy
			proxy.ServeHTTP(responseWrapper, req)


			duration := time.Since(start)
			target.RecordRequest(duration)

			statusCode := responseWrapper.statusCode
            responseSize := responseWrapper.responseSize

			lb.updateBackendMetrics(prefix, target, duration, statusCode)

			RouteRequestsTotal.WithLabelValues(prefix, req.Method, strconv.Itoa(statusCode)).Inc()
            RouteRequestDuration.WithLabelValues(prefix, req.Method).Observe(duration.Seconds())
            
			if responseSize > 0 {
                RouteResponseSize.WithLabelValues(prefix).Observe(float64(responseSize))
            }

			if statusCode >= 400 {
                errorType := "client_error"
                if statusCode >= 500 {
                    errorType = "server_error"
                }
                RouteErrorsTotal.WithLabelValues(prefix, errorType).Inc()
            }
			
			logger.Info(ctx, "Routing request", map[string]string{
				"method":   req.Method,
				"path":     req.URL.Path,
				"target":   target.URL.String(),
				"duration": duration.String(),
			})
			log.Printf("[%s] %s -> %s [strategy=%s] in %v (avg %.2f ms, active %d)", req.Method, req.URL.Path, target.URL, BP.Strategy, duration, target.AvgLatency(), target.ActiveRequests())

			return
		}
	}
	RouteErrorsTotal.WithLabelValues("unknown", "route_not_found").Inc()
    http.Error(w, "No backend found for route", http.StatusNotFound)
}

func (lb *LoadBalancer) updateBackendMetrics(routePrefix string, backend *Backend, duration time.Duration, statusCode int) {
    backendURL := backend.URL.String()
    backendHost := backend.URL.Host
    
    // Update backend request metrics
    BackendRequestsTotal.WithLabelValues(routePrefix, backendURL, backendHost, strconv.Itoa(statusCode)).Inc()
    BackendRequestDuration.WithLabelValues(routePrefix, backendURL, backendHost).Observe(duration.Seconds())
    

    
    // Update load score (active requests + normalized latency)
    active := float64(backend.ActiveRequests())
    latency := backend.AvgLatency()
    loadScore := active + (latency / 100)
    BackendLoadScore.WithLabelValues(routePrefix, backendURL, backendHost).Set(loadScore)
    
    // Record backend failures if status code indicates failure
    if statusCode >= 500 {
        failureType := "server_error"
        BackendFailuresTotal.WithLabelValues(routePrefix, backendURL, backendHost, failureType).Inc()
    }
}

func (lb *LoadBalancer) updateBackendHealthMetrics(routePrefix string, backend *Backend, isAlive bool, healthCheckDuration time.Duration) {
    backendURL := backend.URL.String()
    backendHost := backend.URL.Host
    
    // Update health status
    if isAlive {
        BackendHealthStatus.WithLabelValues(routePrefix, backendURL, backendHost).Set(1)
    } else {
        BackendHealthStatus.WithLabelValues(routePrefix, backendURL, backendHost).Set(0)
        BackendHealthCheckFailures.WithLabelValues(routePrefix, backendURL, backendHost).Inc()
    }
    
    // Update health check duration
    BackendHealthCheckDuration.WithLabelValues(routePrefix, backendURL, backendHost).Observe(healthCheckDuration.Seconds())
}

func (lb *LoadBalancer) StartHealthChecks(interval time.Duration, healthPath string) {
	client := &http.Client{Timeout: 2 * time.Second}
	ticker := time.NewTicker(interval)

	go func() {
		for range ticker.C {
			lb.mux.RLock()
			routesSnapshot := make(map[string]*BackendPool, len(lb.Routes))
			maps.Copy(routesSnapshot, lb.Routes)
			lb.mux.RUnlock()
			for prefix, pool := range routesSnapshot {
				for _, b := range pool.Backends {
					go func(b *Backend, prefix string) {
						ctx := logger.WithRequestID(context.Background())
						healthURL := b.URL.String() + healthPath
						start := time.Now()
						resp, err := client.Get(healthURL)
						healthCheckDuration := time.Since(start)
						isAlive := err == nil && resp != nil && resp.StatusCode == http.StatusOK
						if resp != nil {
							resp.Body.Close()
						}
						if !isAlive && err != nil {
                            failureType := "connection_error"
                            BackendFailuresTotal.WithLabelValues(prefix, b.URL.String(), b.URL.Host, failureType).Inc()
                        }
						b.SetAlive(isAlive)
						lb.updateBackendHealthMetrics(prefix, b, isAlive, healthCheckDuration)
						status := "DOWN"
						if isAlive {
							status = "UP"
						}
						logger.Info(ctx, "Health check result", map[string]string{
							"path":   healthURL,
							"target": b.URL.String(),
							"method": "GET",
							"status": status,
						})
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



// func (lb *LoadBalancer) MetricsHandler(w http.ResponseWriter, r *http.Request) {
// 	type BackendMetrics struct {
// 		URL         string  `json:"url"`
// 		Alive       bool    `json:"alive"`
// 		Requests    int64   `json:"total_requests"`
// 		AvgLatency  float64 `json:"avg_latency_ms"`
// 		Active      int64   `json:"active_requests"`
// 	}

// 	out := map[string][]BackendMetrics{}

// 	lb.mux.RLock()
// 	defer lb.mux.RUnlock()

// 	for prefix, pool := range lb.Routes {
// 		var metrics []BackendMetrics
// 		for _, b := range pool.Backends {
// 			metrics = append(metrics, BackendMetrics{
// 				URL:        b.URL.String(),
// 				Alive:      b.IsAlive(),
// 				Requests:   atomic.LoadInt64(&b.TotalRequests),
// 				AvgLatency: b.AvgLatency(),
// 				Active:     atomic.LoadInt64(&b.Active),
// 			})
// 		}
// 		out[prefix] = metrics
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(out)
// }




func (lb *LoadBalancer) GetRoutesInfo() []map[string]interface{} {
	var result []map[string]interface{}
	lb.mux.RLock()
    defer lb.mux.RUnlock()
	for prefix, pool := range lb.Routes {
		var backends []string
		for _, b := range pool.Backends {
			backends = append(backends, b.URL.String())
		}
		result = append(result, map[string]interface{}{
			"prefix":   prefix,
			"strategy": pool.Strategy,
			"backends": backends,
		})
	}
	return result
}

func (lb *LoadBalancer) UpdateRoute(prefix string, backends []string, strategy string) error {
	lb.mux.Lock()
    defer lb.mux.Unlock()
	pool, ok := lb.Routes[prefix]
	if !ok {
		return fmt.Errorf("route not found: %s", prefix)
	}

	if len(backends) > 0 {
		newPool := NewRoute(backends)
		pool.Backends = newPool.Backends
	}
	if strategy!=""{
		oldStrategy := string(pool.Strategy)
        newStrategy := ParseStrategy(strategy)
		if oldStrategy != string(newStrategy) {
            RouteStrategyChanges.WithLabelValues(prefix, oldStrategy, string(newStrategy)).Inc()
            pool.Strategy = newStrategy
        }
	}
	

	lb.Routes[prefix] = pool
	return nil
}