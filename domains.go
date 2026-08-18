package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
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
	domain := strings.ToLower(strings.TrimSpace(r.FormValue("domain")))
	if !isValidDomain(domain) {
		http.Error(w, "invalid domain", http.StatusBadRequest)
		return
	}
	token := "ns-verify-" + randomHex(12)
	if _, err := db.Exec(`INSERT OR IGNORE INTO domains (user_id, domain, txt_verification_token, status) VALUES (?,?,?,'pending')`,
		claims.Sub, domain, token); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<div class="rounded-lg border border-cyan-500/40 bg-cyan-500/10 p-3 font-mono text-xs">
  <div class="text-cyan-400">Name</div><div class="text-gray-200">_netsekurity</div>
  <div class="mt-2 text-cyan-400">Type</div><div class="text-gray-200">TXT</div>
  <div class="mt-2 text-cyan-400">Value</div><div class="break-all text-gray-200">%s</div>
</div>
<button hx-post="/api/domains/verify" hx-vals='{"domain":"%s"}' hx-swap="outerHTML"
  class="mt-3 w-full rounded-md bg-emerald-500 px-4 py-2 text-sm font-bold text-black hover:bg-emerald-400">Verify now</button>`,
		token, domain)
}

// handleVerifyDomain checks the TXT record in public DNS for the domain.
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
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	txts, _ := (&net.Resolver{}).LookupTXT(ctx, "_netsekurity."+domain)
	matched := false
	for _, t := range txts {
		if strings.Trim(t, `"`) == stored {
			matched = true
			break
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if matched {
		db.Exec(`UPDATE domains SET status='verified', verified_at=? WHERE user_id=? AND domain=?`, now(), claims.Sub, domain)
		fmt.Fprintf(w, `<div class="rounded-lg border border-emerald-500/40 bg-emerald-500/10 px-4 py-3 text-sm text-emerald-300"><strong>Verified!</strong> %s ownership confirmed — a pentest can now be scheduled.</div>`, domain)
		return
	}
	// Not found yet — offer a retry.
	fmt.Fprintf(w, `<div class="space-y-3">
  <div class="rounded-lg border border-red-500/40 bg-red-500/10 px-4 py-3 text-sm text-red-300">TXT record not found yet. Add <code class="font-mono">%s</code> at <code class="font-mono">_netsekurity.%s</code> and retry.</div>
  <button hx-post="/api/domains/verify" hx-vals='{"domain":"%s"}' hx-swap="outerHTML"
    class="rounded-md bg-emerald-500 px-4 py-2 text-sm font-bold text-black hover:bg-emerald-400">Retry verification</button>
</div>`, stored, domain, domain)
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
		fmt.Fprintf(w, `<div class="rounded-lg border border-red-500/40 bg-red-500/10 px-4 py-3 text-sm text-red-300"><strong>Cannot delete.</strong> %s is already verified — contact support to remove it.</div>`, domain)
		return
	}
	db.Exec(`DELETE FROM domains WHERE user_id=? AND domain=?`, claims.Sub, domain)
	fmt.Fprintf(w, `<div class="rounded-lg border border-white/10 bg-white/5 px-4 py-3 text-sm text-gray-300">%s removed.</div>`, domain)
}

func isValidDomain(domain string) bool {
	if len(domain) < 3 || len(domain) > 253 || !strings.Contains(domain, ".") {
		return false
	}
	if strings.HasPrefix(domain, "-") || strings.HasSuffix(domain, "-") {
		return false
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
