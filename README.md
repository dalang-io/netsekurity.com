# netsekurity.com — Pentest as a Service (PTaaS)

On-demand pentesting, per domain, per credit. Add a domain, verify ownership with an
auto-generated TXT record, and get a full, agent-reviewed pentest report.

**Pricing (credits):** $50 = 1 credit · $100 = 3 credits · $500 = 20 credits · $1000 = 50 credits.
**1 credit = 1 pentest · 1 domain.**

---

## Tech Stack

| Layer | Choice |
|-------|--------|
| Language | Go (standard library only — `net/http`) |
| UI | HTMX (server-rendered fragments, no JS framework) |
| Styling | Tailwind CSS (built once into `static/css/styles.css`) |
| Database | **SQLite** |
| Auth | **Google OAuth** (login with Google account) |
| Payments | **Xendit** (invoice-based; paid → credit credited) |
| Hosting | Linux server, systemd unit + Pingora proxy route |

---

## Current Status

- **Live (prod):** marketing landing page + **authenticated dashboard** — `http://netsekurity.com` (via Pingora on `:80`, origin port `8094`, systemd `netsekurity.service`).
- **Implemented:** Google OAuth login, SQLite persistence, Xendit credit top-up (idempotent webhook credit), domain add + TXT verification, credits/dashboard, **and the full on-demand pentest pipeline driven by the scanning agent** (queue → scan → report PDF → download).
- **Scanning agent (bot):** a Hermes agent working the `pentests` queue. When a user clicks **Scan**, 1 credit is consumed and a `queued` job is created; the agent claims it, runs an **enterprise-grade scan**, uploads the report PDF (`YYYYMMDD-hh:mm-<domain>.pdf`), and marks it `completed` for the user to download.

> See [`AGENT_WORKER.md`](AGENT_WORKER.md) for the authoritative behavior reference of the scanning agent.

---

## Architecture

A single Go binary (`netsekurity`) that:

1. **Serves the landing page + static assets** (HTML, CSS) embedded into the binary via `go:embed`.
2. **Serves HTMX fragments** (e.g. `/api/txt`, `/api/verify`, `/api/faq`, `/contact`) — HTML responses swapped into the page by HTMX.
3. (Planned) **REST/HTMX endpoints** for auth, credits, domain verification, and payment callbacks, backed by SQLite.

The app is stateless except for the SQLite file; sessions are JWT (issued after Google OAuth), matching `dalang.io`/`api.dalang.io` conventions.

```
Browser ──HTMX──▶ netsekurity (Go :8094)
                    ├── /             → index.html (embedded)
                    ├── /css/*        → Tailwind CSS
                    ├── /api/*        → HTMX fragments / JSON
                    └── SQLite (data/netsekurity.db)
        ▲
        │
Pingora proxy (:80) ── Host: netsekurity.com → 127.0.0.1:8094
```

---

## Repository Layout

```
netsekurity.com/
├── main.go              # Go stdlib server + HTMX handlers
├── pentests.go          # pentest: start / list / worker claim+report / report download
├── dashboard.go         # authenticated dashboard (HTMX)
├── admin.go             # super-admin panel + report serving
├── payment.go           # Xendit credit top-up + webhook
├── domains.go           # domain add / TXT verify / delete
├── go.mod
├── package.json         # Tailwind build scripts (dev only)
├── tailwind.config.js
├── assets/input.css     # Tailwind source
├── assets/netsekurity_logo.png  # report PDF branding logo
├── AGENT_WORKER.md      # scanning-agent behavior reference
├── nsec_scan.py         # enterprise scanner (agent)
├── nsec_io.sh           # claim/upload I/O helpers (agent)
├── nsec_monitor.sh      # token-saver monitor (agent cron)
├── nsec_worker.sh       # standalone one-shot worker (agent)
├── deploy.sh            # test → build → commit+push → deploy to prod
├── env.example          # environment template (placeholders)
└── static/
    ├── index.html       # landing page (HTMX)
    └── css/styles.css   # built Tailwind output (embedded)
```

Build CSS: `npm run build:css` → `go build -o netsekurity .` — or just `./deploy.sh`.

---

## Database (SQLite)

Planned schema (created via SQL migrations, SQLite single file). Migrations run on startup
through a `schema_migrations` table — same approach as `api.dalang.io`.

