package auth

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/souravkumar/distributed-rate-limiter/internal/constants"
)

// GenerateAPIKey creates a new random API key with the format: rl_<64 hex chars>
func GenerateAPIKey() (string, error) {
	bytes := make([]byte, constants.APIKeyBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return constants.APIKeyPrefix + hex.EncodeToString(bytes), nil
}
