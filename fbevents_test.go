package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// fbToken must be read lazily: package-level init runs before main() calls
// loadEnv, so a var-initialised token would never see a value from .env.
func TestFBTokenReadsEnvLoadedAfterInit(t *testing.T) {
	old, had := env["META_CAPI_TOKEN"]
	defer func() {
		if had {
			env["META_CAPI_TOKEN"] = old
		} else {
			delete(env, "META_CAPI_TOKEN")
		}
	}()
	env["META_CAPI_TOKEN"] = "tok-from-dotenv"
	if got := fbToken(); got != "tok-from-dotenv" {
		t.Errorf("fbToken() = %q, want the value loadEnv put in the env map", got)
	}
}

// Meta rejects a Conversions API call whose event carries no user identifier.
func TestMetaUserPayloadCarriesIdentifiers(t *testing.T) {
	r := httptest.NewRequest("GET", "/auth/google/callback", nil)
	r.Header.Set("CF-Connecting-IP", "203.0.113.9")
	r.Header.Set("User-Agent", "Mozilla/5.0")
	r.AddCookie(&http.Cookie{Name: "_fbp", Value: "fb.1.123.456"})

	p := metaUserFrom(r, " User@Example.COM ").payload()
	if p["client_ip_address"] != "203.0.113.9" {
		t.Errorf("client_ip_address = %v, want the CF-Connecting-IP value", p["client_ip_address"])
	}
	if p["client_user_agent"] != "Mozilla/5.0" {
		t.Errorf("client_user_agent = %v", p["client_user_agent"])
	}
	if p["fbp"] != "fb.1.123.456" {
		t.Errorf("fbp = %v", p["fbp"])
	}
	em, ok := p["em"].([]string)
	if !ok || len(em) != 1 {
		t.Fatalf("em = %v, want a one-element hash slice", p["em"])
	}
	// Lower-cased and trimmed before hashing, as Meta requires.
	if em[0] != fbHash("user@example.com") || len(em[0]) != 64 {
		t.Errorf("em[0] = %q, want the sha256 of the normalised email", em[0])
	}
	if strings.Contains(em[0], "@") {
		t.Error("email was sent unhashed")
	}
}

func TestMetaUserPayloadEmptyWithoutIdentifiers(t *testing.T) {
	if p := (metaUser{}).payload(); len(p) != 0 {
		t.Errorf("payload = %v, want empty so sendMetaEvent can skip the call", p)
	}
}

func TestClientIPPrefersProxyHeaders(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.1:5555"
	if got := clientIP(r); got != "10.0.0.1" {
		t.Errorf("clientIP with no headers = %q, want the RemoteAddr host", got)
	}
	r.Header.Set("X-Forwarded-For", "198.51.100.7, 10.0.0.1")
	if got := clientIP(r); got != "198.51.100.7" {
		t.Errorf("clientIP = %q, want the first X-Forwarded-For hop", got)
	}
	r.Header.Set("CF-Connecting-IP", "203.0.113.4")
	if got := clientIP(r); got != "203.0.113.4" {
		t.Errorf("clientIP = %q, want CF-Connecting-IP to win", got)
	}
}
