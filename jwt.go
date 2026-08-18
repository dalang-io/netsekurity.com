package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// Claims is the JWT payload for an authenticated user.
type Claims struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Exp   int64  `json:"exp"`
}

func b64url(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

// signPayload signs an already-encoded header.payload with the configured secret.
func signPayload(payload string) string {
	mac := hmac.New(sha256.New, []byte(getenv("JWT_SECRET", "dev-secret-change-me")))
	mac.Write([]byte(payload))
	return payload + "." + b64url(mac.Sum(nil))
}

// issueJWT signs an HS256 JWT with the configured JWT_SECRET.
func issueJWT(sub, email string) (string, error) {
	header, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	claims, _ := json.Marshal(Claims{Sub: sub, Email: email, Exp: time.Now().Add(24 * time.Hour).Unix()})
	return signPayload(b64url(header) + "." + b64url(claims)), nil
}

// verifyJWT validates the signature and expiry, returning the claims.
func verifyJWT(token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("malformed token")
	}
	secret := []byte(getenv("JWT_SECRET", "dev-secret-change-me"))
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(parts[0] + "." + parts[1]))
	sig := b64url(mac.Sum(nil))
	if !hmac.Equal([]byte(parts[2]), []byte(sig)) {
		return nil, errors.New("invalid signature")
	}
	pb, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	var c Claims
	if err := json.Unmarshal(pb, &c); err != nil {
		return nil, err
	}
	if c.Exp < time.Now().Unix() {
		return nil, errors.New("token expired")
	}
	return &c, nil
}
