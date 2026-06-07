package auth

import "testing"

func TestHashPassword_AndCheckPasswordHash_Success(t *testing.T) {
	t.Parallel()

	hash, err := HashPassword("secret123")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}

	err = CheckPasswordHash("secret123", hash)
	if err != nil {
		t.Fatalf("expected password hash to match, got error: %v", err)
	}
}

func TestCheckPasswordHash_WrongPassword(t *testing.T) {
	t.Parallel()

	hash, err := HashPassword("secret123")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	err = CheckPasswordHash("wrong-password", hash)
	if err == nil {
		t.Fatal("expected password hash mismatch error")
	}
}

func TestHashPassword_EmptyPassword(t *testing.T) {
	t.Parallel()

	hash, err := HashPassword("")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "" {
		t.Fatal("expected hash even for empty password")
	}
}
