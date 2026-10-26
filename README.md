# netsekurity.com — External Web Security Assessment

> **Positioning (honest):** "You build fast with AI. We check your production attack surface —
> with agent-reviewed findings." Netsekurity is a **continuous external web security assessment**
> (automated, blackbox, unauthenticated, read-only) with agent triage, **not a full
> penetration test**. Deeper whitebox testing is a separate paid offering.

1 credit = 1 external assessment · 1 domain. Verify ownership with an auto-generated TXT record;
get a plain-language, agent-reviewed report. **Pricing:** $50 = 1 · $100 = 3 · $500 = 20 · $1000 = 50 credits.

This README is the **handover/operations reference** for whoever takes over the project.

---

## Quick Facts

| | |
|---|---|
| **Live URL** | `https://netsekurity.com` (public HTTPS via Cloudflare → origin HTTP :80) |
| **Origin service** | systemd `netsekurity.service`, Go binary, listens on `:8094` |
| **Proxy** | Pingora (`dalang-proxy`) routes `netsekurity.com` + `www.netsekurity.com` → `127.0.0.1:8094` |
| **Server** | `root@163.128.54.5` (same box as dalang.io/api.dalang.io) |
| **App dir (server)** | `/home/dalang/netsekurity.com/` |
| **Database** | SQLite `data/netsekurity.db` (+ `data/reports/` for report PDFs) |
| **Repo** | `github.com/dalang-io/netsekurity.com` (branch `main`) |

---

## Tech Stack

| Layer | Choice |
|---|---|
| Language | Go (stdlib `net/http`) + `modernc.org/sqlite` (pure-Go SQLite driver, no CGO) |
| UI | HTMX (server-rendered fragments, no JS framework) |
| Styling | Tailwind CSS (built to `static/css/styles.css`, cache-busted with `?v=<sha>`) |
| Database | SQLite (single file) |
| Auth | Google OAuth 2.0 (Authorization Code + One Tap) + JWT session |
| Payments | Xendit invoices (auto-credit poller; webhook optional) |
| Hosting | Linux + systemd + Pingora proxy + Cloudflare (flexible SSL) |
| Scanning agent | Hermes agent + `nsec_scan.py` (see AGENT_WORKER.md) |
| Analytics | Meta Pixel (ads), consent-gated; Umami (cookieless) + optional Plausible / GA4 via `PLAUSIBLE_DOMAIN` / `GA_MEASUREMENT_ID` |

---

## Architecture

A single Go binary serves everything (landing, dashboard, admin, API, static) via embedded
assets (`go:embed`). It is stateless except for the SQLite file; sessions are JWT.

```
Browser (HTTPS) → Cloudflare (flexible) → Pingora :80
   Host: netsekurity.com → 127.0.0.1:8094 → netsekurity (Go)
     ├── / and /id/           home (English / Indonesian), rendered from site_copy_*.go
     ├── /enterprise, /trust, /external-assessment-vs-penetration-test  (+ /id/… twins)
     ├── /sample-report       public example of what one credit delivers (+ /id/contoh-laporan,
     │                        and a printable PDF of each: /netsekurity-sample-report-{en,id}.pdf)
     ├── /product-scope       PRODUCT_SCOPE[.id].md rendered to semantic HTML (md.go) (+ /id/cakupan-produk)
     ├── /docs                CI/CD + API reference (+ /id/dokumentasi)
     ├── /sitemap.xml, /llms.txt, /llms-full.txt   generated from the same copy and prices
     ├── /css/*, /og.png, /og-id.png, /robots.txt, /security.txt   static
     ├── /auth/*              Google OAuth
     ├── /dashboard           customer app, light theme + sidebar (HTMX): Home, then
     │   /dashboard/websites | /scans | /credits | /automation
     ├── /checkout, /su       order confirmation; super-admin panel
     ├── /api/*               top-up, domains/verify, pentests, api-tokens
     ├── /api/v1/assessments  CI/CD API (X-API-Token): POST starts one, GET /{id} reads status and
     │                        finding counts; /api/v1/pentests is the original name for the POST
     ├── /reports/*           pentest PDFs (owner/admin only, no-store)
     └── SQLite (data/netsekurity.db) + data/reports/
```

