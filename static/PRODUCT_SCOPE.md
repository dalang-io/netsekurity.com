# Product Scope — What's Included in a Netsekurity Pentest

Client-facing scope of the automated pentest delivered per verified domain / credit.

---

## Testing Methodology: Blackbox

Netsekurity scans are **blackbox** — performed from a public, external perspective without
credentials, source code, or internal access. This reflects what an external attacker can
see and exploit.

---

## Whitebox Options (source + credentials + human expert)

For clients who need deeper assurance, **whitebox** testing is available — conducted by a
human security engineer assisted by the agent, with source code access, credentials, and
internal architecture knowledge.

| Tier | What's included | Price |
|------|-----------------|-------|
| **Blackbox (automated)** | External, unauthenticated scan — this product | **Included in credits** ($50/credit) |
| **Whitebox (agent + human expert)** | Source review, authenticated testing, business-logic & exploit-chain analysis, human-written report | **$10,000 USD per app / per domain** |

> Whitebox engagements are scoped per application or per domain, and include a dedicated
> human security engineer plus agent-assisted deep analysis. Contact us to schedule.

---

## In Scope (included in every scan)

| Area | Checks performed |
|------|------------------|
| **Recon & attack-surface mapping** | DNS resolution, subdomain enumeration (passive), open-port discovery, technology & platform fingerprinting |
| **Web application fingerprint** | HTTP/S response analysis, server & WAF detection, CMS/stack identification |
| **Security headers** | Full OWASP set: HSTS, X-Content-Type-Options, X-Frame-Options, Content-Security-Policy, Referrer-Policy, Permissions-Policy; clickjacking check |
| **TLS / PKI** | Certificate validity, expiry window, issuer, SAN coverage (openssl-based) |
| **Information disclosure** | Exposed source/version files (`.git`, `.env`, backups), server-status, phpinfo, robots.txt, SVN metadata |
| **Common web vulnerabilities** | Unauthenticated probing for SQL injection, LFI/RFI, XSS, open redirect, path traversal, and server misconfiguration (Nikto + custom checks) |

## Deliverables

- **Pentest report (PDF, English)** per domain, downloadable from the dashboard:
  - Executive summary & overall security posture score
  - Risk register (likelihood × impact)
  - Detailed findings with severity, CVSS v3.1 scoring, CWE mapping, evidence, reproduction steps, and remediation guidance
  - Compliance & control mapping
- Remediation roadmap (30/60/90) per finding.

## Out of Scope (not included — important to state clearly)

1. **Authenticated testing.** Scans are performed **unauthenticated** (from a public, no-credential perspective). Application logic behind a login/session is not exercised.
2. **Deep / manual exploit-chain assessment.** We detect vulnerabilities and misconfigurations, but do not build multi-stage manual exploit chains (e.g. chained RCE), which require human security-engineering analysis.
3. **Destructive actions.** All testing is **read-only and non-destructive**; no data modification, deletion, or defacement is ever performed.
4. **Extended manual engagement.** This is an automated per-domain assessment, not a multi-day APT-style engagement.
5. **Complex auth-heavy applications.** Web apps with deep role-based access control or extensive authenticated APIs may warrant a supplemental **manual penetration test** for full coverage.

## Best-fit use cases

- Quick, recurring **security posture check** of a public web property.
- **Pre-launch smoke assessment** before a production release.
- **Compliance hygiene** (headers, TLS, exposed data) evidence for audits.
- Baseline layer that can be **escalated to a manual pentest** for critical findings.

## Recommended companion (optional)

For **production transactional applications** (fintech, e-commerce, government-facing), we
recommend pairing the automated scan with a **supplemental manual penetration test** to cover
authenticated business-logic and deeper exploit-path verification.

---

*Scope per domain. 1 credit = 1 pentest · 1 domain. Full technical detail in `AGENT_WORKER.md`.*
