#!/usr/bin/env python3
"""
Netsekurity on-demand scanner — produces a report PDF for ANY domain.
Used by the Hermes cron worker (netsekurity-scan-worker) and nsec_worker.sh.

Flow:
  1. ensure assessment row exists in assessment.db for the domain (create if missing)
  2. run lightweight recon + core checks (throttled, read-only)
  3. insert any NEW findings
  4. call report_auto.py --domain <d>  -> writes <OUTROOT>/<domain>/<domain>_Assessment.pdf
  5. print the produced PDF path (for the worker to rename+upload)

Usage: python3 nsec_scan.py <domain>
"""
import os, sys, re, sqlite3, subprocess, datetime, json

DB = "/opt/data/report/assessment.db"
TEMPLATES = "/opt/data/report/templates"
OUTROOT = "/opt/data/report"
UA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/126.0"

def sh(c, timeout=30):
    return subprocess.run(c, shell=True, capture_output=True, text=True, timeout=timeout).stdout

def ensure_assessment(domain):
    c = sqlite3.connect(DB)
    cur = c.cursor()
    row = cur.execute("SELECT id FROM assessments WHERE domain=?", (domain,)).fetchone()
    if row:
        aid = row[0]
    else:
        now = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        cur.execute("INSERT INTO assessments (domain,agency,target_type,status,summary,assessed_on,created_at) VALUES (?,?,?,?,?,?,?)",
                    (domain, "Netsekurity client (on-demand)", "commercial", "assessed",
                     f"On-demand automated scan — {domain}", datetime.date.today().isoformat(), now))
        aid = cur.lastrowid
        c.commit()
    c.close()
    return aid

def recon(domain):
    """Lightweight recon: resolve, subdomains, http fingerprint. Returns list of (title,sev,category,desc,evidence,impact,recommendation)."""
    findings = []
    def add(sev, title, cat, desc, ev, impact="", rec=""):
        findings.append(dict(sev=sev, title=title, cat=cat, desc=desc, ev=ev, impact=impact, rec=rec))

    # resolve apex
    ips = sh(f"python3 -c \"import socket;print(' '.join(socket.gethostbyname_ex('{domain}')[2]))\"").strip()
    if not ips:
        add("INFO", f"Domain {domain} does not resolve", "recon",
            "The domain has no A/AAAA record; the scan surface may be empty.",
            f"gethostbyname_ex('{domain}') -> empty", "No web surface reachable.",
            "Confirm DNS records exist before commissioning a pentest.")
        return findings

    # http fingerprint
    code = sh(f"curl -sk -m 15 -A '{UA}' -o /dev/null -w '%{{http_code}} %{{content_type}}' https://{domain}/").strip()
    srv = sh(f"curl -sk -m 15 -A '{UA}' -D - -o /dev/null https://{domain}/ 2>/dev/null | grep -i '^server:' | head -1").strip()
    title = sh(f"curl -sk -m 15 -A '{UA}' https://{domain}/ 2>/dev/null | grep -oiE '<title>[^<]*' | head -1").strip()

    # security headers
    hdrs = sh(f"curl -sk -m 15 -A '{UA}' -D - -o /dev/null https://{domain}/ 2>/dev/null")
    hd = {h.lower():v for h,v in [tuple(l.split(':',1)) for l in hdrs.splitlines() if ':' in l]}
    missing = [h for h in ["strict-transport-security","x-content-type-options","x-frame-options","content-security-policy","referrer-policy"] if h not in hd]
    if missing:
        add("LOW", "Missing security headers", "misconfig",
            "Response lacks common hardening headers.",
            f"missing={','.join(missing)}; server={srv}",
            "Reduced browser-side protections (clickjacking, MIME sniffing).",
            "Add HSTS, X-Content-Type-Options, X-Frame-Options, CSP.")

    add("INFO", f"Web surface fingerprint — {domain}", "recon",
        f"HTTP {code}, server={srv}, title={title}, ip={ips[:60]}",
        f"curl https://{domain}/ -> {code}",
        "Establishes the scan surface.",
        "—")
    return findings

def store_findings(aid, findings):
    c = sqlite3.connect(DB)
    cur = c.cursor()
    for f in findings:
        d = dict(assessment_id=aid, host="*", title=f["title"], severity=f["sev"], cvss=None,
                 cvss_vector="", category=f["cat"], description=f["desc"],
                 evidence=f["ev"], impact=f["impact"] or f["desc"],
                 recommendation=f["rec"] or "Review and remediate.", owasp_top10="",
                 regulasi="", control_ref="", status="open", verified=1,
                 created_at=datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S"))
        # avoid duplicates by title+assessment
        if cur.execute("SELECT 1 FROM findings WHERE assessment_id=? AND title=?", (aid, f["title"])).fetchone():
            continue
        cols = list(d.keys())
        cur.execute("INSERT INTO findings ("+",".join(cols)+") VALUES ("+",".join(["?"]*len(cols))+")", [d[k] for k in cols])
    c.commit()
    c.close()

def main():
    domain = sys.argv[1].strip().lower()
    if not domain:
        print("ERR: domain required"); sys.exit(2)
    aid = ensure_assessment(domain)
    print(f"assessment_id={aid} domain={domain}")
    findings = recon(domain)
    store_findings(aid, findings)
    print(f"findings_stored={len(findings)}")
    # build PDF
    r = subprocess.run(["/opt/data/report/.venv/bin/python", os.path.join(TEMPLATES, "report_auto.py"), "--domain", domain],
                       capture_output=True, text=True, timeout=240)
    out = r.stdout.strip().splitlines()[-1] if r.stdout.strip() else r.stderr[-300:]
    print("build:", out)
    # locate produced pdf
    pdfs = [os.path.join(OUTROOT, domain, f) for f in os.listdir(os.path.join(OUTROOT, domain)) if f.endswith("_Assessment.pdf")]
    if pdfs:
        print(f"PDF={pdfs[0]}")
    else:
        print("PDF_ERROR"); sys.exit(1)

if __name__ == "__main__":
    main()
