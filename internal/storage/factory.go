package storage

import (
	"fmt"
)

// StorageType defines the type of storage backend
type StorageType string

const (
	// StorageTypeFile uses file-based JSON storage
	StorageTypeFile StorageType = "file"
	// StorageTypeFileGit uses file-based JSON storage with git versioning
	StorageTypeFileGit StorageType = "file-git"
	// StorageTypeFDB uses FoundationDB
	StorageTypeFDB StorageType = "fdb"
	// StorageTypeCockroach uses CockroachDB
	StorageTypeCockroach StorageType = "cockroach"
)

// StorageConfig holds configuration for all storage types
type StorageConfig struct {
	Type StorageType

	// File storage configuration
	File *FileConfig

	// FoundationDB configuration
	FDB *FDBConfig

	// CockroachDB configuration
	Cockroach *CockroachConfig
}

// FileConfig holds file storage configuration
type FileConfig struct {
	DataDir string
	// Git configuration (optional)
	GitEnabled     bool
	GitAutoCommit  bool
	GitAuthorName  string
	GitAuthorEmail string
}

// FDBConfig holds FoundationDB configuration
type FDBConfig struct {
	ClusterFile string
	APIVersion  int
}

// CockroachConfig holds CockroachDB configuration
type CockroachConfig struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
	SSLMode  string
	SSLCert  string
	SSLKey   string
	SSLRoot  string
}

// RepositoryFactory is a function type for creating repositories
type RepositoryFactory func(interface{}) (Repository, error)

var factories = make(map[StorageType]RepositoryFactory)

// RegisterFactory registers a storage factory
func RegisterFactory(storageType StorageType, factory RepositoryFactory) {
	factories[storageType] = factory
}

// NewRepository creates a new storage repository based on configuration
func NewRepository(cfg *StorageConfig) (Repository, error) {
	factory, ok := factories[cfg.Type]
	if !ok {
		return nil, fmt.Errorf("unknown storage type: %s (not registered)", cfg.Type)
	}

	// Pass the appropriate config based on type
	var config interface{}
	switch cfg.Type {
	case StorageTypeFile, StorageTypeFileGit:
		config = cfg.File
	case StorageTypeFDB:
		config = cfg.FDB
	case StorageTypeCockroach:
		config = cfg.Cockroach
	default:
		return nil, fmt.Errorf("unknown storage type: %s", cfg.Type)
	}

	return factory(config)
}
