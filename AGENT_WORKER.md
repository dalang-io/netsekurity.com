# Netsekurity Agent Worker — Behavior Reference

Authoritative record of how the **scanning agent** (the Hermes "Netsekurity scan worker")
behaves when a user commissions a pentest. Kept in-repo so the platform's scanning
behavior is reproducible and auditable.

## Pipeline (end to end)

```
User (web)                          Platform (Go)              Agent (Hermes)
─────────                          ─────────────             ──────────────
add domain
  → TXT token (ns-verify-xxx)
verify (owner NS check)  ──────────> status='verified'
click "scan" (dashboard) ──────────> /api/pentests/start:
                                        • verified? yes
                                        • in-flight cap (1/user)
                                        • consume 1 credit (atomic)
                                        • insert pentests(status='queued')
then every minute ──────────────────> worker cron polls monitor:
                                        monitor: queued? → "JOB pid domain"
                                                           → scan + upload →
                                        /api/pentests/worker/claim → status='running'
                                        /api/pentests/worker/report → status='completed'
                                        report file: reports/<YYYYMMDD-hh:mm>|<domain>.pdf
User dashboard ────────────────────> /reports/<file>.pdf  (owner/admin only)
```

## Worker cron behavior

- **Trigger:** a Hermes **cron job** (`netsekurity-scan-worker`) running every minute.
- **Monitor (token saver):** `nsec_monitor.sh` runs first and only emits a changing
  token when a pentest is actually `queued`. When idle it emits a stable `IDLE`,
  which suppresses the LLM entirely (no AI tokens burned when nobody scans).
  When a `JOB <pid> <domain>` appears, the agent wakes to process it.
- **Claim/manual:** after waking, the agent confirms via `nsec_io.sh claim`.

## Scanner

- **`nsec_scan.py`** is the deterministic scanner: it
  1. ensures an assessment row exists for the domain,
  2. runs recon (DNS, subdomain enum via assetfinder+crt.sh, and with
     `--deep-scan`: nmap port scan + whatweb fingerprint),
  3. HTTP fingerprint + full OWASP security-header check + clickjacking + server
     version disclosure + WAF detect,
  4. TLS/PKI checks (expiry, issuer, SAN),
  5. exposed-asset checks (13 paths: `.git`, `.env`, backups, admin/login,
     server-status, phpinfo, SVN, robots.txt),
  6. open-redirect discovery (9 parameters), and with `--deep-scan` runs **nikto**,
  7. stores findings into `assessment.db`, then renders the PDF report.
- Findings are CVSS-scored and CWE-mapped (OWASP-family).

  The cron worker runs it **without** `--deep-scan`, so nmap, whatweb and nikto
  never run for customers. The public copy (landing, FAQ, `llms.txt`,
  `PRODUCT_SCOPE.md`) describes exactly the non-deep set — 38 checks — and says
  plainly that no injection payloads (SQLi/XSS/LFI) are sent. Do not widen the
  copy without widening the worker. A `--mode destructive` flag used to exist;
  it changed nothing in the scan and was removed together with the 2-credit
  "destructive" option in the platform.


### Severity breakdown (`findings`)

`nsec_scan.py` prints a machine-readable summary next to `PDF=` and
`TOTAL_FINDINGS=`:

```
SEVERITY=critical=0,high=2,medium=7,low=3,info=9
```

Both upload paths forward it to `/api/pentests/worker/report` as the optional
form field `findings`: `nsec_worker.sh` (the cron path) reads it from the scan
log, and `nsec_io.sh upload <pid> <pdf> [findings]` takes it as a third
argument. The platform stores it on the `pentests` row
(`findings_reported` plus one counter per level) and renders it as chips on the
customer's dashboard card and in `/su`.

The field is **optional and additive**: a worker that omits it leaves
`findings_reported = 0`, and the dashboard then says *"severity breakdown is in
the report"* rather than implying the scan came back clean. Unknown keys and
malformed values are ignored; if nothing parses, the summary counts as absent.
Admins can supply the same string by hand when uploading a report from `/su`.

### Updating the agent

`nsec_worker.sh` resolves `nsec_scan.py` next to itself, so updating the agent is
a `git pull` in its checkout of this repo — no copying into
`/opt/data/netsekurity_draft/` (that path stays as a fallback for older installs).

Full steps, verification and troubleshooting: **[HERMES_UPDATE.md](HERMES_UPDATE.md)**.

## Report

- **Language:** English only.
- **Filename:** `YYYYMMDD-hh:mm-<domain>.pdf` (completion-time timestamp;
  start time stored in `pentests.started_at`).
- **Branding:** "Netsekurity.com Security Agent" (cover, header, document owner,
  auditor, revision, disclaimer). Logo: `assets/netsekurity_logo.png`.
- **Generation:** SQLite(`assessment.db`) → `report_auto.py` → LaTeX → Tectonic.
- Structure: Document Control, Findings Summary, Risk Register, Methodology,
  Findings & Analysis, Remediation Recommendations, Approval & Sign-off,
  Remediation Plan (30/60/90), Compliance Mapping, Appendix A – Evidence.

## Platform endpoints (new in this repo)

