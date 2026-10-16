package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// A pricing click must survive the Google round trip.
func TestSafeNext(t *testing.T) {
	ok := map[string]string{
		"/checkout?package=starter": "/checkout?package=starter",
		"/dashboard":                "/dashboard",
		"  /docs ":                  "/docs",
	}
	for in, want := range ok {
		if got := safeNext(in); got != want {
			t.Errorf("safeNext(%q) = %q, want %q", in, got, want)
		}
	}
	for _, bad := range []string{"", "https://evil.example", "//evil.example", "/\\evil.example", "/login", "/auth/google", "/logout", "/x\r\nSet-Cookie: a=b", "dashboard"} {
		if got := safeNext(bad); got != "" {
			t.Errorf("safeNext(%q) = %q, want rejection", bad, got)
		}
	}
}

func TestRequireAuthCarriesNext(t *testing.T) {
	h := requireAuth(func(w http.ResponseWriter, r *http.Request) { t.Fatal("must not reach handler") })
	rr := httptest.NewRecorder()
	h(rr, httptest.NewRequest("GET", "/checkout?package=professional", nil))
	if rr.Code != http.StatusFound {
		t.Fatalf("status %d", rr.Code)
	}
	loc := rr.Header().Get("Location")
	u, err := url.Parse(loc)
	if err != nil || u.Path != "/login" || u.Query().Get("next") != "/checkout?package=professional" {
		t.Errorf("redirect = %q, want /login?next=/checkout?package=professional", loc)
	}
	if rr.Header().Get("HX-Redirect") != loc {
		t.Errorf("HX-Redirect %q differs from Location %q", rr.Header().Get("HX-Redirect"), loc)
	}
}

func TestConsumeNextClearsCookie(t *testing.T) {
	r := httptest.NewRequest("GET", "/auth/google/callback", nil)
	r.AddCookie(&http.Cookie{Name: nextCookie, Value: "/checkout?package=starter"})
	rr := httptest.NewRecorder()
	if got := consumeNext(rr, r); got != "/checkout?package=starter" {
		t.Errorf("consumeNext = %q", got)
	}
	cleared := false
	for _, c := range rr.Result().Cookies() {
		if c.Name == nextCookie && c.MaxAge < 0 {
			cleared = true
		}
	}
	if !cleared {
		t.Error("next cookie not cleared")
	}
	// No cookie, or a hostile one, falls back to the dashboard.
	r2 := httptest.NewRequest("GET", "/", nil)
	r2.AddCookie(&http.Cookie{Name: nextCookie, Value: "https://evil.example"})
	if got := consumeNext(httptest.NewRecorder(), r2); got != "/dashboard" {
		t.Errorf("hostile next = %q, want /dashboard", got)
	}
}

func TestCallbackRejectsBadState(t *testing.T) {
	rr := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/auth/google/callback?code=abc&state=zzz", nil)
	r.AddCookie(&http.Cookie{Name: "nsk_oauth_state", Value: "expected"})
	handleGoogleCallback(rr, r)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status %d, want 400 on state mismatch", rr.Code)
	}
}

func TestRateLimiter(t *testing.T) {
	rl := newRateLimiter(3, time.Minute)
	now := time.Now()
	for i := 0; i < 3; i++ {
		if !rl.allowAt("k", now) {
			t.Fatalf("call %d should pass", i+1)
		}
	}
	if rl.allowAt("k", now) {
		t.Error("4th call should be refused")
	}
	if !rl.allowAt("other", now) {
		t.Error("other key must be independent")
	}
	if !rl.allowAt("k", now.Add(61*time.Second)) {
		t.Error("window should reset")
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	rl := newRateLimiter(1, time.Minute)
	hits := 0
	h := rateLimit(rl, func(w http.ResponseWriter, r *http.Request) { hits++ })
	post := func() int {
		rr := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api/domains", nil)
		r.RemoteAddr = "203.0.113.5:1234"
		h(rr, r)
		return rr.Code
	}
	if post() != 200 || post() != http.StatusTooManyRequests {
		t.Errorf("second POST should be 429")
	}
	// GET is never limited.
	rr := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/domains", nil)
	r.RemoteAddr = "203.0.113.5:1234"
	h(rr, r)
	if rr.Code != 200 || hits != 2 {
		t.Errorf("GET was throttled (code %d, hits %d)", rr.Code, hits)
	}
}

// The 2-credit destructive mode was the same scan at double the price.
func TestScanModeStandardOnly(t *testing.T) {
	for _, in := range []string{"", "standard", " Standard "} {
		if m, c := scanMode(in); m != "standard" || c != 1 {
			t.Errorf("scanMode(%q) = %q,%v", in, m, c)
		}
	}
	for _, in := range []string{"destructive", "deep", "x"} {
		if m, _ := scanMode(in); m != "" {
			t.Errorf("scanMode(%q) accepted", in)
		}
	}
}

func TestAPIRejectsDestructive(t *testing.T) {
	const uid = "u_api_mode_test"
	db.Exec(`INSERT OR IGNORE INTO users (id, email, name) VALUES (?, 'apimode@test.local', 'A')`, uid)
	secret := generateTokenSecret("nsk_live")
	db.Exec(`INSERT INTO api_tokens (id, user_id, name, token_hash, prefix, expires_at) VALUES (?,?,?,?,?,?)`,
		"tk_mode", uid, "t", tokenHash(secret), secret[:16], time.Now().Add(time.Hour).UTC().Format(time.RFC3339))

	rr := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/v1/pentests", strings.NewReader("domain=example.com&mode=destructive"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("X-API-Token", secret)
	handleAPICreatePentest(rr, r)
	if rr.Code != http.StatusBadRequest || !strings.Contains(rr.Body.String(), "standard only") {
		t.Errorf("destructive accepted: %d %s", rr.Code, rr.Body.String())
	}

	// Query-string tokens are no longer honoured.
	rr = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/api/v1/pentests?api_token="+secret, strings.NewReader("domain=example.com"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	handleAPICreatePentest(rr, r)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("query-string token still accepted: %d", rr.Code)
	}
}

