package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// The CI/CD API: start an assessment, then read its status. Both are
// authenticated by X-API-Token and both answer only for the token's own account.

func apiFixture(t *testing.T, uid string) (token string, mux *http.ServeMux) {
	t.Helper()
	db.Exec(`INSERT OR IGNORE INTO users (id, email, name) VALUES (?, ?, 'A')`, uid, uid+"@test.local")
	// The scan tests elsewhere claim "the next queued scan": leave none behind.
	t.Cleanup(func() {
		db.Exec(`DELETE FROM pentests WHERE user_id=?`, uid)
		db.Exec(`DELETE FROM domains WHERE user_id=?`, uid)
		db.Exec(`DELETE FROM api_tokens WHERE user_id=?`, uid)
		db.Exec(`DELETE FROM credit_balance WHERE user_id=?`, uid)
		db.Exec(`DELETE FROM credit_transactions WHERE user_id=?`, uid)
	})
	token = generateTokenSecret("nsk_live")
	db.Exec(`INSERT INTO api_tokens (id, user_id, name, token_hash, prefix, expires_at) VALUES (?,?,?,?,?,?)`,
		"tk_"+uid, uid, "t", tokenHash(token), token[:16], time.Now().Add(time.Hour).UTC().Format(time.RFC3339))
	mux = http.NewServeMux()
	registerAPIv1(mux, newRateLimiter(100, time.Minute), newRateLimiter(100, time.Minute))
	return token, mux
}

func apiCall(mux *http.ServeMux, method, path, token, body string) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.RemoteAddr = "203.0.113.9:4000"
	if body != "" {
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if token != "" {
		r.Header.Set("X-API-Token", token)
	}
	mux.ServeHTTP(rr, r)
	return rr
}

func decode(t *testing.T, rr *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &m); err != nil {
		t.Fatalf("not JSON: %q (%v)", rr.Body.String(), err)
	}
	return m
}

func TestAssessmentsIsTheNameAndPentestsStillWorks(t *testing.T) {
	const uid = "u_api_alias"
	token, mux := apiFixture(t, uid)
	db.Exec(`INSERT OR REPLACE INTO credit_balance (user_id, balance) VALUES (?, 5)`, uid)
	db.Exec(`INSERT INTO domains (user_id, domain, txt_verification_token, status) VALUES (?, 'alias.example.com', 'x', 'verified')`, uid)

	rr := apiCall(mux, "POST", "/api/v1/assessments", token, "domain=alias.example.com")
	if rr.Code != 200 {
		t.Fatalf("assessments: %d %s", rr.Code, rr.Body)
	}
	m := decode(t, rr)
	id, _ := m["assessment_id"].(string)
	if id == "" || m["pentest_id"] != id || m["status"] != "queued" {
		t.Errorf("response = %v; want assessment_id == pentest_id, status queued", m)
	}

	// One at a time per account: the second start is a 409 JSON error, not a 200.
	rr = apiCall(mux, "POST", "/api/v1/pentests", token, "domain=alias.example.com")
	if rr.Code != http.StatusConflict {
		t.Errorf("second start while queued: %d %s, want 409", rr.Code, rr.Body)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("409 content type = %q", ct)
	}
	var bal float64
	db.QueryRow(`SELECT balance FROM credit_balance WHERE user_id=?`, uid).Scan(&bal)
	if bal != 4 {
		t.Errorf("balance = %v, want 4: the refused start must not cost a credit", bal)
	}
}

