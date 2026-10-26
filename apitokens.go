package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"
)

// tokenHash returns a SHA-256 hex of a token (we only store the hash, never the plaintext).
func tokenHash(tok string) string {
	h := sha256.Sum256([]byte(tok))
	return hex.EncodeToString(h[:])
}

// generateTokenSecret returns a random URL-safe token with the given prefix,
// e.g. "nsk_live_abc...". The plaintext is returned exactly once at creation.
func generateTokenSecret(prefix string) string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(b))
}

// validAPIExpiry returns true if expiryDays is one of the allowed options.
func validAPIExpiry(expiryDays int) bool {
	switch expiryDays {
	case 7, 14, 30, 60, 90:
		return true
	}
	return false
}

// requireAPIToken resolves an API token (via "X-API-Token") to a user id, or writes
// 401 and returns false. Also touches last_used_at and flags the plaintext log out.
func requireAPIToken(w http.ResponseWriter, r *http.Request) (string, bool) {
	// Header only. A query-string fallback used to exist; tokens in URLs end up
	// in access logs, browser history and Referer headers.
	tok := strings.TrimSpace(r.Header.Get("X-API-Token"))
	if tok == "" || !strings.HasPrefix(tok, "nsk_") {
		http.Error(w, `{"error":"missing or invalid API token"}`, http.StatusUnauthorized)
		return "", false
	}
	hash := tokenHash(tok)
	var userID, expiresAt string
	err := db.QueryRow(`SELECT user_id, expires_at FROM api_tokens WHERE token_hash=?`, hash).Scan(&userID, &expiresAt)
	if err != nil {
		http.Error(w, `{"error":"invalid API token"}`, http.StatusUnauthorized)
		return "", false
	}
	// expiry check: expires_at stored as RFC3339 via time.Unix; compare with now.
	exp, _ := time.Parse(time.RFC3339, expiresAt)
	if exp.Before(time.Now()) {
		http.Error(w, `{"error":"API token expired"}`, http.StatusUnauthorized)
		return "", false
	}
	db.Exec(`UPDATE api_tokens SET last_used_at=? WHERE token_hash=?`, now(), hash)
	return userID, true
}

// handleTokensCreate lets an authenticated user generate an API token.
// POST /api/tokens  form: name (optional), expiry_days (7|14|30|60|90)
func handleTokensCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	claims, err := currentUser(r)
	if err != nil {
		w.Header().Set("HX-Redirect", "/login")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		name = "CI/CD"
	}
	expiryDays := parseFloat(r.FormValue("expiry_days"))
	if !validAPIExpiry(int(expiryDays)) {
		http.Error(w, "invalid expiry_days (choose 7, 14, 30, 60, or 90)", http.StatusBadRequest)
		return
	}
	secret := generateTokenSecret("nsk_live")
	prefix := secret[:16]
	id := "tk_" + randomHex(8)
	expires := time.Now().Add(time.Duration(int(expiryDays)) * 24 * time.Hour).UTC().Format(time.RFC3339)
	_, err = db.Exec(`INSERT INTO api_tokens (id, user_id, name, token_hash, prefix, expires_at) VALUES (?,?,?,?,?,?)`,
		id, claims.Sub, name, tokenHash(secret), prefix, expires)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	// Return the plaintext ONCE (HTMX renders it so the user can copy it).
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// The list on the page listens for this, so the new token appears at once
	// instead of after the next 15-second poll.
	w.Header().Set("HX-Trigger", "tokens-changed")
	fmt.Fprint(w, alertHTML("ok", `<strong>Token created. Copy it now — it is shown only once.</strong>
  <div class="nsk-code mt-2 bg-white"><span>`+template.HTMLEscapeString(secret)+`</span>
    <button type="button" data-copy="`+template.HTMLEscapeString(secret)+`" onclick="nskCopy(this)" class="nsk-btn nsk-btn-ghost nsk-btn-sm">`+string(icon("copy", 14))+`<span class="nsk-copy-label">Copy</span></button></div>
  <p class="mt-2 text-sm">Name: `+template.HTMLEscapeString(name)+` · valid until `+string(timeTag(expires))+`.
  Keep it somewhere safe, such as your pipeline's secret settings, and send it as the <code>X-API-Token</code> header.</p>`))
}

