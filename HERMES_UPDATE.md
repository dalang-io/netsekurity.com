# Hermes agent — update runbook

For the operator (human or agent) of the **Hermes scan worker**. Behaviour
reference lives in [AGENT_WORKER.md](AGENT_WORKER.md); this file is only the
"what do I run, and how do I know it worked" note.

Current change: **the scanner now reports a severity breakdown**, and the
dashboard shows it on the scan card. Until you run the steps below, the platform
side is live but receives nothing, and every card says *"severity breakdown is in
the report"*. Nothing is broken in the meantime — the field is optional.

---

## 1. Update

The agent runs from a checkout of this repo, so this is the whole deployment:

```bash
cd <your checkout of netsekurity.com>
git pull
```

No copying into `/opt/data/netsekurity_draft/`. `nsec_worker.sh` now resolves
`nsec_scan.py` next to itself and only falls back to that path when the file is
not beside it.

## 2. Verify statically

```bash
grep -c 'SEVERITY='  nsec_scan.py     # expect 1  — scanner emits the breakdown
grep -c 'findings='  nsec_worker.sh   # expect 1  — cron worker forwards it
grep -c 'findings='  nsec_io.sh       # expect 1  — direct upload path forwards it
grep -qF 'RUNLOG="$WORK/scan-$PID.log"' nsec_worker.sh && echo "per-run log ok"
bash -n nsec_worker.sh && bash -n nsec_io.sh && echo "syntax ok"
python3 -c "import ast;ast.parse(open('nsec_scan.py').read());print('python ok')"
```

A missing count, or no `per-run log ok`, means the pull did not land — check
`git log --oneline -3` shows `801f5c5` or later.

## 3. Verify on a real scan

Commission one scan from the dashboard, then:

```bash
grep -E 'scanner:|severity:|upload-resp:' /tmp/nsk_worker.log | tail -6
```

Expected:

| Line | Expected value |
|---|---|
| `scanner:` | a path **inside your checkout** — not `/opt/data/netsekurity_draft/` |
| `severity:` | `critical=0,high=2,medium=7,low=3,info=9` — **not** `<none>` |
| `upload-resp:` | `{"ok":true,"report":"..."}` |

Then open the dashboard card for that scan: it should show coloured chips
(`2 high`, `7 medium`, …) or a green `no findings`, instead of the fallback note.

---

## Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| `severity: <none>` and `scanner:` points at `/opt/data/netsekurity_draft/` | An old `nsec_scan.py` is being executed, not the pulled one | Confirm `nsec_scan.py` sits beside `nsec_worker.sh` in the checkout, or update the copy at that path too |
| `severity: <none>` and `scanner:` points inside the checkout | The scan errored before printing `SEVERITY=` | Read the run log: `cat /tmp/nsk_worker/scan-<pentest_id>.log` |
| Card still says "severity breakdown is in the report" | Upload succeeded without the field, or the string did not parse | Check `upload-resp:`; the platform ignores unparseable values rather than storing a wrong count |
| Chips show but the numbers look wrong | Stale-log attribution (fixed in `801f5c5`) | Confirm the per-run log check in step 2 prints `per-run log ok` |

---

## The contract

`nsec_scan.py` prints, next to the existing `PDF=` and `TOTAL_FINDINGS=` lines:

```
SEVERITY=critical=0,high=2,medium=7,low=3,info=9
```

Both upload paths forward it to `/api/pentests/worker/report` as the optional
form field `findings`:

- `nsec_worker.sh` — cron path, reads it from that run's own log
- `nsec_io.sh upload <pentest_id> <pdf> [findings]` — third argument

Platform behaviour:

- **Optional and additive.** Omit it and an un-updated agent keeps working.
- Unknown keys, negative numbers and malformed values are ignored. If nothing
  parses, the summary counts as *absent*, never as zero.
- `findings_reported` keeps **"the scan found nothing"** (green `no findings`)
  distinct from **"the agent never told us"** (fallback note). Do not send
  `critical=0,high=0,medium=0,low=0,info=0` unless the scan genuinely found
  nothing — that renders as a clean bill of health.

## Two things worth knowing

**Per-run logs.** `/tmp/nsk_worker.log` is shared and append-only. `PDF=` and
`SEVERITY=` used to be grepped from it with `tail -1`, so a run that failed to
print one picked up a *previous* run's line — attaching another domain's report
or findings to this pentest. Each run now writes `/tmp/nsk_worker/scan-<pid>.log`
and is read from there. Keep it that way.

**Old scans cannot be backfilled.** The counts exist in `assessment.db` on the
agent side, but nothing records which `assessment_id` produced which
`pentest_id`. If backfill matters later, store the assessment id on the
`pentests` row at upload time.
