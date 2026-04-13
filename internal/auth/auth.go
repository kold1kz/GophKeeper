package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"strings"
)

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func secretKey() string {
	return getEnvOrDefault("SECRETKEY", "debugkey")
}

func signUserID(userID string) (string, error) {
	if userID == "" {
		return "", errors.New("empty user id")
	}

	mac := hmac.New(sha256.New, []byte(secretKey()))
	mac.Write([]byte(userID))
	sig := mac.Sum(nil)

	return userID + ":" + hex.EncodeToString(sig), nil
}

func parseToken(token string) (string, error) {
	if token == "" {
		return "", errors.New("empty token")
	}

	parts := strings.Split(token, ":")
	if len(parts) != 2 {
		return "", errors.New("bad token format")
	}

	userID := parts[0]
	if userID == "" {
		return "", errors.New("empty user id")
	}

	sigBytes, err := hex.DecodeString(parts[1])
	if err != nil {
		return "", err
	}

	mac := hmac.New(sha256.New, []byte(secretKey()))
	mac.Write([]byte(userID))
	expected := mac.Sum(nil)

	if !hmac.Equal(sigBytes, expected) {
		return "", errors.New("invalid signature")
	}

	return userID, nil
}

func NewTokenForUserID(userID string) (string, error) {
	return signUserID(userID)
}

func ParseToken(token string) (string, error) {
	return parseToken(token)
}
