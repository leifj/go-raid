package file

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/leifj/go-raid/internal/models"
	"github.com/leifj/go-raid/internal/storage"
)

func init() {
	// Register file storage factory
	storage.RegisterFactory(storage.StorageTypeFile, func(cfg interface{}) (storage.Repository, error) {
		fileCfg, ok := cfg.(*storage.FileConfig)
		if !ok || fileCfg == nil {
			fileCfg = &storage.FileConfig{DataDir: "./data"}
		}
		return New(&Config{DataDir: fileCfg.DataDir})
	})
}

// FileStorage implements storage.Repository using JSON files
type FileStorage struct {
	dataDir         string
	raidDir         string
	servicePointDir string
	mu              sync.RWMutex
	idCounter       int64
}

// Config holds configuration for file-based storage
type Config struct {
	DataDir string
}

// New creates a new file-based storage instance
func New(cfg *Config) (*FileStorage, error) {
	if cfg.DataDir == "" {
		cfg.DataDir = "./data"
	}

	raidDir := filepath.Join(cfg.DataDir, "raids")
	servicePointDir := filepath.Join(cfg.DataDir, "servicepoints")

	// Create directories if they don't exist
	if err := os.MkdirAll(raidDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create raids directory: %w", err)
	}
	if err := os.MkdirAll(servicePointDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create servicepoints directory: %w", err)
	}

	fs := &FileStorage{
		dataDir:         cfg.DataDir,
		raidDir:         raidDir,
		servicePointDir: servicePointDir,
		idCounter:       1000, // Start service point IDs at 1000
	}

	// Load the highest service point ID
	if err := fs.loadMaxServicePointID(); err != nil {
		return nil, err
	}

	return fs, nil
}

// CreateRAiD mints a new RAiD
func (fs *FileStorage) CreateRAiD(ctx context.Context, raid *models.RAiD) (*models.RAiD, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Generate identifier if not present
	if raid.Identifier == nil || raid.Identifier.ID == "" {
		servicePointID := int64(0)
		if raid.Identifier != nil && raid.Identifier.Owner != nil {
			servicePointID = raid.Identifier.Owner.ServicePoint
		}
		prefix, suffix, err := fs.generateIdentifier(ctx, servicePointID)
		if err != nil {
			return nil, err
		}
		if raid.Identifier == nil {
			raid.Identifier = &models.Identifier{}
		}
		raid.Identifier.ID = fmt.Sprintf("https://raid.org/%s/%s", prefix, suffix)
	}

	// Extract prefix and suffix from identifier
	prefix, suffix, err := parseRAiDIdentifier(raid.Identifier.ID)
	if err != nil {
		return nil, err
	}

	// Check if already exists
	filePath := fs.getRaidFilePath(prefix, suffix)
	if _, err := os.Stat(filePath); err == nil {
		return nil, storage.ErrAlreadyExists
	}

	// Set metadata
	now := time.Now()
	if raid.Metadata == nil {
		raid.Metadata = &models.Metadata{}
	}
	raid.Metadata.Created = now
	raid.Metadata.Updated = now

	// Set version
	if raid.Identifier.Version == 0 {
		raid.Identifier.Version = 1
	}

	// Save to file
	if err := fs.saveRAiD(raid, prefix, suffix); err != nil {
		return nil, err
	}

	return raid, nil
}

// GetRAiD retrieves a RAiD by prefix and suffix
func (fs *FileStorage) GetRAiD(ctx context.Context, prefix, suffix string) (*models.RAiD, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	return fs.loadRAiD(prefix, suffix)
}

// GetRAiDVersion retrieves a specific version of a RAiD
func (fs *FileStorage) GetRAiDVersion(ctx context.Context, prefix, suffix string, version int) (*models.RAiD, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	// Load the current version
	raid, err := fs.loadRAiD(prefix, suffix)
	if err != nil {
		return nil, err
	}

	// If requesting current version, return it
	if raid.Identifier.Version == version {
		return raid, nil
	}

	// Try to load historical version
	historyFile := fs.getRaidHistoryFilePath(prefix, suffix, version)
	if _, err := os.Stat(historyFile); err != nil {
		return nil, storage.ErrNotFound
	}

	return fs.loadRAiDFromFile(historyFile)
}

