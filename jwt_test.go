package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestJWT_RoundTrip(t *testing.T) {
	tok, err := issueJWT("u_abc", "a@example.com")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	c, err := verifyJWT(tok)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if c.Sub != "u_abc" || c.Email != "a@example.com" {
		t.Errorf("claims mismatch: %+v", c)
	}
	if c.Exp <= time.Now().Unix() {
		t.Errorf("expiry should be in the future")
	}
}

func TestJWT_TamperedSignatureRejected(t *testing.T) {
	tok, _ := issueJWT("u_abc", "a@example.com")
	parts := strings.Split(tok, ".")
	b := []byte(parts[1])
	if b[0] == 'A' {
		b[0] = 'B'
	} else {
		b[0] = 'A'
	}
	tampered := parts[0] + "." + string(b) + "." + parts[2]
	if _, err := verifyJWT(tampered); err == nil {
		t.Errorf("expected error for tampered token")
	}
}

func TestJWT_MalformedRejected(t *testing.T) {
	if _, err := verifyJWT("not.a.jwt"); err == nil {
		t.Errorf("expected error for malformed token")
	}
}

// mustExpiredToken builds a signed JWT whose exp is already in the past.
func mustExpiredToken(sub, email string) string {
	header, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	claims, _ := json.Marshal(Claims{Sub: sub, Email: email, Exp: time.Now().Add(-time.Hour).Unix()})
	return signPayload(b64url(header) + "." + b64url(claims))
}

func TestJWT_ExpiredRejected(t *testing.T) {
	tok := mustExpiredToken("u", "e@x.com")
	if _, err := verifyJWT(tok); err == nil {
		t.Errorf("expected error for expired token")
	}
}