func TestContactPersists(t *testing.T) {
	form := url.Values{"name": {"Ada"}, "email": {"ADA@Example.com"}, "message": {"wildcard pricing?"}, "source": {"landing"}}
	rr := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/contact", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	handleContact(rr, r)
	if rr.Code != 200 || !strings.Contains(rr.Body.String(), "Message received") {
		t.Fatalf("contact: %d %s", rr.Code, rr.Body.String())
	}
	var n int
	db.QueryRow(`SELECT COUNT(*) FROM contact_messages WHERE email='ada@example.com' AND message='wildcard pricing?'`).Scan(&n)
	if n != 1 {
		t.Errorf("message not stored (n=%d)", n)
	}

	// Honeypot filled: pretend success, store nothing.
	form.Set("website", "http://spam")
	form.Set("email", "bot@example.com")
	rr = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/contact", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	handleContact(rr, r)
	db.QueryRow(`SELECT COUNT(*) FROM contact_messages WHERE email='bot@example.com'`).Scan(&n)
	if rr.Code != 200 || n != 0 {
		t.Errorf("honeypot: code %d stored %d", rr.Code, n)
	}

	// Missing message is a 400, not a silent success.
	rr = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/contact", strings.NewReader("email=x@y.z"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	handleContact(rr, r)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("empty message accepted: %d", rr.Code)
	}
}

func TestWebhookRequiresCallbackKey(t *testing.T) {
	env["XENDIT_CALLBACK_KEY"] = "secret-key"
	defer delete(env, "XENDIT_CALLBACK_KEY")
	for _, tok := range []string{"", "wrong", "secret-ke"} {
		rr := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/webhook/xendit", strings.NewReader(`{"external_id":"x","status":"PAID"}`))
		r.Header.Set("x-callback-token", tok)
		handleXenditWebhook(rr, r)
		if rr.Code != http.StatusUnauthorized {
			t.Errorf("token %q accepted (%d)", tok, rr.Code)
		}
	}
	// An unset key must never match an empty header.
	delete(env, "XENDIT_CALLBACK_KEY")
	rr := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/webhook/xendit", strings.NewReader(`{}`))
	handleXenditWebhook(rr, r)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("empty key accepted empty header (%d)", rr.Code)
	}
}

func TestAnalyticsSnippet(t *testing.T) {
	delete(env, "PLAUSIBLE_DOMAIN")
	delete(env, "GA_MEASUREMENT_ID")
	if got := analyticsSnippet(); got != "" {
		t.Errorf("no config should mean no tag, got %q", got)
	}
	env["PLAUSIBLE_DOMAIN"] = "netsekurity.com"
	defer delete(env, "PLAUSIBLE_DOMAIN")
	if got := analyticsSnippet(); !strings.Contains(got, `data-domain="netsekurity.com"`) {
		t.Errorf("plausible tag missing: %q", got)
	}
}

// Public pages must carry the claims the scanner can back, and none it cannot.
func TestPublicPagesClaims(t *testing.T) {
	pages := map[string]func(http.ResponseWriter, *http.Request){
		"/":              handleIndex,
		"/sample-report": handleSampleReport,
		"/docs":          handleDocs,
		"/product-scope": handleProductScope,
	}
	banned := []string{"Pentest as a Service", "PTaaS", "destructive mode", "60+", "SQL injection on search", "anyone = admin"}
	for path, h := range pages {
		rr := httptest.NewRecorder()
		h(rr, httptest.NewRequest("GET", path, nil))
		if rr.Code != 200 {
			t.Errorf("%s: status %d", path, rr.Code)
		}
		body := rr.Body.String()
		for _, b := range banned {
			if strings.Contains(body, b) {
				t.Errorf("%s still says %q", path, b)
			}
		}
	}
	rr := httptest.NewRecorder()
	handleIndex(rr, httptest.NewRequest("GET", "/", nil))
	body := rr.Body.String()
	for _, want := range []string{`href="/checkout?package=starter"`, `href="/checkout?package=professional"`, `href="/sample-report"`, `hx-post="/contact"`, "whoami", "expert review"} {
		if !strings.Contains(body, want) {
			t.Errorf("landing missing %q", want)
		}
	}
	if strings.Contains(body, "__SCANS_PROOF__") || strings.Contains(body, "__ANALYTICS__") {
		t.Error("landing has an unrendered placeholder")
	}
}

func TestScansProofThreshold(t *testing.T) {
	if scansCompleted() >= scansProofMin {
		t.Skip("test db unexpectedly has many completed scans")
	}
	if got := renderScansProof(); got != "" {
		t.Errorf("proof line rendered below threshold: %q", got)
	}
}