// UpdateRAiD updates an existing RAiD
func (fs *FileStorage) UpdateRAiD(ctx context.Context, prefix, suffix string, raid *models.RAiD) (*models.RAiD, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Load existing RAiD
	existing, err := fs.loadRAiD(prefix, suffix)
	if err != nil {
		return nil, err
	}

	// Save old version to history
	historyFile := fs.getRaidHistoryFilePath(prefix, suffix, existing.Identifier.Version)
	if err := fs.saveRAiDToFile(existing, historyFile); err != nil {
		return nil, fmt.Errorf("failed to save history: %w", err)
	}

	// Update metadata
	now := time.Now()
	if raid.Metadata == nil {
		raid.Metadata = &models.Metadata{}
	}
	raid.Metadata.Created = existing.Metadata.Created
	raid.Metadata.Updated = now

	// Increment version
	raid.Identifier.Version = existing.Identifier.Version + 1

	// Save updated RAiD
	if err := fs.saveRAiD(raid, prefix, suffix); err != nil {
		return nil, err
	}

	return raid, nil
}

// ListRAiDs retrieves RAiDs with filters
func (fs *FileStorage) ListRAiDs(ctx context.Context, filter *storage.RAiDFilter) ([]*models.RAiD, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	raids, err := fs.loadAllRAiDs()
	if err != nil {
		return nil, err
	}

	// Apply filters
	filtered := fs.applyFilters(raids, filter)

	// Apply pagination
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(filtered) {
			filtered = filtered[filter.Offset:]
		}
		if filter.Limit > 0 && filter.Limit < len(filtered) {
			filtered = filtered[:filter.Limit]
		}
	}

	return filtered, nil
}

// ListPublicRAiDs retrieves only public RAiDs
func (fs *FileStorage) ListPublicRAiDs(ctx context.Context, filter *storage.RAiDFilter) ([]*models.RAiD, error) {
	raids, err := fs.ListRAiDs(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Filter for open access only
	public := make([]*models.RAiD, 0)
	for _, raid := range raids {
		if raid.Access != nil && raid.Access.Type != nil && raid.Access.Type.ID == "https://vocabulary.raid.org/access.type.schema/82" {
			public = append(public, raid)
		}
	}

	return public, nil
}

// GetRAiDHistory retrieves version history
func (fs *FileStorage) GetRAiDHistory(ctx context.Context, prefix, suffix string) ([]*models.RAiD, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	// Load current version
	current, err := fs.loadRAiD(prefix, suffix)
	if err != nil {
		return nil, err
	}

	history := []*models.RAiD{current}

	// Load historical versions
	historyDir := fs.getRaidHistoryDir(prefix, suffix)
	entries, err := os.ReadDir(historyDir)
	if err != nil {
		if os.IsNotExist(err) {
			return history, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			filePath := filepath.Join(historyDir, entry.Name())
			raid, err := fs.loadRAiDFromFile(filePath)
			if err != nil {
				continue // Skip corrupted history files
			}
			history = append(history, raid)
		}
	}

	return history, nil
}

// DeleteRAiD soft deletes a RAiD
func (fs *FileStorage) DeleteRAiD(ctx context.Context, prefix, suffix string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	filePath := fs.getRaidFilePath(prefix, suffix)
	deletedPath := filePath + ".deleted"

	return os.Rename(filePath, deletedPath)
}

// GenerateIdentifier generates a unique identifier
func (fs *FileStorage) GenerateIdentifier(ctx context.Context, servicePointID int64) (prefix, suffix string, err error) {
	return fs.generateIdentifier(ctx, servicePointID)
}

// CreateServicePoint creates a new service point
func (fs *FileStorage) CreateServicePoint(ctx context.Context, sp *models.ServicePoint) (*models.ServicePoint, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Generate ID if not set
	if sp.ID == 0 {
		fs.idCounter++
		sp.ID = fs.idCounter
	}

	// Check if already exists
	filePath := fs.getServicePointFilePath(sp.ID)
	if _, err := os.Stat(filePath); err == nil {
		return nil, storage.ErrAlreadyExists
	}

	// Save to file
	if err := fs.saveServicePoint(sp); err != nil {
		return nil, err
	}

	return sp, nil
}

// GetServicePoint retrieves a service point by ID
func (fs *FileStorage) GetServicePoint(ctx context.Context, id int64) (*models.ServicePoint, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	return fs.loadServicePoint(id)
}

// UpdateServicePoint updates a service point
func (fs *FileStorage) UpdateServicePoint(ctx context.Context, id int64, sp *models.ServicePoint) (*models.ServicePoint, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Check if exists
	if _, err := fs.loadServicePoint(id); err != nil {
		return nil, err
	}

	// Ensure ID matches
	sp.ID = id

	// Save to file
	if err := fs.saveServicePoint(sp); err != nil {
		return nil, err
	}

	return sp, nil
}

// ListServicePoints retrieves all service points
func (fs *FileStorage) ListServicePoints(ctx context.Context) ([]*models.ServicePoint, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	entries, err := os.ReadDir(fs.servicePointDir)
	if err != nil {
		return nil, err
	}

	servicePoints := make([]*models.ServicePoint, 0)
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			filePath := filepath.Join(fs.servicePointDir, entry.Name())
			sp, err := fs.loadServicePointFromFile(filePath)
			if err != nil {
				continue // Skip corrupted files
			}
			servicePoints = append(servicePoints, sp)
		}
	}

	return servicePoints, nil
}