// handleTokensList shows the user's existing API tokens (masked).
// GET /api/tokens  (auth) — HTMX fragment or JSON
func handleTokensList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	claims, err := currentUser(r)
	if err != nil {
		w.Header().Set("HX-Redirect", "/login")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	rows, _ := db.Query(`SELECT id, name, prefix, expires_at, COALESCE(last_used_at,'') FROM api_tokens WHERE user_id=? ORDER BY created_at DESC`, claims.Sub)
	defer rows.Close()
	fmt.Fprint(w, `<ul class="divide-y-2 divide-gray-200">`)
	found := false
	for rows.Next() {
		var id, name, prefix, expires, last string
		rows.Scan(&id, &name, &prefix, &expires, &last)
		found = true
		exp, _ := time.Parse(time.RFC3339, expires)
		status := `<span class="nsk-badge nsk-badge-ok">Active</span>`
		if exp.Before(time.Now()) {
			status = `<span class="nsk-badge nsk-badge-muted">Expired</span>`
		}
		used := "Not used yet"
		if last != "" {
			used = "Last used " + string(timeTag(last))
		}
		fmt.Fprintf(w, `<li class="flex flex-wrap items-center justify-between gap-3 py-3 first:pt-0 last:pb-0">
  <div class="min-w-0">
    <p class="font-bold">%s %s</p>
    <p class="nsk-mono text-sm text-gray-700">%s…</p>
    <p class="text-sm text-gray-700">Valid until %s · %s</p>
  </div>
  <button hx-delete="/api/tokens/%s" hx-target="closest li" hx-swap="outerHTML"
    hx-confirm="Delete this token? Anything still using it will stop working."
    class="nsk-btn nsk-btn-danger nsk-btn-sm">Delete</button></li>`,
			template.HTMLEscapeString(name), status, template.HTMLEscapeString(prefix[:14]),
			string(timeTag(expires)), used, template.HTMLEscapeString(id))
	}
	if !found {
		fmt.Fprint(w, `<li class="text-gray-700">No tokens yet. Create one above if a developer needs to connect Netsekurity to your deployment pipeline.</li>`)
	}
	fmt.Fprint(w, `</ul>`)
}

// handleTokensDelete revokes an API token.
// DELETE /api/tokens/{id}
func handleTokensDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	claims, err := currentUser(r)
	if err != nil {
		w.Header().Set("HX-Redirect", "/login")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/tokens/")
	if id == "" {
		http.Error(w, "token id required", http.StatusBadRequest)
		return
	}
	db.Exec(`DELETE FROM api_tokens WHERE id=? AND user_id=?`, id, claims.Sub)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, ``)
}

