# netsekurity.com — External Web Security Assessment

> **Positioning (honest):** "You build fast with AI. We check your production attack surface —
> with agent-reviewed findings." Netsekurity is a **continuous external web security assessment**
> (automated, blackbox, unauthenticated, read-only) with human/agent review, **not a full
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

---

## Architecture

A single Go binary serves everything (landing, dashboard, admin, API, static) via embedded
assets (`go:embed`). It is stateless except for the SQLite file; sessions are JWT.

```
Browser (HTTPS) → Cloudflare (flexible) → Pingora :80
   Host: netsekurity.com → 127.0.0.1:8094 → netsekurity (Go)
     ├── /                    landing page (HTML, injects client id + CSS hash)
     ├── /css/*, /og.png, /robots.txt, /sitemap.xml, /llms*.txt, /security.txt
     ├── /auth/*              Google OAuth + One Tap
     ├── /dashboard, /su      authenticated pages (HTMX)
     ├── /api/*               top-up, domains/verify, pentests, api-tokens
     ├── /reports/*           pentest PDFs (owner/admin only, no-store)
     └── SQLite (data/netsekurity.db) + data/reports/
```

---

## Repository Layout

```
netsekurity.com/
├── main.go            # server bootstrap, routes, handleIndex (injects client id / css hash)
├── auth.go            # Google OAuth + One Tap, JWT session, requireAuth / requireAdmin / isAdmin
├── jwt.go             # HS256 JWT (stdlib)
├── db.go              # SQLite init, schema, seed packages, super-admin seed
├── env.go             # .env loader
├── util.go            # parseFloat, secureCompare
├── dashboard.go       # authenticated dashboard (HTML template + cssHash func)
├── admin.go           # super-admin panel (/su) + report serving + add-credit
├── payment.go         # Xendit top-up, webhook, auto-credit poller (syncPendingPayments)
├── domains.go         # domain add / TXT verify (authoritative NS) / delete
├── pentests.go        # pentest start/list + worker claim/report (queue)
├── apitokens.go       # CI/CD API tokens (expiry 7/14/30/60/90d)
├── marketing.go       # landing HTMX fragments (contact, TXT demo, FAQ)
├── stack.go           # "cat stack.conf" section (real SVG logos + CRT glow tiles)
├── header.go          # shared reusable header (brand + desktop nav + mobile burger)
├── docs.go            # /docs page (CI/CD + API reference)
├── *_test.go          # unit tests (JWT, credit idempotency, topup amount, domain, index, auth)
├── deploy.sh          # test → build → commit+push → deploy (scp + systemd restart)
├── nsec_scan.py       # scanning agent scanner (enterprise-grade, produces PDF)
├── nsec_io.sh         # agent claim/upload I/O helpers
├── nsec_monitor.sh    # agent token-saver monitor (Hermes cron)
├── nsec_worker.sh     # standalone one-shot worker
├── AGENT_WORKER.md    # agent behavior reference (authoritative)
├── HERMES_UPDATE.md   # Hermes agent: update steps + verification runbook
├── marketing-copy.md  # social/ads copy (vibe-coder + CI/CD use cases)
├── env.example        # environment template
├── README.md          # this doc
└── static/
    ├── index.html                 # landing (HTMX, One Tap, JSON-LD)
    ├── css/styles.css             # built Tailwind output (embedded)
    ├── og.png                     # OG/Twitter social image
    ├── sitemap.xml, robots.txt    # SEO/GEO
    ├── llms.txt, llms-full.txt    # LLM/GEO discovery
    ├── security.txt               # + .well-known/security.txt (RFC 9116)
    └── PRODUCT_SCOPE.md           # client-facing scope & methodology
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

### Credit rules
- `1 credit = 1 external assessment · 1 domain`.
- Top-up credits are added by the **auto-credit poller** (`syncPendingPayments`, every 30s + on dashboard load) — the Xendit **webhook is optional** (kept as a fallback).
- Spending a credit writes a `spend` row and decrements `balance` atomically.

---

## Authentication & Roles

- **Google OAuth 2.0** (Authorization Code + **One Tap** suggest-login). Redirect URI:
  `https://netsekurity.com/auth/google/callback` (must be registered in Google Console;
  `https://netsekurity.com` must be in Authorized JavaScript origins for One Tap).
- **JWT session** (HS256, `JWT_SECRET`) in an `httpOnly`, `Secure` cookie (`nsk_session`).
- **Super admin** = `SUPER_ADMIN_EMAIL` (default `hans@dalang.io`) OR `users.role = 'admin'`.
  - `/su` panel: list users, add user, set role, **add credit**, list transactions
    (paid/expired), customer domains + Pentest button, pentests + PDF upload/list.
  - Protected pages set `Cache-Control: no-store`; reports are owner/admin-only (404 otherwise).

---

## Payments — Xendit

- **Top-up:** `POST /api/topup` → creates a Xendit invoice for the selected package.
  Currency = `XENDIT_CURRENCY` (default **IDR**; USD unsupported on this account yet — charging
  USD returns `UNSUPPORTED_CURRENCY`). IDR amount = `usd_price × XENDIT_USD_RATE`.
- **Auto-credit:** the poller checks pending payments against Xendit and credits when
  `PAID`/`SETTLED` (idempotent). Runs every 30s + on dashboard load. Webhook `/webhook/xendit`
  is optional (token-gated by `XENDIT_CALLBACK_KEY`).

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
- API tokens for CI/CD auto-scan: `apitokens.go` (7/14/30/60/90-day expiry).

---

## Marketing & GEO assets

- `sitemap.xml`, `robots.txt` (allows AI crawlers), `llms.txt` + `llms-full.txt` (LLM/GEO),
  `security.txt` (+ `/.well-known/security.txt`), `og.png`, JSON-LD on landing.
- Client-facing scope/methodology: `static/PRODUCT_SCOPE.md`.
- Social/ads copy: `marketing-copy.md` (Vibe Coder + CI/CD use cases).

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
| `XENDIT_CURRENCY` | `IDR` (default) or `USD` (if enabled) |
| `XENDIT_USD_RATE` | IDR-per-USD for IDR charges |
| `BOT_AUTH_TOKEN` | agent↔platform secret (`openssl rand -hex 32`) |

---

## Local Development

```bash
npm install
npm run build:css        # build static/css/styles.css
cp env.example .env      # fill values (or rely on defaults)
go run .                 # serves on :8094 (PORT env to override)
go test ./...            # unit tests
```

---

## Deployment (prod)

Everything is automated in **`deploy.sh`**: `go test` → `npm run build:css` → build linux binary →
commit+push (`main`) → `scp` → `mv` → `systemctl restart netsekurity.service` → health check.

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
- Xendit charges in **IDR** (USD not yet enabled on this account).
- Google One Tap requires `https://netsekurity.com` in Authorized JavaScript origins.

---

## Docs index (for handover reading order)

1. `README.md` — this document.
2. `AGENT_WORKER.md` — scanning agent behavior.
3. `static/PRODUCT_SCOPE.md` — client-facing scope & methodology.
4. `docs.go` → `/docs` — CI/CD integration + HTTP API reference.
5. `env.example` — all environment variables.
