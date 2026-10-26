#!/usr/bin/env bash
# Renders the printable copies of the sample report (static/netsekurity-sample-report-{en,id}.pdf)
# from the live pages with headless Chrome, using the print stylesheet in assets/input.css.
#
# Run it after changing the sample report copy (sampleEN/sampleID in site_copy_*.go), its
# template (siteSampleHTML) or the print rules, then commit the PDFs and the .sources file.
# TestSamplePDFsMatchTheirSource fails until you do. The PDFs carry no price: the buttons
# and the call to action are hidden in print, so they cannot go stale when a price changes.
set -euo pipefail
cd "$(dirname "$0")/../.."
CHROME="${CHROME:-/Applications/Google Chrome.app/Contents/MacOS/Google Chrome}"
PORT="${PORT:-8197}"
TMP="$(mktemp -d)"
trap 'kill "${SERVER_PID:-0}" 2>/dev/null || true; rm -rf "$TMP"' EXIT

npm run build:css >/dev/null
go build -o "$TMP/nsk" .
# An empty working directory, so a developer's .env (which wins over the environment) is not read.
( cd "$TMP" && DATABASE_PATH="$TMP/render.db" JWT_SECRET=render-only PORT="$PORT" SERVER_PORT="$PORT" "$TMP/nsk" >"$TMP/server.log" 2>&1 ) &
SERVER_PID=$!
for _ in $(seq 1 40); do curl -fsS -o /dev/null "http://localhost:$PORT/" 2>/dev/null && break; sleep 0.25; done

for pair in "en:/sample-report" "id:/id/contoh-laporan"; do
  lang="${pair%%:*}"; path="${pair##*:}"
  "$CHROME" --headless=new --disable-gpu --no-pdf-header-footer --virtual-time-budget=8000 \
    --print-to-pdf="$PWD/static/netsekurity-sample-report-$lang.pdf" "http://localhost:$PORT$path" >/dev/null 2>&1
  echo "static/netsekurity-sample-report-$lang.pdf"
done
WRITE_SAMPLE_SOURCES=1 go test -run TestSamplePDFsMatchTheirSource -count=1 . >/dev/null
echo "static/sample-report.sources"
