package auth

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/souravkumar/distributed-rate-limiter/internal/common"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a plaintext password using bcrypt
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), common.BcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword compares a plaintext password with a bcrypt hash
// Returns true if the password matches, false otherwise
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateAPIKey creates a new random API key with the format: rl_<64 hex chars>
func GenerateAPIKey() (string, error) {
	bytes := make([]byte, common.APIKeyBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return common.APIKeyPrefix + hex.EncodeToString(bytes), nil
}
