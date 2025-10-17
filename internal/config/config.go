package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/leifj/go-raid/internal/storage"
)

// Config holds application configuration
type Config struct {
	Server  ServerConfig
	Storage storage.StorageConfig
	Auth    AuthConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host string
	Port int
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret string
	// For future OAuth2/OIDC integration
	Enabled bool
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	port, err := strconv.Atoi(getEnv("SERVER_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid SERVER_PORT: %w", err)
	}

	// Load storage configuration
	storageType := storage.StorageType(getEnv("STORAGE_TYPE", "file"))
	storageCfg, err := loadStorageConfig(storageType)
	if err != nil {
		return nil, fmt.Errorf("failed to load storage config: %w", err)
	}

	return &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: port,
		},
		Storage: *storageCfg,
		Auth: AuthConfig{
			JWTSecret: getEnv("JWT_SECRET", ""),
			Enabled:   getEnv("AUTH_ENABLED", "false") == "true",
		},
	}, nil
}

func loadStorageConfig(storageType storage.StorageType) (*storage.StorageConfig, error) {
	cfg := &storage.StorageConfig{
		Type: storageType,
	}

	switch storageType {
	case storage.StorageTypeFile, storage.StorageTypeFileGit:
		cfg.File = &storage.FileConfig{
			DataDir:        getEnv("STORAGE_FILE_DATADIR", "./data"),
			GitEnabled:     storageType == storage.StorageTypeFileGit,
			GitAutoCommit:  getEnv("STORAGE_GIT_AUTOCOMMIT", "true") == "true",
			GitAuthorName:  getEnv("STORAGE_GIT_AUTHOR_NAME", "RAiD System"),
			GitAuthorEmail: getEnv("STORAGE_GIT_AUTHOR_EMAIL", "raid@example.org"),
		}

	case storage.StorageTypeFDB:
		apiVersion, _ := strconv.Atoi(getEnv("STORAGE_FDB_API_VERSION", "710"))
		cfg.FDB = &storage.FDBConfig{
			ClusterFile: getEnv("STORAGE_FDB_CLUSTER_FILE", ""),
			APIVersion:  apiVersion,
		}

	case storage.StorageTypeCockroach:
		port, _ := strconv.Atoi(getEnv("STORAGE_COCKROACH_PORT", "26257"))
		cfg.Cockroach = &storage.CockroachConfig{
			Host:     getEnv("STORAGE_COCKROACH_HOST", "localhost"),
			Port:     port,
			Database: getEnv("STORAGE_COCKROACH_DATABASE", "raid"),
			User:     getEnv("STORAGE_COCKROACH_USER", "root"),
			Password: getEnv("STORAGE_COCKROACH_PASSWORD", ""),
			SSLMode:  getEnv("STORAGE_COCKROACH_SSLMODE", "disable"),
			SSLCert:  getEnv("STORAGE_COCKROACH_SSLCERT", ""),
			SSLKey:   getEnv("STORAGE_COCKROACH_SSLKEY", ""),
			SSLRoot:  getEnv("STORAGE_COCKROACH_SSLROOT", ""),
		}

	default:
		return nil, fmt.Errorf("unknown storage type: %s", storageType)
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
