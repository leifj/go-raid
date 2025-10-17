package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/leifj/go-raid/internal/models"
	"github.com/leifj/go-raid/internal/storage"
)

// RAiDHandler handles RAiD-related HTTP requests
type RAiDHandler struct {
	storage storage.Repository
}

// NewRAiDHandler creates a new RAiD handler
func NewRAiDHandler(repo storage.Repository) *RAiDHandler {
	return &RAiDHandler{
		storage: repo,
	}
}

// MintRAiD handles POST /raid/ - creates a new RAiD
func (h *RAiDHandler) MintRAiD(w http.ResponseWriter, r *http.Request) {
	var req models.RAiD
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create RAiD using storage
	raid, err := h.storage.CreateRAiD(r.Context(), &req)
	if err != nil {
		if err == storage.ErrAlreadyExists {
			http.Error(w, "RAiD already exists", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(raid)
}

// FindAllRAiDs handles GET /raid/ - lists all RAiDs
func (h *RAiDHandler) FindAllRAiDs(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	filter := &storage.RAiDFilter{
		ContributorID:  r.URL.Query().Get("contributor.id"),
		OrganisationID: r.URL.Query().Get("organisation.id"),
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		filter.Limit, _ = strconv.Atoi(limit)
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		filter.Offset, _ = strconv.Atoi(offset)
	}

	// List RAiDs
	raids, err := h.storage.ListRAiDs(r.Context(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(raids)
}

// FindAllPublicRAiDs handles GET /raid/all-public - lists public RAiDs
func (h *RAiDHandler) FindAllPublicRAiDs(w http.ResponseWriter, r *http.Request) {
	filter := &storage.RAiDFilter{}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		filter.Limit, _ = strconv.Atoi(limit)
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		filter.Offset, _ = strconv.Atoi(offset)
	}

	raids, err := h.storage.ListPublicRAiDs(r.Context(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(raids)
}

// FindRAiDByName handles GET /raid/{prefix}/{suffix} - retrieves a specific RAiD
func (h *RAiDHandler) FindRAiDByName(w http.ResponseWriter, r *http.Request) {
	prefix := chi.URLParam(r, "prefix")
	suffix := chi.URLParam(r, "suffix")

	raid, err := h.storage.GetRAiD(r.Context(), prefix, suffix)
	if err != nil {
		if err == storage.ErrNotFound {
			http.Error(w, "RAiD not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(raid)
}

// UpdateRAiD handles PUT /raid/{prefix}/{suffix} - updates a RAiD
func (h *RAiDHandler) UpdateRAiD(w http.ResponseWriter, r *http.Request) {
	prefix := chi.URLParam(r, "prefix")
	suffix := chi.URLParam(r, "suffix")

	var req models.RAiD
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	raid, err := h.storage.UpdateRAiD(r.Context(), prefix, suffix, &req)
	if err != nil {
		if err == storage.ErrNotFound {
			http.Error(w, "RAiD not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(raid)
}

// PatchRAiD handles PATCH /raid/{prefix}/{suffix} - partially updates a RAiD
func (h *RAiDHandler) PatchRAiD(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement JSON Patch (RFC 6902) support
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "RAiD patch not yet implemented",
	})
}

// FindRAiDByNameAndVersion handles GET /raid/{prefix}/{suffix}/{version}
func (h *RAiDHandler) FindRAiDByNameAndVersion(w http.ResponseWriter, r *http.Request) {
	prefix := chi.URLParam(r, "prefix")
	suffix := chi.URLParam(r, "suffix")
	versionStr := chi.URLParam(r, "version")

	version, err := strconv.Atoi(versionStr)
	if err != nil {
		http.Error(w, "Invalid version number", http.StatusBadRequest)
		return
	}

	raid, err := h.storage.GetRAiDVersion(r.Context(), prefix, suffix, version)
	if err != nil {
		if err == storage.ErrNotFound {
			http.Error(w, "RAiD version not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(raid)
}

// RAiDHistory handles GET /raid/{prefix}/{suffix}/history - retrieves version history
func (h *RAiDHandler) RAiDHistory(w http.ResponseWriter, r *http.Request) {
	prefix := chi.URLParam(r, "prefix")
	suffix := chi.URLParam(r, "suffix")

	history, err := h.storage.GetRAiDHistory(r.Context(), prefix, suffix)
	if err != nil {
		if err == storage.ErrNotFound {
			http.Error(w, "RAiD not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}
