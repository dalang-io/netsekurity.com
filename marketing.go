package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

// handleContact accepts the landing-page contact form (HTMX hx-post). Messages
// are stored in contact_messages and listed on /su; the handler used to answer
// "Message received!" and discard the body, so every lead sent through it was
// lost.
func handleContact(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	_ = r.ParseForm()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	msg := strings.TrimSpace(r.FormValue("message"))
	source := strings.TrimSpace(r.FormValue("source"))
	// Honeypot: bots fill every field, people never see this one.
	if r.FormValue("website") != "" {
		fmt.Fprint(w, contactOK)
		return
	}
	if !strings.Contains(email, "@") || len(email) > 200 || msg == "" || len(msg) > 4000 || len(name) > 200 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `<div class="rounded-lg border border-red-500/40 bg-red-500/10 px-4 py-3 text-sm text-red-300" role="alert">Please enter a valid email and a message.</div>`)
		return
	}
	if len(source) > 100 {
		source = source[:100]
	}
	if _, err := db.Exec(`INSERT INTO contact_messages (name, email, message, source, ip) VALUES (?,?,?,?,?)`,
		name, email, msg, source, clientIP(r)); err != nil {
		log.Printf("contact: store failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `<div class="rounded-lg border border-red-500/40 bg-red-500/10 px-4 py-3 text-sm text-red-300" role="alert">Could not save your message. Please email sales@dalang.io directly.</div>`)
		return
	}
	log.Printf("contact: new message from %s (%s)", email, source)
	fmt.Fprint(w, contactOK)
}

const contactOK = `<div class="rounded-lg border border-emerald-500/40 bg-emerald-500/10 px-4 py-3 text-sm text-emerald-300" role="status">
  <strong>Message received.</strong> We reply from sales@dalang.io within 1 business day.
</div>`

// scansCompleted is the one number the site is allowed to quote as social
// proof: real completed assessments. Rendered only once it is large enough to
// mean something; a "3 scans run" badge would do the opposite of reassure.
func scansCompleted() int {
	var n int
	db.QueryRow(`SELECT COUNT(*) FROM pentests WHERE status='completed'`).Scan(&n)
	return n
}

const scansProofMin = 25

// renderScansProof returns the proof line for the landing page, or "" below the threshold.
func renderScansProof() string {
	n := scansCompleted()
	if n < scansProofMin {
		return ""
	}
	return fmt.Sprintf(`<p class="mt-3 font-mono text-xs text-gray-500"><span class="text-emerald-300">%d</span> assessments delivered so far</p>`, n)
}

var faqAnswers = map[string]string{
	"1": `How does pricing work? <strong>1 credit = 1 external assessment on 1 domain.</strong> Buy a package, add a domain, verify ownership with the auto-generated TXT record, and the assessment is scheduled.`,
	"2": `What is covered in one assessment? An automated, OWASP-mapped assessment of your verified domain's <strong>public, unauthenticated</strong> surface: DNS &amp; passive subdomain discovery, server/WAF fingerprint, the full security-header set, clickjacking and server-version disclosure, TLS certificate validity and expiry, 13 exposed-file and admin-path probes (<code>.git</code>, <code>.env</code>, backups, phpinfo, server-status, login pages), and open-redirect probing on 9 common parameters. You get a prioritized English report with CVSS v3.1 scores and remediation steps. <a href="/sample-report" class="text-cyan-400 hover:underline">See a sample report.</a>`,
	"3": `How do I verify my domain? We auto-generate a unique TXT record per domain. Add it in your DNS provider, hit verify, and ownership is confirmed in seconds. No manual approval.`,
	"4": `Is it really automated? Yes, end-to-end. Our analysis agent deduplicates and ranks the findings and writes the report; no human reads it before delivery. For a security engineer's manual review, see the expert review add-on and the whitebox tier.`,
	"5": `What if we have many subdomains? Each domain (hostname) costs 1 credit. A wildcard or an API surface can be scoped — contact us for a custom plan.`,
	"6": `What is <strong>not</strong> covered? Scans are <strong>unauthenticated</strong> — we do not log in or test behind sessions, and we do not test authorization (IDOR) or business logic. The standard scan does not send injection payloads (SQLi, XSS, LFI) — it finds exposure and misconfiguration, not exploitable input bugs. Testing is read-only and non-destructive. For critical or auth-heavy production apps, pair it with the whitebox tier. Full list: <a href="/product-scope" class="text-cyan-400 hover:underline">product scope</a>.`,
	"7": `Blackbox vs Whitebox? This scan is <strong>blackbox</strong> (external, unauthenticated — no credentials or source). For <strong>whitebox</strong> (source review, authenticated testing, business-logic &amp; exploit-chain analysis by a security engineer + agent), it is <strong>$10,000 USD per app / per domain</strong>. Contact us to schedule.`,
	"8": `I'm not technical / built this with an agent — is this for me? <strong>Yes.</strong> If you vibe-coded an app with an agent and don't have a security background, Netsekurity is exactly for you: add your domain, verify it with an auto-generated TXT record, and get a plain-language, agent-reviewed report telling you what to fix before going to production.`,
	"9": `Is this a full penetration test? <strong>No.</strong> The standard product is an automated, <strong>external, read-only web security assessment</strong> of your public attack surface with agent-reviewed findings. It does not test authenticated areas, authorization (IDOR), or business logic. If your app has critical authenticated flows, pair it with our <strong>whitebox</strong> tier (source + authenticated manual testing by a security engineer, $10,000 USD per app/domain).`,
}

