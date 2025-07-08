package utils

import (
	"fmt"
	"urlShortner/config"

	"github.com/speps/go-hashids/v2"
)

// HashIDConverter provides HashID encoding and decoding functionality
type HashIDConverter struct {
	hashids   *hashids.HashID
	salt      string
	minLength int
}

// NewHashIDConverter creates a new HashIDConverter instance
func NewHashIDConverter(cfg *config.HashIDsConfig) (*HashIDConverter, error) {
	if cfg.Salt == "" {
		return nil, fmt.Errorf("hashids salt cannot be empty")
	}
	
	if cfg.MinLength < 1 {
		return nil, fmt.Errorf("hashids min length must be at least 1")
	}
	
	hd := hashids.NewData()
	hd.Salt = cfg.Salt
	hd.MinLength = cfg.MinLength
	
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return nil, fmt.Errorf("failed to create hashids instance: %w", err)
	}
	
	return &HashIDConverter{
		hashids:   h,
		salt:      cfg.Salt,
		minLength: cfg.MinLength,
	}, nil
}

// CreateHashStringForID converts an ID to a hash string
func (h *HashIDConverter) CreateHashStringForID(id int64) (string, error) {
	if id < 0 {
		return "", fmt.Errorf("id cannot be negative: %d", id)
	}
	
	hashString, err := h.hashids.EncodeInt64([]int64{id})
	if err != nil {
		return "", fmt.Errorf("failed to encode id %d: %w", id, err)
	}
	
	if hashString == "" {
		return "", fmt.Errorf("generated hash string is empty for id %d", id)
	}
	
	return hashString, nil
}

// GetIDForString converts a hash string back to an ID
func (h *HashIDConverter) GetIDForString(hashString string) (int64, error) {
	if hashString == "" {
		return 0, fmt.Errorf("hash string cannot be empty")
	}
	
	ids, err := h.hashids.DecodeInt64WithError(hashString)
	if err != nil {
		return 0, fmt.Errorf("failed to decode hash string '%s': %w", hashString, err)
	}
	
	if len(ids) != 1 {
		return 0, fmt.Errorf("invalid hash string '%s': expected 1 id, got %d", hashString, len(ids))
	}
	
	if ids[0] < 0 {
		return 0, fmt.Errorf("decoded id is negative: %d", ids[0])
	}
	
	return ids[0], nil
}

// IsValidHashString checks if a hash string is valid without decoding
func (h *HashIDConverter) IsValidHashString(hashString string) bool {
	if hashString == "" {
		return false
	}
	
	_, err := h.GetIDForString(hashString)
	return err == nil
}

// GetSalt returns the salt used for hashing (for debugging/testing)
func (h *HashIDConverter) GetSalt() string {
	return h.salt
}

// GetMinLength returns the minimum length configured
func (h *HashIDConverter) GetMinLength() int {
	return h.minLength
}

// CreateMultipleHashStrings creates hash strings for multiple IDs
func (h *HashIDConverter) CreateMultipleHashStrings(ids []int64) ([]string, error) {
	if len(ids) == 0 {
		return []string{}, nil
	}
	
	hashStrings := make([]string, len(ids))
	
	for i, id := range ids {
		hashString, err := h.CreateHashStringForID(id)
		if err != nil {
			return nil, fmt.Errorf("failed to create hash for id %d at index %d: %w", id, i, err)
		}
		hashStrings[i] = hashString
	}
	
	return hashStrings, nil
}

// DecodeMultipleHashStrings decodes multiple hash strings to IDs
func (h *HashIDConverter) DecodeMultipleHashStrings(hashStrings []string) ([]int64, error) {
	if len(hashStrings) == 0 {
		return []int64{}, nil
	}
	
	ids := make([]int64, len(hashStrings))
	
	for i, hashString := range hashStrings {
		id, err := h.GetIDForString(hashString)
		if err != nil {
			return nil, fmt.Errorf("failed to decode hash '%s' at index %d: %w", hashString, i, err)
		}
		ids[i] = id
	}
	
	return ids, nil
}

// ValidateConfiguration validates the HashID configuration
func (h *HashIDConverter) ValidateConfiguration() error {
	// Test encoding and decoding with a sample ID
	testID := int64(12345)
	
	hashString, err := h.CreateHashStringForID(testID)
	if err != nil {
		return fmt.Errorf("configuration validation failed during encoding: %w", err)
	}
	
	decodedID, err := h.GetIDForString(hashString)
	if err != nil {
		return fmt.Errorf("configuration validation failed during decoding: %w", err)
	}
	
	if decodedID != testID {
		return fmt.Errorf("configuration validation failed: expected %d, got %d", testID, decodedID)
	}
	
	// Check minimum length
	if len(hashString) < h.minLength {
		return fmt.Errorf("generated hash string length %d is less than configured minimum %d", len(hashString), h.minLength)
	}
	
	return nil
}
