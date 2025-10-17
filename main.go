package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/leifj/go-raid/internal/config"
	"github.com/leifj/go-raid/internal/handlers"
	"github.com/leifj/go-raid/internal/storage"

	// Import storage implementations to register factories
	_ "github.com/leifj/go-raid/internal/storage/cockroach"
	_ "github.com/leifj/go-raid/internal/storage/fdb"
	_ "github.com/leifj/go-raid/internal/storage/file"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize storage
	repo, err := storage.NewRepository(&cfg.Storage)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer repo.Close()

	// Health check storage
	if err := repo.HealthCheck(nil); err != nil {
		log.Printf("Warning: Storage health check failed: %v", err)
	} else {
		log.Printf("Storage (%s) initialized successfully", cfg.Storage.Type)
	}

	// Create router
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Initialize handlers with storage
	raidHandler := handlers.NewRAiDHandler(repo)
	spHandler := handlers.NewServicePointHandler(repo)

	// Setup routes
	setupRoutes(r, raidHandler, spHandler)

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting go-RAiD server on %s", addr)
	log.Printf("API endpoints available at http://%s/raid/", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func setupRoutes(r chi.Router, raidHandler *handlers.RAiDHandler, spHandler *handlers.ServicePointHandler) {
	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// RAiD endpoints
	r.Route("/raid", func(r chi.Router) {
		r.Post("/", raidHandler.MintRAiD)
		r.Get("/", raidHandler.FindAllRAiDs)
		r.Get("/all-public", raidHandler.FindAllPublicRAiDs)

		r.Route("/{prefix}/{suffix}", func(r chi.Router) {
			r.Get("/", raidHandler.FindRAiDByName)
			r.Put("/", raidHandler.UpdateRAiD)
			r.Patch("/", raidHandler.PatchRAiD)
			r.Get("/history", raidHandler.RAiDHistory)
			r.Get("/{version}", raidHandler.FindRAiDByNameAndVersion)
		})
	})

	// Service Point endpoints
	r.Route("/service-point", func(r chi.Router) {
		r.Post("/", spHandler.CreateServicePoint)
		r.Get("/", spHandler.FindAllServicePoints)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", spHandler.FindServicePointByID)
			r.Put("/", spHandler.UpdateServicePoint)
		})
	})
}
