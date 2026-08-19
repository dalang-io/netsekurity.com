package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
	tok := strings.TrimSpace(r.Header.Get("X-API-Token"))
	if tok == "" {
		tok = strings.TrimSpace(r.URL.Query().Get("api_token"))
	}
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
	fmt.Fprintf(w, `<div class="rounded border border-emerald-500/30 bg-emerald-500/10 px-3 py-2 text-xs text-emerald-300">
  <div class="font-bold">Token created (shown once):</div>
  <code class="block break-all font-mono text-[11px] text-white">%s</code>
  <div class="text-[10px] text-gray-400">expires %s · name: %s · store it securely — it will not be shown again.
  Add it to your CI/CD as <span class="text-cyan-300">X-API-Token</span>.</div>
  <button hx-post="/api/tokens" hx-vals='{"name":"%s","expiry_days":"7"}' hx-target="#token-result" hx-swap="innerHTML"
    class="mt-2 rounded border border-emerald-400/60 px-2 py-1 text-[10px] font-bold text-emerald-300 hover:bg-emerald-500/15">generate another</button>
</div>`, secret, expires, name, name)
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
	fmt.Fprint(w, `<div class="mt-2"><div class="mb-1 font-mono text-[11px] text-gray-500"># api tokens</div><ul class="space-y-1.5">`)
	found := false
	for rows.Next() {
		var id, name, prefix, expires, last string
		rows.Scan(&id, &name, &prefix, &expires, &last)
		found = true
		exp, _ := time.Parse(time.RFC3339, expires)
		fmt.Fprintf(w, `<li class="rounded border border-white/10 bg-ink px-2.5 py-1.5 font-mono text-[11px]">
  <div class="flex items-center justify-between gap-2"><span class="text-white">%s</span>
  <span class="text-gray-500">%s…</span></div>
  <div class="flex items-center justify-between gap-2 text-[10px]"><span class="text-gray-500">expires %s%s</span>
  <button hx-delete="/api/tokens/%s" hx-target="closest li" hx-swap="outerHTML" hx-confirm="Delete this API token?"
    class="text-red-300 hover:underline">revoke</button></div></li>`,
			name, prefix[:14], exp.Format("2006-01-02"), func() string {
				if exp.Before(time.Now()) {
					return " (expired)"
				} else if last != "" {
					return " · used " + last
				} else {
					return ""
				}
			}(), id)
	}
	if !found {
		fmt.Fprint(w, `<li class="text-[11px] text-gray-600">no API tokens yet — generate one for CI/CD</li>`)
	}
	fmt.Fprintf(w, `</ul></div>`)
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
// POST /api/v1/pentests  JSON/Form: domain, mode (standard|destructive, optional)
// Returns JSON {pentest_id, domain, mode, status}.
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
	mode := strings.ToLower(strings.TrimSpace(r.FormValue("mode")))
	if mode == "" {
		mode = "standard"
	}
	if domain == "" {
		var b struct {
			Domain string `json:"domain"`
			Mode   string `json:"mode"`
		}
		if err := json.NewDecoder(r.Body).Decode(&b); err == nil {
			domain = strings.TrimSpace(b.Domain)
			if b.Mode != "" {
				mode = strings.ToLower(b.Mode)
			}
		}
	}
	if domain == "" {
		http.Error(w, `{"error":"domain required"}`, http.StatusBadRequest)
		return
	}
	if mode != "standard" && mode != "destructive" {
		http.Error(w, `{"error":"invalid mode"}`, http.StatusBadRequest)
		return
	}
	// find a verified domain owned by this user
	var domainID int64
	var status string
	err := db.QueryRow(`SELECT id, status FROM domains WHERE user_id=? AND domain=?`, userID, domain).Scan(&domainID, &status)
	if err != nil || status != "verified" {
		http.Error(w, `{"error":"domain not verified for this account"}`, http.StatusBadRequest)
		return
	}
	// in-flight cap
	var inflight int
	db.QueryRow(`SELECT COUNT(*) FROM pentests WHERE user_id=? AND status IN ('queued','running')`, userID).Scan(&inflight)
	if inflight > 0 {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"error":"a pentest is already queued or running; wait for it to finish"}`)
		return
	}
	cost := 1.0
	if mode == "destructive" {
		cost = 2.0
	}
	desc := "Pentest scheduled (CI/CD)"
	if mode == "destructive" {
		desc = "Destructive pentest scheduled (CI/CD)"
	}
	if !consumeCredit(userID, desc, fmt.Sprintf("pentest-api-%d", domainID), cost) {
		http.Error(w, `{"error":"insufficient credits"}`, http.StatusPaymentRequired)
		return
	}
	pid := "pt_" + randomHex(8)
	db.Exec(`INSERT INTO pentests (id, user_id, domain_id, mode, status) VALUES (?,?,?,?,'queued')`, pid, userID, domainID, mode)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"pentest_id": pid, "domain": domain, "mode": mode, "status": "queued",
	})
}
