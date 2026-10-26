package main

import (
	"html/template"
	"net/http"
)

// /sample-report — what a customer gets for a credit, before they pay for one.
//
// The product is a PDF. Nothing on the site showed one, so a visitor was asked
// to spend $50 on a document they had never seen. This page mirrors the real
// report's structure (Document Control, Findings Summary, Risk Register,
// Findings & Analysis, Remediation Plan) with findings drawn only from checks
// nsec_scan.py actually performs, against a fictional domain. The full PDF also
// carries a methodology section and an evidence appendix.
//
// Wording rules (CLAIMS.md): no statistic without a source, and nothing the
// external scan cannot establish (it cannot say a site is, or is not,
// compromised). Exposure probes discard the response body, so the sample says
// "kept", not "downloaded".
func handleSampleReport(w http.ResponseWriter, r *http.Request) {
	body := template.HTML(loadSitePricing().replacer().Replace(sampleReportBody))
	serveRawPage(w, r, rawPage{
		Key: "sample", Path: "/sample-report", Crumb: "Sample report", Body: body,
		Title:       "Sample Security Assessment Report: What One Credit Delivers",
		Description: "A worked example of a Netsekurity external web security assessment report: document control, findings summary, risk register, evidence and a 30/60/90-day plan.",
	})
}

