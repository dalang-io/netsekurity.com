#!/usr/bin/env bash
# Netsekurity worker MONITOR — suppresses LLM runs when there's no queued job.
# Emits a stable token when the queue is empty (=> cron skips the LLM entirely,
# saving AI tokens). When a queued pentest appears, the output changes => the
# LLM agent wakes, runs the scan, and uploads the report.
# After the job is claimed (status -> running), output returns to the stable
# "IDLE" token, so subsequent minute-ticks stay silent again.
#
# Output contract (stable, deterministic, no timestamps):
#   IDLE                 -> no queued/running job (LLM suppressed)
#   RUNNING              -> a job is in-flight (LLM suppressed, don't double-trigger)
#   JOB <pid> <domain>   -> exactly one queued job (LLM wakes to process it)
set -uo pipefail
H=163.128.54.5
KEY=/opt/data/recon/keys/dalang_key
# Cheap: read the queue state from the production DB over SSH (no token leaked).
STATE=$(ssh -i "$KEY" -o StrictHostKeyChecking=no -o ConnectTimeout=10 root@$H \
  "python3 -c \"
import sqlite3
c=sqlite3.connect('/home/dalang/netsekurity.com/data/netsekurity.db')
q=c.execute(\\\"SELECT COUNT(*) FROM pentests WHERE status='queued'\\\").fetchone()[0]
r=c.execute(\\\"SELECT COUNT(*) FROM pentests WHERE status='running'\\\").fetchone()[0]
if r!=0: print('RUNNING')
elif q!=0:
    row=c.execute(\\\"SELECT p.id,d.domain FROM pentests p JOIN domains d ON d.id=p.domain_id WHERE p.status='queued' ORDER BY p.created_at ASC LIMIT 1\\\").fetchone()
    print('JOB %s %s' % (row[0], row[1]))
else: print('IDLE')
c.close()
\"" 2>/dev/null)
echo "${STATE:-IDLE}"