| Endpoint | Method | Auth | Purpose |
|---|---|---|---|
| `/api/pentests/start` | POST | session (user) | trigger scan; verify + cap + credit |
| `/api/pentests/list` | GET | session (user) | HTMX list of user's pentests |
| `/api/pentests/worker/claim` | POST | `X-Bot-Token` | agent claims next queued job |
| `/api/pentests/worker/report` | POST | `X-Bot-Token` | agent uploads completed PDF (+ optional `findings`) |
| `/api/pentests/worker/fail` | POST | `X-Bot-Token` | agent closes a scan it cannot finish; refunds credits only for self-serve scans |
| `/api/partner/v1/scans/reserve` | POST | partner HMAC | platform reserves a scan (not queued, no credit) |
| `/api/partner/v1/scans/{id}/confirm` | POST | partner HMAC | platform confirms payment; scan enters the queue |
| `/api/partner/v1/scans/{id}/release` | POST | partner HMAC | platform abandons an unpaid reservation |
| `/api/partner/v1/scans/{id}` | GET | partner HMAC | status, attempts, deadline, findings, signed report URL |
| `/api/partner/v1/scans/{id}/report?exp=&sig=` | GET | signed URL | PDF download without a session |

## Dalang platform scans

Dalang platform scans are created by `api.dalang.io` through `partner.go`. They never
consume Netsekurity credits and never write a `credit_transactions` row.

- **States:** `reserved` (not claimable) → `queued` (after `confirm`) →
  `running` → `completed` / `failed`; `reserved` → `cancelled` on `release` or
  reservation expiry.
- **Billing:** `pentests.billing_source='partner'` and `origin='dalang'`. The
  worker refuses to refund credits for these rows; a failed scan is compensated
  on the Dalang wallet by the platform.
- **Deadline:** `deadline_at` is the reservation TTL while `reserved`, then
  `queued_at + 6h`. `sweepPartnerScans()` (called by the claim handler and the
  payment poller) fails queued/running rows past the deadline and cancels
  expired reservations.
- **Attempts:** `attempts` increments on every claim. A platform-created scan that has
  been claimed 3 times is marked `failed` instead of being requeued, so a dying
  worker cannot spin the order forever.
- **Ownership proof:** `proof_type=txt` requires an existing verified domain;
  `proof_type=cloudflare` records a Cloudflare-activated hostname, and
  `proof_type=registrar` an assertion from the hosting panel that the domain was
  sold/verified at the registrar (`domains.verified_by='registrar'`). Platform-owned hostnames
  (`*.dalang.io`, `*.netsekurity.com`) are refused.
- **Admin view:** `/su` labels Dalang-created rows `dalang · <external id>` (users,
  domains, and pentests). The synthetic `…@partner.invalid` address is an
  internal placeholder and never reaches the console.
- **Auth:** HMAC-SHA256 over `partner_id \n timestamp \n nonce \n body` with
  `PARTNER_SECRET`, a ±5-minute window and a single-use nonce ledger
  (`partner_nonces`). A request is replayed only via its idempotency key
  (`partner_requests`), which returns the original response.

## Security controls

- `BOT_AUTH_TOKEN` (`.env`) guards the worker intake endpoints (constant-time compare).
- Credit consume is atomic (`UPDATE ... WHERE balance >= 1`) — no overdraft.
- In-flight cap: 1 queued/running pentest per user.
- Report filenames strictly validated (`^[0-9]{8}-[0-9]{2}:[0-9]{2}-.+\.pdf$`), no
  path traversal, PDF-only.
- `/reports/` serves only to the owning user or an admin.

## Files

- `pentests.go` — the Go handlers above.
- `nsec_io.sh` — claim/upload I/O helpers (token read from `.env` via SSH, never logged).
- `nsec_scan.py` — deterministic enterprise scanner.
- `nsec_worker.sh` — standalone one-shot worker (alternate to the cron).
- `nsec_monitor.sh` — token-saver monitor (suppresses idle LLM runs).
- `assets/netsekurity_logo.png` — report branding logo.

---

## Closing a scan the worker cannot finish: `POST /api/pentests/worker/fail`

A claimed pentest that cannot be completed (target unreachable, report refused by the
generator's freshness gate, scan error) must be closed explicitly. Leaving it in
`running` only lets the 45-minute stale recovery push it back to `queued`, which turns
into an endless claim/fail loop with no outcome for the customer and no refund for the
credit already spent.

```
POST /api/pentests/worker/fail        (header: X-Bot-Token)
  pentest_id = pt_xxxxxxxx
  reason     = short human-readable cause
-> {"ok":true,"pentest_id":"pt_xxx","status":"failed","refunded":1}
```

- Only a pentest in `running` may fail; anything else returns `409`.
- For self-serve scans the credit is refunded automatically: 1 for a standard scan, 2
  for a destructive one, recorded as an `admin_credit` transaction (the type the
  dashboard lists) with the pentest id as reference. Dalang platform scans
  (`billing_source='partner'`) are paid on the Dalang wallet and get `refunded:0` —
  the platform issues that refund.
- The dashboard shows a red `failed` badge plus *"scan failed — credit refunded"*, and
  never offers a report download for a failed scan.
