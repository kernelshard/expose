package server

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateRandomSubdomain() (string, error) {
	bytes := make([]byte, 4) // 4 bytes = 8 hex chars
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
