// Package auth содержит функции аутентификации и авторизации.
//
// Пакет отвечает за:
//   - генерацию токенов доступа;
//   - проверку подлинности токенов;
//   - хэширование и проверку паролей.
//
// Для защиты паролей используется алгоритм bcrypt.
package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword хэширует пароль пользователя с использованием bcrypt.
//
// Возвращает строковое представление хэша, которое может быть безопасно
// сохранено в базе данных.
//
// В случае ошибки генерации хэша возвращает ошибку.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPasswordHash сравнивает пароль в открытом виде с его bcrypt-хэшем.
//
// Возвращает nil, если пароль соответствует хэшу.
// Если пароль некорректен или хэш поврежден, возвращает ошибку.
func CheckPasswordHash(password, passwordHash string) error {
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
}
