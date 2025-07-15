package repository_test

import (
	"testing"
	"urlShortner/models"
	"urlShortner/repository"

	"github.com/stretchr/testify/assert"
)


func TestNewPostgresRepository(t *testing.T){
	repo,err := repository.NewPostgresRepository()
	assert.NoError(t,err)
	assert.NotNil(t,repo)

}

func cleanTable(t *testing.T, repo *repository.PostgresRepository) {
	_, err := repo.DB().Exec("DELETE FROM eg_url_shortener")
	assert.NoError(t, err)
}

func TestGetOrCreateShortkey(t *testing.T) {
	repo, err := repository.NewPostgresRepository()
	assert.NoError(t, err)
	assert.NotNil(t, repo)

	// Clean up table before test
	_, err = repo.DB().Exec("DELETE FROM eg_url_shortener")
	assert.NoError(t, err)

	req := models.ShortenRequest{
		URL:       "https://test-url.com",
		ValidFrom: 1000,
		ValidTill: 2000,
	}

	key, err := repo.GetOrCreateShortKey(req)
	assert.NoError(t, err)
	assert.NotEmpty(t, key)

	// Idempotency: same URL should return same key
	key2, err := repo.GetOrCreateShortKey(req)
	assert.NoError(t, err)
	assert.Equal(t, key, key2)

	// Test GetURL
	url, err := repo.GetURL(key)
	assert.NoError(t, err)
	assert.Equal(t, req.URL, url)

	// Test GetURL with non-existent key
	_, err = repo.GetURL("nonexistentkey")
	assert.Error(t, err)
}

// Optionally, add a DB() getter to PostgresRepository for test access
// func (p *PostgresRepository) DB() *sql.DB { return p.db }
	