// handleAPICreatePentest is the CI/CD entrypoint. Authenticated by X-API-Token.
// POST /api/v1/assessments (alias: /api/v1/pentests, the original name)
// JSON/Form: domain, mode (optional, "standard" is the only value)
// Returns JSON {assessment_id, pentest_id, domain, mode, status}; the two ids are the
// same value, and the second is kept so existing pipelines keep working.
func handleAPICreatePentest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, ok := requireAPIToken(w, r)
	if !ok {
		return
	}
	// allow JSON or form body
	domain := strings.TrimSpace(r.FormValue("domain"))
	authorised := r.FormValue("authorised") == "true"
	mode := strings.ToLower(strings.TrimSpace(r.FormValue("mode")))
	if mode == "" {
		mode = "standard"
	}
	if domain == "" {
		var b struct {
			Domain     string `json:"domain"`
			Mode       string `json:"mode"`
			Authorised bool   `json:"authorised"`
		}
		if err := json.NewDecoder(r.Body).Decode(&b); err == nil {
			domain = strings.TrimSpace(b.Domain)
			authorised = b.Authorised
			if b.Mode != "" {
				mode = strings.ToLower(b.Mode)
			}
		}
	}
	if domain == "" {
		http.Error(w, `{"error":"domain required"}`, http.StatusBadRequest)
		return
	}
	var cost float64
	if mode, cost = scanMode(mode); mode == "" {
		http.Error(w, `{"error":"invalid mode (standard only)"}`, http.StatusBadRequest)
		return
	}
	if !authorised {
		apiJSON(w, http.StatusBadRequest, map[string]string{"error": "authorised=true is required: confirm permission to test this hostname within the external, read-only scope"})
		return
	}
	pid, err := queueVerifiedAssessment(r.Context(), userID, domain, mode, cost)
	if err != nil {
		code := http.StatusConflict
		if errors.Is(err, errAssessmentCredit) {
			code = http.StatusPaymentRequired
		}
		apiJSON(w, code, map[string]string{"error": err.Error()})
		return
	}

	apiJSON(w, http.StatusOK, map[string]string{
		"assessment_id": pid, "pentest_id": pid, "domain": domain, "mode": mode, "status": "queued",
	})
}

// apiJSON writes one JSON response with its status code.
func apiJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

// handleAPIAssessment is the read side of the CI/CD API.
// GET /api/v1/assessments/{id}, authenticated by X-API-Token.
//
// It answers only for the token's own account: an id that belongs to someone else
// is a 404, the same as an id that does not exist, so ids cannot be probed. The
// finding counts appear once the report has been delivered with a breakdown; until
// then "findings" is null, never zeros, because zero would read as a clean result.
// The PDF itself is not served here: it stays in the signed-in dashboard.
func handleAPIAssessment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, ok := requireAPIToken(w, r)
	if !ok {
		return
	}
	id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/assessments/"), "/")
	if id == "" || strings.Contains(id, "/") {
		apiJSON(w, http.StatusNotFound, map[string]string{"error": "assessment not found"})
		return
	}
	var domain, status, started, completed string
	var reported, crit, high, med, low, info int
	err := db.QueryRow(`SELECT COALESCE(d.domain,''), p.status, COALESCE(p.started_at,''), COALESCE(p.completed_at,''),
			p.findings_reported, p.findings_critical, p.findings_high, p.findings_medium, p.findings_low, p.findings_info
		FROM pentests p LEFT JOIN domains d ON d.id=p.domain_id
		WHERE p.id=? AND p.user_id=?`, id, userID).Scan(&domain, &status, &started, &completed, &reported, &crit, &high, &med, &low, &info)
	if err != nil {
		apiJSON(w, http.StatusNotFound, map[string]string{"error": "assessment not found"})
		return
	}
	out := map[string]any{
		"assessment_id": id, "domain": domain, "status": status,
		"started_at": isoTime(started), "completed_at": isoTime(completed), "findings": nil,
	}
	if status == "completed" && reported == 1 {
		out["findings"] = map[string]int{"critical": crit, "high": high, "medium": med, "low": low, "info": info}
	}
	apiJSON(w, http.StatusOK, out)
}

// isoTime normalises a stored timestamp (RFC3339 from Go, "YYYY-MM-DD HH:MM:SS" from
// SQLite) to RFC3339, or nil when there is none, so the JSON has one format.
func isoTime(s string) any {
	if t, ok := parseStoredTime(s); ok {
		return t.Format(time.RFC3339)
	}
	return nil
}

// registerAPIv1 mounts the CI/CD API. "assessments" is the public name; "pentests"
// is the original one, kept so pipelines written against it do not break.
func registerAPIv1(mux *http.ServeMux, write, read *rateLimiter) {
	mux.HandleFunc("/api/v1/assessments", rateLimit(write, handleAPICreatePentest))
	mux.HandleFunc("/api/v1/pentests", rateLimit(write, handleAPICreatePentest))
	mux.HandleFunc("/api/v1/assessments/", rateLimitAll(read, "api-v1-read", handleAPIAssessment))
}
