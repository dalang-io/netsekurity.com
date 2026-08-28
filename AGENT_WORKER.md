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
  5. exposed-asset checks (`.git`, `.env`, backups, admin/login, server-status,
     phpinfo),
  6. open-redirect discovery, and with `--deep-scan` runs **nikto**,
  7. stores findings into `assessment.db`, then renders the PDF report.
- Findings are CVSS-scored and CWE-mapped (OWASP-family).

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

```bash
cd <hermes checkout of netsekurity.com> && git pull
grep -c 'SEVERITY=' nsec_scan.py     # 1 = the scanner emits the breakdown
grep -c 'findings=' nsec_worker.sh   # 1 = the worker forwards it
```

Then watch one real scan: `grep -E 'scanner:|severity:|upload-resp:' /tmp/nsk_worker.log`.

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
| `/reports/{name}.pdf` | GET | session (owner/admin) | report download |

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
