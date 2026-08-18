#!/usr/bin/env bash
# Netsekurity scan worker — one-shot: claim a queued pentest, scan, upload report.
# Runs under Hermes cron (scheduled every minute). Uses BOT_AUTH_TOKEN from .env.
set -uo pipefail

APP_DIR=/home/dalang/netsekurity.com
DB="$APP_DIR/data/netsekurity.db"
REPORTS="$APP_DIR/data/reports"
API="https://netsekurity.com"
ENVFILE="$APP_DIR/.env"
TOKEN=$(sed -n 's/^BOT_AUTH_TOKEN=//p' "$ENVFILE")
WORK=/tmp/nsk_worker
LOG=/tmp/nsk_worker.log
mkdir -p "$WORK"
echo "--- nsk worker $(date -u +%FT%TZ) ---" >> "$LOG"

# 1) Claim a job
CLAIM=$(curl -s -m 20 -X POST -H "X-Bot-Token: $TOKEN" "$API/api/pentests/worker/claim")
echo "claim: $CLAIM" >> "$LOG"
if echo "$CLAIM" | grep -q '"empty":true'; then
  echo "no job" >> "$LOG"; exit 0
fi
PID=$(echo "$CLAIM" | python3 -c "import sys,json;print(json.load(sys.stdin)['pentest_id'])")
DOMAIN=$(echo "$CLAIM" | python3 -c "import sys,json;print(json.load(sys.stdin)['domain'])")
echo "job: pid=$PID domain=$DOMAIN" >> "$LOG"
[ -n "$PID" ] && [ -n "$DOMAIN" ] || { echo "bad claim" >> "$LOG"; exit 1; }

# 2) Run scan. Use the on-demand scanner (creates assessment + PDF for ANY domain).
REPORT_ROOT=/opt/data/report
VE=$REPORT_ROOT/.venv/bin/python
[ -x "$VE" ] || VE=python3
"$VE" /opt/data/netsekurity_draft/nsec_scan.py "$DOMAIN" >> "$LOG" 2>&1
# nsec_scan prints PDF=<path>; extract it
SRC=$(grep -oE '^PDF=.*' "$LOG" | tail -1 | cut -d= -f2-)
[ -n "$SRC" ] && [ -f "$SRC" ] || SRC=$(ls /opt/data/report/$DOMAIN/*Assessment*.pdf 2>/dev/null | head -1)
if [ -z "$SRC" ] || [ ! -f "$SRC" ]; then
  echo "scan produced no PDF" >> "$LOG"
  exit 0
fi
echo "pdf: $SRC" >> "$LOG"

# 3) Rename to completion-time format and upload
TS=$(date -u +%Y%m%d-%H:%M)   # completion time
OUT="$WORK/$TS-$DOMAIN.pdf"
cp "$SRC" "$OUT"
echo "upload $OUT ..." >> "$LOG"
RES=$(curl -s -m 60 -X POST -H "X-Bot-Token: $TOKEN" \
  -F "pentest_id=$PID" \
  -F "report=@$OUT;type=application/pdf;filename=$(basename "$OUT")" \
  "$API/api/pentests/worker/report")
echo "upload-resp: $RES" >> "$LOG"
# clean scratch
rm -f "$OUT" 2>/dev/null
echo "done pid=$PID" >> "$LOG"
exit 0