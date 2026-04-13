// Package crypto содержит функции клиентского шифрования данных.
//
// Пакет отвечает за:
//   - вывод симметричного ключа из мастер-пароля;
//   - шифрование полезной нагрузки;
//   - расшифровку полезной нагрузки;
//   - вычисление хэша данных.
//
// Для вывода ключа используется scrypt.
// Для шифрования используется алгоритм AES-GCM.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/scrypt"
)

const (
	// keyLen определяет длину симметричного ключа в байтах.
	keyLen = 32
	// nonceSize определяет длину nonce для AES-GCM в байтах.
	nonceSize = 12
)

// DeriveKey выводит 32-байтный симметричный ключ из мастер-пароля и соли.
//
// Функция используется для получения ключа, которым затем шифруются
// и расшифровываются пользовательские данные на стороне клиента.
//
// Возвращает ошибку, если мастер-пароль пустой, соль отсутствует
// или не удалось выполнить вывод ключа.
func DeriveKey(masterPassword string, salt []byte) ([]byte, error) {
	if masterPassword == "" {
		return nil, fmt.Errorf("master password is empty")
	}
	if len(salt) == 0 {
		return nil, fmt.Errorf("salt is empty")
	}

	key, err := scrypt.Key([]byte(masterPassword), salt, 1<<15, 8, 1, keyLen)
	if err != nil {
		return nil, fmt.Errorf("derive key with scrypt: %w", err)
	}

	return key, nil
}

// Encrypt шифрует открытый текст с помощью AES-GCM.
//
// Функция принимает симметричный ключ и данные в открытом виде,
// затем возвращает шифртекст и nonce, необходимый для последующей
// расшифровки.
//
// Возвращает ошибку, если ключ имеет неверную длину, данные пусты
// или не удалось создать криптографические примитивы.
func Encrypt(key, plaintext []byte) ([]byte, []byte, error) {
	if len(key) != keyLen {
		return nil, nil, fmt.Errorf("invalid key length: got %d want %d", len(key), keyLen)
	}
	if len(plaintext) == 0 {
		return nil, nil, fmt.Errorf("plaintext is empty")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, fmt.Errorf("create aes cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("create gcm: %w", err)
	}

	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := aead.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

// Decrypt расшифровывает данные, зашифрованные с помощью AES-GCM.
//
// Для успешной расшифровки необходимо передать тот же ключ и nonce,
// которые использовались при шифровании.
//
// Возвращает ошибку, если ключ или nonce имеют неверный размер,
// если шифртекст пустой или если расшифровка не удалась.
func Decrypt(key, ciphertext, nonce []byte) ([]byte, error) {
	if len(key) != keyLen {
		return nil, fmt.Errorf("invalid key length: got %d want %d", len(key), keyLen)
	}
	if len(ciphertext) == 0 {
		return nil, fmt.Errorf("ciphertext is empty")
	}
	if len(nonce) != nonceSize {
		return nil, fmt.Errorf("invalid nonce length: got %d want %d", len(nonce), nonceSize)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create aes cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}

	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt payload: %w", err)
	}

	return plaintext, nil
}

// Hash вычисляет SHA-256 хэш для переданных данных.
//
// Хэш может использоваться для контроля целостности, обнаружения изменений
// и как вспомогательное значение для дедупликации полезной нагрузки.
//
// Возвращает хэш в шестнадцатеричном строковом формате.
func Hash(data []byte) string {
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum[:])
}
