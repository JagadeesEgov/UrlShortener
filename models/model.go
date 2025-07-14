package models

type ShortenRequest struct {
	URL       string `json:"url" binding:"required"`
	ValidFrom int64  `json:"validFrom"`
	ValidTill int64  `json:"validTill"`
}