// DeleteServicePoint removes a service point
func (fs *FileStorage) DeleteServicePoint(ctx context.Context, id int64) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	filePath := fs.getServicePointFilePath(id)
	return os.Remove(filePath)
}

// Close closes the storage
func (fs *FileStorage) Close() error {
	return nil // No resources to clean up
}

// HealthCheck verifies storage is accessible
func (fs *FileStorage) HealthCheck(ctx context.Context) error {
	// Try to write a test file
	testFile := filepath.Join(fs.dataDir, ".healthcheck")
	if err := os.WriteFile(testFile, []byte("ok"), 0644); err != nil {
		return fmt.Errorf("storage not writable: %w", err)
	}
	return os.Remove(testFile)
}

// Helper methods

func (fs *FileStorage) generateIdentifier(ctx context.Context, servicePointID int64) (string, string, error) {
	// Load service point to get prefix
	prefix := "10.25.1.1" // Default prefix
	if servicePointID > 0 {
		sp, err := fs.loadServicePoint(servicePointID)
		if err == nil && sp.Prefix != "" {
			prefix = sp.Prefix
		}
	}

	// Generate suffix using timestamp + random component
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	return prefix, suffix, nil
}

func (fs *FileStorage) getRaidFilePath(prefix, suffix string) string {
	// Sanitize prefix to create directory structure
	dirPath := filepath.Join(fs.raidDir, sanitizePath(prefix))
	os.MkdirAll(dirPath, 0755)
	return filepath.Join(dirPath, sanitizePath(suffix)+".json")
}

func (fs *FileStorage) getRaidHistoryDir(prefix, suffix string) string {
	dirPath := filepath.Join(fs.raidDir, sanitizePath(prefix), ".history", sanitizePath(suffix))
	os.MkdirAll(dirPath, 0755)
	return dirPath
}

func (fs *FileStorage) getRaidHistoryFilePath(prefix, suffix string, version int) string {
	historyDir := fs.getRaidHistoryDir(prefix, suffix)
	return filepath.Join(historyDir, fmt.Sprintf("v%d.json", version))
}

func (fs *FileStorage) getServicePointFilePath(id int64) string {
	return filepath.Join(fs.servicePointDir, fmt.Sprintf("%d.json", id))
}

func (fs *FileStorage) saveRAiD(raid *models.RAiD, prefix, suffix string) error {
	filePath := fs.getRaidFilePath(prefix, suffix)
	return fs.saveRAiDToFile(raid, filePath)
}