---

## Repository Layout

```
netsekurity.com/
├── main.go            # server bootstrap, routes; handleIndex dispatches to the page registry
├── auth.go            # Google OAuth + One Tap, JWT session, requireAuth / requireAdmin / isAdmin
├── jwt.go             # HS256 JWT (stdlib)
├── db.go              # SQLite init, schema, seed packages, super-admin seed
├── env.go             # .env loader
├── util.go            # parseFloat, secureCompare
├── dashboard.go       # customer dashboard: data loading, page handlers, template funcs, sidebar nav
├── dashboard_html.go  # dashboard templates: shell (top bar + sidebar + toast) and the five pages
├── ui.go              # shared UI helpers: credit/time formatting, icons, alertHTML()
├── admin.go           # super-admin panel (/su) + report serving + add-credit
├── payment.go         # Xendit top-up, webhook, auto-credit poller (syncPendingPayments)
├── domains.go         # domain add / TXT verify (authoritative NS) / delete
├── pentests.go        # pentest start/list + worker claim/report (queue)
├── apitokens.go       # CI/CD API tokens (expiry 7/14/30/60/90d) and the /api/v1 endpoints
├── security.go        # security headers on every response; strict nonce CSP for public pages
├── site.go            # public site engine: page registry, view model, SEO head, JSON-LD, sitemap
├── site_copy.go       # copy types + helpers (token expansion, completeness checks)
├── site_copy_en.go    # English copy (all pages)         ┐ same structure, enforced
├── site_copy_id.go    # Indonesian copy (all pages)      ┘ by TestSiteCopyIsComplete
├── site_html.go       # layout, shared partials, home / enterprise / trust / guide templates
├── site_evidence_html.go # sample report, docs and product scope templates
├── site_svg.go        # GENERATED isometric illustrations (tools/svg/iso.py); never edit by hand
├── site_llms.go       # /llms.txt and /llms-full.txt, generated from the copy and prices
├── md.go              # tiny Markdown renderer (for PRODUCT_SCOPE.md)
├── marketing.go       # contact form (→ contact_messages, localised), scan-count proof
├── analytics.go       # optional Plausible / GA4 tag for public pages
├── consent.go         # consent gate: Meta Pixel loads only after Accept (nsk_consent cookie)
├── fbevents.go        # Meta Pixel base code + Conversions API (server-side Lead/Purchase)
├── ratelimit.go       # per-IP fixed-window limiter on write endpoints
├── docs.go            # the CI examples (data, shared by both languages; notes are copy)
├── *_test.go          # unit tests (JWT, credit idempotency, topup amount, domain, index, auth)
├── deploy.sh          # test → build → commit+push → deploy (scp + systemd stop/start, PID check)
├── nsec_scan.py       # scanning agent scanner (enterprise-grade, produces PDF)
├── nsec_io.sh         # agent claim/upload I/O helpers
├── nsec_monitor.sh    # agent token-saver monitor (Hermes cron)
├── nsec_worker.sh     # standalone one-shot worker
├── AGENT_WORKER.md    # agent behavior reference (authoritative)
├── HERMES_UPDATE.md   # Hermes agent: update steps + verification runbook
├── CLAIMS.md          # every public claim with its source, and what is deliberately not claimed
├── marketing-copy.md  # older social/ads copy (vibe-coder angle; not used by the site)
├── tools/og/          # Open Graph card sources + render.sh (headless Chrome → static/og*.png)
├── tools/pdf/         # render.sh: the printable sample reports (static/netsekurity-sample-report-*.pdf)
├── tools/svg/         # iso.py: generates site_svg.go (isometric, three tones + ink outline)
├── env.example        # environment template
├── README.md          # this doc
└── static/
    ├── css/styles.css             # built Tailwind output (embedded)
    ├── og.png, og-id.png          # OG/Twitter share cards, 1200x630 (tools/og/render.sh)
    ├── fonts/, js/                # Plus Jakarta Sans (latin, variable) and htmx 1.9.12, self-hosted;
    │                              #   immutable cache; the htmx file's sha256 is pinned by a test
    ├── netsekurity-sample-report-{en,id}.pdf   # price-free printable sample (tools/pdf/render.sh)
    ├── sample-report.sources      # hash of the copy the PDFs were made from (TestSamplePDFsMatchTheirSource)
    ├── robots.txt                 # one shared rule set for search and AI crawlers
    ├── security.txt               # + .well-known/security.txt (RFC 9116); Policy → /trust#disclosure
    └── PRODUCT_SCOPE.md, PRODUCT_SCOPE.id.md   # client-facing scope & methodology, one per language
```

