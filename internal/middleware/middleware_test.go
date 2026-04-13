package middleware

import "testing"

func TestInitLogger(t *testing.T) {
	logger := InitLogger()
	if logger == nil {
		t.Fatal("expected logger")
	}
}
