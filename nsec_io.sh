#!/usr/bin/env bash
# Netsekurity worker: claim + upload helpers (runs in Hermes container).
# The actual scan (recon + report_auto.py) is done by the Hermes agent between
# claim and upload. This script only handles the platform I/O via SSH + curl.
#
# Usage:
#   nsec_claim.sh                          -> prints JSON from /worker/claim
#   nsec_upload.sh <pid> <pdf> [findings]  -> uploads completed PDF
#
# findings is the optional severity breakdown, in the same format nsec_scan.py
# prints on its SEVERITY= line:
#   critical=0,high=2,medium=7,low=3,info=9
# Omit it and the dashboard says the breakdown is in the report, rather than
# implying the scan came back clean.
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
    PID="${2:-}"; PDF="${3:-}"; FINDINGS="${4:-}"
    [ -n "$PID" ] && [ -f "$PDF" ] || { echo '{"error":"bad args"}'; exit 1; }
    NAME=$(basename "$PDF")
    curl -s -m 90 -X POST -H "X-Bot-Token: $TOKEN" \
      -F "pentest_id=$PID" \
      -F "findings=$FINDINGS" \
      -F "report=@$PDF;type=application/pdf;filename=$NAME" \
      "$API/api/pentests/worker/report"
    ;;
  *) echo 'usage: nsec_io.sh {claim|upload <pid> <pdf> [findings]}' ;;
esac
