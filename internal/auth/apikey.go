package auth

import (
	"crypto/rand"
	"encoding/hex"
)

const (
	// apiKeyPrefix is prepended to all API keys for easy identification
	apiKeyPrefix = "rl_"
	// apiKeyBytes is the number of random bytes (32 bytes = 64 hex chars)
	apiKeyBytes = 32
)

// GenerateAPIKey creates a new random API key with the format: rl_<64 hex chars>
func GenerateAPIKey() (string, error) {
	bytes := make([]byte, apiKeyBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return apiKeyPrefix + hex.EncodeToString(bytes), nil
}
