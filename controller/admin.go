package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	core "github.com/shashankk204/load_balancer/pkg"
	"github.com/shashankk204/load_balancer/utils"
)


type AdminHandler struct {
	LB *core.LoadBalancer
}

func (a *AdminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/admin/add-backend":
		a.handleAddBackend(w, r)
	case r.Method == http.MethodPost && r.URL.Path == "/admin/remove-backend":
		a.handleRemoveBackend(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/admin/list":
		a.handleListRoutes(w, r)
	case r.Method == http.MethodPut && r.URL.Path == "/admin/update":
		a.handleUpdateRoute(w, r)
	default:
		http.Error(w, "Unknown or unsupported admin endpoint", http.StatusNotFound)
	}
}

func (a *AdminHandler) handleAddBackend(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Prefix string `json:"prefix"`
		URL    string `json:"url"`
		Strategy string   `json:"strategy,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	strategy := core.ParseStrategy(req.Strategy)
	if err := a.LB.AddBackendToRoute(req.Prefix, req.URL,strategy); err != nil {
		http.Error(w, "Failed to add backend", http.StatusInternalServerError)
		return
	}
	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"status":   "success",
		"action":   "add-backend",
		"prefix":   req.Prefix,
		"url":      req.URL,
		"strategy": strategy,
	})
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
	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"action": "remove-backend",
		"prefix": req.Prefix,
		"url":    req.URL,
	})
}


func (a *AdminHandler) handleListRoutes(w http.ResponseWriter, r *http.Request) {
	routes := a.LB.GetRoutesInfo()
	utils.RespondJSON(w, http.StatusOK, routes)
}

func (a *AdminHandler) handleUpdateRoute(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Prefix   string   `json:"prefix"`
		Backends []string `json:"backends,omitempty"`
		Strategy string   `json:"strategy,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := a.LB.UpdateRoute(req.Prefix, req.Backends, req.Strategy); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update route: %v", err), http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"status":   "success",
		"action":   "update-route",
		"prefix":   req.Prefix,
		"strategy": req.Strategy,
	})
}