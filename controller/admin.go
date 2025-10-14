package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	core "github.com/shashankk204/load_balancer/pkg"
)


type AdminHandler struct {
	LB *core.LoadBalancer
}

func (a *AdminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/admin/add-backend":
		a.handleAddBackend(w, r)
	case "/admin/remove-backend":
		a.handleRemoveBackend(w, r)
	default:
		http.Error(w, "Unknown admin endpoint", http.StatusNotFound)
	}
}

func (a *AdminHandler) handleAddBackend(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Prefix string `json:"prefix"`
		URL    string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if err := a.LB.AddBackendToRoute(req.Prefix, req.URL); err != nil {
		http.Error(w, "Failed to add backend", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Backend %s added to route %s\n", req.URL, req.Prefix)
}

func (a *AdminHandler) handleRemoveBackend(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Prefix string `json:"prefix"`
		URL    string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	a.LB.RemoveBackendFromRoute(req.Prefix, req.URL)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Backend %s removed from route %s\n", req.URL, req.Prefix)
}