// faqOrder fixes the display order of the questions (Go maps are unordered).
var faqOrder = []struct{ ID, Q string }{
	{"1", "How does pricing work?"},
	{"2", "What is covered in one assessment?"},
	{"3", "How do I verify my domain?"},
	{"4", "Is it really automated?"},
	{"5", "What if we have many subdomains?"},
	{"6", "What is NOT covered?"},
	{"7", "Blackbox vs whitebox?"},
	{"8", "I'm not technical — is this for me?"},
	{"9", "Is this a full penetration test?"},
}

// renderFAQ renders the FAQ as native <details> disclosures. It used to be nine
// HTMX endpoints — one request to open each answer and another to close it — with
// no open/closed affordance on the question itself. The answers are static, so
// they belong in the page.
func renderFAQ() string {
	var b strings.Builder
	for _, f := range faqOrder {
		ans, ok := faqAnswers[f.ID]
		if !ok {
			continue
		}
		b.WriteString(`<details class="faq-item rounded border border-white/10 bg-[#04060c]">
  <summary class="cursor-pointer list-none px-4 py-3 font-mono text-sm text-gray-200 marker:content-[''] hover:text-emerald-300">
    <span class="text-emerald-400">?</span> ` + template.HTMLEscapeString(f.Q) + `
  </summary>
  <div class="copy border-t border-white/10 px-4 py-3 text-gray-300">` + ans + `</div>
</details>
`)
	}
	return b.String()
}

// handleProductScope renders PRODUCT_SCOPE.md inside the site shell. It used to be
// linked as a raw .md file, which browsers show unrendered or download outright —
// for an explicitly non-technical audience.
func handleProductScope(w http.ResponseWriter, r *http.Request) {
	md, err := staticFS.ReadFile("static/PRODUCT_SCOPE.md")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en"><head>
<meta charset="UTF-8"/><meta name="viewport" content="width=device-width, initial-scale=1.0"/>
<title>product scope — netsekurity</title>
<link rel="stylesheet" href="/css/styles.css?v=%s"/>
<link rel="icon" type="image/svg+xml" href="/favicon.svg"/>%s
</head>
<body class="scanlines bg-ink text-gray-300">
<main class="mx-auto max-w-3xl px-4 py-12 sm:px-6">
  <a href="/" class="font-mono text-sm font-bold text-white"><span class="text-emerald-400">net</span>sekurity<span class="text-emerald-500">.com</span></a>
  <h1 class="mt-6 font-mono text-2xl font-bold text-white"><span class="text-emerald-400">$</span> cat PRODUCT_SCOPE.md</h1>
  <pre class="copy mt-6 whitespace-pre-wrap break-words rounded border border-white/10 bg-[%s] p-5 text-gray-300">%s</pre>
  <p class="mt-6 font-mono text-xs"><a href="/#faq" class="text-cyan-400 hover:underline">&larr; back to faq</a></p>
</main>
</body></html>`, cssHash, analyticsSnippet(), "#04060c", template.HTMLEscapeString(string(md)))
}

var _ = strings.TrimSpace