Build everything: `./deploy.sh`.

---

## Database (SQLite)

Created automatically at startup (`db.go`); migrations run idempotently via `schema_migrations`.

### Tables

| Table | Purpose | Key columns |
|---|---|---|
| `users` | Accounts (Google) | `id`, `email`, `name`, `google_sub`, **`role`** (`user`/`admin`), `created_at` |
| `credit_balance` | Spendable credits | `user_id`, `balance`, `updated_at` |
| `credit_transactions` | Credit ledger | `user_id`, `type` (`topup`/`spend`/`admin_credit`), `amount`, `description`, `reference_id` |
| `credit_packages` | Bundles | `id`, `name`, `usd_price`, `credits`, `is_active` |
| `domains` | User domains | `id`, `user_id`, `domain`, `txt_verification_token`, `status` (`pending`/`verified`), `verified_at` |
| `pentests` | Scan jobs | `id`, `user_id`, `domain_id`, `status` (`queued`/`running`/`completed`), `report_ref`, `started_at`, `completed_at` |
| `payments` | Xendit invoices | `user_id`, `external_id`, `xendit_invoice_id`, `package_id`, `amount_usd`, `credits`, `status` (`pending`/`paid`/`expired`), `currency`, `paid_at` |
| `api_tokens` | CI/CD tokens | `user_id`, `token_hash`, `label`, `expires_at` |
| `contact_messages` | Landing-page leads | `name`, `email`, `message`, `source`, `ip`, `handled_at` — listed on `/su` |

### Credit rules
- `1 credit = 1 external assessment · 1 domain`.
- Top-up credits are added by the **auto-credit poller** (`syncPendingPayments`, every 30s + when Home or Credits loads) — the Xendit **webhook is optional** (kept as a fallback).
- Spending a credit writes a `spend` row and decrements `balance` atomically.

---

## Authentication & Roles

- **Google OAuth 2.0** (Authorization Code + **One Tap** suggest-login). Redirect URI:
  `https://netsekurity.com/auth/google/callback` (must be registered in Google Console;
  `https://netsekurity.com` must be in Authorized JavaScript origins for One Tap).
- **JWT session** (HS256, `JWT_SECRET`) in an `httpOnly`, `Secure` cookie (`nsk_session`).
- `/login?next=/checkout?package=starter` returns the visitor to `next` after Google
  (same-site paths only, carried in the `nsk_next` cookie). The OAuth `state` is verified on
  the callback.
- With `ENVIRONMENT=production` the binary refuses to start unless `JWT_SECRET`,
  `BOT_AUTH_TOKEN` and `XENDIT_CALLBACK_KEY` are set.
- **Super admin** = `SUPER_ADMIN_EMAIL` (default `hans@dalang.io`) OR `users.role = 'admin'`.
  - `/su` panel: list users, add user, set role, **add credit**, list transactions
    (paid/expired), customer domains + Pentest button, pentests + PDF upload/list.
  - Protected pages set `Cache-Control: no-store`; reports are owner/admin-only (404 otherwise).

---

## Payments — Xendit