func (fs *FileStorage) saveRAiDToFile(raid *models.RAiD, filePath string) error {
	data, err := json.MarshalIndent(raid, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal RAiD: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write RAiD file: %w", err)
	}

	return nil
}

func (fs *FileStorage) loadRAiD(prefix, suffix string) (*models.RAiD, error) {
	filePath := fs.getRaidFilePath(prefix, suffix)
	return fs.loadRAiDFromFile(filePath)
}

func (fs *FileStorage) loadRAiDFromFile(filePath string) (*models.RAiD, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, storage.ErrNotFound
		}
		return nil, fmt.Errorf("failed to read RAiD file: %w", err)
	}

	var raid models.RAiD
	if err := json.Unmarshal(data, &raid); err != nil {
		return nil, fmt.Errorf("failed to unmarshal RAiD: %w", err)
	}

	return &raid, nil
}

func (fs *FileStorage) loadAllRAiDs() ([]*models.RAiD, error) {
	raids := make([]*models.RAiD, 0)

	err := filepath.Walk(fs.raidDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".json") && !strings.Contains(path, ".history") && !strings.HasSuffix(path, ".deleted") {
			raid, err := fs.loadRAiDFromFile(path)
			if err == nil {
				raids = append(raids, raid)
			}
		}
		return nil
	})

	return raids, err
}

func (fs *FileStorage) saveServicePoint(sp *models.ServicePoint) error {
	filePath := fs.getServicePointFilePath(sp.ID)
	data, err := json.MarshalIndent(sp, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal service point: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write service point file: %w", err)
	}

	return nil
}

func (fs *FileStorage) loadServicePoint(id int64) (*models.ServicePoint, error) {
	filePath := fs.getServicePointFilePath(id)
	return fs.loadServicePointFromFile(filePath)
}

func (fs *FileStorage) loadServicePointFromFile(filePath string) (*models.ServicePoint, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, storage.ErrNotFound
		}
		return nil, fmt.Errorf("failed to read service point file: %w", err)
	}

	var sp models.ServicePoint
	if err := json.Unmarshal(data, &sp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal service point: %w", err)
	}

	return &sp, nil
}

func (fs *FileStorage) loadMaxServicePointID() error {
	entries, err := os.ReadDir(fs.servicePointDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	maxID := int64(1000)
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			filePath := filepath.Join(fs.servicePointDir, entry.Name())
			sp, err := fs.loadServicePointFromFile(filePath)
			if err == nil && sp.ID > maxID {
				maxID = sp.ID
			}
		}
	}

	fs.idCounter = maxID
	return nil
}

func (fs *FileStorage) applyFilters(raids []*models.RAiD, filter *storage.RAiDFilter) []*models.RAiD {
	if filter == nil {
		return raids
	}

	filtered := make([]*models.RAiD, 0)
	for _, raid := range raids {
		// Filter by contributor ID
		if filter.ContributorID != "" {
			found := false
			for _, contributor := range raid.Contributor {
				if contributor.ID == filter.ContributorID {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Filter by organisation ID
		if filter.OrganisationID != "" {
			found := false
			for _, org := range raid.Organisation {
				if org.ID == filter.OrganisationID {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		filtered = append(filtered, raid)
	}

	return filtered
}

func parseRAiDIdentifier(id string) (prefix, suffix string, err error) {
	// Expected format: https://raid.org/{prefix}/{suffix}
	parts := strings.Split(id, "/")
	if len(parts) < 5 {
		return "", "", fmt.Errorf("invalid RAiD identifier format: %s", id)
	}
	return parts[3], parts[4], nil
}

func sanitizePath(s string) string {
	// Replace characters that are problematic in file paths
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	s = strings.ReplaceAll(s, ":", "_")
	s = strings.ReplaceAll(s, "*", "_")
	s = strings.ReplaceAll(s, "?", "_")
	s = strings.ReplaceAll(s, "\"", "_")
	s = strings.ReplaceAll(s, "<", "_")
	s = strings.ReplaceAll(s, ">", "_")
	s = strings.ReplaceAll(s, "|", "_")
	return s
}
