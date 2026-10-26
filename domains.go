package main

import (
	"fmt"
	"html/template"
	"net"
	"net/http"
	"strings"
)

// handleAddDomain registers a domain and generates its TXT verification token.
func handleAddDomain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	claims, err := currentUser(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	domain := normalizeDomain(r.FormValue("domain"))
	if !isValidDomain(domain) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, alertHTML("bad", `<strong>That doesn't look like a website address.</strong> Type just the address, like <code>www.mybusiness.com</code> — no spaces, and nothing after the .com.`))
		return
	}
	token := "ns-verify-" + randomHex(12)
	res, err := db.Exec(`INSERT OR IGNORE INTO domains (user_id, domain, txt_verification_token, status) VALUES (?,?,?,'pending')`,
		claims.Sub, domain, token)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// The verification steps live on the domain's own card, which carries the
	// stored token. Echoing the freshly generated one here was wrong whenever the
	// domain already existed: INSERT OR IGNORE keeps the original token.
	if n, _ := res.RowsAffected(); n == 0 {
		fmt.Fprint(w, alertHTML("info", `<strong>`+template.HTMLEscapeString(domain)+`</strong> is already in your list.`))
	} else {
		fmt.Fprint(w, alertHTML("ok", `<strong>`+template.HTMLEscapeString(domain)+`</strong> added. Next, prove it is yours: follow the steps on its card below.`))
	}
	renderDomainListOOB(w, claims.Sub)
}

// handleVerifyDomain checks the selected proof for this account and hostname.
func handleVerifyDomain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	claims, err := currentUser(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	domain := strings.ToLower(strings.TrimSpace(r.FormValue("domain")))
	var stored string
	err = db.QueryRow(`SELECT txt_verification_token FROM domains WHERE user_id=? AND domain=?`, claims.Sub, domain).Scan(&stored)
	if err != nil {
		http.Error(w, "domain not found", http.StatusNotFound)
		return
	}
	method := verificationMethod(r.FormValue("method"))
	if method == "" {
		http.Error(w, "invalid verification method", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := checkDomainControl(r.Context(), domain, stored, method); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, alertHTML("warn", `<strong>We could not confirm your website yet.</strong> `+template.HTMLEscapeString(err.Error())+`. Check the instructions for your chosen method and try again. No credits have been used.`))
		return
	}
	if _, err := db.Exec(`UPDATE domains SET status='verified', verified_at=?, verified_by=? WHERE user_id=? AND domain=?`, now(), method, claims.Sub, domain); err != nil {
		http.Error(w, "could not save verification", http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, alertHTML("ok", `<strong>Verified!</strong> Control of `+template.HTMLEscapeString(domain)+` confirmed. We will check the proof again before each scan; leave it in place.`))
	renderDomainListOOB(w, claims.Sub)
}

// handleDeleteDomain removes a domain — but only while it is NOT verified.
func handleDeleteDomain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	claims, err := currentUser(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	domain := strings.ToLower(strings.TrimSpace(r.FormValue("domain")))
	var status string
	err = db.QueryRow(`SELECT status FROM domains WHERE user_id=? AND domain=?`, claims.Sub, domain).Scan(&status)
	if err != nil {
		http.Error(w, "domain not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if status == "verified" {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, alertHTML("bad", `<strong>This website can't be removed here.</strong> `+template.HTMLEscapeString(domain)+` is already verified. Contact us and we will remove it for you.`))
		return
	}
	db.Exec(`DELETE FROM domains WHERE user_id=? AND domain=?`, claims.Sub, domain)
	fmt.Fprint(w, alertHTML("info", `<strong>`+template.HTMLEscapeString(domain)+`</strong> was removed.`))
	renderDomainListOOB(w, claims.Sub)
}

// normalizeDomain reduces whatever the customer pasted to a bare hostname.
// Customers paste full URLs: "https://app.example.com/" used to pass
// isValidDomain (splitting on "." yields plausible-looking labels), get stored,
// and then never verify, because the TXT lookup ran against a string containing
// a scheme and a slash.
func normalizeDomain(raw string) string {
	d := strings.ToLower(strings.TrimSpace(raw))
	if i := strings.Index(d, "://"); i >= 0 {
		d = d[i+3:]
	}
	if i := strings.IndexAny(d, "/?#"); i >= 0 {
		d = d[:i]
	}
	if i := strings.Index(d, "@"); i >= 0 { // user:pass@host
		d = d[i+1:]
	}
	if i := strings.LastIndex(d, ":"); i >= 0 { // :port
		d = d[:i]
	}
	return strings.Trim(strings.TrimSuffix(d, "."), ".")
}

func isValidDomain(domain string) bool {
	if net.ParseIP(domain) != nil {
		return false
	}
	if len(domain) < 3 || len(domain) > 253 || !strings.Contains(domain, ".") {
		return false
	}
	// Reject anything a hostname can never contain, so a malformed paste fails
	// loudly at the form instead of silently at DNS-lookup time.
	if strings.ContainsAny(domain, "/:?#@ _") {
		return false
	}
	if strings.HasPrefix(domain, "-") || strings.HasSuffix(domain, "-") {
		return false
	}
	for _, c := range domain {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '.' || c == '-') {
			return false
		}
	}
	for _, label := range strings.Split(domain, ".") {
		if label == "" || len(label) > 63 {
			return false
		}
		if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return false
		}
	}
	return true
}

// handleVerificationFile serves only the signed-in account's own proof file.
func handleVerificationFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	claims, err := currentUser(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var token string
	if err := db.QueryRow(`SELECT txt_verification_token FROM domains WHERE user_id=? AND domain=?`, claims.Sub, r.URL.Query().Get("domain")).Scan(&token); err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="netsekurity-verification.txt"`)
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	fmt.Fprintln(w, token)
}
