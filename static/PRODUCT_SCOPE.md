# Product Scope — Netsekurity External Web Security Assessment

Client-facing scope of the **automated external web security assessment** delivered per
verified domain / credit. 1 credit = 1 external assessment · 1 domain.

> **Plainly stated:** this is an **automated, external, blackbox, unauthenticated, read-only
> assessment** with human-reviewed findings — **not a full penetration test**. It does not
> test authenticated application logic, authorization (IDOR), or business logic. For that,
> see the **Whitebox** tier below.

---

## Methodology (blackbox, external)

1. **Recon & attack-surface mapping** — DNS resolution, passive subdomain enumeration,
   open-port discovery, technology & platform fingerprinting.
2. **Web fingerprint** — HTTP/S response analysis, server & WAF detection, CMS/stack identification.
3. **Security headers** — full OWASP set: HSTS, X-Content-Type-Options, X-Frame-Options,
   Content-Security-Policy, Referrer-Policy, Permissions-Policy; clickjacking check.
4. **TLS / PKI** — certificate validity, expiry window, issuer, SAN coverage.
5. **Information disclosure** — exposed `.git`, `.env`, backups, `server-status`, `phpinfo`,
   `robots.txt`, SVN metadata, source maps, debug endpoints, secrets.
6. **Common web vulnerability probing** — SQLi, LFI/RFI, XSS, open redirect, path traversal,
   server misconfiguration.
7. **Human validation & curation** — a security engineer reviews every finding before
   delivery: removes false positives, merges duplicates, checks evidence, adjusts severity,
   re-tests critical items where feasible, and curates the final report.

Findings are **mapped to OWASP Top 10 categories** and scored with **CVSS v3.1**, each with
CWE mapping, reproduction steps, and remediation guidance.

---

## In Scope (included in every scan)

| Area | Checks performed |
|------|------------------|
| **Recon & attack-surface mapping** | DNS, passive subdomain enumeration, open ports, tech/WAF fingerprint |
| **Web application fingerprint** | HTTP/S response, server & WAF, CMS/stack identification |
| **Security headers** | HSTS, X-Content-Type-Options, X-Frame-Options, CSP, Referrer-Policy, Permissions-Policy; clickjacking |
| **TLS / PKI** | Certificate validity, expiry, issuer, SAN coverage |
| **Information disclosure** | `.git`, `.env`, backups, server-status, phpinfo, robots.txt, SVN metadata, source maps, debug endpoints, secrets |
| **Common web vulnerabilities** | Unauthenticated probing for SQLi, LFI/RFI, XSS, open redirect, path traversal, server misconfiguration |

## Coverage note (be precise)

- **"OWASP-mapped"** means findings are *categorized* per OWASP Top 10. It is **not** a claim
  that every OWASP category is exhaustively tested — some (e.g. broken access control)
  require authentication and multiple identities that an external scan cannot reach.
- **Not claimed:** SSRF, GraphQL, WebSocket, request-smuggling, web-cache-poisoning, and other
  classes not listed above. Their absence from a report means "not scanned for," not "verified safe."
- **"Read-only"** is safe for production: we may indicate a parameter appears injectable but
  do not extract live database records as destructive proof.

## Out of Scope (not included — clearly stated)

1. **Authenticated testing.** Scans are unauthenticated; logic behind a login/session is not exercised.
2. **Authorization / IDOR / privilege escalation.** Requires multiple identities; not tested.
3. **Business-logic, workflow, payment, or race-condition testing.** Not included.
4. **Deep / manual exploit chains.** We detect vulnerabilities; we do not build multi-stage exploit chains.
5. **Source-code analysis.** Blackbox only.
6. **Destructive actions.** All testing is read-only and non-destructive.
7. **Extended manual engagement.** Not a multi-day APT-style engagement.

## Deliverables

- **Report (PDF, English)** per domain, downloadable from the dashboard:
  - Executive summary & overall security posture score
  - Risk register (likelihood × impact)
  - Detailed findings: severity, CVSS v3.1, CWE mapping, evidence, reproduction steps, remediation guidance
  - Compliance & control mapping
- Remediation roadmap (30/60/90) per finding.

## Whitebox (deeper assurance — separate offering)

For authenticated testing, authorization/business-logic analysis, and exploit-path validation:

| Tier | What's included | Price |
|------|-----------------|-------|
| **Blackbox (automated)** | External, unauthenticated assessment — this product | **Included in credits** ($50/credit) |
| **Whitebox (agent + human expert)** | Source review, authenticated testing, business-logic & exploit-chain analysis, human-written report | **$10,000 USD per app / per domain** |

> Whitebox engagements are scoped per app/domain and include a dedicated human security
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

*Scope per domain. 1 credit = 1 external assessment · 1 domain. Full technical detail in `AGENT_WORKER.md`.*