- **Top-up:** `POST /api/topup` → creates a Xendit invoice for the selected package.
  Currency = `XENDIT_CURRENCY` (code default **USD**). This Xendit account is not enabled
  for USD — a USD invoice is rejected with `UNSUPPORTED_CURRENCY` — so production sets
  `XENDIT_CURRENCY=IDR`; the checkout footnote and the `/su` banner state the converted
  amount. IDR amount = `usd_price × XENDIT_USD_RATE`. Switch back to USD only once Xendit
  enables it on the account.
- **Auto-credit:** the poller checks pending payments against Xendit and credits when
  `PAID`/`SETTLED` (idempotent). Runs every 30s + when Home or Credits loads. Webhook `/webhook/xendit`
  is optional (token-gated by `XENDIT_CALLBACK_KEY`).

### Dalang platform scan API

`api.dalang.io` sells scans to its own customers and uses Netsekurity as
fulfilment. The platform debits its wallet, then drives this API; no credit is
consumed here and a failed platform scan is refunded on the Dalang side.

Netsekurity is a **Dalang Pte Ltd product**, not a third-party partner. `partner`
survives only in internal identifiers (`/api/partner/v1`, `PARTNER_SECRET`,
`billing_source='partner'`).

| Endpoint | Auth | Purpose |
|---|---|---|
| `POST /api/partner/v1/scans/reserve` | HMAC | reserve a scan (not queued, not claimable) |
| `POST /api/partner/v1/scans/{id}/confirm` | HMAC | payment confirmed → `queued`, deadline starts (6h) |
| `POST /api/partner/v1/scans/{id}/release` | HMAC | abandon an unpaid reservation |
| `GET /api/partner/v1/scans/{id}` | HMAC | status, attempts, deadline, findings, signed report URL |
| `GET /api/partner/v1/scans/{id}/report?exp=&sig=` | signed URL | PDF download without a session |

HMAC is `sha256(PARTNER_SECRET, partner_id \n timestamp \n nonce \n body)`, with a
±5-minute window and a single-use nonce. `proof_type=txt` requires a verified
domain; `proof_type=cloudflare` records a Cloudflare-activated hostname asserted
by the platform; `proof_type=registrar` records the hosting panel's assertion
that the domain was sold/verified at the registrar
(`domains.verified_by='registrar'`). Platform-owned hostnames (`*.dalang.io`, `*.netsekurity.com`)
are refused. Attempts are capped at 3 and a scan that passes its deadline is
failed; see `AGENT_WORKER.md` and `PLAN-dalang-integration.md`.

---

## Domain Verification (TXT)

1. User adds a domain → server generates `ns-verify-<hex>` token.
2. Dashboard shows the record: **`_netsekurity.<domain>` TXT `<token>`** — **always visible**
   (pending & verified) so users can re-add/update DNS later.
3. Verify checks **recursive resolvers AND authoritative nameservers** (raw DNS query) so a
   fresh record verifies immediately. Verified domains can't be deleted.

---

## Pentest Pipeline (agent-driven)

```
User clicks "scan" → /api/pentests/start → 1 credit → pentests(status='queued')
Agent (every minute, Hermes cron) → nsec_monitor.sh → sees queue → nsec_worker.sh
  → /api/pentests/worker/claim        → status='running'
  → nsec_scan.py <domain>             → enterprise scan → PDF
  → /api/pentests/worker/report       → status='completed', report saved
User dashboard → /reports/YYYYMMDD-hh:mm-<domain>.pdf  (owner/admin only)
```

- Worker endpoints guarded by `X-Bot-Token` (`BOT_AUTH_TOKEN`).
- **Standard mode only.** A 2-credit "destructive" mode used to be accepted by the dashboard,
  `/su` and the API; the worker never passed it to the scanner, so it was the same scan at
  double the price. Removed 2026-09-10. Old rows keep `mode='destructive'` for history.
- **What the scan actually does** (38 checks, no `--deep-scan`, no injection payloads) is
  spelled out in `AGENT_WORKER.md`; every public claim must match it.
