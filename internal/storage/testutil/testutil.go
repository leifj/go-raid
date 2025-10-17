package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/leifj/go-raid/internal/models"
	"github.com/leifj/go-raid/internal/storage"
)

// Simple mock repository for testing handlers
// This provides configurable behavior via function fields

type MockRepository struct {
	mu sync.RWMutex
	
	// RAiD operations
	CreateRAiDFunc         func(context.Context, *models.RAiD) (*models.RAiD, error)
	GetRAiDFunc            func(context.Context, string, string) (*models.RAiD, error)
	GetRAiDVersionFunc     func(context.Context, string, string, int) (*models.RAiD, error)
	UpdateRAiDFunc         func(context.Context, string, string, *models.RAiD) (*models.RAiD, error)
	ListRAiDsFunc          func(context.Context, *storage.RAiDFilter) ([]*models.RAiD, error)
	ListPublicRAiDsFunc    func(context.Context, *storage.RAiDFilter) ([]*models.RAiD, error)
	GetRAiDHistoryFunc     func(context.Context, string, string) ([]*models.RAiD, error)
	DeleteRAiDFunc         func(context.Context, string, string) error
	GenerateIdentifierFunc func(context.Context, int64) (string, string, error)
	
	// ServicePoint operations
	CreateServicePointFunc func(context.Context, *models.ServicePoint) (*models.ServicePoint, error)
	GetServicePointFunc    func(context.Context, int64) (*models.ServicePoint, error)
	UpdateServicePointFunc func(context.Context, int64, *models.ServicePoint) (*models.ServicePoint, error)
	ListServicePointsFunc  func(context.Context) ([]*models.ServicePoint, error)
	DeleteServicePointFunc func(context.Context, int64) error
	
	// Repository operations
	CloseFunc       func() error
	HealthCheckFunc func(context.Context) error
	
	// Call counters
	CreateRAiDCalls         int
	GetRAiDCalls            int
	UpdateRAiDCalls         int
	DeleteRAiDCalls         int
	ListRAiDsCalls          int
	GetRAiDHistoryCalls     int
	GenerateIdentifierCalls int
	
	CreateServicePointCalls int
	GetServicePointCalls    int
	UpdateServicePointCalls int
	ListServicePointsCalls  int
	DeleteServicePointCalls int
}

// NewMockRepository creates a new mock repository with default implementations
func NewMockRepository() *MockRepository {
	return &MockRepository{}
}

// RAiD operations

func (m *MockRepository) CreateRAiD(ctx context.Context, raid *models.RAiD) (*models.RAiD, error) {
	m.mu.Lock()
	m.CreateRAiDCalls++
	m.mu.Unlock()
	if m.CreateRAiDFunc != nil {
		return m.CreateRAiDFunc(ctx, raid)
	}
	return raid, nil
}

func (m *MockRepository) GetRAiD(ctx context.Context, prefix, suffix string) (*models.RAiD, error) {
	m.mu.Lock()
	m.GetRAiDCalls++
	m.mu.Unlock()
	if m.GetRAiDFunc != nil {
		return m.GetRAiDFunc(ctx, prefix, suffix)
	}
	return NewTestRAiD(prefix, suffix), nil
}

func (m *MockRepository) GetRAiDVersion(ctx context.Context, prefix, suffix string, version int) (*models.RAiD, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.GetRAiDVersionFunc != nil {
		return m.GetRAiDVersionFunc(ctx, prefix, suffix, version)
	}
	raid := NewTestRAiD(prefix, suffix)
	if raid.Identifier != nil {
		raid.Identifier.Version = version
	}
	return raid, nil
}

func (m *MockRepository) UpdateRAiD(ctx context.Context, prefix, suffix string, raid *models.RAiD) (*models.RAiD, error) {
	m.mu.Lock()
	m.UpdateRAiDCalls++
	m.mu.Unlock()
	if m.UpdateRAiDFunc != nil {
		return m.UpdateRAiDFunc(ctx, prefix, suffix, raid)
	}
	return raid, nil
}

func (m *MockRepository) ListRAiDs(ctx context.Context, filter *storage.RAiDFilter) ([]*models.RAiD, error) {
	m.mu.Lock()
	m.ListRAiDsCalls++
	m.mu.Unlock()
	if m.ListRAiDsFunc != nil {
		return m.ListRAiDsFunc(ctx, filter)
	}
	return []*models.RAiD{}, nil
}

