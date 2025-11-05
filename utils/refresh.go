package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// returns the plain refresh token and its hashed version
func GenerateRefreshToken() (string, string, error) {
	// Tạo một chuỗi ngẫu nhiên
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", "", err
	}
	plainToken := base64.RawURLEncoding.EncodeToString(randomBytes)
	hashedToken := HashRefreshToken(plainToken)
	return plainToken, hashedToken, nil
}

func HashRefreshToken(plainToken string) string {
	sum := sha256.Sum256([]byte(plainToken))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
