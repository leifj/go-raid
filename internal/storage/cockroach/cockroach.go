package cockroach

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/leifj/go-raid/internal/models"
	"github.com/leifj/go-raid/internal/storage"
	_ "github.com/lib/pq" // PostgreSQL/CockroachDB driver
)

func init() {
	// Register CockroachDB storage factory
	storage.RegisterFactory(storage.StorageTypeCockroach, func(cfg interface{}) (storage.Repository, error) {
		crdbCfg, ok := cfg.(*storage.CockroachConfig)
		if !ok || crdbCfg == nil {
			return nil, fmt.Errorf("CockroachDB configuration is required")
		}
		return New(&Config{
			Host:     crdbCfg.Host,
			Port:     crdbCfg.Port,
			Database: crdbCfg.Database,
			User:     crdbCfg.User,
			Password: crdbCfg.Password,
			SSLMode:  crdbCfg.SSLMode,
			SSLCert:  crdbCfg.SSLCert,
			SSLKey:   crdbCfg.SSLKey,
			SSLRoot:  crdbCfg.SSLRoot,
		})
	})
}

// CockroachStorage implements storage.Repository using CockroachDB
type CockroachStorage struct {
	db *sql.DB
}

// Config holds CockroachDB configuration
type Config struct {
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

// New creates a new CockroachDB storage instance
func New(cfg *Config) (*CockroachStorage, error) {
	// Build connection string
	connStr := buildConnString(cfg)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	cs := &CockroachStorage{
		db: db,
	}

	// Initialize schema
	if err := cs.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return cs, nil
}

// Initialize database schema
func (cs *CockroachStorage) initSchema() error {
	schema := `
	-- RAiD table
	CREATE TABLE IF NOT EXISTS raids (
		prefix TEXT NOT NULL,
		suffix TEXT NOT NULL,
		version INT NOT NULL,
		is_current BOOLEAN NOT NULL DEFAULT true,
		is_deleted BOOLEAN NOT NULL DEFAULT false,
		data JSONB NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
		PRIMARY KEY (prefix, suffix, version),
		INDEX raids_current_idx (prefix, suffix) WHERE is_current = true,
		INDEX raids_deleted_idx (is_deleted),
		INVERTED INDEX raids_data_idx (data)
	);

	-- Service Point table
	CREATE TABLE IF NOT EXISTS service_points (
		id SERIAL PRIMARY KEY,
		data JSONB NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
		INVERTED INDEX service_points_data_idx (data)
	);

	-- ID counter table for service points
	CREATE TABLE IF NOT EXISTS id_counters (
		name TEXT PRIMARY KEY,
		value INT NOT NULL DEFAULT 1000
	);
	`

	_, err := cs.db.Exec(schema)
	return err
}

// CreateRAiD creates a new RAiD
func (cs *CockroachStorage) CreateRAiD(ctx context.Context, raid *models.RAiD) (*models.RAiD, error) {
	// Generate identifier if not present
	if raid.Identifier == nil || raid.Identifier.ID == "" {
		servicePointID := int64(0)
		if raid.Identifier != nil && raid.Identifier.Owner != nil {
			servicePointID = raid.Identifier.Owner.ServicePoint
		}
		prefix, suffix, err := cs.GenerateIdentifier(ctx, servicePointID)
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

	// Serialize to JSON
	data, err := json.Marshal(raid)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal RAiD: %w", err)
	}

	// Insert into database
	tx, err := cs.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Check if exists
	var exists bool
	err = tx.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM raids WHERE prefix = $1 AND suffix = $2 AND is_current = true)`,
		prefix, suffix,
	).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, storage.ErrAlreadyExists
	}

	// Insert
	_, err = tx.ExecContext(ctx,
		`INSERT INTO raids (prefix, suffix, version, is_current, data, created_at, updated_at) 
		 VALUES ($1, $2, $3, true, $4, $5, $6)`,
		prefix, suffix, raid.Identifier.Version, data, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert RAiD: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return raid, nil
}

// GetRAiD retrieves a RAiD
func (cs *CockroachStorage) GetRAiD(ctx context.Context, prefix, suffix string) (*models.RAiD, error) {
	var data []byte

	err := cs.db.QueryRowContext(ctx,
		`SELECT data FROM raids WHERE prefix = $1 AND suffix = $2 AND is_current = true AND is_deleted = false`,
		prefix, suffix,
	).Scan(&data)

	if err == sql.ErrNoRows {
		return nil, storage.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	var raid models.RAiD
	if err := json.Unmarshal(data, &raid); err != nil {
		return nil, fmt.Errorf("failed to unmarshal RAiD: %w", err)
	}

	return &raid, nil
}

// GetRAiDVersion retrieves a specific version
func (cs *CockroachStorage) GetRAiDVersion(ctx context.Context, prefix, suffix string, version int) (*models.RAiD, error) {
	var data []byte

	err := cs.db.QueryRowContext(ctx,
		`SELECT data FROM raids WHERE prefix = $1 AND suffix = $2 AND version = $3`,
		prefix, suffix, version,
	).Scan(&data)

	if err == sql.ErrNoRows {
		return nil, storage.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	var raid models.RAiD
	if err := json.Unmarshal(data, &raid); err != nil {
		return nil, fmt.Errorf("failed to unmarshal RAiD: %w", err)
	}

	return &raid, nil
}

// UpdateRAiD updates a RAiD
func (cs *CockroachStorage) UpdateRAiD(ctx context.Context, prefix, suffix string, raid *models.RAiD) (*models.RAiD, error) {
	tx, err := cs.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Get current version
	var currentVersion int
	var createdAt time.Time
	err = tx.QueryRowContext(ctx,
		`SELECT version, created_at FROM raids WHERE prefix = $1 AND suffix = $2 AND is_current = true`,
		prefix, suffix,
	).Scan(&currentVersion, &createdAt)

	if err == sql.ErrNoRows {
		return nil, storage.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// Update metadata
	now := time.Now()
	if raid.Metadata == nil {
		raid.Metadata = &models.Metadata{}
	}
	raid.Metadata.Created = createdAt
	raid.Metadata.Updated = now
	raid.Identifier.Version = currentVersion + 1

	// Serialize
	data, err := json.Marshal(raid)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal RAiD: %w", err)
	}

	// Mark old version as not current
	_, err = tx.ExecContext(ctx,
		`UPDATE raids SET is_current = false WHERE prefix = $1 AND suffix = $2 AND is_current = true`,
		prefix, suffix,
	)
	if err != nil {
		return nil, err
	}

	// Insert new version
	_, err = tx.ExecContext(ctx,
		`INSERT INTO raids (prefix, suffix, version, is_current, data, created_at, updated_at) 
		 VALUES ($1, $2, $3, true, $4, $5, $6)`,
		prefix, suffix, raid.Identifier.Version, data, createdAt, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert new version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return raid, nil
}

// ListRAiDs lists RAiDs with filters
func (cs *CockroachStorage) ListRAiDs(ctx context.Context, filter *storage.RAiDFilter) ([]*models.RAiD, error) {
	query := `SELECT data FROM raids WHERE is_current = true AND is_deleted = false`
	args := make([]interface{}, 0)
	argCount := 1

	// Build dynamic query based on filters
	if filter != nil {
		if filter.ContributorID != "" {
			query += fmt.Sprintf(` AND data->'contributor' @> '[{"id": "%s"}]'`, filter.ContributorID)
		}
		if filter.OrganisationID != "" {
			query += fmt.Sprintf(` AND data->'organisation' @> '[{"id": "%s"}]'`, filter.OrganisationID)
		}
		if filter.Limit > 0 {
			query += fmt.Sprintf(` LIMIT $%d`, argCount)
			args = append(args, filter.Limit)
			argCount++
		}
		if filter.Offset > 0 {
			query += fmt.Sprintf(` OFFSET $%d`, argCount)
			args = append(args, filter.Offset)
		}
	}

	rows, err := cs.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	raids := make([]*models.RAiD, 0)
	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			continue
		}

		var raid models.RAiD
		if err := json.Unmarshal(data, &raid); err != nil {
			continue
		}

		raids = append(raids, &raid)
	}

	return raids, rows.Err()
}

// ListPublicRAiDs lists only public RAiDs
func (cs *CockroachStorage) ListPublicRAiDs(ctx context.Context, filter *storage.RAiDFilter) ([]*models.RAiD, error) {
	query := `SELECT data FROM raids 
	          WHERE is_current = true 
	          AND is_deleted = false 
	          AND data->'access'->'type'->>'id' = 'https://vocabulary.raid.org/access.type.schema/82'`
	args := make([]interface{}, 0)
	argCount := 1

	if filter != nil {
		if filter.Limit > 0 {
			query += fmt.Sprintf(` LIMIT $%d`, argCount)
			args = append(args, filter.Limit)
			argCount++
		}
		if filter.Offset > 0 {
			query += fmt.Sprintf(` OFFSET $%d`, argCount)
			args = append(args, filter.Offset)
		}
	}

	rows, err := cs.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	raids := make([]*models.RAiD, 0)
	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			continue
		}

		var raid models.RAiD
		if err := json.Unmarshal(data, &raid); err != nil {
			continue
		}

		raids = append(raids, &raid)
	}

	return raids, rows.Err()
}

// GetRAiDHistory retrieves version history
func (cs *CockroachStorage) GetRAiDHistory(ctx context.Context, prefix, suffix string) ([]*models.RAiD, error) {
	rows, err := cs.db.QueryContext(ctx,
		`SELECT data FROM raids WHERE prefix = $1 AND suffix = $2 ORDER BY version DESC`,
		prefix, suffix,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	history := make([]*models.RAiD, 0)
	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			continue
		}

		var raid models.RAiD
		if err := json.Unmarshal(data, &raid); err != nil {
			continue
		}

		history = append(history, &raid)
	}

	return history, rows.Err()
}

// DeleteRAiD soft deletes a RAiD
func (cs *CockroachStorage) DeleteRAiD(ctx context.Context, prefix, suffix string) error {
	result, err := cs.db.ExecContext(ctx,
		`UPDATE raids SET is_deleted = true WHERE prefix = $1 AND suffix = $2 AND is_current = true`,
		prefix, suffix,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return storage.ErrNotFound
	}

	return nil
}

// GenerateIdentifier generates a unique identifier
func (cs *CockroachStorage) GenerateIdentifier(ctx context.Context, servicePointID int64) (prefix, suffix string, err error) {
	// Get prefix from service point
	prefix = "10.25.1.1" // Default
	if servicePointID > 0 {
		sp, err := cs.GetServicePoint(ctx, servicePointID)
		if err == nil && sp.Prefix != "" {
			prefix = sp.Prefix
		}
	}

	// Generate suffix using database sequence
	tx, err := cs.db.BeginTx(ctx, nil)
	if err != nil {
		return "", "", err
	}
	defer tx.Rollback()

	counterName := fmt.Sprintf("raid_%s", strings.ReplaceAll(prefix, ".", "_"))

	// Ensure counter exists
	_, err = tx.ExecContext(ctx,
		`INSERT INTO id_counters (name, value) VALUES ($1, 1) ON CONFLICT (name) DO NOTHING`,
		counterName,
	)
	if err != nil {
		return "", "", err
	}

	// Increment and get counter
	var counter int64
	err = tx.QueryRowContext(ctx,
		`UPDATE id_counters SET value = value + 1 WHERE name = $1 RETURNING value`,
		counterName,
	).Scan(&counter)
	if err != nil {
		return "", "", err
	}

	if err := tx.Commit(); err != nil {
		return "", "", err
	}

	suffix = fmt.Sprintf("%d", counter)
	return prefix, suffix, nil
}

// CreateServicePoint creates a service point
func (cs *CockroachStorage) CreateServicePoint(ctx context.Context, sp *models.ServicePoint) (*models.ServicePoint, error) {
	// Serialize
	data, err := json.Marshal(sp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal service point: %w", err)
	}

	// Insert and get generated ID
	var id int64
	err = cs.db.QueryRowContext(ctx,
		`INSERT INTO service_points (data, created_at, updated_at) 
		 VALUES ($1, NOW(), NOW()) 
		 RETURNING id`,
		data,
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("failed to insert service point: %w", err)
	}

	sp.ID = id
	return sp, nil
}

// GetServicePoint retrieves a service point
func (cs *CockroachStorage) GetServicePoint(ctx context.Context, id int64) (*models.ServicePoint, error) {
	var data []byte

	err := cs.db.QueryRowContext(ctx,
		`SELECT data FROM service_points WHERE id = $1`,
		id,
	).Scan(&data)

	if err == sql.ErrNoRows {
		return nil, storage.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	var sp models.ServicePoint
	if err := json.Unmarshal(data, &sp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal service point: %w", err)
	}

	return &sp, nil
}

// UpdateServicePoint updates a service point
func (cs *CockroachStorage) UpdateServicePoint(ctx context.Context, id int64, sp *models.ServicePoint) (*models.ServicePoint, error) {
	sp.ID = id

	// Serialize
	data, err := json.Marshal(sp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal service point: %w", err)
	}

	result, err := cs.db.ExecContext(ctx,
		`UPDATE service_points SET data = $1, updated_at = NOW() WHERE id = $2`,
		data, id,
	)
	if err != nil {
		return nil, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		return nil, storage.ErrNotFound
	}

	return sp, nil
}

// ListServicePoints lists all service points
func (cs *CockroachStorage) ListServicePoints(ctx context.Context) ([]*models.ServicePoint, error) {
	rows, err := cs.db.QueryContext(ctx, `SELECT data FROM service_points ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sps := make([]*models.ServicePoint, 0)
	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			continue
		}

		var sp models.ServicePoint
		if err := json.Unmarshal(data, &sp); err != nil {
			continue
		}

		sps = append(sps, &sp)
	}

	return sps, rows.Err()
}

