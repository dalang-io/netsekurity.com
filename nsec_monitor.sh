#!/usr/bin/env bash
# Netsekurity worker MONITOR — decides whether the cron should wake the LLM agent.
#
# Emits a stable, deterministic token (no timestamps) so the cron suppresses the
# LLM when idle, and wakes it exactly when a queued job appears:
#   IDLE     -> no queued/running job (LLM suppressed)
#   RUNNING  -> a job is in-flight (LLM suppressed; don't double-trigger)
#   JOB <pid> <domain> -> exactly one queued job (LLM wakes to process it)
#
# The actual SQL lives in nsec_state.py (deployed to the host). We just SSH and run
# it — NO inline python/escaping here, so a failure must NOT silently print IDLE.
set -uo pipefail

H=163.128.54.5
KEY=/opt/data/recon/keys/dalang_key
STATE=$(ssh -i "$KEY" -o StrictHostKeyChecking=no -o ConnectTimeout=10 root@$H \
  "python3 /home/dalang/netsekurity.com/nsec_state.py" 2>/dev/null)

if [ -z "$STATE" ]; then
  # ssh/query failed. Do NOT pretend idle or we stall the queue forever.
  # Emit a value guaranteed to differ from the last stable IDLE baseline so the
  # cron wakes the agent and it can self-heal (requeue stale / process queue).
  echo "ERROR"
  exit 0
fi

echo "$STATE"