package auth

import "testing"

func TestNewTokenForUserID_AndParseToken_Success(t *testing.T) {
	t.Parallel()

	token, err := NewTokenForUserID("123")
	if err != nil {
		t.Fatalf("NewTokenForUserID returned error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	userID, err := ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken returned error: %v", err)
	}
	if userID != "123" {
		t.Fatalf("expected userID 123, got %q", userID)
	}
}

func TestNewTokenForUserID_EmptyUserID(t *testing.T) {
	t.Parallel()

	_, err := NewTokenForUserID("")
	if err == nil {
		t.Fatal("expected error for empty user id")
	}
}

func TestParseToken_EmptyToken(t *testing.T) {
	t.Parallel()

	_, err := ParseToken("")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestParseToken_BadFormat(t *testing.T) {
	t.Parallel()

	_, err := ParseToken("not-a-valid-token")
	if err == nil {
		t.Fatal("expected error for bad token format")
	}
}

func TestParseToken_InvalidSignature(t *testing.T) {
	t.Parallel()

	token, err := NewTokenForUserID("123")
	if err != nil {
		t.Fatalf("NewTokenForUserID returned error: %v", err)
	}

	// Ломаем токен
	bad := token + "broken"

	_, err = ParseToken(bad)
	if err == nil {
		t.Fatal("expected error for invalid signature")
	}
}