### Tables

| Table | Purpose | Key columns |
|-------|---------|-------------|
| `users` | Accounts from Google OAuth | `id`, `email`, `google_sub`, `name`, `created_at` |
| `credit_balance` | Current spendable credit per user | `user_id`, `balance` (fractional credits), `updated_at` |
| `credit_transactions` | Credit ledger (topup / spend) | `user_id`, `type` (`topup`/`spend`), `amount`, `reference_id`, `created_at` |
| `credit_packages` | Configurable credit bundles | `id`, `name`, `usd_price`, `credits`, `is_active` |
| `domains` | Domains pending/verified for pentest | `id`, `user_id`, `domain`, `txt_verification_token`, `status` (`pending`/`verified`/`failed`), `verified_at` |
| `pentests` | One job per credit per domain | `id`, `user_id`, `domain_id`, `status`, `report_ref`, `started_at`, `completed_at` |
| `payments` | Xendit invoice ↔ credits | `id`, `user_id`, `external_id`, `xendit_invoice_id`, `status` (`pending`/`paid`), `amount_usd`, `credits`, `paid_at` |

### Credit rules

- `1 credit = 1 pentest · 1 domain`.
- Buying a package debits nothing up front; on Xendit **paid** callback, the credits are
  credited to `credit_balance` and a `credit_transactions` `topup` row is written
  (dedup by `reference_id`/`external_id` for idempotency).
- Spending a credit on a pentest writes a `spend` row and decrements `balance` atomically.

---

## Authentication — Google OAuth

Login uses **Google OAuth 2.0** (Authorization Code flow), identical to `dalang.io`.

1. User clicks **Sign in with Google** → redirect to Google with `GOOGLE_CLIENT_ID` + `GOOGLE_REDIRECT_URL`.
2. Google redirects back with an authorization code.
3. Server exchanges the code for a token, fetches the Google profile (`sub`, `email`, `name`).
4. Upserts the `users` row by `google_sub`/`email` and issues a signed **JWT** (`JWT_SECRET`) stored as an `httpOnly` cookie.
5. All protected endpoints verify the JWT.

Required env: `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URL`, `JWT_SECRET`, `ALLOWED_ORIGINS`.

---

## Payments — Xendit

Credits are purchased with **Xendit invoices** (`POST https://api.xendit.co/v2/invoices`).

1. User selects a `credit_package` → server creates an Xendit invoice (`amount`, `external_id`, `success_redirect_url`).
2. Xendit webhook (`POST /webhook/xendit`) with callback token `XENDIT_CALLBACK_KEY` reports status:
   - `PAID`/`SETTLED` → mark `payments.status='paid'`, credit `credit_transactions` + `credit_balance` (idempotent on `external_id`).
   - `EXPIRED`/`FAILED` → leave pending, no credit.
3. User sees credits reflected immediately on the dashboard.

Required env: `XENDIT_SECRET_KEY`, `XENDIT_CALLBACK_KEY`, `XENDIT_SUCCESS_URL`, `XENDIT_FAILURE_URL`.

> The same key convention is already used in `api.dalang.io` (`XENDIT_SECRET_KEY`).

---

## Domain Verification (TXT)

Verification is required before a pentest can run — we only test domains the user owns.

1. User adds a domain → server generates a unique token (`ns-verify-<hex>`) and stores it in `domains.txt_verification_token`.
2. The dashboard shows the record to add:
   - **Name:** `_netsekurity` · **Type:** TXT · **Value:** `<token>`
3. User adds the TXT to their DNS and clicks **Verify**.
4. Server resolves `TXT _netsekurity.<domain>` and compares to the stored token. On match → `domains.status='verified'` and the domain becomes eligible for a pentest.

> A live demo of this flow already exists on the landing page (`/api/txt` → generate, `/api/verify` → verified).

---

## Agent Integration (automated pentest)

The platform is driven end-to-end by a **scanning agent** (a Hermes agent). The web app only
owns the queue + credits + storage; the agent owns the scan + report.

### Flow

