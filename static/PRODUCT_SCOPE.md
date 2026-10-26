# Product Scope — Netsekurity External Web Security Assessment

Client-facing scope of the **automated external web security assessment** delivered per
verified domain / credit. 1 credit = 1 external assessment · 1 domain.

> **Plainly stated:** this is an **automated, external, blackbox, unauthenticated, read-only
> assessment** with agent-reviewed findings — **not a full penetration test**. It does not
> test authenticated application logic, authorisation (IDOR), or business logic. For that,
> see the **Whitebox** tier below.

---

## Methodology (blackbox, external)

1. **Recon & attack-surface mapping** — DNS resolution and passive subdomain listing from
   certificate-transparency and public sources. Subdomains are listed for your information;
   only the domain you verified is assessed.
2. **Web fingerprint** — the server header, server-version disclosure and web application
   firewall detection, read from the HTTP/S response.
3. **Security headers** — full OWASP set: HSTS, X-Content-Type-Options, X-Frame-Options,
   Content-Security-Policy, Referrer-Policy, Permissions-Policy; clickjacking check.
4. **TLS / PKI** — certificate validity, expiry window, issuer, SAN coverage.
5. **Information disclosure** — 13 probes: exposed `.git`, `.env`, SQL/zip backups,
   `wp-config` backup, `server-status`, `phpinfo`, SVN metadata, admin/login paths, `robots.txt`.
6. **Open redirect** — 9 common redirect parameters (`redirect`, `next`, `url`, `return`, …)
   tested for an external 302.
7. **Agent triage** — an analysis agent deduplicates the findings, ranks them by severity and
   business impact, and writes the report. No human reads the standard report before delivery;
   a security engineer's re-test is the separate **expert review add-on** (below).

Findings are **mapped to OWASP Top 10 categories** and scored with **CVSS v3.1**, each with
CWE mapping, reproduction steps, and remediation guidance.

---

## In Scope (included in every scan)

| Area | Checks performed |
|------|------------------|
| **Recon & attack-surface mapping** | DNS resolution, passive subdomain listing (subdomains are listed, not assessed) |
| **Web application fingerprint** | Server header, server-version disclosure, WAF detection |
| **Security headers** | HSTS, X-Content-Type-Options, X-Frame-Options, CSP, Referrer-Policy, Permissions-Policy; clickjacking |
| **TLS / PKI** | Certificate validity, expiry, issuer, SAN coverage |
| **Information disclosure** | `.git`, `.env`, backups, server-status, phpinfo, robots.txt, SVN metadata, admin/login paths |
| **Open redirect** | 9 common redirect parameters tested for an external 302 |

## Coverage note (be precise)

- **"OWASP-mapped"** means findings are *categorised* per OWASP Top 10. It is **not** a claim
  that every OWASP category is exhaustively tested — some (e.g. broken access control)
  require authentication and multiple identities that an external scan cannot reach.
- **Not claimed:** SSRF, GraphQL, WebSocket, request-smuggling, web-cache-poisoning, and other
  classes not listed above. Their absence from a report means "not scanned for," not "verified safe."
- **"Read-only"** is safe for production: every request is a plain GET; nothing is written,
  uploaded or exploited.

## Out of Scope (not included — clearly stated)

1. **Authenticated testing.** Scans are unauthenticated; logic behind a login/session is not exercised.
2. **Authorisation / IDOR / privilege escalation.** Requires multiple identities; not tested.
3. **Injection payload testing.** The standard scan sends no SQLi, XSS, LFI/RFI or path-traversal
   payloads. It finds exposure and misconfiguration, not exploitable input bugs.
4. **Business-logic, workflow, payment, or race-condition testing.** Not included.
5. **Network port scanning.** The standard assessment does not scan ports.
6. **Deep / manual exploit chains.** We detect vulnerabilities; we do not build multi-stage exploit chains.
7. **Source-code analysis.** Blackbox only.
8. **Destructive actions.** All testing is read-only and non-destructive.
9. **Extended manual engagement.** Not a multi-day APT-style engagement.

## Deliverables

- **Report (PDF, English)** per domain, downloadable from the dashboard:
  - Executive summary & overall security posture score
  - Risk register (likelihood × impact)
  - Detailed findings: severity, CVSS v3.1, CWE mapping, evidence, reproduction steps, remediation guidance
  - CWE / CVE references and a proof of concept (PoC) for each finding
- Remediation roadmap (30/60/90) per finding.

## Whitebox (deeper assurance — separate offering)

For authenticated testing, authorisation/business-logic analysis, and exploit-path validation:

| Tier | What's included | Price |
|------|-----------------|-------|
| **Blackbox (automated)** | External, unauthenticated assessment — this product | **Included in credits** (prices are on the pricing page) |
| **Expert review (add-on)** | A security engineer re-tests every finding in one completed report, removes false positives, and writes a fix plan for your stack | **Priced on request** |
| **Whitebox (agent + expert)** | Source review, authenticated testing, business-logic & exploit-chain analysis, agent-written report | **$10,000 USD per app / per domain** |

> Whitebox engagements are scoped per app/domain and include a dedicated security
> engineer plus agent-assisted deep analysis. Contact us to schedule.

## Best-fit use cases

- Quick, recurring **external security posture check** of a public web property.
- **Pre-launch smoke assessment** before a production release.
- **Compliance hygiene** (headers, TLS, exposed data) evidence for audits.
- Baseline layer that can be **escalated to a manual/whitebox pentest** for critical findings.

## Recommended companion (optional)

For **production transactional applications** (fintech, e-commerce, government-facing), pair the
automated external scan with the **whitebox** tier or a **supplemental manual pentest** to cover
authenticated business-logic and deeper exploit-path verification.

---

## Vulnerability disclosure (our own systems)

Found something in netsekurity.com or dalang.io itself? Email **security@dalang.io**. We
acknowledge within 2 business days, do not pursue good-faith researchers, and credit fixes on
request. Please do not test against other customers' domains.

---

*Scope per domain. 1 credit = 1 external assessment · 1 domain. A worked example is at
https://netsekurity.com/sample-report.*