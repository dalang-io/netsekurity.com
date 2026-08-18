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
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	// Check both recursive resolvers and the authoritative nameservers directly,
	// so a freshly-added TXT record verifies immediately (no waiting for
	// recursive-cache propagation).
	txts, _ := (&net.Resolver{}).LookupTXT(ctx, "_netsekurity."+domain)
	txts = append(txts, lookupTXTAuth("_netsekurity."+domain)...)
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

// lookupTXTAuth queries the TXT records for a name directly from the domain's
// authoritative nameservers (bypassing recursive-cache propagation), so a
// freshly-added verification record is seen immediately.
func lookupTXTAuth(name string) []string {
	// The NS records belong to the base domain, not the full record name
	// (e.g. _netsekurity.netsekurity.com -> netsekurity.com).
	base := name
	if i := strings.Index(base, "."); i >= 0 {
		base = base[i+1:]
	}
	ns, err := net.LookupNS(base)
	if err != nil || len(ns) == 0 {
		return nil
	}
	var out []string
	for _, s := range ns {
		out = append(out, dnsQueryTXT(strings.TrimSuffix(s.Host, "."), name)...)
	}
	return out
}

// dnsQueryTXT sends a minimal UDP DNS TXT query to a specific nameserver and
// returns the TXT record strings (stdlib-only, no external DNS library).
func dnsQueryTXT(nameserver, name string) []string {
	buf := []byte{0x12, 0x34, 0x01, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0}
	for _, label := range strings.Split(name, ".") {
		if len(label) == 0 {
			continue
		}
		buf = append(buf, byte(len(label)))
		buf = append(buf, []byte(label)...)
	}
	buf = append(buf, 0, 0, 16, 0, 1) // qname null + TXT + IN

	conn, err := net.Dial("udp", net.JoinHostPort(nameserver, "53"))
	if err != nil {
		return nil
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	if _, err := conn.Write(buf); err != nil {
		return nil
	}
	resp := make([]byte, 4096)
	n, err := conn.Read(resp)
	if err != nil || n < 12 {
		return nil
	}
	resp = resp[:n]
	ancount := int(resp[6])<<8 | int(resp[7])
	off := 12
	off = skipDNSName(resp, off) // question name
	off += 4                      // qtype + qclass
	var txts []string
	for i := 0; i < ancount && off < len(resp); i++ {
		off = skipDNSName(resp, off)
		if off+10 > len(resp) {
			break
		}
		rtype := int(resp[off])<<8 | int(resp[off+1])
		rdlen := int(resp[off+8])<<8 | int(resp[off+9])
		off += 10
		if rtype == 16 { // TXT
			end := off + rdlen
			var s []byte
			for off < end && off < len(resp) {
				l := int(resp[off])
				off++
				if off+l <= len(resp) {
					s = append(s, resp[off:off+l]...)
				}
				off += l
			}
			txts = append(txts, string(s))
		} else {
			off += rdlen
		}
	}
	return txts
}

// skipDNSName advances past a (possibly compression-pointer) DNS name.
func skipDNSName(msg []byte, off int) int {
	for off < len(msg) {
		l := msg[off]
		if l == 0 {
			return off + 1
		}
		if l&0xC0 == 0xC0 {
			return off + 2
		}
		off += int(l) + 1
	}
	return off
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
