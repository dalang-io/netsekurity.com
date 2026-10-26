#!/usr/bin/env bash
# Renders the Open Graph cards (1200x630) from tools/og/og-*.html with headless
# Chrome. Run after changing the headline, the sample card or the brand, then
# commit static/og.png and static/og-id.png. The pages append ?v=<content hash of the
# file> (assets.go), so crawlers refetch the image exactly when it changes.
set -euo pipefail
cd "$(dirname "$0")/../.."
CHROME="${CHROME:-/Applications/Google Chrome.app/Contents/MacOS/Google Chrome}"
for pair in "en:og.png" "id:og-id.png"; do
  lang="${pair%%:*}"; out="${pair##*:}"
  "$CHROME" --headless=new --disable-gpu --hide-scrollbars --force-device-scale-factor=1 \
    --window-size=1200,630 --virtual-time-budget=8000 \
    --screenshot="$PWD/static/$out" "file://$PWD/tools/og/og-$lang.html" >/dev/null 2>&1
  echo "static/$out"
done
