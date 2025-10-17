package fdb

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	"github.com/leifj/go-raid/internal/models"
	"github.com/leifj/go-raid/internal/storage"
)

func init() {
	// Register FDB storage factory
	storage.RegisterFactory(storage.StorageTypeFDB, func(cfg interface{}) (storage.Repository, error) {
		fdbCfg, ok := cfg.(*storage.FDBConfig)
		if !ok || fdbCfg == nil {
			fdbCfg = &storage.FDBConfig{}
		}
		return New(&Config{
			ClusterFile: fdbCfg.ClusterFile,
			APIVersion:  fdbCfg.APIVersion,
		})
	})
}

// FDBStorage implements storage.Repository using FoundationDB
type FDBStorage struct {
	db              fdb.Database
	raidDir         directory.DirectorySubspace
	servicePointDir directory.DirectorySubspace
	counterDir      directory.DirectorySubspace
}

// Config holds FoundationDB configuration
type Config struct {
	ClusterFile string // Path to fdb.cluster file, empty for default
	APIVersion  int    // FDB API version, 0 for latest
}

// New creates a new FoundationDB storage instance
func New(cfg *Config) (*FDBStorage, error) {
	// Set API version
	apiVersion := cfg.APIVersion
	if apiVersion == 0 {
		apiVersion = 710 // FDB 7.1
	}

	if err := fdb.APIVersion(apiVersion); err != nil {
		return nil, fmt.Errorf("failed to set FDB API version: %w", err)
	}

	// Open database
	var db fdb.Database
	var err error

	if cfg.ClusterFile != "" {
		db, err = fdb.OpenDatabase(cfg.ClusterFile)
	} else {
		db, err = fdb.OpenDefault()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open FDB database: %w", err)
	}

	fs := &FDBStorage{
		db: db,
	}

	// Initialize directory structure
	if err := fs.initDirectories(); err != nil {
		return nil, err
	}

	return fs, nil
}

// Initialize directory structure in FDB
func (fs *FDBStorage) initDirectories() error {
	_, err := fs.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		// Create raid directory
		raidDir, err := directory.CreateOrOpen(tr, []string{"raid"}, nil)
		if err != nil {
			return nil, err
		}
		fs.raidDir = raidDir

		// Create service point directory
		spDir, err := directory.CreateOrOpen(tr, []string{"servicepoint"}, nil)
		if err != nil {
			return nil, err
		}
		fs.servicePointDir = spDir

		// Create counter directory for ID generation
		counterDir, err := directory.CreateOrOpen(tr, []string{"counters"}, nil)
		if err != nil {
			return nil, err
		}
		fs.counterDir = counterDir

		return nil, nil
	})

	return err
}

// CreateRAiD creates a new RAiD
func (fs *FDBStorage) CreateRAiD(ctx context.Context, raid *models.RAiD) (*models.RAiD, error) {
	// Generate identifier if not present
	if raid.Identifier == nil || raid.Identifier.ID == "" {
		servicePointID := int64(0)
		if raid.Identifier != nil && raid.Identifier.Owner != nil {
			servicePointID = raid.Identifier.Owner.ServicePoint
		}
		prefix, suffix, err := fs.GenerateIdentifier(ctx, servicePointID)
		if err != nil {
			return nil, err
		}
		if raid.Identifier == nil {
			raid.Identifier = &models.Identifier{}
		}
		raid.Identifier.ID = fmt.Sprintf("https://raid.org/%s/%s", prefix, suffix)
	}

	// Extract prefix and suffix
	prefix, suffix, err := parseRAiDIdentifier(raid.Identifier.ID)
	if err != nil {
		return nil, err
	}

	// Set metadata
	now := time.Now()
	if raid.Metadata == nil {
		raid.Metadata = &models.Metadata{}
	}
	raid.Metadata.Created = now
	raid.Metadata.Updated = now

	if raid.Identifier.Version == 0 {
		raid.Identifier.Version = 1
	}

	// Store in FDB
	_, err = fs.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		key := fs.raidDir.Pack(tuple.Tuple{prefix, suffix, "current"})

		// Check if exists
		existing := tr.Get(key).MustGet()
		if existing != nil {
			return nil, storage.ErrAlreadyExists
		}

		// Serialize raid
		data, err := json.Marshal(raid)
		if err != nil {
			return nil, err
		}

		// Store current version
		tr.Set(key, data)

		// Store in version history
		versionKey := fs.raidDir.Pack(tuple.Tuple{prefix, suffix, "version", raid.Identifier.Version})
		tr.Set(versionKey, data)

		return nil, nil
	})

	if err != nil {
		return nil, err
	}

	return raid, nil
}

