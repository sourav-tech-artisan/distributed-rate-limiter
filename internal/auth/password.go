package auth

import (
	"golang.org/x/crypto/bcrypt"
)

const (
	// bcryptCost is the cost factor for bcrypt hashing
	// 12 is a good balance between security and performance
	bcryptCost = 12
)

// HashPassword hashes a plaintext password using bcrypt
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
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
