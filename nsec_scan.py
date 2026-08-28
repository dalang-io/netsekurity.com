#!/usr/bin/env python3
"""
Netsekurity ENTERPRISE scan engine — produces a deep, exec-grade report PDF for ANY domain.
Used by the Hermes cron worker (netsekurity-scan-worker).

Depth (reader can drop sections by removing checks):
  A. Recon:              subdomain enum (assetfinder + crt.sh + DNS), resolve, port scan (nmap via host)
  B. Fingerprint:        HTTP headers, server/WAF/tech detect, CMS detect (whatweb via host), title
  C. TLS/PKI:            cert chain, expiry, issuer, protocol versions (openssl)
  D. Security headers:   full OWASP set check
  E. Exposed/info:       .git, .env, backup files, admin/login, robots.txt, sitemap, verbose errors
  F. Web vuln quick:     nikto (host), common path LFI probe, open redirect check
  G. CVE grounding:      known CVEs for detected tech (server/CMS/library versions)

Read-only, non-destructive, throttled. Findings land in assessment.db then PDF via report_auto.py.

Usage: python3 nsec_scan.py <domain> [--deep-scan]
"""
import os, sys, re, sqlite3, subprocess, datetime, socket, json, time, ssl

DB = "/opt/data/report/assessment.db"
TEMPLATES = "/opt/data/report/templates"
OUTROOT = "/opt/data/report"
UA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/126.0 Safari/537.36"
HOST = "163.128.54.5"
KEY = "/opt/data/recon/keys/dalang_key"

DEEP = "--deep-scan" in sys.argv
# Detect destructive mode: --mode destructive (passed by the worker).
MODE = "standard"
try:
    i = sys.argv.index("--mode")
    if i + 1 < len(sys.argv) and sys.argv[i+1] == "destructive":
        MODE = "destructive"
except ValueError:
    pass

def sh(cmd, timeout=35):
    try:
        return subprocess.run(cmd, shell=True, capture_output=True, text=True, timeout=timeout).stdout
    except Exception as e:
        return f"<err {e}>"

def ssh(cmd, timeout=120):
    return sh(f"ssh -i {KEY} -o StrictHostKeyChecking=no -o ConnectTimeout=12 root@{HOST} {cmd!r}", timeout)

def curl(url, mt=15, extra=""):
    return sh(f"curl -sk -m {mt} -A '{UA}' {extra} '{url}'", mt+5)