// GetRAiD retrieves a RAiD
func (fs *FDBStorage) GetRAiD(ctx context.Context, prefix, suffix string) (*models.RAiD, error) {
	result, err := fs.db.ReadTransact(func(rtr fdb.ReadTransaction) (interface{}, error) {
		key := fs.raidDir.Pack(tuple.Tuple{prefix, suffix, "current"})
		data := rtr.Get(key).MustGet()

		if data == nil {
			return nil, storage.ErrNotFound
		}

		var raid models.RAiD
		if err := json.Unmarshal(data, &raid); err != nil {
			return nil, err
		}

		return &raid, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*models.RAiD), nil
}

// GetRAiDVersion retrieves a specific version
func (fs *FDBStorage) GetRAiDVersion(ctx context.Context, prefix, suffix string, version int) (*models.RAiD, error) {
	result, err := fs.db.ReadTransact(func(rtr fdb.ReadTransaction) (interface{}, error) {
		key := fs.raidDir.Pack(tuple.Tuple{prefix, suffix, "version", version})
		data := rtr.Get(key).MustGet()

		if data == nil {
			return nil, storage.ErrNotFound
		}

		var raid models.RAiD
		if err := json.Unmarshal(data, &raid); err != nil {
			return nil, err
		}

		return &raid, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*models.RAiD), nil
}

// UpdateRAiD updates a RAiD
func (fs *FDBStorage) UpdateRAiD(ctx context.Context, prefix, suffix string, raid *models.RAiD) (*models.RAiD, error) {
	_, err := fs.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		// Load existing
		key := fs.raidDir.Pack(tuple.Tuple{prefix, suffix, "current"})
		existingData := tr.Get(key).MustGet()

		if existingData == nil {
			return nil, storage.ErrNotFound
		}

		var existing models.RAiD
		if err := json.Unmarshal(existingData, &existing); err != nil {
			return nil, err
		}

		// Update metadata
		now := time.Now()
		if raid.Metadata == nil {
			raid.Metadata = &models.Metadata{}
		}
		raid.Metadata.Created = existing.Metadata.Created
		raid.Metadata.Updated = now
		raid.Identifier.Version = existing.Identifier.Version + 1

		// Serialize
		data, err := json.Marshal(raid)
		if err != nil {
			return nil, err
		}

		// Update current version
		tr.Set(key, data)

		// Store in version history
		versionKey := fs.raidDir.Pack(tuple.Tuple{prefix, suffix, "version", raid.Identifier.Version})
		tr.Set(versionKey, data)

		return nil, nil
	})

	if err != nil {
		return nil, err
	}

	return raid, nil
}

// ListRAiDs lists RAiDs with filters
func (fs *FDBStorage) ListRAiDs(ctx context.Context, filter *storage.RAiDFilter) ([]*models.RAiD, error) {
	result, err := fs.db.ReadTransact(func(rtr fdb.ReadTransaction) (interface{}, error) {
		// Get all current RAiDs
		prefix := fs.raidDir.Pack(tuple.Tuple{})

		iter := rtr.GetRange(fdb.KeyRange{
			Begin: fdb.Key(append(prefix, 0x00)),
			End:   fdb.Key(append(prefix, 0xFF)),
		}, fdb.RangeOptions{}).Iterator()

		raids := make([]*models.RAiD, 0)

		for iter.Advance() {
			kv := iter.MustGet()

			// Only process "current" keys
			t, err := fs.raidDir.Unpack(kv.Key)
			if err != nil {
				continue
			}
			if len(t) >= 3 && t[2].(string) == "current" {
				var raid models.RAiD
				if err := json.Unmarshal(kv.Value, &raid); err != nil {
					continue
				}
				raids = append(raids, &raid)
			}
		}

		return raids, nil
	})

	if err != nil {
		return nil, err
	}

	raids := result.([]*models.RAiD)

	// Apply filters
	raids = applyFilters(raids, filter)

	// Apply pagination
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(raids) {
			raids = raids[filter.Offset:]
		}
		if filter.Limit > 0 && filter.Limit < len(raids) {
			raids = raids[:filter.Limit]
		}
	}

	return raids, nil
}

