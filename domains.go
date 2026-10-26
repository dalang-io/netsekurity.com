package main

import (
	"context"
	"fmt"
	"html/template"
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
		fmt.Fprint(w, alertHTML("ok", `<strong>Verified!</strong> We confirmed you own `+template.HTMLEscapeString(domain)+`. You can scan it now.`))
		renderDomainListOOB(w, claims.Sub)
		return
	}
	// Not found yet. The domain's card still has the record and its Check now
	// button, so this only has to say what to do — not repeat the controls.
	fmt.Fprint(w, alertHTML("warn", `<strong>We can't see the record yet.</strong> DNS changes can take a few minutes to spread. `+
		`Check that the TXT record <code>_netsekurity</code> (or <code>_netsekurity.`+template.HTMLEscapeString(domain)+`</code>) has exactly the value shown on the card, then press <strong>Check now</strong> again.`))
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
	off += 4                     // qtype + qclass
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