// DeleteServicePoint deletes a service point
func (cs *CockroachStorage) DeleteServicePoint(ctx context.Context, id int64) error {
	result, err := cs.db.ExecContext(ctx,
		`DELETE FROM service_points WHERE id = $1`,
		id,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return storage.ErrNotFound
	}

	return nil
}

// Close closes the database connection
func (cs *CockroachStorage) Close() error {
	return cs.db.Close()
}

// HealthCheck verifies database is accessible
func (cs *CockroachStorage) HealthCheck(ctx context.Context) error {
	return cs.db.PingContext(ctx)
}

// Helper functions

func buildConnString(cfg *Config) string {
	parts := []string{
		fmt.Sprintf("host=%s", cfg.Host),
		fmt.Sprintf("port=%d", cfg.Port),
		fmt.Sprintf("user=%s", cfg.User),
		fmt.Sprintf("dbname=%s", cfg.Database),
	}

	if cfg.Password != "" {
		parts = append(parts, fmt.Sprintf("password=%s", cfg.Password))
	}

	if cfg.SSLMode != "" {
		parts = append(parts, fmt.Sprintf("sslmode=%s", cfg.SSLMode))
	}

	if cfg.SSLCert != "" {
		parts = append(parts, fmt.Sprintf("sslcert=%s", cfg.SSLCert))
	}

	if cfg.SSLKey != "" {
		parts = append(parts, fmt.Sprintf("sslkey=%s", cfg.SSLKey))
	}

	if cfg.SSLRoot != "" {
		parts = append(parts, fmt.Sprintf("sslrootcert=%s", cfg.SSLRoot))
	}

	return strings.Join(parts, " ")
}

func parseRAiDIdentifier(id string) (prefix, suffix string, err error) {
	parts := strings.Split(id, "/")
	if len(parts) < 5 {
		return "", "", fmt.Errorf("invalid RAiD identifier format: %s", id)
	}
	return parts[3], parts[4], nil
}

// Verify CockroachStorage implements storage.Repository
var _ storage.Repository = (*CockroachStorage)(nil)
