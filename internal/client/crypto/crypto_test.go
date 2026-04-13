package crypto

import (
	"bytes"
	"testing"
)

func TestDeriveKey_SameInputSameOutput(t *testing.T) {
	t.Parallel()

	salt := []byte("0123456789abcdef")

	key1, err := DeriveKey("master-password", salt)
	if err != nil {
		t.Fatalf("DeriveKey returned error: %v", err)
	}

	key2, err := DeriveKey("master-password", salt)
	if err != nil {
		t.Fatalf("DeriveKey returned error: %v", err)
	}

	if !bytes.Equal(key1, key2) {
		t.Fatal("expected equal keys for same password and salt")
	}
}

func TestDeriveKey_DifferentSaltDifferentOutput(t *testing.T) {
	t.Parallel()

	key1, err := DeriveKey("master-password", []byte("0123456789abcdef"))
	if err != nil {
		t.Fatalf("DeriveKey returned error: %v", err)
	}

	key2, err := DeriveKey("master-password", []byte("fedcba9876543210"))
	if err != nil {
		t.Fatalf("DeriveKey returned error: %v", err)
	}

	if bytes.Equal(key1, key2) {
		t.Fatal("expected different keys for different salts")
	}
}

func TestDeriveKey_EmptyPassword(t *testing.T) {
	t.Parallel()

	_, err := DeriveKey("", []byte("0123456789abcdef"))
	if err == nil {
		t.Fatal("expected error for empty password")
	}
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	t.Parallel()

	key, err := DeriveKey("master-password", []byte("0123456789abcdef"))
	if err != nil {
		t.Fatalf("DeriveKey returned error: %v", err)
	}

	plaintext := []byte("hello secret world")

	ciphertext, nonce, err := Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}

	if bytes.Equal(ciphertext, plaintext) {
		t.Fatal("ciphertext must differ from plaintext")
	}

	decrypted, err := Decrypt(key, ciphertext, nonce)
	if err != nil {
		t.Fatalf("Decrypt returned error: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Fatalf("decrypted plaintext mismatch: got %q want %q", decrypted, plaintext)
	}
}

func TestEncrypt_GeneratesDifferentCiphertexts(t *testing.T) {
	t.Parallel()

	key, err := DeriveKey("master-password", []byte("0123456789abcdef"))
	if err != nil {
		t.Fatalf("DeriveKey returned error: %v", err)
	}

	plaintext := []byte("same plaintext")

	ciphertext1, nonce1, err := Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}

	ciphertext2, nonce2, err := Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}

	if bytes.Equal(nonce1, nonce2) {
		t.Fatal("expected different nonces")
	}

	if bytes.Equal(ciphertext1, ciphertext2) {
		t.Fatal("expected different ciphertexts with different nonces")
	}
}

func TestDecrypt_WrongKey(t *testing.T) {
	t.Parallel()

	key1, err := DeriveKey("master-password", []byte("0123456789abcdef"))
	if err != nil {
		t.Fatalf("DeriveKey returned error: %v", err)
	}

	key2, err := DeriveKey("other-password", []byte("0123456789abcdef"))
	if err != nil {
		t.Fatalf("DeriveKey returned error: %v", err)
	}

	ciphertext, nonce, err := Encrypt(key1, []byte("hello"))
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}

	_, err = Decrypt(key2, ciphertext, nonce)
	if err == nil {
		t.Fatal("expected decrypt error for wrong key")
	}
}

func TestHash_Deterministic(t *testing.T) {
	t.Parallel()

	h1 := Hash([]byte("abc"))
	h2 := Hash([]byte("abc"))

	if h1 != h2 {
		t.Fatal("expected deterministic hash")
	}
}
