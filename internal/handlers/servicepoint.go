package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/leifj/go-raid/internal/models"
	"github.com/leifj/go-raid/internal/storage"
)

// ServicePointHandler handles service point-related HTTP requests
type ServicePointHandler struct {
	storage storage.Repository
}

// NewServicePointHandler creates a new service point handler
func NewServicePointHandler(repo storage.Repository) *ServicePointHandler {
	return &ServicePointHandler{
		storage: repo,
	}
}

// CreateServicePoint handles POST /service-point/
func (h *ServicePointHandler) CreateServicePoint(w http.ResponseWriter, r *http.Request) {
	var req models.ServicePoint
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	sp, err := h.storage.CreateServicePoint(r.Context(), &req)
	if err != nil {
		if err == storage.ErrAlreadyExists {
			http.Error(w, "Service point already exists", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sp)
}

// FindAllServicePoints handles GET /service-point/
func (h *ServicePointHandler) FindAllServicePoints(w http.ResponseWriter, r *http.Request) {
	servicePoints, err := h.storage.ListServicePoints(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(servicePoints)
}

// FindServicePointByID handles GET /service-point/{id}
func (h *ServicePointHandler) FindServicePointByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid service point ID", http.StatusBadRequest)
		return
	}

	sp, err := h.storage.GetServicePoint(r.Context(), id)
	if err != nil {
		if err == storage.ErrNotFound {
			http.Error(w, "Service point not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sp)
}

// UpdateServicePoint handles PUT /service-point/{id}
func (h *ServicePointHandler) UpdateServicePoint(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid service point ID", http.StatusBadRequest)
		return
	}

	var req models.ServicePoint
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	sp, err := h.storage.UpdateServicePoint(r.Context(), id, &req)
	if err != nil {
		if err == storage.ErrNotFound {
			http.Error(w, "Service point not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sp)
}
