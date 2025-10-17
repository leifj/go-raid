package storage
package storage

import (
	"context"
	"errors"

	"github.com/leifj/go-raid/internal/models"
)

var (
	// ErrNotFound is returned when a resource is not found
	ErrNotFound = errors.New("resource not found")
	// ErrAlreadyExists is returned when attempting to create a resource that already exists
	ErrAlreadyExists = errors.New("resource already exists")
	// ErrInvalidVersion is returned when version mismatch occurs
	ErrInvalidVersion = errors.New("invalid version")
	// ErrAccessDenied is returned when access is denied
	ErrAccessDenied = errors.New("access denied")
)

// RAiDRepository defines operations for RAiD persistence
type RAiDRepository interface {
	// CreateRAiD mints a new RAiD with a unique identifier
	CreateRAiD(ctx context.Context, raid *models.RAiD) (*models.RAiD, error)

	// GetRAiD retrieves a RAiD by its prefix and suffix
	GetRAiD(ctx context.Context, prefix, suffix string) (*models.RAiD, error)

	// GetRAiDVersion retrieves a specific version of a RAiD
	GetRAiDVersion(ctx context.Context, prefix, suffix string, version int) (*models.RAiD, error)

	// UpdateRAiD updates an existing RAiD (creates new version)
	UpdateRAiD(ctx context.Context, prefix, suffix string, raid *models.RAiD) (*models.RAiD, error)

	// ListRAiDs retrieves RAiDs with optional filters
	ListRAiDs(ctx context.Context, filter *RAiDFilter) ([]*models.RAiD, error)

	// ListPublicRAiDs retrieves only publicly accessible RAiDs
	ListPublicRAiDs(ctx context.Context, filter *RAiDFilter) ([]*models.RAiD, error)

	// GetRAiDHistory retrieves the version history of a RAiD
	GetRAiDHistory(ctx context.Context, prefix, suffix string) ([]*models.RAiD, error)

	// DeleteRAiD removes a RAiD (soft delete, keeps history)
	DeleteRAiD(ctx context.Context, prefix, suffix string) error

	// GenerateIdentifier generates a unique identifier for a new RAiD
	GenerateIdentifier(ctx context.Context, servicePointID int64) (prefix, suffix string, err error)
}

// ServicePointRepository defines operations for ServicePoint persistence
type ServicePointRepository interface {
	// CreateServicePoint creates a new service point
	CreateServicePoint(ctx context.Context, sp *models.ServicePoint) (*models.ServicePoint, error)

	// GetServicePoint retrieves a service point by ID
	GetServicePoint(ctx context.Context, id int64) (*models.ServicePoint, error)

	// UpdateServicePoint updates an existing service point
	UpdateServicePoint(ctx context.Context, id int64, sp *models.ServicePoint) (*models.ServicePoint, error)

	// ListServicePoints retrieves all service points
	ListServicePoints(ctx context.Context) ([]*models.ServicePoint, error)

	// DeleteServicePoint removes a service point
	DeleteServicePoint(ctx context.Context, id int64) error
}

// Repository combines all repository interfaces
type Repository interface {
	RAiDRepository
	ServicePointRepository

	// Close closes the storage backend connection
	Close() error

	// HealthCheck verifies the storage backend is accessible
	HealthCheck(ctx context.Context) error
}

// RAiDFilter contains filtering options for RAiD queries
type RAiDFilter struct {
	// ContributorID filters by contributor ORCID
	ContributorID string
	// OrganisationID filters by organisation ROR ID
	OrganisationID string
	// IncludeFields specifies which fields to return (nil = all fields)
	IncludeFields []string
	// Limit specifies maximum number of results
	Limit int
	// Offset specifies number of results to skip
	Offset int
}