func TestAssessmentStatusAndFindings(t *testing.T) {
	const uid = "u_api_status"
	token, mux := apiFixture(t, uid)
	res, _ := db.Exec(`INSERT INTO domains (user_id, domain, txt_verification_token, status) VALUES (?, 'status.example.com', 'x', 'verified')`, uid)
	did, _ := res.LastInsertId()

	cases := []struct {
		id, status string
		reported   int
		wantFind   bool
	}{
		{"pt_s_queued", "queued", 0, false},
		{"pt_s_running", "running", 0, false},
		{"pt_s_done", "completed", 1, true},
		{"pt_s_done_nobreakdown", "completed", 0, false}, // report delivered without counts
		{"pt_s_failed", "failed", 0, false},
	}
	for _, c := range cases {
		db.Exec(`INSERT INTO pentests (id, user_id, domain_id, status, findings_reported, findings_critical, findings_high, findings_medium, findings_low, findings_info,
			started_at, completed_at) VALUES (?,?,?,?,?,0,1,2,2,4,'2026-10-26 09:00:00','2026-10-26T09:14:00Z')`, c.id, uid, did, c.status, c.reported)
		rr := apiCall(mux, "GET", "/api/v1/assessments/"+c.id, token, "")
		if rr.Code != 200 {
			t.Fatalf("%s: %d %s", c.id, rr.Code, rr.Body)
		}
		m := decode(t, rr)
		if m["status"] != c.status || m["domain"] != "status.example.com" {
			t.Errorf("%s: %v", c.id, m)
		}
		f, has := m["findings"].(map[string]any)
		if has != c.wantFind {
			t.Errorf("%s: findings present=%v, want %v (%v)", c.id, has, c.wantFind, m["findings"])
		}
		if c.wantFind && (f["high"] != float64(1) || f["medium"] != float64(2) || f["info"] != float64(4) || f["critical"] != float64(0)) {
			t.Errorf("%s: findings %v", c.id, f)
		}
		// Both stored formats come out as RFC3339.
		if m["started_at"] != "2026-10-26T09:00:00Z" || m["completed_at"] != "2026-10-26T09:14:00Z" {
			t.Errorf("%s: times %v %v", c.id, m["started_at"], m["completed_at"])
		}
		if rr.Header().Get("Cache-Control") != "no-store" {
			t.Errorf("%s: result must not be cached", c.id)
		}
	}
}

// Another account's assessment is indistinguishable from one that does not exist.
func TestAssessmentStatusIsOwnerOnly(t *testing.T) {
	tokenA, mux := apiFixture(t, "u_api_a")
	_, _ = apiFixture(t, "u_api_b")
	res, _ := db.Exec(`INSERT INTO domains (user_id, domain, txt_verification_token, status) VALUES ('u_api_b', 'b.example.com', 'x', 'verified')`)
	did, _ := res.LastInsertId()
	db.Exec(`INSERT INTO pentests (id, user_id, domain_id, status) VALUES ('pt_b_secret', 'u_api_b', ?, 'completed')`, did)

	other := apiCall(mux, "GET", "/api/v1/assessments/pt_b_secret", tokenA, "")
	missing := apiCall(mux, "GET", "/api/v1/assessments/pt_does_not_exist", tokenA, "")
	if other.Code != 404 || missing.Code != 404 || other.Body.String() != missing.Body.String() {
		t.Errorf("other account: %d %q, missing: %d %q", other.Code, other.Body, missing.Code, missing.Body)
	}
	if strings.Contains(other.Body.String(), "b.example.com") {
		t.Error("another account's domain leaked")
	}
}

func TestAssessmentStatusNeedsAToken(t *testing.T) {
	_, mux := apiFixture(t, "u_api_auth")
	for _, tok := range []string{"", "nsk_wrong", "not-a-token"} {
		if rr := apiCall(mux, "GET", "/api/v1/assessments/pt_x", tok, ""); rr.Code != http.StatusUnauthorized {
			t.Errorf("token %q: %d, want 401", tok, rr.Code)
		}
	}
	// A token in the URL is not honoured, as for the POST.
	token, mux := apiFixture(t, "u_api_auth2")
	rr := apiCall(mux, "GET", "/api/v1/assessments/pt_x?api_token="+token, "", "")
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("query-string token: %d, want 401", rr.Code)
	}
	if rr := apiCall(mux, "POST", "/api/v1/assessments/pt_x", token, "x=1"); rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST to a status URL: %d, want 405", rr.Code)
	}
	if rr := apiCall(mux, "GET", "/api/v1/assessments/", token, ""); rr.Code != http.StatusNotFound {
		t.Errorf("empty id: %d, want 404", rr.Code)
	}
}

func TestAssessmentPollingIsRateLimitedPerClient(t *testing.T) {
	token, _ := apiFixture(t, "u_api_rl")
	mux := http.NewServeMux()
	registerAPIv1(mux, newRateLimiter(100, time.Minute), newRateLimiter(3, time.Minute))
	got := make([]int, 0, 5)
	for i := 0; i < 5; i++ {
		// A different id each time: the bucket is per client, not per URL.
		got = append(got, apiCall(mux, "GET", "/api/v1/assessments/pt_rl_"+itoa(i), token, "").Code)
	}
	if got[3] != http.StatusTooManyRequests || got[4] != http.StatusTooManyRequests || got[0] == http.StatusTooManyRequests {
		t.Errorf("codes = %v, want the 4th and 5th to be 429", got)
	}
}
