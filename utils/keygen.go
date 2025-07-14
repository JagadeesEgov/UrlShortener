package utils

import (
	"crypto/rand"
	"math/big"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateShortKey(length int) (string, error) {
	key := make([]byte, length)
	for i := range key {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		key[i] = alphabet[n.Int64()]
	}
	return string(key), nil
} 