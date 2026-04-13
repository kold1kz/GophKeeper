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
	keyLen    = 32
	nonceSize = 12
)

// DeriveKey derives a 32-byte symmetric key from a master password and salt.
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

// Encrypt encrypts plaintext using AES-GCM and returns ciphertext and nonce.
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

// Decrypt decrypts ciphertext using AES-GCM.
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

// Hash returns sha256 hex-like bytes source for payload integrity/dedup basis.
func Hash(data []byte) string {
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum[:])
}