// ListPublicRAiDs lists only public RAiDs
func (fs *FDBStorage) ListPublicRAiDs(ctx context.Context, filter *storage.RAiDFilter) ([]*models.RAiD, error) {
	raids, err := fs.ListRAiDs(ctx, filter)
	if err != nil {
		return nil, err
	}

	public := make([]*models.RAiD, 0)
	for _, raid := range raids {
		if raid.Access != nil && raid.Access.Type != nil && raid.Access.Type.ID == "https://vocabulary.raid.org/access.type.schema/82" {
			public = append(public, raid)
		}
	}

	return public, nil
}

// GetRAiDHistory retrieves version history
func (fs *FDBStorage) GetRAiDHistory(ctx context.Context, prefix, suffix string) ([]*models.RAiD, error) {
	result, err := fs.db.ReadTransact(func(rtr fdb.ReadTransaction) (interface{}, error) {
		keyPrefix := fs.raidDir.Pack(tuple.Tuple{prefix, suffix, "version"})

		iter := rtr.GetRange(fdb.KeyRange{
			Begin: fdb.Key(append(keyPrefix, 0x00)),
			End:   fdb.Key(append(keyPrefix, 0xFF)),
		}, fdb.RangeOptions{}).Iterator()

		history := make([]*models.RAiD, 0)

		for iter.Advance() {
			kv := iter.MustGet()
			var raid models.RAiD
			if err := json.Unmarshal(kv.Value, &raid); err != nil {
				continue
			}
			history = append(history, &raid)
		}

		return history, nil
	})

	if err != nil {
		return nil, err
	}

	return result.([]*models.RAiD), nil
}

// DeleteRAiD soft deletes a RAiD
func (fs *FDBStorage) DeleteRAiD(ctx context.Context, prefix, suffix string) error {
	_, err := fs.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		key := fs.raidDir.Pack(tuple.Tuple{prefix, suffix, "current"})
		deletedKey := fs.raidDir.Pack(tuple.Tuple{prefix, suffix, "deleted"})

		// Move to deleted
		data := tr.Get(key).MustGet()
		if data == nil {
			return nil, storage.ErrNotFound
		}

		tr.Set(deletedKey, data)
		tr.Clear(key)

		return nil, nil
	})

	return err
}

// GenerateIdentifier generates a unique identifier
func (fs *FDBStorage) GenerateIdentifier(ctx context.Context, servicePointID int64) (prefix, suffix string, err error) {
	// Load service point to get prefix
	prefix = "10.25.1.1" // Default
	if servicePointID > 0 {
		sp, err := fs.GetServicePoint(ctx, servicePointID)
		if err == nil && sp.Prefix != "" {
			prefix = sp.Prefix
		}
	}

	// Generate suffix using FDB atomic counter
	result, err := fs.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		counterKey := fs.counterDir.Pack(tuple.Tuple{"raid", prefix})

		// Atomic add
		tr.Add(counterKey, []byte{1, 0, 0, 0, 0, 0, 0, 0})

		// Read new value
		val := tr.Get(counterKey).MustGet()
		if val == nil {
			return int64(1), nil
		}

		// Decode little-endian int64
		var counter int64
		for i := 0; i < 8 && i < len(val); i++ {
			counter |= int64(val[i]) << (i * 8)
		}

		return counter, nil
	})

	if err != nil {
		return "", "", err
	}

	suffix = fmt.Sprintf("%d", result.(int64))
	return prefix, suffix, nil
}

// CreateServicePoint creates a service point
func (fs *FDBStorage) CreateServicePoint(ctx context.Context, sp *models.ServicePoint) (*models.ServicePoint, error) {
	// Generate ID if not set
	if sp.ID == 0 {
		id, err := fs.generateServicePointID(ctx)
		if err != nil {
			return nil, err
		}
		sp.ID = id
	}

	_, err := fs.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		key := fs.servicePointDir.Pack(tuple.Tuple{sp.ID})

		// Check if exists
		existing := tr.Get(key).MustGet()
		if existing != nil {
			return nil, storage.ErrAlreadyExists
		}

		// Serialize
		data, err := json.Marshal(sp)
		if err != nil {
			return nil, err
		}

		tr.Set(key, data)
		return nil, nil
	})

	if err != nil {
		return nil, err
	}

	return sp, nil
}

