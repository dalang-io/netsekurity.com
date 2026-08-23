#!/usr/bin/env python3
# Netsekurity queue monitor state — run on the production HOST.
# Emits one stable token (no timestamps) so the cron can decide whether to wake
# the LLM worker. Also recovers stale-running jobs (reset to queued if > 45min).
import sqlite3, sys

DB = "/home/dalang/netsekurity.com/data/netsekurity.db"

def main():
    c = sqlite3.connect(DB)
    # Stale-running recovery (no LLM needed).
    c.execute(
        "UPDATE pentests SET status='queued', started_at=NULL "
        "WHERE status='running' AND started_at IS NOT NULL AND started_at != '' "
        "AND replace(replace(started_at,'T',' '),'Z','') < datetime('now','-45 minutes')"
    )
    c.commit()
    q = c.execute("SELECT COUNT(*) FROM pentests WHERE status='queued'").fetchone()[0]
    r = c.execute("SELECT COUNT(*) FROM pentests WHERE status='running'").fetchone()[0]
    if r != 0:
        print("RUNNING")
    elif q != 0:
        row = c.execute(
            "SELECT p.id, COALESCE(d.domain,'') FROM pentests p "
            "LEFT JOIN domains d ON d.id=p.domain_id "
            "WHERE p.status='queued' ORDER BY p.created_at ASC LIMIT 1"
        ).fetchone()
        print("JOB %s %s" % (row[0], row[1]))
    else:
        print("IDLE")
    c.close()

if __name__ == "__main__":
    main()