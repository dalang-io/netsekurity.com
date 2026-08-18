package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
)

// handleContact accepts the landing-page contact form (HTMX hx-post) and
// returns a success fragment. Marketing scope: nothing is persisted yet.
func handleContact(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	_ = r.ParseForm()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<div class="rounded-lg border border-green-500/40 bg-green-500/10 px-4 py-3 text-sm text-green-300" role="status">
  <strong>Message received!</strong> Our security team will get back to you within 1 business day.
</div>`))
}

// handleTXT generates a mock auto-verification TXT record (HTMX hx-get).
func handleTXT(w http.ResponseWriter, r *http.Request) {
	token := "ns-verify-" + hex.EncodeToString(mustRand(16))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<div class="space-y-3">
  <div class="rounded-lg border border-cyan-500/40 bg-cyan-500/10 p-3 font-mono text-xs">
    <div class="text-cyan-400">Name</div><div class="text-gray-200">_netsekurity</div>
    <div class="mt-2 text-cyan-400">Type</div><div class="text-gray-200">TXT</div>
    <div class="mt-2 text-cyan-400">Value</div><div class="break-all text-gray-200" id="txt-value">` + token + `</div>
  </div>
  <button type="button" aria-label="Copy TXT value" class="w-full rounded-md border border-white/20 bg-white/5 px-4 py-2 text-sm font-semibold text-white hover:bg-white/10 transition-colors"
    onclick="navigator.clipboard.writeText(document.getElementById('txt-value').textContent); this.textContent='Copied!'; setTimeout(()=>this.textContent='Copy TXT value',1500)">
    Copy TXT value
  </button>
  <button type="button" class="w-full rounded-md bg-emerald-500 px-4 py-2 text-sm font-bold text-black hover:bg-emerald-400 transition-colors"
    hx-post="/api/verify" hx-swap="outerHTML">Verify DNS record</button>
</div>`))
}

// handleVerify simulates the DNS verification completing (HTMX hx-post).
func handleVerify(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<div class="rounded-lg border border-emerald-500/40 bg-emerald-500/10 px-4 py-3 text-sm text-emerald-300" role="status">
  <strong>Verified!</strong> Domain ownership confirmed. Your pentest credit is now active.
</div>`))
}

var faqAnswers = map[string]string{
	"1": `How does pricing work? <strong>1 credit = 1 pentest on 1 domain.</strong> Buy a package, add a domain, verify ownership with the auto-generated TXT record, and a pentest is scheduled.`,
	"2": `What is covered in one pentest? An automated, OWASP-mapped vulnerability assessment of your verified domain's <strong>public, unauthenticated</strong> surface: recon &amp; subdomain discovery, tech/WAF fingerprint, full security-header review, TLS/PKI validation, exposed files &amp; secrets, and common web-vuln probing (SQLi, LFI/RFI, XSS, open redirect, path traversal, misconfiguration). You get a prioritized English report with CVSS v3.1 scores and remediation steps.`,
	"3": `How do I verify my domain? We auto-generate a unique TXT record per domain. Add it in your DNS provider, hit verify, and ownership is confirmed in seconds. No manual approval.`,
	"4": `Is it really automated? The scan is fully automated end-to-end. A human security engineer reviews and curates the findings before the report is delivered.`,
	"5": `What if we have many subdomains? Each domain (hostname) costs 1 credit. A wildcard or an API surface can be scoped — contact us for a custom plan.`,
	"6": `What is <strong>not</strong> covered? Scans are <strong>unauthenticated</strong> — we do not log in or test behind sessions. We detect vulnerabilities and misconfiguration but do not build deep multi-stage manual exploit chains. Testing is read-only and non-destructive. For critical or auth-heavy production apps we recommend pairing with a supplemental manual penetration test. See the full scope in the dashboard &amp; report.`,
}

// handleFAQ toggles FAQ answers (HTMX hx-get) into an element id "faq-a-<q>".
func handleFAQ(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if r.URL.Query().Get("open") != "1" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(``))
		return
	}
	ans, ok := faqAnswers[q]
	if !ok {
		w.Write([]byte(``))
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<div class="mt-2 rounded border border-emerald-500/30 bg-emerald-500/5 px-3 py-2 text-xs leading-relaxed text-gray-300">%s
  <button class="mt-1 text-[10px] text-cyan-400 hover:underline"
    hx-get="/api/faq?q=%s&amp;open=0" hx-target="#faq-a-%s" hx-swap="innerHTML">[x] hide</button>
</div>`, ans, q, q)
}

func mustRand(n int) []byte {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return b
}

var _ = strings.TrimSpace
