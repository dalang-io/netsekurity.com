# CLAIMS.md — what netsekurity.com may say, and where each statement comes from

The public site is read by boards, lead engineers and procurement teams, who test
claims. A claim that fails a test costs more trust than the page ever earned. So:

1. **Never invent a number, a certification, a customer or a duration.** If it is
   not in the table below, it does not go on a page.
2. **Company facts** come from the Dalang Company Profile 2026
   (`marketing.dalang.io/company_decks/`), as of **20 October 2026**. When a deck
   and a page disagree, the deck wins (workspace `CLAUDE.md`).
3. **Product facts** come from the code (`nsec_scan.py`, the Go handlers) and
   `static/PRODUCT_SCOPE.md`. Where the two disagreed on 2026-10-26, the code won
   and the scope document was corrected.
4. **Both languages move together.** Copy lives in `site_copy_en.go` and
   `site_copy_id.go`; `TestSiteCopyIsComplete` fails if they differ in shape.
5. **Prices are never typed.** Pages use `{single}`, `{volume}`, `{volumeCredits}`,
   `{whitebox}` and `{packs}`, filled from `credit_packages` at request time. The
   same goes for `/llms.txt` and `/llms-full.txt`, which are generated.
6. After changing a claim: update the source, this file, both languages, bump
   `siteModified` in `site.go`, run `go test ./...`, and run `tools/og/render.sh`
   if the headline or sample card changed.

## Claims the site makes

| Claim | Source | Notes |
|---|---|---|
| Operator is Dalang Pte. Ltd., UEN 202622331M, 160 Robinson Road #14-04, Singapore 068914 | Company Profile 2026, p. 5 | as of 20 Oct 2026 |
| Netsekurity is a Dalang product, not a partner | Owner ruling 2026-10-23; `PLAN-dalang-integration.md` | |
| The Dalang group operates its own network, AS154547, checkable at bgp.he.net/AS154547 | Company Profile 2026, pp. 5, 17, 24 | Stated as a **group** fact; the AS belongs to PT Mashanz Sumber Terbuka. Never attributed to the scanner. |
| More than 35 read-only checks | `nsec_scan.py`, non-deep set; `AGENT_WORKER.md` counts 38 | "35+" is a safe lower bound. |
| 13 exposed-file and path probes | `nsec_scan.py` `exposed()` | includes robots.txt |
| 9 open-redirect parameters | `nsec_scan.py` `open_redirect()` | |
| 6 security headers, clickjacking, server-version disclosure, WAF detection | `nsec_scan.py` `http_fp()` | WAF detection is cookie-based. |
| TLS validity, days to expiry, issuer, SAN count, handshake failure | `nsec_scan.py` `tls()` | |
| Subdomains are **listed**, not assessed (capped at 80); one hostname = one credit | `nsec_scan.py` `recon()`; owner FAQ copy | |
| No network port scanning in the standard assessment | `AGENT_WORKER.md`: nmap runs only with `--deep-scan`, which the customer cron never passes | **Corrected** in `PRODUCT_SCOPE.md` 2026-10-26 (it listed open-port discovery). |
| No CMS or stack identification beyond the Server header | whatweb is deep-only | **Corrected** in `PRODUCT_SCOPE.md`. |
| Requests are read-only GETs, a DNS lookup and a TLS handshake | `PRODUCT_SCOPE.md`; `nsec_scan.py` | |
| Every HTTP request carries the User-Agent `NetSekurity Scanner` | **Owner statement 2026-10-26**; `nsec_scan.py` `UA` constant | Changed in this repo on 2026-10-26 (it was a Chrome string). **True for customers only once the agent host has pulled it** (`HERMES_UPDATE.md`). `TestScannerUserAgentMatchesThePublicClaim` ties the pages to the source. |
| Exposure and redirect probes are spaced 0.3 seconds apart | `nsec_scan.py` `time.sleep(0.3)` in `exposed()` and `open_redirect()` | |
| Exposure probes record status code and size and discard the body | `nsec_scan.py`: `curl -o /dev/null -w "%{http_code} %{size_download}"` | Exception: robots.txt, whose disallowed paths are summarised (first 200 characters). The body is transferred, then discarded, so the sample report says "did not keep", never "did not download". |
| Standard report is agent-triaged and no human reads it before delivery | `PRODUCT_SCOPE.md` methodology step 7 | |
| Report: PDF, English only, one per domain | `AGENT_WORKER.md` § Report | The full PDF also has a methodology section and an evidence appendix. |
| Findings carry severity, CVSS v3.1 score where one applies, CWE and CVE references, OWASP category, a proof of concept (the request that shows the issue), impact, fix | `nsec_scan.py` (`cvss`, `cwe`); sample report; **owner statement 2026-10-26** for CWE/CVE and PoC | The scanner itself assigns only CWE and records the request as evidence. CVE references, the PoC text and the OWASP mapping are produced by `report_auto.py`, which is not in this repo. This replaces "Compliance & control mapping" in `PRODUCT_SCOPE.md`. |
| 30/60/90-day plan; executive summary; risk register; scope statement | Sample report | |
| 1 credit = 1 assessment of 1 verified domain | `README.md`; `PRODUCT_SCOPE.md` | |
| Prices of the packs | `credit_packages` (seeds in `db.go`) | Rendered from the database. |
| Whitebox is $10,000 per app or domain; expert review is priced on request | `PRODUCT_SCOPE.md` | `whiteboxUSD` in `site.go`. |
| Credits do not expire | `checkout.go` copy ("never expire") | |
| A failed scan is refunded automatically | `pentests.go` worker-fail path; `pentests_fail_test.go` | Self-serve scans only. |
| A scan takes 10 minutes to 24 hours, depending on what it finds | **Owner statement, 2026-10-26** | Not measured by anything in this repo. Always quote the whole range; never a single figure or a narrower one (`TestScanDurationIsAlwaysTheStatedRange`). |
| Ownership proven by a DNS TXT record at `_netsekurity.<domain>`, checked against public resolvers and authoritative name servers | `domains.go` `handleVerifyDomain` | |
| One assessment at a time per account | `pentests.go` in-flight cap | |
| API tokens shown once, stored as SHA-256 hash, expire after 7–90 days | `apitokens.go` | |
| Session cookie: signed, 24 h, HttpOnly, Secure, SameSite=Lax | `auth.go` | |
| Reports served only to the owning account and operators, `no-store`, noindex | `admin.go` `handleReport` | |
| What is kept (account, domains, findings and evidence strings, report PDF, invoice records, contact messages with IP) | `db.go` schema; `handleContact` | |
| Umami is cookieless; the Meta pixel loads only after consent | `consent.go` (owner decision 2026-10-16) | |
| Card details are handled by Xendit, never by us | `checkout.go` copy | |
| Disclosure: security@dalang.io, acknowledged within 2 business days, good-faith researchers not pursued, credit on request | `PRODUCT_SCOPE.md` § Vulnerability disclosure | |
| Sales replies within 1 business day | Owner copy (contact form, earlier `llms.txt`) | |
| Accounts are individual; sign-in is with Google | `auth.go`; no organisation model in `db.go` | |
| Dalang's privacy policy, terms of use and security and trust center are at dalang.io/privacy, /terms-of-use and /security | **Owner statement 2026-10-26**; all three return HTTP 200 | They are Dalang's documents. As of 2026-10-26 none of them names Netsekurity, so the site links them as Dalang's and does not say they cover anything in particular. |

