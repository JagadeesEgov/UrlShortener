package repository

import (
	"context"
	"urlShortner/models"
)

// URLRepository defines the interface for URL storage operations
type URLRepository interface {
	// IncrementID returns the next available ID
	IncrementID(ctx context.Context) (int64, error)
	
	// SaveURL saves a URL with the given key and request
	SaveURL(ctx context.Context, key string, request *models.ShortenRequest) error
	
	// GetURL retrieves a URL by ID
	GetURL(ctx context.Context, id int64) (string, error)
	
	// GetURLDetails retrieves full URL details by ID
	GetURLDetails(ctx context.Context, id int64) (*models.ShortenRequest, error)
	
	// DeleteURL deletes a URL by ID
	DeleteURL(ctx context.Context, id int64) error
	
	// CheckURLExists checks if a URL exists for the given ID
	CheckURLExists(ctx context.Context, id int64) (bool, error)
	
	// HealthCheck performs a health check on the repository
	HealthCheck(ctx context.Context) error
	
	// Close closes the repository connection
	Close() error
}

// RepositoryType represents the type of repository
type RepositoryType string

const (
	// Redis repository type
	Redis RepositoryType = "redis"
	// Postgres repository type
	Postgres RepositoryType = "postgres"
)

// RepositoryConfig holds common repository configuration
type RepositoryConfig struct {
	Type    RepositoryType
	Timeout int // Timeout in seconds
}
