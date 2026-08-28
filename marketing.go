package main

import (
	"fmt"
	"html/template"
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

var faqAnswers = map[string]string{
	"1": `How does pricing work? <strong>1 credit = 1 pentest on 1 domain.</strong> Buy a package, add a domain, verify ownership with the auto-generated TXT record, and a pentest is scheduled.`,
	"2": `What is covered in one pentest? An automated, OWASP-mapped vulnerability assessment of your verified domain's <strong>public, unauthenticated</strong> surface: recon &amp; subdomain discovery, tech/WAF fingerprint, full security-header review, TLS/PKI validation, exposed files &amp; secrets, and common web-vuln probing (SQLi, LFI/RFI, XSS, open redirect, path traversal, misconfiguration). You get a prioritized English report with CVSS v3.1 scores and remediation steps.`,
	"3": `How do I verify my domain? We auto-generate a unique TXT record per domain. Add it in your DNS provider, hit verify, and ownership is confirmed in seconds. No manual approval.`,
	"4": `Is it really automated? The scan is fully automated end-to-end. A security engineer reviews and curates the findings before the report is delivered.`,
	"5": `What if we have many subdomains? Each domain (hostname) costs 1 credit. A wildcard or an API surface can be scoped — contact us for a custom plan.`,
	"6": `What is <strong>not</strong> covered? Scans are <strong>unauthenticated</strong> — we do not log in or test behind sessions. We detect vulnerabilities and misconfiguration but do not build deep multi-stage manual exploit chains. Testing is read-only and non-destructive. For critical or auth-heavy production apps we recommend pairing with a supplemental manual penetration test. See the full scope in the dashboard &amp; report.`,
	"7": `Blackbox vs Whitebox? This scan is <strong>blackbox</strong> (external, unauthenticated — no credentials or source). For <strong>whitebox</strong> (source review, authenticated testing, business-logic &amp; exploit-chain analysis by a security engineer + agent), it is <strong>$10,000 USD per app / per domain</strong>. Contact us to schedule.`,
	"8": `I'm not technical / built this with an agent — is this for me? <strong>Yes.</strong> If you vibe-coded an app with an agent and don't have a security background, Netsekurity is exactly for you: add your domain, verify it with an auto-generated TXT record, and get a plain-language, agent-reviewed report telling you what to fix before going to production.`,
	"9": `Is this a full penetration test? <strong>No.</strong> The standard product is an automated, <strong>external, read-only web security assessment</strong> of your public attack surface with agent-reviewed findings. It does not test authenticated areas, authorization (IDOR), or business logic. If your app has critical authenticated flows, pair it with our <strong>whitebox</strong> tier (source + authenticated manual testing by a security engineer, $10,000 USD per app/domain).`,
}

// faqOrder fixes the display order of the questions (Go maps are unordered).
var faqOrder = []struct{ ID, Q string }{
	{"1", "How does pricing work?"},
	{"2", "What is covered in one pentest?"},
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
<link rel="icon" type="image/svg+xml" href="/favicon.svg"/>
</head>
<body class="scanlines bg-ink text-gray-300">
<main class="mx-auto max-w-3xl px-4 py-12 sm:px-6">
  <a href="/" class="font-mono text-sm font-bold text-white"><span class="text-emerald-400">net</span>sekurity<span class="text-emerald-500">.com</span></a>
  <h1 class="mt-6 font-mono text-2xl font-bold text-white"><span class="text-emerald-400">$</span> cat PRODUCT_SCOPE.md</h1>
  <pre class="copy mt-6 whitespace-pre-wrap break-words rounded border border-white/10 bg-[%s] p-5 text-gray-300">%s</pre>
  <p class="mt-6 font-mono text-xs"><a href="/#faq" class="text-cyan-400 hover:underline">&larr; back to faq</a></p>
</main>
</body></html>`, cssHash, "#04060c", template.HTMLEscapeString(string(md)))
}

var _ = strings.TrimSpace
