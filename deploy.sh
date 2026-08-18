#!/usr/bin/env bash
# netsekurity deploy script: test → build → commit+push → deploy to prod.
#
# Usage:
#   ./deploy.sh                    # auto commit message
#   ./deploy.sh "feat: my change"  # custom commit message
#   ./deploy.sh --no-push          # build only, don't commit/push/deploy
set -euo pipefail
cd "$(dirname "$0")"

SERVER=root@163.128.54.5
APP_DIR=/home/dalang/netsekurity.com
MSG="${1:-deploy: netsekurity update}"
if [ "$1" = "--no-push" ]; then MSG="skip"; fi

log() { printf '\n\033[1;33m==> %s\033[0m\n' "$*"; }

# 1. Tests
log "go test"
go test ./... -count=1

# 2. Build (CSS + linux binary)
log "build css + linux binary"
npm run build:css >/dev/null
GOOS=linux GOARCH=amd64 go build -o netsekurity-linux .

if [ "$1" = "--no-push" ]; then
  log "build only — done"
  exit 0
fi

# 3. Commit + push (rebase if partner pushed meanwhile)
log "commit + push"
git add -A
if ! git diff --cached --quiet; then
  git commit -m "$MSG"
fi
git pull --rebase origin main || { echo "rebase conflict — fix manually"; exit 1; }
git push origin main

# 4. Deploy
log "deploy to prod"
scp -o BatchMode=yes netsekurity-linux "$SERVER:/tmp/netsekurity.new"
ssh -o BatchMode=yes "$SERVER" "mv /tmp/netsekurity.new $APP_DIR/netsekurity && \
  chown dalang:dalang $APP_DIR/netsekurity && chmod 755 $APP_DIR/netsekurity && \
  systemctl restart netsekurity.service && sleep 3 && \
  echo -n 'service: ' && systemctl is-active netsekurity.service && \
  curl -s -o /dev/null -w 'landing HTTP %{http_code}\n' http://127.0.0.1:8094/"

log "deployed OK — https://netsekurity.com"
