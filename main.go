package main

import (
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
)

//go:embed static
var staticFS embed.FS

func main() {
	port := "8090"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatalf("embed fs: %v", err)
	}

	mux := http.NewServeMux()
	// Static assets (index.html, css, js) served from the embedded FS.
	mux.Handle("/", http.FileServer(http.FS(sub)))
	// HTMX endpoints (marketing scope).
	mux.HandleFunc("/contact", handleContact)
	mux.HandleFunc("/api/txt", handleTXT)
	mux.HandleFunc("/api/verify", handleVerify)
	mux.HandleFunc("/api/faq", handleFAQ)

	addr := ":" + port
	log.Printf("netsekurity.com landing listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

// handleContact accepts the landing-page contact form (HTMX hx-post) and
// returns a success fragment. Marketing scope: nothing is persisted yet.
func handleContact(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.TrimSpace(r.FormValue("email"))
	_ = email
	_ = name
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<div class="rounded-lg border border-green-500/40 bg-green-500/10 px-4 py-3 text-sm text-green-300" role="status">
  <strong>Message received!</strong> Our security team will get back to you within 1 business day.
</div>`))
}

// handleTXT generates a mock auto-verification TXT record (HTMX hx-get) so the
// landing page can demo the "Domain verification by TXT, auto-generated" flow.
func handleTXT(w http.ResponseWriter, r *http.Request) {
	token := randomToken(16)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<div class="space-y-3">
  <div class="rounded-lg border border-cyan-500/40 bg-cyan-500/10 p-3 font-mono text-xs">
    <div class="text-cyan-400">Name</div>
    <div class="text-gray-200">_netsekurity</div>
    <div class="mt-2 text-cyan-400">Type</div>
    <div class="text-gray-200">TXT</div>
    <div class="mt-2 text-cyan-400">Value</div>
    <div class="break-all text-gray-200" id="txt-value">` + token + `</div>
  </div>
  <button class="w-full rounded-md border border-white/20 bg-white/5 px-4 py-2 text-sm font-semibold text-white hover:bg-white/10 transition-colors"
    onclick="navigator.clipboard.writeText(document.getElementById('txt-value').textContent); this.textContent='Copied!'; setTimeout(()=>this.textContent='Copy TXT value',1500)">
    Copy TXT value
  </button>
  <button class="w-full rounded-md bg-emerald-500 px-4 py-2 text-sm font-bold text-black hover:bg-emerald-400 transition-colors"
    hx-post="/api/verify" hx-swap="outerHTML">
    Verify DNS record
  </button>
</div>`))
}

// handleVerify simulates the DNS verification completing (HTMX hx-post).
func handleVerify(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<div class="rounded-lg border border-emerald-500/40 bg-emerald-500/10 px-4 py-3 text-sm text-emerald-300" role="status">
  <strong>Verified!</strong> Domain ownership confirmed. Your pentest credit is now active.
</div>`))
}

func randomToken(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return "ns-verify-" + hex.EncodeToString(b)
}

var faqAnswers = map[string]string{
	"1": `How does pricing work? <strong>1 credit = 1 pentest on 1 domain.</strong> Buy a package, add a domain, verify ownership with the auto-generated TXT record, and a pentest is scheduled.`,
	"2": `What is covered in one pentest? An automated vulnerability assessment mapped to OWASP Top 10: injection, auth flaws, misconfiguration, exposed secrets, and more — with a prioritized report.`,
	"3": `How do I verify my domain? We auto-generate a unique TXT record per domain. Add it in your DNS provider, hit verify, and ownership is confirmed in seconds. No manual approval.`,
	"4": `Is it really automated? The scan is fully automated end-to-end. A human security engineer reviews and curates the findings before the report is delivered.`,
	"5": `What if we have many subdomains? Each domain (hostname) costs 1 credit. A wildcard or an API surface can be scoped — contact us for a custom plan.`,
}

// handleFAQ toggles FAQ answers (HTMX hx-get). open=1 returns the answer,
// open=0 (or a second click) collapses it.
func handleFAQ(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	open := r.URL.Query().Get("open") == "1"
	ans, ok := faqAnswers[q]
	if !ok || !open {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(``))
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<div class="mt-2 rounded-md border border-white/10 bg-white/5 px-4 py-3 text-sm leading-relaxed text-gray-300">` + ans + `</div>`))
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}