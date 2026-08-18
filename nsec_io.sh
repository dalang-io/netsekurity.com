#!/usr/bin/env bash
# Netsekurity worker: claim + upload helpers (runs in Hermes container).
# The actual scan (recon + report_auto.py) is done by the Hermes agent between
# claim and upload. This script only handles the platform I/O via SSH + curl.
#
# Usage:
#   nsec_claim.sh            -> prints JSON from /worker/claim
#   nsec_upload.sh <pid> <pdf>  -> uploads completed PDF
set -uo pipefail
H=163.128.54.5
KEY=/opt/data/recon/keys/dalang_key
API="https://netsekurity.com"
TOKEN=$(ssh -i "$KEY" -o StrictHostKeyChecking=no root@$H "sed -n 's/^BOT_AUTH_TOKEN=//p' /home/dalang/netsekurity.com/.env" 2>/dev/null | tr -d '\r\n ')
case "${1:-}" in
  claim)
    curl -s -m 20 -X POST -H "X-Bot-Token: $TOKEN" "$API/api/pentests/worker/claim"
    ;;
  upload)
    PID="${2:-}"; PDF="${3:-}"
    [ -n "$PID" ] && [ -f "$PDF" ] || { echo '{"error":"bad args"}'; exit 1; }
    NAME=$(basename "$PDF")
    curl -s -m 90 -X POST -H "X-Bot-Token: $TOKEN" \
      -F "pentest_id=$PID" \
      -F "report=@$PDF;type=application/pdf;filename=$NAME" \
      "$API/api/pentests/worker/report"
    ;;
  *) echo 'usage: nsec_claim.sh {claim|upload <pid> <pdf>}' ;;
esac
