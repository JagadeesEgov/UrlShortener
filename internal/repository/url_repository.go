package repository

import "urlShortner/internal/models"

type URLRepository interface {
	GetOrCreateShortKey(req models.ShortenRequest) (string, error)
	GetURL(shortKey string) (string, error)
}