func (m *MockRepository) ListPublicRAiDs(ctx context.Context, filter *storage.RAiDFilter) ([]*models.RAiD, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ListPublicRAiDsFunc != nil {
		return m.ListPublicRAiDsFunc(ctx, filter)
	}
	return []*models.RAiD{}, nil
}

func (m *MockRepository) GetRAiDHistory(ctx context.Context, prefix, suffix string) ([]*models.RAiD, error) {
	m.mu.Lock()
	m.GetRAiDHistoryCalls++
	m.mu.Unlock()
	if m.GetRAiDHistoryFunc != nil {
		return m.GetRAiDHistoryFunc(ctx, prefix, suffix)
	}
	return []*models.RAiD{}, nil
}

func (m *MockRepository) DeleteRAiD(ctx context.Context, prefix, suffix string) error {
	m.mu.Lock()
	m.DeleteRAiDCalls++
	m.mu.Unlock()
	if m.DeleteRAiDFunc != nil {
		return m.DeleteRAiDFunc(ctx, prefix, suffix)
	}
	return nil
}

func (m *MockRepository) GenerateIdentifier(ctx context.Context, servicePointID int64) (string, string, error) {
	m.mu.Lock()
	m.GenerateIdentifierCalls++
	m.mu.Unlock()
	if m.GenerateIdentifierFunc != nil {
		return m.GenerateIdentifierFunc(ctx, servicePointID)
	}
	return "10.12345", fmt.Sprintf("%d", time.Now().UnixNano()), nil
}

// ServicePoint operations

func (m *MockRepository) CreateServicePoint(ctx context.Context, sp *models.ServicePoint) (*models.ServicePoint, error) {
	m.mu.Lock()
	m.CreateServicePointCalls++
	m.mu.Unlock()
	if m.CreateServicePointFunc != nil {
		return m.CreateServicePointFunc(ctx, sp)
	}
	return sp, nil
}

func (m *MockRepository) GetServicePoint(ctx context.Context, id int64) (*models.ServicePoint, error) {
	m.mu.Lock()
	m.GetServicePointCalls++
	m.mu.Unlock()
	if m.GetServicePointFunc != nil {
		return m.GetServicePointFunc(ctx, id)
	}
	return NewTestServicePoint(id), nil
}

func (m *MockRepository) UpdateServicePoint(ctx context.Context, id int64, sp *models.ServicePoint) (*models.ServicePoint, error) {
	m.mu.Lock()
	m.UpdateServicePointCalls++
	m.mu.Unlock()
	if m.UpdateServicePointFunc != nil {
		return m.UpdateServicePointFunc(ctx, id, sp)
	}
	return sp, nil
}

func (m *MockRepository) ListServicePoints(ctx context.Context) ([]*models.ServicePoint, error) {
	m.mu.Lock()
	m.ListServicePointsCalls++
	m.mu.Unlock()
	if m.ListServicePointsFunc != nil {
		return m.ListServicePointsFunc(ctx)
	}
	return []*models.ServicePoint{}, nil
}

func (m *MockRepository) DeleteServicePoint(ctx context.Context, id int64) error {
	m.mu.Lock()
	m.DeleteServicePointCalls++
	m.mu.Unlock()
	if m.DeleteServicePointFunc != nil {
		return m.DeleteServicePointFunc(ctx, id)
	}
	return nil
}

// Repository operations

func (m *MockRepository) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func (m *MockRepository) HealthCheck(ctx context.Context) error {
	if m.HealthCheckFunc != nil {
		return m.HealthCheckFunc(ctx)
	}
	return nil
}

// Ensure MockRepository implements storage.Repository
var _ storage.Repository = (*MockRepository)(nil)

// Test Data Fixtures

