// Package auth содержит функции аутентификации и авторизации.
//
// Пакет отвечает за:
//   - генерацию токенов доступа;
//   - проверку подлинности токенов;
//   - обеспечение целостности данных с помощью HMAC.
//
// Используется симметричная подпись на основе SHA-256.
// Секретный ключ задается через переменную окружения SECRETKEY.
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"strings"
)

// getEnvOrDefault возвращает значение переменной окружения.
//
// Если переменная не задана, возвращается значение по умолчанию.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// secretKey возвращает секретный ключ для подписи токенов.
//
// Ключ берется из переменной окружения SECRETKEY.
// Если переменная не задана, используется значение "debugkey".
func secretKey() string {
	return getEnvOrDefault("SECRETKEY", "debugkey")
}

// signUserID создает токен на основе userID.
//
// Формирует строку вида:
//
//	userID:signature
//
// Где signature — HMAC-SHA256 подпись userID.
//
// Возвращает ошибку, если userID пустой.
func signUserID(userID string) (string, error) {
	if userID == "" {
		return "", errors.New("empty user id")
	}

	mac := hmac.New(sha256.New, []byte(secretKey()))
	mac.Write([]byte(userID))
	sig := mac.Sum(nil)

	return userID + ":" + hex.EncodeToString(sig), nil
}

// parseToken разбирает и проверяет токен.
//
// Проверяет:
//   - формат токена;
//   - наличие userID;
//   - корректность подписи.
//
// Возвращает userID, если токен валиден.
//
// Возможные ошибки:
//   - пустой токен;
//   - некорректный формат;
//   - ошибка декодирования подписи;
//   - неверная подпись.
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

// NewTokenForUserID создает новый токен для пользователя.
//
// Является публичной оберткой над signUserID.
// Используется сервисным слоем для выдачи токена после логина.
func NewTokenForUserID(userID string) (string, error) {
	return signUserID(userID)
}

// ParseToken выполняет разбор и проверку токена.
//
// Возвращает userID, если токен валиден.
// Используется на сервере для аутентификации запросов.
func ParseToken(token string) (string, error) {
	return parseToken(token)
}
