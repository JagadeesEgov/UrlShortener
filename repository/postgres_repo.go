package repository

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"urlShortner/models"
	"urlShortner/utils"

	_ "github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository() (*PostgresRepository, error) {
	host := getenv("DATABASE_HOST", "localhost")
	port := getenv("DATABASE_PORT", "5432")
	name := getenv("DATABASE_NAME", "devdb")
	user := getenv("DATABASE_USERNAME", "postgres")
	pass := getenv("DATABASE_PASSWORD", "postgres")
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, name)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return &PostgresRepository{db: db}, nil
}

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func getMinLength() int {
	minLenStr := os.Getenv("MIN_LENGTH")
	if minLenStr == "" {
		return 6
	}
	minLen, err := strconv.Atoi(minLenStr)
	if err != nil || minLen < 1 {
		return 6
	}
	return minLen
}

// Get or create a short key for a long URL
func (p *PostgresRepository) GetOrCreateShortKey(req models.ShortenRequest) (string, error) {
	// 1. Check if URL already exists
	var existingKey string
	err := p.db.QueryRow("SELECT short_key FROM eg_url_shortener WHERE url = $1", req.URL).Scan(&existingKey)
	if err == nil {
		return existingKey, nil
	}
	if err != sql.ErrNoRows {
		return "", err
	}

	// 2. Generate unique short key
	var shortKey string
	minLen := getMinLength()
	for {
		shortKey, err = utils.GenerateShortKey(minLen)
		if err != nil {
			return "", err
		}
		var exists int
		err = p.db.QueryRow("SELECT 1 FROM eg_url_shortener WHERE short_key = $1", shortKey).Scan(&exists)
		if err == sql.ErrNoRows {
			break // unique
		}
	}

	// 3. Insert new row
	_, err = p.db.Exec(
		"INSERT INTO eg_url_shortener (short_key, url, validfrom, validto) VALUES ($1, $2, $3, $4)",
		shortKey, req.URL, req.ValidFrom, req.ValidTill,
	)
	if err != nil {
		return "", err
	}
	return shortKey, nil
}

func (p *PostgresRepository) GetURL(shortKey string) (string, error) {
	var url string
	err := p.db.QueryRow("SELECT url FROM eg_url_shortener WHERE short_key = $1", shortKey).Scan(&url)
	if err != nil {
		return "", err
	}
	return url, nil
}
