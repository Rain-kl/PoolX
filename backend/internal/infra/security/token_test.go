package security

import (
	"testing"
	"time"
)

func TestAccessTokenRoundTrip(t *testing.T) {
	service := NewTokenService("12345678901234567890123456789012")
	token, expiresAt, err := service.CreateAccessToken(7, 11, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Fatal("expected non-empty access token")
	}
	if expiresAt.Before(time.Now().UTC()) {
		t.Fatalf("expiresAt is in the past: %s", expiresAt)
	}
	identity, err := service.ParseAccessToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if identity.AdminID != 7 || identity.SessionID != 11 {
		t.Fatalf("identity = %#v", identity)
	}
	if _, err := service.ParseAccessToken("not-a-jwt"); err == nil {
		t.Fatal("expected invalid token to fail")
	}
}

func TestOpaqueAndHashToken(t *testing.T) {
	raw, err := NewOpaqueToken(16)
	if err != nil {
		t.Fatal(err)
	}
	if raw == "" {
		t.Fatal("expected opaque token")
	}
	hexValue, err := NewHexToken(8)
	if err != nil {
		t.Fatal(err)
	}
	if len(hexValue) != 16 {
		t.Fatalf("hex token length = %d", len(hexValue))
	}
	if HashToken(raw) == raw || HashToken(raw) == "" {
		t.Fatal("expected hashed token digest")
	}
	if HashToken(raw) != HashToken(raw) {
		t.Fatal("hash should be deterministic")
	}
}
