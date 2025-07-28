package repository

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"
	"urlShortner/internal/models"
	"urlShortner/internal/utils"

	_ "github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}


func NewPostgresRepository() (*PostgresRepository, error) {
    dsn := fmt.Sprintf(
        "postgres://%s:%s@%s:%s/%s?sslmode=disable",
        getenv("DATABASE_USERNAME", "postgres"),
        getenv("DATABASE_PASSWORD", "postgres"),
        getenv("DATABASE_HOST", "localhost"),
        getenv("DATABASE_PORT", "5432"),
        getenv("DATABASE_NAME", "devdb"),
    )

    var db *sql.DB
    var err error

    // Retry loop (max 10 attempts with 2s interval)
    for i := 0; i < 10; i++ {
        db, err = sql.Open("postgres", dsn)
        if err == nil {
            if pingErr := db.Ping(); pingErr == nil {
                break
            }
        }
        fmt.Println("Waiting for DB to be ready...")
        time.Sleep(2 * time.Second)
    }

    if err != nil {
        return nil, fmt.Errorf("failed to connect to DB after retries: %w", err)
    }

    // schema setup ...
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
		return 4
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

func (p *PostgresRepository) DB() *sql.DB {
	return p.db
}
