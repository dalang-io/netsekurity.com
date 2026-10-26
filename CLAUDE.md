# Project: netsekurity.com

External, read-only, unauthenticated web security assessment, sold by **Dalang Pte. Ltd.** One Go
binary (stdlib `net/http`, SQLite WAL, Tailwind via `npm run build:css`, HTMX), deployed with
`./deploy.sh` to `root@163.128.54.5` (systemd `netsekurity.service`, port 8094, behind Pingora and
Cloudflare). A separate scanning agent (Hermes) runs `nsec_*` from its own checkout of this repo.

Read in this order: `HANDOVER.md` (state, traps, decisions) → `TODO.md` → `README.md` → `CLAIMS.md`.

## Rules that are not negotiable

- **Never state what `CLAIMS.md` does not list.** No number, certification, customer, duration, or
  provider without a source. Both languages move together (`site_copy_en.go`, `site_copy_id.go`);
  prices are `{tokens}` filled from `credit_packages`, never typed.
- **Public pages run a strict nonce CSP.** No inline `on…=` handler, no `style="…"`, no ETag, pages are
  `private, no-store`. The dashboard, checkout and `/su` still use the baseline policy (TODO T-09).
- **Static assets are `?v=<content hash>`** through the `asset` template func (`assets.go`). Do not write a
  raw `/favicon…`, `/css/…` or `/og…` URL (a test fails). The font and htmx are versioned by file name and
  pinned by sha256.
- **Generated files are not edited by hand:** `site_svg.go` (`python3 tools/svg/iso.py`),
  `static/css/styles.css` (`npm run build:css`), `static/og*.png` (`tools/og/render.sh`),
  `static/netsekurity-sample-report-*.pdf` + `static/sample-report.sources` (`tools/pdf/render.sh`).
- **`.env` in the repo is real and beats environment variables.** Run local servers from an empty directory
  with a throwaway `DATABASE_PATH` and `JWT_SECRET` (see `HANDOVER.md` §4). Never press Accept on the cookie
  banner against the real Meta pixel.
- **No hot patches on hosts.** Everything, the agent's scripts included, goes through `main`.
- Do not name the AI model provider anywhere until the owner supplies it (TODO T-03).

## Commands

```bash
go build ./... && go vet ./... && go test ./... -count=1   # the truth; ignore stale editor diagnostics
npm run build:css                                           # after any template or class change
./deploy.sh "msg"                                           # test, build, commit+push main, deploy, PID check
git tag -a vYYYY.MM.DD.N -m "..." && git push origin vYYYY.MM.DD.N   # after a deploy only
```

Tests: `site_test.go` (metadata, hreflang, claims, links, sitemap), `site_quality_test.go` (spelling, terms,
icons, mobile tables, bilingual evidence pages, CI gate script, sample PDFs), `security_test.go` (headers, CSP,
assets), `api_v1_test.go` (CI/CD API). A change to a claim updates the source, `CLAIMS.md`, both languages and
the tests in the same commit.

## Workspace

This repo is one of several siblings (see `../CLAUDE.md`). The frontend of `dalang.io` and `api.dalang.io` are
separate; Netsekurity talks to `api.dalang.io` only through the partner API (`partner.go`, HMAC).