// GetServicePoint retrieves a service point
func (fs *FDBStorage) GetServicePoint(ctx context.Context, id int64) (*models.ServicePoint, error) {
	result, err := fs.db.ReadTransact(func(rtr fdb.ReadTransaction) (interface{}, error) {
		key := fs.servicePointDir.Pack(tuple.Tuple{id})
		data := rtr.Get(key).MustGet()

		if data == nil {
			return nil, storage.ErrNotFound
		}

		var sp models.ServicePoint
		if err := json.Unmarshal(data, &sp); err != nil {
			return nil, err
		}

		return &sp, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*models.ServicePoint), nil
}

// UpdateServicePoint updates a service point
func (fs *FDBStorage) UpdateServicePoint(ctx context.Context, id int64, sp *models.ServicePoint) (*models.ServicePoint, error) {
	sp.ID = id

	_, err := fs.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		key := fs.servicePointDir.Pack(tuple.Tuple{id})

		// Check if exists
		existing := tr.Get(key).MustGet()
		if existing == nil {
			return nil, storage.ErrNotFound
		}

		// Serialize
		data, err := json.Marshal(sp)
		if err != nil {
			return nil, err
		}

		tr.Set(key, data)
		return nil, nil
	})

	if err != nil {
		return nil, err
	}

	return sp, nil
}

// ListServicePoints lists all service points
func (fs *FDBStorage) ListServicePoints(ctx context.Context) ([]*models.ServicePoint, error) {
	result, err := fs.db.ReadTransact(func(rtr fdb.ReadTransaction) (interface{}, error) {
		prefix := fs.servicePointDir.Pack(tuple.Tuple{})

		iter := rtr.GetRange(fdb.KeyRange{
			Begin: fdb.Key(append(prefix, 0x00)),
			End:   fdb.Key(append(prefix, 0xFF)),
		}, fdb.RangeOptions{}).Iterator()

		sps := make([]*models.ServicePoint, 0)

		for iter.Advance() {
			kv := iter.MustGet()
			var sp models.ServicePoint
			if err := json.Unmarshal(kv.Value, &sp); err != nil {
				continue
			}
			sps = append(sps, &sp)
		}

		return sps, nil
	})

	if err != nil {
		return nil, err
	}

	return result.([]*models.ServicePoint), nil
}

// DeleteServicePoint deletes a service point
func (fs *FDBStorage) DeleteServicePoint(ctx context.Context, id int64) error {
	_, err := fs.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		key := fs.servicePointDir.Pack(tuple.Tuple{id})
		tr.Clear(key)
		return nil, nil
	})

	return err
}

// Close closes the FDB connection
func (fs *FDBStorage) Close() error {
	// FDB database handles don't need explicit closing
	return nil
}

// HealthCheck verifies FDB is accessible
func (fs *FDBStorage) HealthCheck(ctx context.Context) error {
	_, err := fs.db.ReadTransact(func(rtr fdb.ReadTransaction) (interface{}, error) {
		// Try to read a key
		testKey := fs.counterDir.Pack(tuple.Tuple{"healthcheck"})
		rtr.Get(testKey).MustGet()
		return nil, nil
	})
	return err
}

// Helper methods

func (fs *FDBStorage) generateServicePointID(ctx context.Context) (int64, error) {
	result, err := fs.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		counterKey := fs.counterDir.Pack(tuple.Tuple{"servicepoint_id"})

		// Atomic add
		tr.Add(counterKey, []byte{1, 0, 0, 0, 0, 0, 0, 0})

		// Read new value
		val := tr.Get(counterKey).MustGet()
		if val == nil {
			return int64(1001), nil
		}

		// Decode little-endian int64
		var counter int64
		for i := 0; i < 8 && i < len(val); i++ {
			counter |= int64(val[i]) << (i * 8)
		}

		if counter < 1000 {
			counter = 1000
		}

		return counter, nil
	})

	if err != nil {
		return 0, err
	}

	return result.(int64), nil
}

func parseRAiDIdentifier(id string) (prefix, suffix string, err error) {
	parts := strings.Split(id, "/")
	if len(parts) < 5 {
		return "", "", fmt.Errorf("invalid RAiD identifier format: %s", id)
	}
	return parts[3], parts[4], nil
}

func applyFilters(raids []*models.RAiD, filter *storage.RAiDFilter) []*models.RAiD {
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

// Verify FDBStorage implements storage.Repository
var _ storage.Repository = (*FDBStorage)(nil)
