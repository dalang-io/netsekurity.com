package main

import (
	"html/template"
	"net/http"
	"strings"
)

// /sample-report — what a customer gets for a credit, before they pay for one.
//
// The product is a PDF. Nothing on the site showed one, so a visitor was asked
// to spend $50 on a document they had never seen. This page mirrors the real
// report's structure (Document Control, Findings Summary, Risk Register,
// Findings & Analysis, Remediation Plan) with findings drawn only from checks
// nsec_scan.py actually performs, against a fictional domain.
func handleSampleReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	authNav := `<a href="/login" class="whitespace-nowrap rounded border border-emerald-400 bg-emerald-500/10 px-3 py-1.5 text-xs font-bold text-emerald-300 hover:bg-emerald-500/20 glow">login</a>`
	if _, err := currentUser(r); err == nil {
		authNav = `<a href="/dashboard" class="whitespace-nowrap rounded border border-emerald-400 bg-emerald-500/10 px-3 py-1.5 text-xs font-bold text-emerald-300 hover:bg-emerald-500/20 glow">./dashboard<span class="cursor"></span></a>`
	}
	nav := []hdrLink{
		{Href: "/", Text: "home"},
		{Href: "/#how", Text: "how it works"},
		{Href: "/#pricing", Text: "pricing"},
		{Href: "/product-scope", Text: "scope"},
		{Href: "/docs", Text: "docs"},
	}
	out := strings.ReplaceAll(sampleReportHTML, "__HEADER__", string(renderHeader(nav, template.HTML(authNav), "/"))+headerMobileJS)
	out = strings.ReplaceAll(out, "__CSS_HASH__", cssHash)
	out = strings.ReplaceAll(out, "__FBPIXEL__", consentGate(true))
	out = strings.ReplaceAll(out, "__ANALYTICS__", analyticsSnippet())
	w.Write([]byte(out))
}

const sampleReportHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1"/>
<title>sample report — netsekurity.com</title>
<meta name="description" content="A sample Netsekurity external web security assessment report: what one credit buys, section by section, for a fictional domain."/>
<link rel="canonical" href="https://netsekurity.com/sample-report"/>
<link rel="icon" type="image/svg+xml" href="/favicon.svg"/>
__FBPIXEL__
__ANALYTICS__
<link rel="stylesheet" href="/css/styles.css?v=__CSS_HASH__"/>
</head>
<body class="scanlines bg-ink text-gray-300 min-h-screen overflow-x-hidden">
__HEADER__
<main class="mx-auto max-w-4xl px-4 py-10 sm:px-6">
  <p class="font-mono text-xs text-cyan-300">[ sample — fictional domain, real report structure ]</p>
  <h1 class="mt-2 font-mono text-2xl font-bold text-white sm:text-3xl"><span class="text-emerald-400">$</span> cat 20260901-09:14-shop.example.com.pdf</h1>
  <p class="copy mt-3 text-gray-400">This is what one credit delivers. Every finding below comes from a check the scanner really performs; the domain and evidence are invented. Your report is a PDF with the same sections, in English, downloadable from the dashboard.</p>

  <!-- Document control -->
  <section class="mt-8 rounded border border-emerald-500/30 bg-[#04060c]">
    <div class="border-b border-emerald-500/20 px-4 py-2 font-mono text-[11px] text-gray-500">$ 1 · document control</div>
    <dl class="grid grid-cols-1 gap-x-6 gap-y-2 px-4 py-4 font-mono text-xs sm:grid-cols-2">
      <div><dt class="text-gray-500">target</dt><dd class="text-white">shop.example.com</dd></div>
      <div><dt class="text-gray-500">assessment type</dt><dd class="text-white">external · unauthenticated · read-only</dd></div>
      <div><dt class="text-gray-500">ownership</dt><dd class="text-white">verified via TXT <span class="text-gray-500">_netsekurity</span></dd></div>
      <div><dt class="text-gray-500">completed</dt><dd class="text-white">2026-09-01 09:14 UTC</dd></div>
      <div><dt class="text-gray-500">auditor</dt><dd class="text-white">Netsekurity.com Security Agent</dd></div>
      <div><dt class="text-gray-500">classification</dt><dd class="text-white">confidential — owner &amp; admins only</dd></div>
    </dl>
  </section>

  <!-- Findings summary -->
  <section class="mt-6 rounded border border-white/10 bg-[#04060c]">
    <div class="border-b border-white/10 px-4 py-2 font-mono text-[11px] text-gray-500">$ 2 · findings summary</div>
    <div class="px-4 py-4">
      <div class="flex flex-wrap items-center gap-1">
        <span class="rounded bg-red-500/20 px-2 py-0.5 font-mono text-[11px] font-bold text-red-300">1 high</span>
        <span class="rounded bg-yellow-500/20 px-2 py-0.5 font-mono text-[11px] font-bold text-yellow-300">2 medium</span>
        <span class="rounded bg-cyan-500/20 px-2 py-0.5 font-mono text-[11px] font-bold text-cyan-300">2 low</span>
        <span class="rounded bg-white/10 px-2 py-0.5 font-mono text-[11px] font-bold text-gray-400">4 info</span>
      </div>
      <p class="copy mt-3 text-gray-400">Overall posture: <span class="text-yellow-300">needs attention</span>. One file exposure can leak credentials today; the remaining items are hardening. Nothing found suggests the site is currently compromised.</p>
    </div>
  </section>

  <!-- Risk register -->
  <section class="mt-6 rounded border border-white/10 bg-[#04060c]">
    <div class="border-b border-white/10 px-4 py-2 font-mono text-[11px] text-gray-500">$ 3 · risk register</div>
    <div class="overflow-x-auto p-3">
      <table class="w-full font-mono text-xs">
        <thead><tr class="text-left text-gray-500"><th class="py-1 pr-3">#</th><th class="pr-3">finding</th><th class="pr-3">severity</th><th class="pr-3">cvss</th><th class="pr-3">cwe</th><th>owasp</th></tr></thead>
        <tbody class="text-gray-300">
          <tr class="border-t border-white/10"><td class="py-1.5 pr-3">F-01</td><td class="pr-3 text-white">Environment file exposed at /.env</td><td class="pr-3 text-red-300">HIGH</td><td class="pr-3">7.5</td><td class="pr-3">CWE-538</td><td>A05 · A02</td></tr>
          <tr class="border-t border-white/10"><td class="py-1.5 pr-3">F-02</td><td class="pr-3 text-white">Missing security headers (4 of 6)</td><td class="pr-3 text-yellow-300">MEDIUM</td><td class="pr-3">5.3</td><td class="pr-3">CWE-1021</td><td>A05</td></tr>
          <tr class="border-t border-white/10"><td class="py-1.5 pr-3">F-03</td><td class="pr-3 text-white">Open redirect via <code>?next=</code></td><td class="pr-3 text-yellow-300">MEDIUM</td><td class="pr-3">5.4</td><td class="pr-3">CWE-601</td><td>A01</td></tr>
          <tr class="border-t border-white/10"><td class="py-1.5 pr-3">F-04</td><td class="pr-3 text-white">Server version disclosed (nginx/1.18.0)</td><td class="pr-3 text-cyan-300">LOW</td><td class="pr-3">3.7</td><td class="pr-3">CWE-200</td><td>A05</td></tr>
          <tr class="border-t border-white/10"><td class="py-1.5 pr-3">F-05</td><td class="pr-3 text-white">TLS certificate expires in 41 days</td><td class="pr-3 text-cyan-300">LOW</td><td class="pr-3">3.7</td><td class="pr-3">CWE-295</td><td>A02</td></tr>
          <tr class="border-t border-white/10 text-gray-500"><td class="py-1.5 pr-3">I-01…04</td><td class="pr-3">DNS resolution · 7 subdomains · WAF detected · robots.txt disallow entries</td><td class="pr-3">INFO</td><td class="pr-3">—</td><td class="pr-3">—</td><td>—</td></tr>
        </tbody>
      </table>
    </div>
  </section>

  <!-- Findings & analysis -->
  <section class="mt-6 rounded border border-white/10 bg-[#04060c]">
    <div class="border-b border-white/10 px-4 py-2 font-mono text-[11px] text-gray-500">$ 4 · findings &amp; analysis</div>
    <div class="space-y-5 px-4 py-4">

      <article class="rounded border border-red-500/30 bg-red-500/5 p-4">
        <div class="flex flex-wrap items-baseline justify-between gap-2">
          <h3 class="font-mono text-sm font-bold text-white">F-01 · Environment file exposed at /.env</h3>
          <span class="rounded bg-red-500/20 px-2 py-0.5 font-mono text-[11px] font-bold text-red-300">HIGH · CVSS 7.5</span>
        </div>
        <dl class="mt-3 space-y-2 font-mono text-xs">
          <div><dt class="text-gray-500">what we saw</dt><dd class="copy text-gray-300">The path <code>/.env</code> returned HTTP 200 with a 1.2 KB body. Files with this name normally hold database passwords, API keys and signing secrets. We did not download or store the contents.</dd></div>
          <div><dt class="text-gray-500">evidence</dt><dd><pre class="mt-1 overflow-x-auto rounded bg-black/60 p-2 text-[11px] text-emerald-300">GET /.env -> 200 1247</pre></dd></div>
          <div><dt class="text-gray-500">impact</dt><dd class="copy text-gray-300">Anyone can read your production secrets. This is the single most common way vibe-coded deployments get taken over.</dd></div>
          <div><dt class="text-gray-500">fix</dt><dd class="copy text-gray-300">Block the file at the web server (nginx: <code>location ~ /\.env { deny all; }</code>), move it outside the web root, then <strong>rotate every secret it contains</strong> — assume it has already been read.</dd></div>
        </dl>
      </article>

      <article class="rounded border border-yellow-500/30 bg-yellow-500/5 p-4">
        <div class="flex flex-wrap items-baseline justify-between gap-2">
          <h3 class="font-mono text-sm font-bold text-white">F-02 · Missing security headers</h3>
          <span class="rounded bg-yellow-500/20 px-2 py-0.5 font-mono text-[11px] font-bold text-yellow-300">MEDIUM · CVSS 5.3</span>
        </div>
        <dl class="mt-3 space-y-2 font-mono text-xs">
          <div><dt class="text-gray-500">what we saw</dt><dd class="copy text-gray-300">The response lacks <code>strict-transport-security</code>, <code>content-security-policy</code>, <code>referrer-policy</code> and <code>permissions-policy</code>. <code>x-frame-options</code> and <code>x-content-type-options</code> are present.</dd></div>
          <div><dt class="text-gray-500">evidence</dt><dd><pre class="mt-1 overflow-x-auto rounded bg-black/60 p-2 text-[11px] text-emerald-300">missing=strict-transport-security, content-security-policy, referrer-policy, permissions-policy</pre></dd></div>
          <div><dt class="text-gray-500">fix</dt><dd class="copy text-gray-300">Add the four headers at the proxy or framework. Start HSTS at <code>max-age=300</code>, raise to a year once nothing breaks. A CSP of <code>default-src 'self'</code> plus your CDN hosts is a safe first version.</dd></div>
        </dl>
      </article>

      <article class="rounded border border-yellow-500/30 bg-yellow-500/5 p-4">
        <div class="flex flex-wrap items-baseline justify-between gap-2">
          <h3 class="font-mono text-sm font-bold text-white">F-03 · Open redirect via <code>?next=</code></h3>
          <span class="rounded bg-yellow-500/20 px-2 py-0.5 font-mono text-[11px] font-bold text-yellow-300">MEDIUM · CVSS 5.4</span>
        </div>
        <dl class="mt-3 space-y-2 font-mono text-xs">
          <div><dt class="text-gray-500">what we saw</dt><dd class="copy text-gray-300">Requesting <code>/?next=https://evil.example</code> answered with a 302 to that external address.</dd></div>
          <div><dt class="text-gray-500">evidence</dt><dd><pre class="mt-1 overflow-x-auto rounded bg-black/60 p-2 text-[11px] text-emerald-300">GET /?next=https://evil.example -> 302 https://evil.example</pre></dd></div>
          <div><dt class="text-gray-500">impact</dt><dd class="copy text-gray-300">Phishing links can carry your real domain and still land on an attacker's page.</dd></div>
          <div><dt class="text-gray-500">fix</dt><dd class="copy text-gray-300">Only accept a relative path that starts with a single <code>/</code>, or check the target against an allowlist of your own hosts.</dd></div>
        </dl>
      </article>

      <details class="rounded border border-white/10 p-4">
        <summary class="cursor-pointer font-mono text-xs text-gray-400 hover:text-white">F-04, F-05 and the four informational items — expand</summary>
        <div class="mt-3 space-y-3 font-mono text-xs text-gray-300">
          <p><span class="text-cyan-300">F-04 · Server version disclosed.</span> <code>server: nginx/1.18.0</code> tells an attacker exactly which CVEs to try. Set <code>server_tokens off;</code>.</p>
          <p><span class="text-cyan-300">F-05 · Certificate expires in 41 days.</span> Issuer Let's Encrypt, 3 SAN entries. Confirm auto-renewal is running; an expired certificate takes the site offline.</p>
          <p><span class="text-gray-500">I-01 · DNS.</span> Apex resolves to 203.0.113.10. <span class="text-gray-500">I-02 · Subdomains.</span> 7 hosts found via certificate transparency, including <code>staging.shop.example.com</code> — check it is meant to be public. <span class="text-gray-500">I-03 · WAF.</span> Cloudflare fingerprint present. <span class="text-gray-500">I-04 · robots.txt.</span> Disallows <code>/admin/</code>, which confirms the path exists.</p>
        </div>
      </details>
    </div>
  </section>

  <!-- Remediation plan -->
  <section class="mt-6 rounded border border-white/10 bg-[#04060c]">
    <div class="border-b border-white/10 px-4 py-2 font-mono text-[11px] text-gray-500">$ 5 · remediation plan (30 / 60 / 90)</div>
    <div class="grid gap-3 px-4 py-4 sm:grid-cols-3">
      <div class="rounded border border-red-500/30 p-3 font-mono text-xs"><div class="font-bold text-red-300">today</div><p class="copy mt-1 text-gray-300">Block <code>/.env</code>. Rotate every secret in it.</p></div>
      <div class="rounded border border-yellow-500/30 p-3 font-mono text-xs"><div class="font-bold text-yellow-300">this month</div><p class="copy mt-1 text-gray-300">Fix the redirect. Add HSTS, CSP, Referrer-Policy, Permissions-Policy.</p></div>
      <div class="rounded border border-cyan-500/30 p-3 font-mono text-xs"><div class="font-bold text-cyan-300">next quarter</div><p class="copy mt-1 text-gray-300">Hide the server version. Decide whether <code>staging.</code> should be public. Re-scan (1 credit) to confirm.</p></div>
    </div>
  </section>

  <section class="mt-6 rounded border border-white/10 bg-[#04060c] px-4 py-4 font-mono text-xs text-gray-400">
    <div class="text-gray-500">$ 6 · what this report does not cover</div>
    <p class="copy mt-2">Nothing behind a login, no authorization (IDOR) or business-logic testing, and no injection payloads (SQLi, XSS, LFI). Those need the <a href="/#pricing" class="text-cyan-400 hover:underline">whitebox tier</a>. Full list on the <a href="/product-scope" class="text-cyan-400 hover:underline">product scope</a> page.</p>
  </section>

  <div class="mt-8 flex flex-wrap items-center gap-3">
    <a href="/checkout?package=starter" class="rounded border-2 border-emerald-400 bg-emerald-500/10 px-5 py-2.5 font-mono text-sm font-bold text-emerald-300 hover:bg-emerald-500/20 glow">get this for your domain — $50</a>
    <a href="/#pricing" class="font-mono text-xs text-gray-500 hover:text-white">all packages →</a>
  </div>
</main>
<footer class="mt-10 border-t border-white/10 py-6 text-center font-mono text-[11px] text-gray-600">
  &copy; 2026 Dalang Pte Ltd · <a href="/" class="text-cyan-400 hover:underline">netsekurity.com</a>
</footer>
</body>
</html>`