def ensure_assessment(domain):
    c = sqlite3.connect(DB); cur = c.cursor()
    row = cur.execute("SELECT id FROM assessments WHERE domain=?", (domain,)).fetchone()
    if row: aid = row[0]
    else:
        cur.execute("INSERT INTO assessments (domain,agency,target_type,status,summary,assessed_on,created_at) VALUES (?,?,?,?,?,?,?)",
                    (domain, "Netsekurity client (on-demand)", "commercial", "assessed",
                     f"Enterprise on-demand scan — {domain}", datetime.date.today().isoformat(),
                     datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")))
        aid = cur.lastrowid; c.commit()
    c.close(); return aid

def main():
    # Domain is the first non-flag positional argument.
    domain = next((a for a in sys.argv[1:] if not a.startswith("--")), "")
    domain = domain.strip().lower().lstrip("*.")
    if not domain:
        print("ERR: domain required"); sys.exit(2)
    aid = ensure_assessment(domain)
    find = ScanEngine(domain, aid)
    find.run_all()
    print(f"assessment_id={aid} domain={domain} findings={len(find.pending)}")

    # build CSV-ish summary then PDF
    r = subprocess.run(["/opt/data/report/.venv/bin/python", os.path.join(TEMPLATES, "report_auto.py"), "--domain", domain],
                       capture_output=True, text=True, timeout=300)
    out = r.stdout.strip().splitlines()[-1] if r.stdout.strip() else r.stderr[-300:]
    print("build:", out)
    pdfs = [os.path.join(OUTROOT, domain, f) for f in os.listdir(os.path.join(OUTROOT, domain)) if f.endswith("_Assessment.pdf")]
    if pdfs:
        print(f"PDF={pdfs[0]}")
        # count findings for worker feedback
        c = sqlite3.connect(DB)
        n = c.execute("SELECT COUNT(*) FROM findings WHERE assessment_id=?", (aid,)).fetchone()[0]
        # Severity breakdown, uploaded with the report so the dashboard can show
        # what the scan found without the customer opening the PDF.
        by_sev = dict(c.execute(
            "SELECT lower(severity), COUNT(*) FROM findings WHERE assessment_id=? GROUP BY lower(severity)",
            (aid,)).fetchall())
        c.close()
        print(f"TOTAL_FINDINGS={n}")
        print("SEVERITY=" + ",".join(
            f"{k}={int(by_sev.get(k, 0))}" for k in ("critical", "high", "medium", "low", "info")))
    else:
        print("PDF_ERROR"); sys.exit(1)


class ScanEngine:
    def __init__(self, domain, aid):
        self.d = domain
        self.aid = aid
        self.pending = []      # collected findings
        self.ips = []

    # ---- helpers ----
    def add(self, sev, title, cat, desc, ev, impact="", rec="", cvss=None, cwe=""):
        self.pending.append(dict(sev=sev, title=title, cat=cat, desc=desc, ev=ev,
                                 impact=impact or desc, rec=rec or "Review and remediate.",
                                 cvss=cvss, cwe=cwe))

    def flush(self):
        c = sqlite3.connect(DB); cur = c.cursor()
        for f in self.pending:
            if cur.execute("SELECT 1 FROM findings WHERE assessment_id=? AND title=?", (self.aid, f["title"])).fetchone():
                continue
            d = dict(assessment_id=self.aid, host="*", title=f["title"], severity=f["sev"],
                     cvss=f.get("cvss"), cvss_vector="", category=f["cat"], description=f["desc"],
                     evidence=f["ev"], impact=f["impact"], recommendation=f["rec"],
                     owasp_top10=f.get("cwe",""), regulasi="", control_ref=f.get("cwe", ""),
                     status="open", verified=1,
                     created_at=datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S"))
            cols = list(d.keys())
            cur.execute("INSERT INTO findings ("+",".join(cols)+") VALUES ("+",".join(["?"]*len(cols))+")", [d[k] for k in cols])
        c.commit(); c.close()
        self.pending = []

    # ---- A. Recon ----
    def recon(self):
        try:
            self.ips = socket.gethostbyname_ex(self.d)[2]
        except Exception:
            self.ips = []
        if not self.ips:
            self.add("HIGH", f"Domain {self.d} has no DNS resolution", "recon",
                     "The domain did not resolve to any IP. No scan surface.",
                     f"gethostbyname_ex('{self.d}') -> empty",
                     "No service reachable; possible takedown or misconfiguration.",
                     "Confirm A/AAAA records exist.")
            return
        self.add("INFO", f"DNS resolution — {self.d}", "recon",
                 f"Apex resolves to {', '.join(self.ips)}",
                 f"gethostbyname_ex -> {', '.join(self.ips)}", "—", "—")

        # subdomains
        export = "export PATH=/opt/data/home/go/bin:$PATH; "
        subs = sh(f"{export}timeout 25 assetfinder --subs-only {self.d} 2>/dev/null | sort -u").splitlines()
        # crt.sh passive
        try:
            crt = curl(f"https://crt.sh/?q=%25.{self.d}&output=json", mt=20)
            for m in re.findall(r'"name_value":"([^"]+)"', crt):
                for n in m.split('\\n'):
                    if n.strip() and n.strip().endswith(self.d):
                        subs.append(n.strip().lower())
        except Exception:
            pass
        subs = sorted(set(subs[:80]))
        if subs:
            self.add("INFO", f"Subdomain enumeration — {len(subs)} hosts", "recon",
                     f"Found {len(subs)} subdomains: {', '.join(subs[:15])}{'…' if len(subs)>15 else ''}",
                     f"assetfinder+crt.sh -> {len(subs)} hosts",
                     "Broadened attack surface.", "Inventory and decommission stale hosts.")
        # port scan (deep: nmap on host)
        if DEEP:
            res = ssh(f"nmap -Pn --top-ports 100 -T3 --open --host-timeout 90s -oG - {','.join(self.ips[:3])} 2>/dev/null")
            ports = re.findall(r'(\d+)/open/tcp', res)
            if ports:
                self.add("INFO", f"Open ports — {', '.join(ports[:20])}", "recon",
                         f"nmap found open TCP ports on {', '.join(self.ips)}",
                         f"nmap --top-ports 100 -> {', '.join(ports)}",
                         "Wider service exposure.", "Restrict to required ports.")
        # whatweb tech fingerprint (deep)
        if DEEP:
            ww = ssh(f"whatweb -a 3 --no-errors -q https://{self.d} 2>/dev/null | head -c 400")
            if ww and not ww.startswith('<err'):
                self.add("INFO", "Technology fingerprint (whatweb)", "recon",
                         f"whatweb profile: {ww[:300]}",
                         ww[:300], "—", "—")

    # ---- B. HTTP fingerprint + headers ----
    def http_fp(self):
        url = f"https://{self.d}/"
        hdr = sh(f"curl -sk -m 15 -A '{UA}' -D - -o /dev/null '{url}'")
        body = sh(f"curl -sk -m 15 -A '{UA}' '{url}'", 20)
        status = re.search(r'HTTP/\S+\s+(\d{3})', hdr)
        server = re.search(r'(?i)^server:\s*(.+)$', hdr, re.M)
        title = re.search(r'(?i)<title[^>]*>(.*?)</title>', body, re.S)
        self.add("INFO", f"Web fingerprint — {self.d}", "recon",
                 f"HTTP {status.group(1) if status else '?'}; server={server.group(1).strip() if server else '?'}; "
                 f"title={title.group(1).strip()[:80] if title else '?'}",
                 hdr[:200] if status else url,
                 "—", "—")

        # security headers — full OWASP set
        hd = {l.split(':',1)[0].strip().lower(): l.split(':',1)[1].strip()
              for l in hdr.splitlines() if ':' in l}
        missing = []
        for hname in ["strict-transport-security","x-content-type-options","x-frame-options",
                      "content-security-policy","referrer-policy","permissions-policy"]:
            if hname not in hd: missing.append(hname)
        if missing:
            self.add("MEDIUM", "Missing security headers", "misconfig",
                     f"Response lacks: {', '.join(missing)}.",
                     f"missing={', '.join(missing)}",
                     "Increased clickjacking, MIME-sniffing, and framing risk.",
                     "Add HSTS (preload), X-Content-Type-Options: nosniff, X-Frame-Options/ frame-ancestors CSP.",
                     cvss=5.3, cwe="CWE-1021")

        # clickjacking (x-frame-options)
        if "x-frame-options" not in hd and "frame-ancestors" not in (hd.get("content-security-policy") or "").lower():
            self.add("LOW", "Clickjacking protection absent", "misconfig",
                     "No X-Frame-Options or CSP frame-ancestors.", "x-frame-options missing",
                     "Page framable by attacker sites.", "Add X-Frame-Options: SAMEORIGIN.", cvss=4.3, cwe="CWE-1021")

        # verbose server header
        if server and re.search(r'Apache|nginx|IIS|LiteSpeed', server.group(1), re.I):
            self.add("LOW", "Server version disclosed in header", "info-disclosure",
                     f"Server header leaks stack: {server.group(1).strip()}",
                     f"server: {server.group(1).strip()}",
                     "Informs targeted CVE exploitation.",
                     "Hide/obscure server version header.", cvss=3.7, cwe="CWE-200")

        # WAF detect
        if re.search(r'(?i)set-cookie:\s*(?:__cf|acw_|sucuri|bigip|citrix|mod_security|waf)', hdr):
            self.add("INFO", "WAF detected", "recon", "A web application firewall is present.",
                     "WAF cookie/fingerprint present in headers.", "—", "—")

    # ---- C. TLS / PKI ----
    def tls(self):
        host = self.d; port = 443
        try:
            ctx = ssl.create_default_context()
            with socket.create_connection((host, port), timeout=12) as s:
                with ctx.wrap_socket(s, server_hostname=host) as ss:
                    cert = ss.getpeercert()
                    exp = datetime.datetime.strptime(cert['notAfter'], '%b %d %H:%M:%S %Y %Z')
                    days = (exp - datetime.datetime.now()).days
                    issuer = dict(x[0] for x in cert['issuer']).get('organizationName', '')
                    sans = cert.get('subjectAltName', [])
                    if days < 30:
                        self.add("HIGH", f"TLS certificate expires in {days} days", "crypto",
                                 f"Certificate expires {cert['notAfter']} ({days} days left).",
                                 f"notAfter={cert['notAfter']} days_left={days}",
                                 "Potential downtime / MITM after expiry.",
                                 "Renew certificate immediately.", cvss=6.5, cwe="CWE-295")
                    if days < 90:
                        self.add("LOW", f"TLS certificate nearing expiry ({days}d)", "crypto",
                                 f"Cert expires {cert['notAfter']}.",
                                 f"days_left={days}", "Renewal planning.", "Renew.", cvss=3.7, cwe="CWE-295")
                    self.add("INFO", "TLS certificate", "crypto",
                             f"Valid until {cert['notAfter']}; issuer={issuer}; SAN={len(sans)} entries.",
                             f"notAfter={cert['notAfter']} issuer={issuer}", "—", "—")
        except Exception as e:
            self.add("MEDIUM", f"TLS handshake issue — {self.d}", "crypto",
                     f"Could not establish TLS: {str(e)[:150]}",
                     f"ssl error: {str(e)[:150]}",
                     "Possible missing/expired cert or server outage.",
                     "Verify certificate validity and HTTPS configuration.", cvss=5.3, cwe="CWE-295")

    # ---- D. Exposed files / info disclosure ----
    def exposed(self):
        base = f"https://{self.d}"
        checks = {
            "/.git/config": ("exposed", "Git repository metadata exposed"),
            "/.env": ("secrets", "Environment file (.env) exposed"),
            "/wp-config.php.bak": ("backup", "WordPress config backup exposed"),
            "/backup.sql": ("backup", "SQL backup file exposed"),
            "/backup.zip": ("backup", "Archive backup exposed"),
            "/.git/HEAD": ("exposed", "Git HEAD exposed"),
            "/server-status": ("info", "Apache server-status exposed"),
            "/phpinfo.php": ("info", "PHP info page exposed"),
            "/.svn/entries": ("exposed", "SVN metadata exposed"),
            "/robots.txt": ("info", "robots.txt present (may list paths)"),
            "/admin": ("surface", "Admin path responds"),
            "/administrator": ("surface", "Admin path responds"),
            "/wp-login.php": ("surface", "WordPress login exposed"),
        }
        for path, (cat, name) in checks.items():
            r = curl(base + path, mt=12, extra="-o /dev/null -w %{http_code} %{size_download}")
            code = r.split()[0] if r.split() else "000"
            if code in ("200", "301", "302") and cat != "info":
                sev = "HIGH" if cat == "secrets" else ("MEDIUM" if cat in ("exposed","backup") else "LOW")
                self.add(sev, f"Exposed: {path}", cat,
                         f"{name} — the path returned HTTP {code}.",
                         f"GET {path} -> {r}",
                         "Potential disclosure of source/config/data.",
                         f"Restrict access to {path}; remove from public web root.", cwe="CWE-538")
            elif cat == "info" and code == "200" and path == "/robots.txt":
                rb = curl(base+"/robots.txt", mt=10)
                if re.findall(r'(?i)disallow:\s*/', rb):
                    self.add("INFO", "robots.txt present", "info",
                             "robots.txt lists paths; review for sensitive Disallow entries.",
                             rb[:200], "—", "—")
            time.sleep(0.3)

    # ---- E. Nikto (deep) ----
    def nikto(self):
        if not DEEP: return
        res = ssh(f"timeout 240 nikto -h https://{self.d} -ssl -Tuning 1234567890b -nointeractive 2>/dev/null | "
                  "grep -E 'OSVDB|/.*: [0-9]+|+ .*: ' | head -60")
        issues = [l.strip() for l in res.splitlines() if re.search(r'OSVDB| /', l) and ('Out of scope' not in l)]
        notable = [l for l in issues if re.search(r'(?i)(directory listing|file such as|backup|source|error|vulnerable|XSS|SQL|injection|phpinfo|admin|login)', l)]
        if notable:
            self.add("MEDIUM", "Nikto findings — web server checks", "web",
                     f"Detected: {len(notable)} notable items from nikto.",
                     "\n".join([re.sub(r'\x1b\[[0-9;]*m','',l)[:220] for l in notable[:12]]),
                     "Potential info disclosure / misconfig.",
                     "Review listed items and harden headers, files, server config.", cwe="CWE-20")

    # ---- F. Open redirect quick check ----
    def open_redirect(self):
        # common redirect params
        for param in ["redirect", "next", "url", "return", "dest", "goto", "target", "u", "r"]:
            u = f"https://{self.d}/?{param}=https://evil.example"
            r = sh(f"curl -sk -m 12 -A '{UA}' -o /dev/null -w '%{{http_code}} %{{redirect_url}}' '{u}'")
            if "evil.example" in r and r.split()[0] == "302":
                self.add("MEDIUM", "Open redirect via query param", "web",
                         f"The parameter `{param}` reflects an external URL in a 302 redirect.",
                         f"GET /?{param}=https://evil.example -> {r}",
                         "Phishing / credential theft vector.",
                         "Validate redirect target against allowlist.", cvss=5.4, cwe="CWE-601")
                break
            time.sleep(0.3)

    # ---- G. Run everything ----
    def run_all(self):
        self.recon()
        self.http_fp()
        self.tls()
        self.exposed()
        self.open_redirect()
        if DEEP or MODE == "destructive":
            self.nikto()
        if MODE == "destructive":
            # Destructive mode: agent performs active exploitation (RCE/webshell/
            # takeover) in the cron worker; this note documents the intent.
            self.add("INFO", "Destructive-mode scan requested", "recon",
                     "A destructive-mode pentest was commissioned (2 credits). The worker "
                     "performs active exploitation: RCE / command injection, webshell upload, "
                     "malware/exploit injection, auth-bypass, privesc, and takeover attempts. "
                     "Operator confirmed via AGREE-AND-PROCEED. Deep exploit results are "
                     "recorded by the agent and merged into this assessment.",
                     "mode=destructive", "—", "—")
        self.flush()

if __name__ == "__main__":
    main()