- API tokens for CI/CD auto-scan: `apitokens.go` (7/14/30/60/90-day expiry).

---

## Marketing & GEO assets

- `sitemap.xml`, `robots.txt` (allows AI crawlers), `llms.txt` + `llms-full.txt` (LLM/GEO),
  `security.txt` (+ `/.well-known/security.txt`), `og.png`, JSON-LD on landing.
- Client-facing scope/methodology: `static/PRODUCT_SCOPE.md`.
- Social/ads copy: `marketing-copy.md` (Vibe Coder + CI/CD use cases).

### Tracking & consent

No tracker talks to anyone before the visitor agrees (Owner decision, 2026-10-16;
same rule as `dalang.io`'s `consent.js`). `consent.go` is the netsekurity port:

- The Meta Pixel base code lives in `fbevents.go:metaPixelBaseJS` and is shipped
  as a **JSON string**, not a `<script>`. `consentGate()` puts it into the document
  only after Accept, so `connect.facebook.net` is never requested first — and with
  it go `www.facebook.com/tr` and the CAPI Gateway POST to `*.on.aws/events`, which
  Meta's own config script derives from the pixel.
- Every client event goes through `nskTrack()` (no-op without consent) instead of
  calling `fbq` directly, and the `<noscript>` tracking pixel is gone: it fired for
  visitors with JS off, who had no banner to answer.
- Server-side Conversions API events obey the same gate: `Lead` (`auth.go`) and
  `Purchase` (`payment.go`) are skipped unless the request — or, for the Xendit
  webhook, the `payments.meta_consent` stored at checkout — says `granted`.
- The choice is a cookie, `nsk_consent=granted|denied` (180 days, `SameSite=Lax`),
  not localStorage, because Go needs to read it too. Absent ≠ granted.
- Umami stays ungated: cookieless, no personal data — the exemption `dalang.io`
  documents in `src/lib/analytics/consent.js`.
- **Residual, not ours to switch off:** Cloudflare Web Analytics (RUM) is injected
  by the CDN, not by this app, and still fires before consent. It is turned off in
  the Cloudflare dashboard (Analytics → Web Analytics → Manage site → Advanced
  options), not in this repo.

---

## Environment Variables (`.env`)

Full template: [`env.example`](env.example). Key vars:

| Var | Notes |
|---|---|
| `PORT` | default `8094` |
| `DATABASE_PATH` | SQLite path (default `./data/netsekurity.db`) |
| `GOOGLE_CLIENT_ID/SECRET/REDIRECT_URL` | Google OAuth (reuse `dalang.io/.env`) |
| `JWT_SECRET` | session signing (reuse `api.dalang.io/.env`) |
| `SUPER_ADMIN_EMAIL` | default `hans@dalang.io` |
| `XENDIT_SECRET_KEY/CALLBACK_KEY` | Xendit (reuse `api.dalang.io/.env`) |
| `XENDIT_CURRENCY` | `USD` (code default) or `IDR` (required while the account is IDR-only) |
| `XENDIT_USD_RATE` | IDR-per-USD for IDR charges |
| `XENDIT_API_BASE_URL` | invoice API origin; default `https://api.xendit.co`. Set only to point tests at a local stub |
| `PARTNER_ID` / `PARTNER_SECRET` | Dalang partner API; empty secret disables the endpoints (503) |
| `BOT_AUTH_TOKEN` | agent↔platform secret (`openssl rand -hex 32`) |
| `PLAUSIBLE_DOMAIN` / `GA_MEASUREMENT_ID` | optional analytics tag on public pages (set one) |

---

## Customer UI (dashboard + checkout)

Light theme that mirrors `dalang.io/dashboard`: white surfaces, yellow `#facc15` primary,
near-black `#1c1a17` borders with hard offset shadows, Plus Jakarta Sans, a left sidebar.
The public pages (home, enterprise, trust, guide, sample report, docs, scope) use the same
light theme since 2026-10-26; only `/su` keeps its own look.

- **Scope.** Everything is under `.nsk-app` (a class on `<body>`); the component classes
  (`nsk-card`, `nsk-btn`, `nsk-badge-*`, `nsk-alert-*`, `nsk-nav-link`, `nsk-input`, …) live in
  `assets/input.css` under `@layer components`. The shared stylesheet is global, so never put
  a bare rule there that the landing page would pick up.
- **Pages.** One task per page: Home (next step, balance, latest scans), My websites, Security
  scans, Credits & payments, Automation (API tokens, for developers). Only Home and Credits call
  `syncPendingPayments` — it hits Xendit once per pending invoice and must not run on every click.
- **Results are a toast.** Every action posts to `#flash`, a fixed, dismissible, auto-clearing
  message, so nothing underneath shifts. Handlers answer a refused request with `alertHTML(...)`
  **and an error status**: the shell's `htmx:beforeSwap` swaps any error response that contains
  `nsk-alert` (HTMX drops 4xx/5xx bodies by default, which used to hide "not a hostname" and
  "cannot delete"). Other failures (401, 429, network) get a generic message from the shell.
- **Dynamic class names are purged.** `alertHTML` builds `nsk-alert-<kind>` at runtime, which
  Tailwind's content scan cannot see; `tailwind.config.js` carries a `safelist` for the
  `nsk-alert-*` / `nsk-badge-*` variants. Add new dynamic variants there.
- **Don't name a class after a severity.** `TestSeverityChipsDistinguishCleanFromUnknown`
  looks for the substrings `info`/`medium` in the chips to prove a zero row is omitted.
- **Times** are stored in UTC and rendered by `timeTag` as `<time datetime=…>`; the shell
  script localises them in the browser. SQLite writes `YYYY-MM-DD HH:MM:SS`, Go writes RFC3339 —
  `parseStoredTime` accepts both.
- **Fonts.** Plus Jakarta Sans is self-hosted (`static/fonts/`, one variable latin file), so no
  visitor IP goes to Google for a typeface. htmx is vendored the same way (`static/js/`).
- After changing any template or class, run `npm run build:css` and commit `static/css/styles.css`
  (it is embedded in the binary).

## Public site (SEO and GEO)

Written for boards, lead engineers and procurement. Read `CLAIMS.md` before touching
copy: every statement has a source, and the tests refuse the ones that do not.

- **One source per language.** All copy is Go values in `site_copy_en.go` and
  `site_copy_id.go`; one template per page (`site_html.go`). A section missing from one
  language fails the build; an empty string or a differing list length fails
  `TestSiteCopyIsComplete`.
- **Prices are tokens**, never typed: `{single}`, `{volume}`, `{volumeCredits}`,
  `{whitebox}`, `{packs}`, filled from `credit_packages`. Change a price in the database
  and the pages, the JSON-LD `Offer`s and `/llms.txt` follow.
- **The FAQ is one slice** that renders the visible `<details>` list *and* the `FAQPage`
  JSON-LD, so they cannot disagree (`TestFAQMarkupMatchesVisibleFAQ`).
- **Generated, not hand-kept:** `/sitemap.xml` (from the page registry, with hreflang),
  `/llms.txt` and `/llms-full.txt` (from the copy). A new page is one entry in `sitePages`.
- **Adding a page:** add a `sitePage`, a copy struct per language, a template body, and a
  line in `buildSiteView`. The tests then cover metadata, hreflang, headings, links,
  anchors, sitemap, banned claims and structured data automatically.
- **Every page is bilingual**, the evidence pages included (`/sample-report`, `/docs`,
  `/product-scope` and their `/id/…` twins). The sample report and the docs have copy structs
  (`sampleCopy`, `docsCopy`); the scope is one Markdown file per language, kept the same shape by
  `TestScopeMarkdownTranslationsHaveTheSameShape`. Code, evidence strings and API names are the
  same in both languages. The product itself is still English-only: the report PDF, and the
  dashboard, which the docs name by its English menu labels.
- **Security headers and CSP** (`security.go`): HSTS, nosniff, X-Frame-Options, Referrer-Policy,
  Permissions-Policy and COOP on every response. Public pages get a strict nonce CSP
  (`script-src 'nonce-…' 'strict-dynamic'`), so they must contain no inline handler and no
  `style="…"`; `security_test.go` enforces that. Everything else (dashboard, checkout, `/su`)
  still has inline handlers and gets a baseline CSP. **Because of the nonce, public pages are
  `private, no-store` with no ETag** (a 304 would pair a cached body with a new nonce).
- **Illustrations** are generated: `python3 tools/svg/iso.py` rewrites `site_svg.go`
  (`--preview <file>` writes a contact sheet). Edit the Python, not the Go. They are decoration
  (`aria-hidden`), contain no words except the digits 1-5, and each idea in them is also in text.
- **Print.** The sample report and the procurement fact sheet have print buttons; the print
  rules are at the end of `assets/input.css`. The sample is also offered as a PDF: run
  `tools/pdf/render.sh` after changing the sample copy, its template or the print rules, and
  commit the PDFs and `static/sample-report.sources`. The test fails on a stale copy.
- **The CI/CD API.** `POST /api/v1/assessments` (alias `/api/v1/pentests`) starts one, `GET
  /api/v1/assessments/{id}` reads status and finding counts for the token's own account (another
  account's id is a 404). A second start while one runs is a 409. `findings` is `null`, never
  zeros, until the report was delivered with a breakdown. The docs' CI gate script is executed
  by `TestDocsGateScriptRunsInAllFourOutcomes` (needs `sh`, `curl`, `jq`; skipped without them).
- **Spelling.** English is British, Indonesian terms are fixed (`TestSpellingIsBritish`,
  `TestIndonesianTermsAreConsistent`). The brand is "Netsekurity"; only the scanner's
  User-Agent is "NetSekurity Scanner".
- **After a content change:** bump `siteModified` in `site.go` (it is the sitemap
  `lastmod` and every `dateModified`), run the tests, and run `tools/og/render.sh` if the
  headline or sample card changed.
- **Assets never need purging.** Every static file a page loads is addressed as
  `/file?v=<12 hex of its sha256>` (`assets.go`, template func `asset`): the stylesheet, favicons,
  share cards and sample PDFs. The hash is the file's content, so a release that changes a file
  changes its URL, and one that does not leaves it (and its year-long `immutable` cache) alone.
  The web font and htmx cannot carry a query (the preload must match the `@font-face` URL), so
  they are versioned by file name and their sha256 is pinned by a test: change the bytes, rename
  the file. `TestEveryStaticAssetIsAddressedByItsContentHash` fails on any raw `/favicon…`,
  `/css/…` or `/og…` URL. Pages themselves are `no-store`, so HTML is always fresh.
- **Policies** are Dalang's own, at dalang.io/privacy, /terms-of-use and /security, linked from
  every footer and from the trust page. They do not name Netsekurity. Nothing states a data
  retention period, a DPA, an SLA or the VAT treatment for Netsekurity: `CLAIMS.md` lists these gaps.
- **The scanner's User-Agent** (`NetSekurity Scanner`) is stated on the trust and enterprise
  pages and checked against `nsec_scan.py` by a test. The agent host runs a separate checkout:
  after changing the scanner, follow `HERMES_UPDATE.md` there, or the pages describe behaviour
  that is not live.

## Local Development

```bash
npm install
npm run build:css        # build static/css/styles.css
cp env.example .env      # fill values (or rely on defaults)
go run .                 # serves on :8094 (PORT env to override)
go test ./...            # unit tests
```

> **Trap — `.env` beats real environment variables.** `env.go` says the opposite in a comment,
> but `getenv` reads the `.env` map first. If `.env` holds real secrets (Xendit, Google,
> `DATABASE_PATH`), `go run .` uses them whatever you export. To run an isolated copy for UI
> work, `go build -o /tmp/nsk-dev .` and start it from an **empty directory** with
> `PORT`, `DATABASE_PATH`, `JWT_SECRET` and `SUPER_ADMIN_EMAIL` set; the assets are embedded,
> so it needs nothing else. Mint a session with `issueJWT` (cookie `nsk_session`).

---

## Deployment (prod)

Everything is automated in **`deploy.sh`**: `go test` → `npm run build:css` → build linux binary →
commit+push (`main`) → `scp` → keep `netsekurity.prev` → `mv` → `systemctl stop` + `start` →
verify `MainPID` changed → health check. (Not `restart`: it races `Restart=always`.)

```bash
./deploy.sh                    # auto commit message
./deploy.sh "feat: my change"  # custom message
./deploy.sh --no-push          # build only
```

Manual steps the first time on the server (`root@163.128.54.5`):
1. `mkdir -p /home/dalang/netsekurity.com && chown dalang:dalang /home/dalang/netsekurity.com`
2. Copy binary + create `/home/dalang/netsekurity.com/.env`.
3. systemd unit `/etc/systemd/system/netsekurity.service` (User=dalang, WorkingDirectory=app dir,
   `Environment=PORT=8094`, `Restart=always`).
4. Pingora route in `/home/dalang/dev/proxy/config.yaml`:
   `- host: netsekurity.com / www.netsekurity.com → 127.0.0.1:8094` then `systemctl restart proxy.service`.

**Cloudflare:** flexible SSL (public HTTPS, origin HTTP :80). DNS A record → `163.128.54.5`.

---

## Operations / Handover

**Service management (on server):**
```bash
systemctl status netsekurity.service        # status
systemctl restart netsekurity.service       # restart after binary swap
journalctl -u netsekurity.service -f        # logs
systemctl status proxy.service              # Pingora
```

**Common tasks**
- Deploy: `./deploy.sh`
- Rollback: keep the previous binary (e.g. `cp netsekurity netsekurity.prev` before a swap) and
  restore it, then `systemctl restart netsekurity.service`; or redeploy the previous git commit.
- **Purge Cloudflare cache** (after report/static changes): Cloudflare dashboard → netsekurity.com →
  Caching → Purge Everything. (The API token in `api.dalang.io/.env` can read the zone but lacks
  Cache Purge permission.)
- **Credits auto-add:** poller every 30s + dashboard load; if a payment isn't credited, check
  `journalctl -u netsekurity.service` for `sync payments` lines and confirm the invoice is
  `PAID` in Xendit.
- **Verify a domain:** TXT is checked via authoritative NS, so it should be instant once the
  record is published; recursive resolvers may lag a few minutes.
- **Reports:** stored in `data/reports/`; served at `/reports/<name>.pdf` (owner/admin only,
  `no-store`). Back up `data/` regularly.

**Backups**
- The single source of truth is `data/netsekurity.db` + `data/reports/`. Copy them off-server
  regularly (or snapshot the app dir).

**Known limitations (be honest with users/AI)**
- External, unauthenticated, read-only assessment — **not** authenticated authorization,
  business-logic, or exploit-chain testing. Whitebox tier ($10k) covers deeper testing.
- Xendit bills in `XENDIT_CURRENCY` (default USD). If the account is not enabled for USD,
  invoice creation fails with `UNSUPPORTED_CURRENCY`; the customer now sees a clear message and
  the journal gets a `xendit: create invoice failed` line. Verify with one real top-up after
  any currency change.
- The standard scan sends **no injection payloads** (SQLi/XSS/LFI). Do not claim it does.
- Google One Tap requires `https://netsekurity.com` in Authorized JavaScript origins.

---

## Docs index (for handover reading order)

1. `README.md` — this document.
2. `AGENT_WORKER.md` — scanning agent behavior.
3. `static/PRODUCT_SCOPE.md` (+ `.id.md`) — client-facing scope & methodology.
4. `/docs` (`docsEN` / `docsID` in `site_copy_*.go`, examples in `docs.go`) — CI/CD integration + HTTP API reference.
5. `env.example` — all environment variables.