## What the site deliberately does not claim

| Not claimed | Why |
|---|---|
| ISO 27001, ISO 9001 | The certificates belong to the **IDC Indonesia facility**, not to any Dalang entity and not to Netsekurity (Company Profile 2026, p. 18). |
| Komdigi Private PSE registration | Registered to **PT Mashanz Sumber Terbuka**, not to Dalang Pte. Ltd. |
| Customers (Telkom, BINUS, TrustMedis …) | They are customers of Dalang **infrastructure**, not of Netsekurity. Using them here would be borrowed proof. |
| Facility, uptime or Tier figures | They describe Dalang's data centres, not where or how the scanner runs. |
| Compliance-framework mapping (ISO 27001, PCI DSS …) in reports | The report carries CWE/CVE references and a PoC, not a framework mapping. Do not describe it as one. |
| Scanner source IP addresses | None is published. The User-Agent is the only identifier the site states. |
| Data retention or deletion period, DPA, MSA, SLA, uptime commitment specific to Netsekurity | Nothing is published. Dalang's privacy policy and terms exist at dalang.io but do not name Netsekurity. |
| SSO, teams, roles | Accounts are individual. |
| Tax (VAT/PPN) treatment, tax invoices | Not stated for Netsekurity; the site tells buyers to ask before purchase. |
| Which languages sales or support answer in | No source (virus.ai.id applies the same rule). |

## Enforced by tests

`site_test.go` and `dashboard_test.go` fail on: a banned phrase (`bannedClaims`),
an unfilled `{token}`, a title or description out of range, a missing canonical or
hreflang, broken structured data, a visible FAQ that differs from its JSON-LD, a
broken internal link or anchor, a sitemap that differs from the pages, a scan
duration in customer copy, or an unmarked terminal-style class.