const sampleReportBody = `
<section class="border-b-2 border-black bg-[#fffbeb]">
  <div class="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:py-16">
    <span class="nsk-tag">Sample · fictional domain, real report structure</span>
    <h1 class="mt-4 max-w-4xl text-4xl font-extrabold leading-[1.1] tracking-tight sm:text-5xl">What one credit delivers</h1>
    <p class="mt-5 max-w-3xl text-lg text-gray-800">Every finding below comes from a check the scanner really performs. The domain and the evidence are invented. Your report is a PDF in English with this structure, downloadable from the dashboard; the full PDF also contains the methodology and an evidence appendix.</p>
    <p class="mt-5"><a href="/dashboard" class="nsk-btn">Get this for your domain · {single}</a> <a href="/#pricing" class="ml-2 font-bold underline">All packages →</a></p>
  </div>
</section>

<div class="mx-auto max-w-7xl space-y-8 px-4 py-12 sm:px-6">

  <section class="nsk-card" aria-labelledby="s1">
    <h2 id="s1" class="nsk-card-head">1 · Document control</h2>
    <dl class="grid divide-y-2 divide-gray-200 text-sm sm:grid-cols-2 sm:divide-y-0">
      <div class="flex justify-between gap-4 px-5 py-3"><dt class="font-extrabold">Target</dt><dd class="nsk-mono">shop.example.com</dd></div>
      <div class="flex justify-between gap-4 px-5 py-3"><dt class="font-extrabold">Assessment type</dt><dd>External · unauthenticated · read-only</dd></div>
      <div class="flex justify-between gap-4 px-5 py-3"><dt class="font-extrabold">Ownership</dt><dd>Verified via TXT <span class="nsk-mono">_netsekurity</span></dd></div>
      <div class="flex justify-between gap-4 px-5 py-3"><dt class="font-extrabold">Completed</dt><dd>2026-09-01 09:14 UTC</dd></div>
      <div class="flex justify-between gap-4 px-5 py-3"><dt class="font-extrabold">Prepared by</dt><dd>Netsekurity.com Security Agent</dd></div>
      <div class="flex justify-between gap-4 px-5 py-3"><dt class="font-extrabold">Classification</dt><dd>Confidential: owner and operators only</dd></div>
    </dl>
  </section>

  <section class="nsk-card" aria-labelledby="s2">
    <h2 id="s2" class="nsk-card-head">2 · Findings summary</h2>
    <div class="p-5">
      <p class="flex flex-wrap gap-2">
        <span class="nsk-badge nsk-badge-bad">1 High</span>
        <span class="nsk-badge nsk-badge-warn">2 Medium</span>
        <span class="nsk-badge nsk-badge-blue">2 Low</span>
        <span class="nsk-badge nsk-badge-muted">4 Info</span>
      </p>
      <p class="mt-4 text-gray-800"><strong class="font-extrabold">Overall posture: needs attention.</strong> One file exposure can leak credentials today; the remaining items are hardening.</p>
    </div>
  </section>

  <section class="nsk-card" aria-labelledby="s3">
    <h2 id="s3" class="nsk-card-head">3 · Risk register</h2>
    <div class="overflow-x-auto">
      <table class="w-full min-w-[640px] text-left text-sm">
        <thead><tr class="border-b-2 border-black bg-[#fffbeb]"><th scope="col" class="px-4 py-3">#</th><th scope="col" class="px-4 py-3">Finding</th><th scope="col" class="px-4 py-3">Severity</th><th scope="col" class="px-4 py-3">CVSS</th><th scope="col" class="px-4 py-3">CWE</th><th scope="col" class="px-4 py-3">OWASP</th></tr></thead>
        <tbody class="[&>tr]:border-b-2 [&>tr]:border-gray-200">
          <tr><td class="px-4 py-3 font-extrabold">F-01</td><td class="px-4 py-3">Environment file exposed at /.env</td><td class="px-4 py-3"><span class="nsk-badge nsk-badge-bad">High</span></td><td class="px-4 py-3">7.5</td><td class="px-4 py-3">CWE-538</td><td class="px-4 py-3">A05 · A02</td></tr>
          <tr><td class="px-4 py-3 font-extrabold">F-02</td><td class="px-4 py-3">Missing security headers (4 of 6)</td><td class="px-4 py-3"><span class="nsk-badge nsk-badge-warn">Medium</span></td><td class="px-4 py-3">5.3</td><td class="px-4 py-3">CWE-1021</td><td class="px-4 py-3">A05</td></tr>
          <tr><td class="px-4 py-3 font-extrabold">F-03</td><td class="px-4 py-3">Open redirect via ?next=</td><td class="px-4 py-3"><span class="nsk-badge nsk-badge-warn">Medium</span></td><td class="px-4 py-3">5.4</td><td class="px-4 py-3">CWE-601</td><td class="px-4 py-3">A01</td></tr>
          <tr><td class="px-4 py-3 font-extrabold">F-04</td><td class="px-4 py-3">Server version disclosed (nginx/1.18.0)</td><td class="px-4 py-3"><span class="nsk-badge nsk-badge-blue">Low</span></td><td class="px-4 py-3">3.7</td><td class="px-4 py-3">CWE-200</td><td class="px-4 py-3">A05</td></tr>
          <tr><td class="px-4 py-3 font-extrabold">F-05</td><td class="px-4 py-3">TLS certificate expires in 41 days</td><td class="px-4 py-3"><span class="nsk-badge nsk-badge-blue">Low</span></td><td class="px-4 py-3">3.7</td><td class="px-4 py-3">CWE-295</td><td class="px-4 py-3">A02</td></tr>
          <tr><td class="px-4 py-3 font-extrabold">I-01…04</td><td class="px-4 py-3">DNS resolution · 7 subdomains · WAF detected · robots.txt disallow entries</td><td class="px-4 py-3"><span class="nsk-badge nsk-badge-muted">Info</span></td><td class="px-4 py-3">—</td><td class="px-4 py-3">—</td><td class="px-4 py-3">—</td></tr>
        </tbody>
      </table>
    </div>
  </section>

  <section class="nsk-card" aria-labelledby="s4">
    <h2 id="s4" class="nsk-card-head">4 · Findings and analysis</h2>
    <div class="divide-y-2 divide-gray-200">
      <article class="space-y-2 p-5">
        <h3 class="text-lg font-extrabold">F-01 · Environment file exposed at /.env <span class="nsk-badge nsk-badge-bad ml-2">High · CVSS 7.5</span></h3>
        <dl class="space-y-2 text-sm text-gray-800">
          <div><dt class="inline font-extrabold">What we saw: </dt><dd class="inline">The path /.env returned HTTP 200 with a 1.2 KB body. Files with this name normally hold database passwords, API keys and signing secrets. We did not keep the contents.</dd></div>
          <div><dt class="inline font-extrabold">Evidence: </dt><dd class="inline nsk-mono">GET /.env -&gt; 200 1247</dd></div>
          <div><dt class="inline font-extrabold">Impact: </dt><dd class="inline">Anyone can read your production secrets.</dd></div>
          <div><dt class="inline font-extrabold">Fix: </dt><dd class="inline">Block the file at the web server (nginx: <span class="nsk-mono">location ~ /\.env { deny all; }</span>), move it outside the web root, then rotate every secret it contains: assume it has already been read.</dd></div>
        </dl>
      </article>
      <article class="space-y-2 p-5">
        <h3 class="text-lg font-extrabold">F-02 · Missing security headers <span class="nsk-badge nsk-badge-warn ml-2">Medium · CVSS 5.3</span></h3>
        <dl class="space-y-2 text-sm text-gray-800">
          <div><dt class="inline font-extrabold">What we saw: </dt><dd class="inline">The response lacks strict-transport-security, content-security-policy, referrer-policy and permissions-policy. x-frame-options and x-content-type-options are present.</dd></div>
          <div><dt class="inline font-extrabold">Evidence: </dt><dd class="inline nsk-mono">missing=strict-transport-security, content-security-policy, referrer-policy, permissions-policy</dd></div>
          <div><dt class="inline font-extrabold">Fix: </dt><dd class="inline">Add the four headers at the proxy or framework. Start HSTS at max-age=300, raise it to a year once nothing breaks. A CSP of <span class="nsk-mono">default-src 'self'</span> plus your CDN hosts is a safe first version.</dd></div>
        </dl>
      </article>
      <article class="space-y-2 p-5">
        <h3 class="text-lg font-extrabold">F-03 · Open redirect via ?next= <span class="nsk-badge nsk-badge-warn ml-2">Medium · CVSS 5.4</span></h3>
        <dl class="space-y-2 text-sm text-gray-800">
          <div><dt class="inline font-extrabold">What we saw: </dt><dd class="inline">Requesting /?next=https://evil.example answered with a 302 to that external address.</dd></div>
          <div><dt class="inline font-extrabold">Evidence: </dt><dd class="inline nsk-mono">GET /?next=https://evil.example -&gt; 302 https://evil.example</dd></div>
          <div><dt class="inline font-extrabold">Impact: </dt><dd class="inline">Phishing links can carry your real domain and still land on an attacker's page.</dd></div>
          <div><dt class="inline font-extrabold">Fix: </dt><dd class="inline">Only accept a relative path that starts with a single /, or check the target against an allowlist of your own hosts.</dd></div>
        </dl>
      </article>
      <details class="p-5">
        <summary class="cursor-pointer font-extrabold">F-04, F-05 and the four informational items</summary>
        <div class="mt-3 space-y-2 text-sm text-gray-800">
          <p><strong class="font-extrabold">F-04 · Server version disclosed.</strong> <span class="nsk-mono">server: nginx/1.18.0</span> tells an attacker exactly which CVEs to try. Set <span class="nsk-mono">server_tokens off;</span>.</p>
          <p><strong class="font-extrabold">F-05 · Certificate expires in 41 days.</strong> Issuer Let's Encrypt, 3 SAN entries. Confirm auto-renewal is running; an expired certificate takes the site offline.</p>
          <p><strong class="font-extrabold">I-01 · DNS.</strong> Apex resolves to 203.0.113.10. <strong class="font-extrabold">I-02 · Subdomains.</strong> 7 hosts found via certificate transparency, including staging.shop.example.com: check it is meant to be public. <strong class="font-extrabold">I-03 · WAF.</strong> Cloudflare fingerprint present. <strong class="font-extrabold">I-04 · robots.txt.</strong> Disallows /admin/, which confirms the path exists.</p>
        </div>
      </details>
    </div>
  </section>

  <section class="nsk-card" aria-labelledby="s5">
    <h2 id="s5" class="nsk-card-head">5 · Remediation plan (30 / 60 / 90)</h2>
    <div class="grid gap-0 divide-y-2 divide-gray-200 md:grid-cols-3 md:divide-x-2 md:divide-y-0">
      <div class="p-5"><p class="text-sm font-extrabold uppercase tracking-wide text-gray-700">Today</p><p class="mt-2 text-gray-800">Block /.env. Rotate every secret in it.</p></div>
      <div class="p-5"><p class="text-sm font-extrabold uppercase tracking-wide text-gray-700">This month</p><p class="mt-2 text-gray-800">Fix the redirect. Add HSTS, CSP, Referrer-Policy and Permissions-Policy.</p></div>
      <div class="p-5"><p class="text-sm font-extrabold uppercase tracking-wide text-gray-700">Next quarter</p><p class="mt-2 text-gray-800">Hide the server version. Decide whether staging. should be public. Re-scan (1 credit) to confirm.</p></div>
    </div>
  </section>

  <section class="nsk-alert nsk-alert-warn" aria-labelledby="s6">
    <h2 id="s6" class="text-lg font-extrabold">6 · What this report does not cover</h2>
    <p class="mt-2">Nothing behind a login, no authorization (IDOR) or business-logic testing, and no injection payloads (SQLi, XSS, LFI). Those need the whitebox tier. A finding that does not appear means “not scanned for”, not “verified safe”. The full list is on the <a class="font-bold underline" href="/product-scope">product scope</a> page.</p>
  </section>

  <p class="flex flex-wrap items-center gap-3"><a href="/dashboard" class="nsk-btn">Get this for your domain · {single}</a><a href="/#pricing" class="font-bold underline">All packages →</a></p>
</div>
`
