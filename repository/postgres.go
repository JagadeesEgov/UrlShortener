package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"urlShortner/config"
	"urlShortner/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// PostgresRepository implements URLRepository using PostgreSQL
type PostgresRepository struct {
	db  *pgxpool.Pool
	log *logrus.Logger
}

// NewPostgresRepository creates a new PostgreSQL repository instance
func NewPostgresRepository(cfg *config.DatabaseConfig, log *logrus.Logger) (*PostgresRepository, error) {
	if cfg == nil {
		return nil, fmt.Errorf("database config cannot be nil")
	}

	// Use the connection string method from config
	connString := cfg.GetConnectionString()
	
	// Configure connection pool
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Configure pool settings
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = time.Minute * 30
	poolConfig.HealthCheckPeriod = time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	repo := &PostgresRepository{
		db:  db,
		log: log,
	}

	// Test the connection
	if err := repo.HealthCheck(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Info("Successfully connected to PostgreSQL database")
	return repo, nil
}

// IncrementID increments and returns the next available ID using the sequence
func (p *PostgresRepository) IncrementID(ctx context.Context) (int64, error) {
	var id int64
	err := p.db.QueryRow(ctx, "SELECT nextval('eg_url_shorter_id')").Scan(&id)
	if err != nil {
		p.log.WithError(err).Error("Failed to get next sequence value")
		return 0, fmt.Errorf("failed to get next sequence value: %w", err)
	}

	// Return id-1 to match Java implementation
	result := id - 1
	p.log.WithField("id", result).Debug("Incremented ID")
	return result, nil
}

// SaveURL saves a URL with the given key and request
func (p *PostgresRepository) SaveURL(ctx context.Context, key string, request *models.ShortenRequest) error {
	// Validate input
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if request.URL == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	// Extract ID from key (format: "url:123")
	var id string
	if len(key) > 4 && key[:4] == "url:" {
		id = key[4:]
	} else {
		return fmt.Errorf("invalid key format: %s", key)
	}

	// Convert Unix timestamps to milliseconds if provided
	var validFromMs, validToMs *int64
	if request.ValidFrom != nil {
		validFromMs = request.ValidFrom
	}
	if request.ValidTill != nil {
		validToMs = request.ValidTill
	}

	// Insert or update the URL entry
	query := `
		INSERT INTO eg_url_shortener (id, url, validform, validto) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) 
		DO UPDATE SET url = $2, validform = $3, validto = $4`

	_, err := p.db.Exec(ctx, query, id, request.URL, validFromMs, validToMs)
	if err != nil {
		p.log.WithError(err).WithField("key", key).Error("Failed to save URL")
		return fmt.Errorf("failed to save URL: %w", err)
	}

	p.log.WithFields(logrus.Fields{
		"url": request.URL,
		"key": key,
	}).Info("URL saved successfully")

	return nil
}

// GetURL retrieves a URL by ID
func (p *PostgresRepository) GetURL(ctx context.Context, id int64) (string, error) {
	request, err := p.GetURLDetails(ctx, id)
	if err != nil {
		return "", err
	}

	// Check if URL is active
	if !request.IsActive() {
		if request.IsExpired() {
			p.log.WithField("id", id).Warn("URL has expired")
			return "", fmt.Errorf("URL at key %d has expired", id)
		}
		p.log.WithField("id", id).Warn("URL is not yet active")
		return "", fmt.Errorf("URL at key %d is not yet active", id)
	}

	return request.URL, nil
}

// GetURLDetails retrieves full URL details by ID
func (p *PostgresRepository) GetURLDetails(ctx context.Context, id int64) (*models.ShortenRequest, error) {
	p.log.WithField("id", id).Debug("Retrieving URL details")

	var url string
	var validFrom, validTo sql.NullInt64
	
	query := "SELECT url, validform, validto FROM eg_url_shortener WHERE id = $1"
	err := p.db.QueryRow(ctx, query, fmt.Sprintf("%d", id)).Scan(&url, &validFrom, &validTo)
	if err != nil {
		if err == pgx.ErrNoRows {
			p.log.WithField("id", id).Debug("URL not found")
			return nil, fmt.Errorf("URL at key %d does not exist", id)
		}
		p.log.WithError(err).WithField("id", id).Error("Failed to retrieve URL")
		return nil, fmt.Errorf("failed to retrieve URL: %w", err)
	}

	// Build the response
	request := &models.ShortenRequest{
		ID:  fmt.Sprintf("%d", id),
		URL: url,
	}

	if validFrom.Valid {
		request.ValidFrom = &validFrom.Int64
	}
	if validTo.Valid {
		request.ValidTill = &validTo.Int64
	}

	p.log.WithFields(logrus.Fields{
		"url": url,
		"id":  id,
	}).Debug("Retrieved URL details")

	return request, nil
}

// DeleteURL deletes a URL by ID
func (p *PostgresRepository) DeleteURL(ctx context.Context, id int64) error {
	query := "DELETE FROM eg_url_shortener WHERE id = $1"
	result, err := p.db.Exec(ctx, query, fmt.Sprintf("%d", id))
	if err != nil {
		p.log.WithError(err).WithField("id", id).Error("Failed to delete URL")
		return fmt.Errorf("failed to delete URL: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("URL at key %d does not exist", id)
	}

	p.log.WithField("id", id).Info("URL deleted successfully")
	return nil
}

// CheckURLExists checks if a URL exists for the given ID
func (p *PostgresRepository) CheckURLExists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM eg_url_shortener WHERE id = $1)"
	err := p.db.QueryRow(ctx, query, fmt.Sprintf("%d", id)).Scan(&exists)
	if err != nil {
		p.log.WithError(err).WithField("id", id).Error("Failed to check URL existence")
		return false, fmt.Errorf("failed to check URL existence: %w", err)
	}

	return exists, nil
}