```
User clicks "scan"  →  /api/pentests/start   →  1 credit consumed, pentests(status='queued')
Agent (every minute) →  nsec_monitor.sh      →  sees "JOB <pid> <domain>" (queue non-empty)
                       →  /api/pentests/worker/claim  →  status='running'
                       →  nsec_scan.py <domain>       →  enterprise scan → PDF
                       →  /api/pentests/worker/report →  status='completed', report saved
User dashboard       →  /reports/YYYYMMDD-hh:mm-<domain>.pdf  (owner/admin only)
```

### Worker endpoints (guarded by `X-Bot-Token`, from `BOT_AUTH_TOKEN`)

| Endpoint | Method | Purpose |
|---|---|---|
| `/api/pentests/start` | POST | user triggers a scan (auth) |
| `/api/pentests/list` | GET | user's pentests + download links (auth) |
| `/api/pentests/worker/claim` | POST | agent claims next queued job |
| `/api/pentests/worker/report` | POST | agent uploads completed PDF |
| `/reports/{name}.pdf` | GET | report download (owner/admin) |

### Files (in repo)

- `pentests.go` — Go handlers for the endpoints above.
- `nsec_scan.py` — deterministic enterprise scanner (recon, fingerprint, headers,
  TLS/PKI, exposed assets, nikto/open-redirect on `--deep-scan`; CVSS+CWE grounded).
- `nsec_io.sh` — claim/upload I/O helpers (token read from `.env` via SSH, never logged).
- `nsec_monitor.sh` — token-saver monitor; suppresses the agent LLM when the queue is idle
  (no agent tokens burned when nobody scans).
- `nsec_worker.sh` — standalone one-shot worker (alternative to the cron).
- `AGENT_WORKER.md` — full behavior reference.

### Running the agent

The agent is scheduled as a Hermes cron job (every minute). The monitor runs first and only
wakes the LLM when a job is actually queued, so idle minutes cost nothing.

---

## Environment Variables (`.env`)

The app reads a `.env` file in its working directory. **Credentials can be reused from the
existing `dalang.io` / `api.dalang.io` `.env`** (they share the same Google OAuth app and
Xendit account).

```env
# Server
PORT=8094
ENVIRONMENT=production
ALLOWED_ORIGINS=https://netsekurity.com

# Database
DATABASE_PATH=./data/netsekurity.db

# Auth (Google OAuth + JWT) — reuse from dalang.io/.env or api.dalang.io/.env
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
GOOGLE_REDIRECT_URL=https://netsekurity.com/auth/google/callback
JWT_SECRET=
JWT_EXPIRY=24h

# Payments (Xendit) — reuse from api.dalang.io/.env
XENDIT_SECRET_KEY=
XENDIT_CALLBACK_KEY=
XENDIT_SUCCESS_URL=https://netsekurity.com/dashboard
XENDIT_FAILURE_URL=https://netsekurity.com/pricing

# Scan Worker — shared secret for agent→platform auth (X-Bot-Token)
BOT_AUTH_TOKEN=
```

| Var | Source |
|-----|--------|
| `GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET` | `dalang.io/.env` |
| `XENDIT_SECRET_KEY` / `XENDIT_CALLBACK_KEY` | `api.dalang.io/.env` |
| `JWT_SECRET` | `api.dalang.io/.env` |
| `BOT_AUTH_TOKEN` | generate fresh — `openssl rand -hex 32` |

> A full template with every variable is in [`env.example`](env.example).

---

## Local Development

```bash
npm install            # Tailwind (dev only)
npm run build:css      # build static/css/styles.css
go run .               # serves on :8090 (override with PORT)
# or
PORT=18090 go run .
```

---

## Deployment (prod)

```bash
# Build linux binary (embedding Tailwind CSS + HTML)
GOOS=linux GOARCH=amd64 go build -o netsekurity-linux .

# On server 163.128.54.5
scp netsekurity-linux root@163.128.54.5:/home/dalang/netsekurity.com/netsekurity

# systemd
systemctl daemon-reload
systemctl enable --now netsekurity.service   # listens on :8094

# Pingora route in /home/dalang/dev/proxy/config.yaml
#   - host: "netsekurity.com"   → 127.0.0.1:8094
#   - host: "www.netsekurity.com" → 127.0.0.1:8094
systemctl restart proxy.service
```

`netsekurity.service` runs as user `dalang`, `WorkingDirectory=/home/dalang/netsekurity.com`,
`Environment=PORT=8094`, with `Restart=always`.
