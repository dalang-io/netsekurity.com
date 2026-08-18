# netsekurity.com — Pentest as a Service (PTaaS)

On-demand pentesting, per domain, per credit. Add a domain, verify ownership with an
auto-generated TXT record, and get a full, human-reviewed pentest report.

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
- **Implemented:** Google OAuth login, SQLite persistence, Xendit credit top-up (idempotent webhook credit), domain add + TXT verification, credits/dashboard.
- **Planned:** the actual pentest scanning engine, report delivery, and retest workflow.

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
├── go.mod
├── package.json         # Tailwind build scripts (dev only)
├── tailwind.config.js
├── assets/input.css     # Tailwind source
└── static/
    ├── index.html       # landing page (HTMX)
    └── css/styles.css   # built Tailwind output (embedded)
```

Build CSS: `npm run build:css` → `go build -o netsekurity .`

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
```

| Var | Source |
|-----|--------|
| `GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET` | `dalang.io/.env` |
| `XENDIT_SECRET_KEY` / `XENDIT_CALLBACK_KEY` | `api.dalang.io/.env` |
| `JWT_SECRET` | `api.dalang.io/.env` |

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