// HealthCheck performs a health check on the repository
func (p *PostgresRepository) HealthCheck(ctx context.Context) error {
	// Test the connection with a simple query
	var result int
	err := p.db.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("database health check returned unexpected result: %d", result)
	}

	// Check if the required table exists
	var tableExists bool
	err = p.db.QueryRow(ctx, 
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'eg_url_shortener')").Scan(&tableExists)
	if err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if !tableExists {
		return fmt.Errorf("required table 'eg_url_shortener' does not exist")
	}

	// Check if sequence exists
	var sequenceExists bool
	err = p.db.QueryRow(ctx,
		"SELECT EXISTS (SELECT FROM information_schema.sequences WHERE sequence_name = 'eg_url_shorter_id')").Scan(&sequenceExists)
	if err != nil {
		return fmt.Errorf("failed to check sequence existence: %w", err)
	}

	if !sequenceExists {
		return fmt.Errorf("required sequence 'eg_url_shorter_id' does not exist")
	}

	return nil
}

// Close closes the database connection pool
func (p *PostgresRepository) Close() error {
	if p.db != nil {
		p.db.Close()
		p.log.Info("PostgreSQL connection pool closed")
	}
	return nil
}

// GetStats returns database connection pool statistics
func (p *PostgresRepository) GetStats() *pgxpool.Stat {
	return p.db.Stat()
}

// ExecuteMigration executes a database migration
func (p *PostgresRepository) ExecuteMigration(ctx context.Context, migration string) error {
	_, err := p.db.Exec(ctx, migration)
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}
	return nil
}

// GetTotalURLCount returns the total number of URLs stored
func (p *PostgresRepository) GetTotalURLCount(ctx context.Context) (int64, error) {
	var count int64
	err := p.db.QueryRow(ctx, "SELECT COUNT(*) FROM eg_url_shortener").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get URL count: %w", err)
	}
	return count, nil
}

// CleanupExpiredURLs removes expired URLs from the database
func (p *PostgresRepository) CleanupExpiredURLs(ctx context.Context) (int64, error) {
	now := time.Now().UnixMilli()
	query := "DELETE FROM eg_url_shortener WHERE validto IS NOT NULL AND validto < $1"
	
	result, err := p.db.Exec(ctx, query, now)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired URLs: %w", err)
	}

	deleted := result.RowsAffected()
	if deleted > 0 {
		p.log.WithField("deleted_count", deleted).Info("Cleaned up expired URLs")
	}

	return deleted, nil
}