// NewTestRAiD creates a test RAiD with the given prefix and suffix
func NewTestRAiD(prefix, suffix string) *models.RAiD {
	now := time.Now()
	id := fmt.Sprintf("https://raid.org/%s/%s", prefix, suffix)
	
	return &models.RAiD{
		Identifier: &models.Identifier{
			ID:        id,
			SchemaURI: "https://raid.org/",
			RegistrationAgency: &models.RegistrationAgency{
				ID:        "https://ror.org/038sjwq14",
				SchemaURI: "https://ror.org/",
			},
			Owner: &models.Owner{
				ID:           "https://ror.org/0384j8v12",
				SchemaURI:    "https://ror.org/",
				ServicePoint: 1,
			},
			License: "https://creativecommons.org/licenses/by/4.0/",
			Version: 1,
		},
		Title: []models.Title{
			{
				Text: fmt.Sprintf("Test RAiD %s/%s", prefix, suffix),
				Type: &models.IDSchema{
					ID:        "https://vocabulary.raid.org/title.type.schema/318",
					SchemaURI: "https://vocabulary.raid.org/title.type.schema",
				},
				StartDate: now.Format("2006-01-02"),
				Language: &models.Language{
					ID:        "eng",
					SchemaURI: "https://www.iso.org/standard/39534.html",
				},
			},
		},
		Date: &models.Date{
			StartDate: now.Format("2006-01-02"),
		},
		Description: []models.Description{
			{
				Text: fmt.Sprintf("Test description for RAiD %s/%s", prefix, suffix),
				Type: &models.IDSchema{
					ID:        "https://vocabulary.raid.org/description.type.schema/318",
					SchemaURI: "https://vocabulary.raid.org/description.type.schema",
				},
			},
		},
		Access: &models.Access{
			Type: &models.IDSchema{
				ID:        "https://vocabulary.raid.org/access.type.schema/53",
				SchemaURI: "https://vocabulary.raid.org/access.type.schema",
			},
		},
	}
}

// NewTestServicePoint creates a test ServicePoint
func NewTestServicePoint(id int64) *models.ServicePoint {
	return &models.ServicePoint{
		ID:               id,
		Name:             fmt.Sprintf("Test Service Point %d", id),
		IdentifierOwner:  fmt.Sprintf("Owner %d", id),
		RepositoryID:     fmt.Sprintf("repo-%d", id),
		Prefix:           fmt.Sprintf("10.%d", 10000+id),
		GroupID:          fmt.Sprintf("group-%d", id),
		TechEmail:        fmt.Sprintf("tech%d@example.com", id),
		AdminEmail:       fmt.Sprintf("admin%d@example.com", id),
		Enabled:          true,
		AppWritesEnabled: true,
	}
}

// Test Helpers

// CreateTempDirectory creates a temporary directory for testing
func CreateTempDirectory(t *testing.T, prefix string) string {
	t.Helper()
	dir, err := os.MkdirTemp("", prefix)
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	return dir
}

// WriteTestFile writes data to a file in the test directory
func WriteTestFile(t *testing.T, dir, filename string, data []byte) string {
	t.Helper()
	path := filepath.Join(dir, filename)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	return path
}

// AssertRAiDEqual asserts that two RAiDs are equal (basic check)
func AssertRAiDEqual(t *testing.T, expected, actual *models.RAiD) {
	t.Helper()
	if expected == nil && actual == nil {
		return
	}
	if expected == nil || actual == nil {
		t.Fatalf("One RAiD is nil: expected=%v, actual=%v", expected, actual)
	}
	if expected.Identifier != nil && actual.Identifier != nil {
		if expected.Identifier.ID != actual.Identifier.ID {
			t.Errorf("Identifier mismatch: expected=%s, actual=%s", 
				expected.Identifier.ID, actual.Identifier.ID)
		}
	}
	// Compare title count
	if len(expected.Title) != len(actual.Title) {
		t.Errorf("Title count mismatch: expected=%d, actual=%d", 
			len(expected.Title), len(actual.Title))
	}
}

// AssertServicePointEqual asserts that two ServicePoints are equal
func AssertServicePointEqual(t *testing.T, expected, actual *models.ServicePoint) {
	t.Helper()
	if expected == nil && actual == nil {
		return
	}
	if expected == nil || actual == nil {
		t.Fatalf("One ServicePoint is nil: expected=%v, actual=%v", expected, actual)
	}
	if expected.ID != actual.ID {
		t.Errorf("ID mismatch: expected=%d, actual=%d", expected.ID, actual.ID)
	}
	if expected.Name != actual.Name {
		t.Errorf("Name mismatch: expected=%s, actual=%s", expected.Name, actual.Name)
	}
}

// AssertErrorContains checks if error message contains expected string
func AssertErrorContains(t *testing.T, err error, expectedMsg string) {
	t.Helper()
	if err == nil {
		t.Fatalf("Expected error containing %q, got nil", expectedMsg)
	}
	if !containsString(err.Error(), expectedMsg) {
		t.Errorf("Error %q does not contain %q", err.Error(), expectedMsg)
	}
}

// AssertNoError checks that there is no error
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

// AssertError checks that there is an error
func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// Helper functions

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && indexOfString(s, substr) >= 0
}

func indexOfString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// ContextWithTimeout returns a context with a reasonable timeout for tests
func ContextWithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

// ContextWithShortTimeout returns a context with a short timeout
func ContextWithShortTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 500*time.Millisecond)
}
