package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
)

type dashboardData struct {
	Name         string
	Email        string
	Balance      float64
	IsAdmin      bool
	Packages     []pkg
	Transactions []txn
	Payments     []pymt // pending + paid
	Expired      []pymt // expired, collapsed behind a disclosure
	Domains      []dom
	RateNote     string
}

type pkg struct {
	ID, Name string
	USD      float64
	Credits  float64
	// Local is the amount Xendit will actually charge, e.g. "Rp 825.000".
	Local string
}

type txn struct {
	Type, Description, CreatedAt string
	Amount                       float64
}

// pymt is an Xendit invoice shown in the transaction history (incl. pending/expired).
type pymt struct {
	ExternalID, Package, Status, CreatedAt, URL string
	AmountUSD                                   float64
	Credits                                     float64
}

type dom struct {
	Domain, Status, TXT string
}

// renderDomainListOOB writes the shared domain list as an HTMX out-of-band swap.
// Handlers that change a domain append it to their response so the list refreshes
// in place — the dashboard used to force a full location.reload() instead, which
// threw away the very message the user had just been shown.
func renderDomainListOOB(w io.Writer, userID string) {
	var d dashboardData
	db.QueryRow(`SELECT balance FROM credit_balance WHERE user_id=?`, userID).Scan(&d.Balance)
	rows, err := db.Query(`SELECT domain, status, txt_verification_token FROM domains WHERE user_id=? ORDER BY created_at DESC`, userID)
	if err != nil {
		return
	}
	for rows.Next() {
		var dm dom
		rows.Scan(&dm.Domain, &dm.Status, &dm.TXT)
		d.Domains = append(d.Domains, dm)
	}
	rows.Close()
	var b strings.Builder
	if err := tmpl.ExecuteTemplate(&b, "domainlist", d); err != nil {
		log.Printf("domainlist render error: %v", err)
		return
	}
	io.WriteString(w, strings.Replace(b.String(),
		`<div id="domain-list">`, `<div id="domain-list" hx-swap-oob="true">`, 1))
}

// renderBalanceOOB writes the credit balance as an HTMX out-of-band swap so an
// action that spends credits updates the figure without a page reload.
func renderBalanceOOB(w io.Writer, userID string) {
	var balance float64
	db.QueryRow(`SELECT balance FROM credit_balance WHERE user_id=?`, userID).Scan(&balance)
	fmt.Fprintf(w, `<span id="credit-balance" hx-swap-oob="true" class="font-mono text-3xl font-bold text-emerald-300 glow">%.1f</span>`, balance)
}

// handleDashboard renders the authenticated dashboard.
func handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store, private")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	// Sync any newly-paid payments so credits appear immediately (no webhook needed).
	syncPendingPayments()
	claims, err := currentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	var name, email string
	var balance float64
	db.QueryRow(`SELECT name, email FROM users WHERE id=?`, claims.Sub).Scan(&name, &email)
	db.QueryRow(`SELECT balance FROM credit_balance WHERE user_id=?`, claims.Sub).Scan(&balance)

	data := dashboardData{Name: name, Email: email, Balance: balance, IsAdmin: isAdmin(email), RateNote: localRateNote()}

	rows, _ := db.Query(`SELECT id, name, usd_price, credits FROM credit_packages WHERE is_active=1 ORDER BY usd_price`)
	for rows.Next() {
		var p pkg
		rows.Scan(&p.ID, &p.Name, &p.USD, &p.Credits)
		p.Local = localPrice(p.USD)
		data.Packages = append(data.Packages, p)
	}
	rows.Close()

	trows, _ := db.Query(`SELECT type, amount, description, created_at FROM credit_transactions WHERE user_id=? ORDER BY created_at DESC LIMIT 20`, claims.Sub)
	for trows.Next() {
		var t txn
		trows.Scan(&t.Type, &t.Amount, &t.Description, &t.CreatedAt)
		data.Transactions = append(data.Transactions, t)
	}
	trows.Close()

	// Mark overdue pending invoices as expired, then load the user's payment history
	// (instances/pending/paid/expired) so the dashboard reflects real invoice state.
	db.Exec(`UPDATE payments SET status='expired' WHERE user_id=? AND status='pending' AND created_at < datetime('now','-1 day')`, claims.Sub)
	ptrows, _ := db.Query(`SELECT external_id, COALESCE(package_id,''), status, COALESCE(paid_at,''), amount_usd, credits, created_at, COALESCE(url,'')
		FROM payments WHERE user_id=? ORDER BY id DESC LIMIT 20`, claims.Sub)
	for ptrows.Next() {
		var p pymt
		var paidSrc string
		_ = ptrows.Scan(&p.ExternalID, &p.Package, &p.Status, &paidSrc, &p.AmountUSD, &p.Credits, &p.CreatedAt, &p.URL)
		if p.Status == "paid" && paidSrc != "" {
			p.CreatedAt = paidSrc
		}
		// Expired invoices are noise on every page load — collapse them so the
		// history panel leads with what the customer can still act on.
		if p.Status == "expired" {
			data.Expired = append(data.Expired, p)
		} else {
			data.Payments = append(data.Payments, p)
		}
	}
	ptrows.Close()

	drows, _ := db.Query(`SELECT domain, status, txt_verification_token FROM domains WHERE user_id=? ORDER BY created_at DESC`, claims.Sub)
	for drows.Next() {
		var d dom
		drows.Scan(&d.Domain, &d.Status, &d.TXT)
		data.Domains = append(data.Domains, d)
	}
	drows.Close()

	if err := tmpl.ExecuteTemplate(w, "dashboard", data); err != nil {
		log.Printf("dashboard render error: %v", err)
	}
}

const dashboardHTML = `{{define "dashboard"}}<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
<title>dashboard — netsekurity</title>
<meta name="robots" content="noindex, nofollow"/>
<link rel="stylesheet" href="/css/styles.css?v={{cssHash}}"/>
<link rel="icon" type="image/svg+xml" href="/favicon.svg" />{{fbpixel}}
<script src="https://unpkg.com/htmx.org@1.9.12/dist/htmx.min.js" defer></script>
</head>
<body class="scanlines bg-ink text-gray-300 min-h-screen overflow-x-hidden">
<header class="border-b border-emerald-500/25 bg-ink/85">
  <div class="mx-auto flex max-w-6xl items-center justify-between px-4 py-3">
    <a href="/" class="font-mono text-base font-bold text-white"><span class="glow text-emerald-400">net</span>sekurity<span class="text-emerald-500">.com</span> <span class="text-xs text-cyan-300">/dashboard</span></a>
    <div class="flex items-center gap-3 font-mono text-xs">
      {{if .IsAdmin}}<a href="/su" class="text-yellow-300 hover:text-yellow-200">su</a>{{end}}
      <span class="hidden text-gray-500 sm:inline">{{.Name}} · {{.Email}}</span>
      <a href="/logout" class="prompt text-gray-500 hover:text-white">logout</a>
    </div>
  </div>
</header>
<main id="main" class="mx-auto max-w-6xl overflow-x-hidden px-4 py-5 sm:px-6">
  <div class="grid w-full gap-4 lg:grid-cols-2">
    <!-- Left: balance + topup + transactions -->
    <div class="min-w-0 w-full space-y-4">
      <section class="min-w-0 w-full rounded border border-emerald-500/30 bg-[#04060c]">
        <div class="flex items-center gap-1.5 border-b border-emerald-500/20 px-3 py-2">
          <span class="h-2.5 w-2.5 rounded-full bg-emerald-500/70"></span>
          <span class="ml-1 font-mono text-[11px] text-gray-500">$ balance</span>
        </div>
        <div class="px-4 py-3">
          <div class="flex items-baseline justify-between">
            <span class="font-mono text-xs text-gray-400">credits</span>
            <span id="credit-balance" class="font-mono text-3xl font-bold text-emerald-300 glow">{{printf "%.1f" .Balance}}</span>
          </div>
          <p class="mt-1 font-mono text-[11px] text-gray-600"># 1 credit = 1 pentest · 1 domain</p>
          {{if le .Balance 0.0}}
          <p class="mt-2 rounded border border-yellow-500/30 bg-yellow-500/10 px-2 py-1.5 font-mono text-[11px] text-yellow-300">
            No credits yet — buy a package below to run your first scan.
          </p>
          {{end}}
        </div>
      </section>

      <section class="min-w-0 w-full rounded border border-white/10 bg-[#04060c]">
        <div class="flex items-center gap-1.5 border-b border-white/10 px-3 py-2">
          <span class="h-2.5 w-2.5 rounded-full bg-red-500/70"></span>
          <span class="ml-1 font-mono text-[11px] text-gray-500">$ topup</span>
        </div>
        <div class="p-3">
          <div class="grid grid-cols-1 gap-2 sm:grid-cols-2">
            {{range .Packages}}
            <form hx-post="/api/topup" hx-target="#topup-result" hx-swap="innerHTML" class="rounded border border-white/15 bg-ink p-3 hover:border-emerald-500/40">
              <input type="hidden" name="package_id" value="{{.ID}}"/>
              <div class="flex items-baseline justify-between">
                <span class="font-mono text-sm text-white">{{.Name}}</span>
                <span class="font-mono text-sm text-emerald-400">${{printf "%.0f" .USD}}</span>
              </div>
              <div class="font-mono text-[11px] text-gray-500">{{printf "%.0f" .Credits}} credits</div>
              {{if .Local}}<div class="mt-0.5 font-mono text-[11px] text-cyan-300">you pay {{.Local}}</div>{{end}}
              <button onclick="fbq('track','InitiateCheckout',{content_name:'{{.Name}}',content_type:'product',value:{{.USD}},currency:'USD'})"
                class="mt-2 w-full rounded border border-emerald-400 bg-emerald-500/10 px-2 py-1.5 font-mono text-xs font-bold text-emerald-300 hover:bg-emerald-500/20">buy{{if .Local}} — {{.Local}}{{end}}</button>
            </form>
            {{end}}
          </div>
          {{if .RateNote}}<p class="mt-2 font-mono text-[11px] leading-relaxed text-gray-500">{{.RateNote}}</p>{{end}}
          <div id="topup-result" class="mt-2"></div>
        </div>
      </section>

      <section class="rounded border border-white/10 bg-[#04060c]">
        <div class="flex items-center gap-1.5 border-b border-white/10 px-3 py-2">
          <span class="h-2.5 w-2.5 rounded-full bg-yellow-500/70"></span>
          <span class="ml-1 font-mono text-[11px] text-gray-500">$ history</span>
        </div>
        <div class="p-3">
          {{if .Transactions}}
          <div class="overflow-x-auto">
          <table class="w-full min-w-[320px] font-mono text-xs">
            <thead><tr class="text-left text-gray-500"><th class="py-1">type</th><th>desc</th><th class="text-right">amt</th></tr></thead>
            <tbody>
            {{range .Transactions}}
            <tr class="border-t border-white/10">
              <td class="py-1 {{if eq .Type "topup"}}text-emerald-300{{else}}text-yellow-300{{end}}">{{.Type}}</td>
              <td class="text-gray-500">{{.Description}}</td>
              <td class="text-right font-mono {{if eq .Type "topup"}}text-emerald-400{{else}}text-gray-300{{end}}">{{if eq .Type "topup"}}+{{end}}{{printf "%.1f" .Amount}}</td>
            </tr>
            {{end}}
            </tbody>
          </table>
          </div>
          {{else}}
          <p class="font-mono text-xs text-gray-600"># no transactions yet</p>
          {{end}}

          {{if .Payments}}
          <div class="mt-4 border-t border-white/10 pt-2">
            <div class="mb-1 font-mono text-[11px] text-gray-500"># invoices</div>
            <div class="overflow-x-auto">
            <table class="w-full min-w-[320px] font-mono text-xs">
              <thead><tr class="text-left text-gray-500"><th class="py-1">invoice</th><th>status</th><th class="text-right">credits</th></tr></thead>
              <tbody>
              {{range .Payments}}
              <tr class="border-t border-white/10">
                <td class="py-1" title="{{.ExternalID}}">
                  <span class="text-gray-200">{{if .Package}}{{.Package}}{{else}}top-up{{end}}</span>
                  <span class="text-gray-500"> · ${{printf "%.0f" .AmountUSD}}</span>
                  <span class="block text-[10px] text-gray-600">{{.CreatedAt}}</span>
                </td>
                <td class="py-1"><span class="rounded px-1.5 py-0.5 text-[10px] font-bold {{if eq .Status "paid"}}bg-emerald-500/20 text-emerald-300{{else}}bg-yellow-500/20 text-yellow-300{{end}}">{{.Status}}</span>{{if eq .Status "pending"}}{{if .URL}}&nbsp;<a href="{{.URL}}" target="_blank" rel="noopener" class="rounded border border-emerald-400 bg-emerald-500/10 px-2 py-0.5 text-[10px] font-bold text-emerald-300 hover:bg-emerald-500/20">pay →</a>{{end}}{{end}}</td>
                <td class="text-right font-mono text-emerald-400">{{printf "%.0f" .Credits}}</td>
              </tr>
              {{end}}
              </tbody>
            </table>
            </div>
          </div>
          {{end}}

          {{if .Expired}}
          <details class="mt-3 border-t border-white/10 pt-2">
            <summary class="cursor-pointer font-mono text-[11px] text-gray-600 hover:text-gray-400">
              # {{len .Expired}} expired invoice(s) — unpaid within 24h
            </summary>
            <div class="mt-1 overflow-x-auto">
            <table class="w-full min-w-[320px] font-mono text-xs">
              <tbody>
              {{range .Expired}}
              <tr class="border-t border-white/10 text-gray-600">
                <td class="py-1" title="{{.ExternalID}}">{{if .Package}}{{.Package}}{{else}}top-up{{end}} · ${{printf "%.0f" .AmountUSD}}</td>
                <td class="py-1 text-[10px]">{{.CreatedAt}}</td>
                <td class="text-right">{{printf "%.0f" .Credits}}</td>
              </tr>
              {{end}}
              </tbody>
            </table>
            </div>
          </details>
          {{end}}
        </div>
      </section>

      <section class="rounded border border-white/10 bg-[#04060c]">
        <div class="flex items-center gap-1.5 border-b border-white/10 px-3 py-2">
          <span class="h-2.5 w-2.5 rounded-full bg-cyan-500/70"></span>
          <span class="ml-1 font-mono text-[11px] text-gray-500">$ api tokens (CI/CD)</span>
        </div>
        <div class="p-3">
          <p class="font-mono text-[11px] text-gray-500">Generate a token for your CI/CD pipeline. On every
            deploy, your pipeline calls <code class="text-cyan-300">POST /api/v1/pentests</code> with
            <code class="text-cyan-300">X-API-Token</code> to start a pentest automatically.</p>
          <div class="mt-2 flex items-end gap-2">
            <div class="flex-1">
              <label class="font-mono text-[10px] text-gray-500">expiry</label>
              <select id="token-expiry" class="w-full rounded border border-white/15 bg-ink px-2 py-1.5 font-mono text-xs text-white focus:border-cyan-400 focus:outline-none">
                <option value="7">7 days</option>
                <option value="14">14 days</option>
                <option value="30" selected>30 days</option>
                <option value="60">60 days</option>
                <option value="90">90 days</option>
              </select>
            </div>
            <div class="flex-1">
              <label class="font-mono text-[10px] text-gray-500">name</label>
              <input id="token-name" type="text" value="CI/CD" class="w-full rounded border border-white/15 bg-ink px-2 py-1.5 font-mono text-xs text-white focus:border-cyan-400 focus:outline-none"/>
            </div>
            <button
              hx-post="/api/tokens/create"
              hx-vals='js:{"name": document.getElementById("token-name").value, "expiry_days": document.getElementById("token-expiry").value}'
              hx-target="#token-result" hx-swap="innerHTML"
              class="rounded border border-cyan-400 bg-cyan-500/10 px-3 py-1.5 font-mono text-xs font-bold text-cyan-300 hover:bg-cyan-500/20">generate</button>
          </div>
          <div id="token-result" class="mt-2"></div>
          <div id="token-list">
            <div hx-get="/api/tokens" hx-trigger="load, every 15s" hx-swap="innerHTML"></div>
          </div>
        </div>
      </section>
    </div>

    <!-- Right: scans first (what the customer paid for), then domains -->
    <div class="min-w-0 w-full space-y-4">

      <section class="min-w-0 w-full rounded border border-emerald-500/30 bg-[#04060c]">
        <div class="flex items-center gap-1.5 border-b border-emerald-500/20 px-3 py-2">
          <span class="h-2.5 w-2.5 rounded-full bg-emerald-500/70"></span>
          <span class="ml-1 font-mono text-[11px] text-gray-500">$ scans</span>
        </div>
        <div class="p-3">
          <div id="pentest-result"></div>
          <div id="pentest-list" hx-get="/api/pentests/list" hx-trigger="load, every 10s, refresh" hx-swap="innerHTML">
            <p class="font-mono text-xs text-gray-600"># loading…</p>
          </div>
        </div>
      </section>

      <section class="min-w-0 w-full rounded border border-cyan-500/30 bg-[#04060c] h-fit">
        <div class="flex items-center gap-1.5 border-b border-cyan-500/20 px-3 py-2">
          <span class="h-2.5 w-2.5 rounded-full bg-cyan-500/70"></span>
          <span class="ml-1 font-mono text-[11px] text-gray-500">$ domains</span>
        </div>
        <div class="p-3">
          <form class="flex min-w-0 gap-2" hx-post="/api/domains" hx-target="#domain-result" hx-swap="innerHTML">
            <input name="domain" required placeholder="app.example.com" aria-label="Domain to verify"
              pattern="[A-Za-z0-9]([A-Za-z0-9\-]*[A-Za-z0-9])?(\.[A-Za-z0-9]([A-Za-z0-9\-]*[A-Za-z0-9])?)+"
              title="Hostname only — no https:// and no trailing slash"
              class="flex-1 rounded border border-white/15 bg-ink px-2 py-1.5 font-mono text-xs text-white focus:border-cyan-400 focus:outline-none"/>
            <button class="rounded border border-cyan-400 bg-cyan-500/10 px-3 py-1.5 font-mono text-xs font-bold text-cyan-300 hover:bg-cyan-500/20">add</button>
          </form>
          <p class="mt-1 font-mono text-[10px] text-gray-600">hostname only — <span class="text-gray-500">app.example.com</span>, not https://app.example.com/</p>
          <div id="domain-result" class="mt-2"></div>
          {{template "domainlist" .}}
        </div>
      </section>

    </div>
  </div>
</main>

<!-- Destructive mode confirm modal -->
<div id="destructive-modal" class="fixed inset-0 z-[100] hidden items-center justify-center bg-black/70 p-4">
  <div class="w-full max-w-lg rounded-lg border border-red-500/40 bg-[#0a0d16] p-5 shadow-[0_0_40px_rgba(239,68,68,0.25)]">
    <div class="mb-2 font-mono text-sm font-bold text-red-300 glow">⚠ DESTRUCTIVE PENTEST — EXTREME CAUTION</div>
    <div class="space-y-2 font-mono text-xs leading-relaxed text-gray-300">
      <p><span class="text-red-400">You are about to run a DESTRUCTIVE-mode scan.</span> This performs
      active exploitation including <b>RCE, webshell upload, malware/exploit injection, and takeover attempts</b>.</p>
      <p class="text-yellow-300">⚠ It can modify, corrupt, or take down the target. It costs <b>2 credits</b>
      (standard is 1).</p>
      <p class="text-yellow-300">Strongly recommended: run against a <b>development / staging server</b>,
      never production, unless explicitly authorized.</p>
      <p class="text-gray-400">You are responsible for authorization and any damage caused.</p>
      <label class="mt-3 block">
        <span class="text-gray-400">Type <b class="text-red-300">AGREE AND PROCEED</b> to confirm:</span>
        <input id="destructive-agree" type="text" autocomplete="off"
          class="mt-1 w-full rounded border border-red-500/40 bg-black px-3 py-2 font-mono text-sm text-white focus:border-red-400 focus:outline-none"
          placeholder="AGREE AND PROCEED"/>
      </label>
      <p id="destructive-domain-tag" class="mt-2 font-mono text-[10px] text-gray-500"></p>
    </div>
    <div class="mt-4 flex items-center justify-end gap-2">
      <button onclick="closeDestructiveModal()" class="rounded border border-white/20 px-3 py-1.5 font-mono text-xs text-gray-400 hover:bg-white/5">cancel</button>
      <button id="destructive-confirm-btn" onclick="startDestructive()"
        class="rounded border border-red-500 bg-red-600/20 px-3 py-1.5 font-mono text-xs font-bold text-red-200 hover:bg-red-600/30 disabled:cursor-not-allowed disabled:opacity-40"
        disabled>run destructive →</button>
    </div>
  </div>
</div>

<script>
let destructiveDomain = null;
function openDestructiveModal(domain) {
  destructiveDomain = domain;
  document.getElementById('destructive-domain-tag').textContent = 'target: ' + domain + ' · cost: 2 credits';
  document.getElementById('destructive-agree').value = '';
  document.getElementById('destructive-confirm-btn').disabled = true;
  var m = document.getElementById('destructive-modal');
  m.classList.remove('hidden'); m.classList.add('flex');
}
function closeDestructiveModal() {
  var m = document.getElementById('destructive-modal');
  m.classList.add('hidden'); m.classList.remove('flex');
  destructiveDomain = null;
}
document.getElementById('destructive-agree').addEventListener('input', function () {
  document.getElementById('destructive-confirm-btn').disabled =
    (this.value.trim() !== 'AGREE AND PROCEED');
});
function startDestructive() {
  if (!destructiveDomain) return;
  var btn = document.getElementById('destructive-confirm-btn');
  btn.disabled = true;
  btn.textContent = 'running…';
  var xhr = new XMLHttpRequest();
  var body = 'domain=' + encodeURIComponent(destructiveDomain) + '&mode=destructive';
  xhr.open('POST', '/api/pentests/start');
  xhr.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
  xhr.setRequestHeader('HX-Request', 'true');
  xhr.onload = function () {
    // htmx.swap processes the out-of-band balance and domain-list fragments the
    // handler returns alongside the message, so nothing needs a page reload.
    if (window.htmx) {
      htmx.swap('#pentest-result', xhr.responseText, { swapStyle: 'innerHTML' });
      htmx.trigger('#pentest-list', 'refresh');
    } else {
      document.getElementById('pentest-result').innerHTML = xhr.responseText;
    }
    closeDestructiveModal();
  };
  xhr.send(body);
}
</script>
</body>
</html>{{end}}

{{define "domainlist"}}<div id="domain-list">
{{if .Domains}}
<ul class="mt-2 space-y-1.5">
  {{range .Domains}}
  <li class="rounded border border-white/10 bg-ink px-2.5 py-1.5 font-mono text-xs">
    <div class="flex items-center justify-between gap-2">
      <span class="truncate text-white">{{.Domain}}</span>
      <span class="flex items-center gap-1.5">
        {{if ne .Status "verified"}}
        <button hx-post="/api/domains/verify" hx-vals='{"domain":"{{.Domain}}"}'
          hx-target="#domain-result" hx-swap="innerHTML"
          class="rounded border border-emerald-400/60 px-2 py-0.5 text-[11px] font-bold text-emerald-300 hover:bg-emerald-500/15">verify</button>
        <button hx-post="/api/domains/delete" hx-vals='{"domain":"{{.Domain}}"}'
          hx-target="#domain-result" hx-swap="innerHTML"
          hx-confirm="Remove {{.Domain}}?"
          class="rounded border border-red-400/50 px-2 py-0.5 text-[11px] font-bold text-red-300 hover:bg-red-500/15">del</button>
        {{else}}
        <span class="rounded bg-emerald-500/20 px-2 py-0.5 text-[11px] font-bold text-emerald-300">[verified]</span>
        {{if le $.Balance 0.0}}
        <button type="button" disabled title="No credits — buy a package to run a scan."
          class="cursor-not-allowed rounded border border-white/15 px-2 py-0.5 text-[11px] font-bold text-gray-600">scan</button>
        {{else}}
        <button hx-post="/api/pentests/start" hx-vals='{"domain":"{{.Domain}}","mode":"standard"}'
          hx-target="#pentest-result" hx-swap="innerHTML"
          title="Standard scan — 1 credit (read-only, non-destructive)"
          class="rounded border border-cyan-400/60 px-2 py-0.5 text-[11px] font-bold text-cyan-300 hover:bg-cyan-500/15">scan</button>
        {{end}}
        {{end}}
      </span>
    </div>
    <div class="mt-1 flex items-center gap-1 text-[11px] text-gray-500">
      <span class="text-cyan-400">_netsekurity</span>
      <span class="text-gray-600">TXT</span>
      <span class="truncate">{{.TXT}}</span>
      <button type="button" class="shrink-0 text-cyan-400 hover:underline"
        onclick="navigator.clipboard.writeText('{{.TXT}}');this.textContent='copied';setTimeout(()=>this.textContent='copy',1200)">copy</button>
    </div>
    {{if eq .Status "verified"}}
    <div class="mt-1 border-t border-white/5 pt-1">
      <button type="button" onclick="openDestructiveModal('{{.Domain}}')"
        {{if le $.Balance 2.0}}disabled title="Destructive mode costs 2 credits."{{end}}
        class="font-mono text-[10px] text-red-400/70 underline decoration-dotted hover:text-red-300 disabled:cursor-not-allowed disabled:text-gray-700 disabled:no-underline">
        advanced: destructive scan (2 credits)
      </button>
    </div>
    {{end}}
  </li>
  {{end}}
</ul>
{{else}}
<p class="mt-2 font-mono text-xs text-gray-600"># no domains — add one to start a scan</p>
{{end}}
</div>{{end}}`

var tmpl = template.Must(template.New("dashboard").Funcs(template.FuncMap{
	"cssHash": func() string { return cssHash },
	// template.HTML, not string: html/template escapes a plain string in an
	// element context and the pixel would render as literal &lt;script&gt; text.
	"fbpixel": func() template.HTML { return template.HTML(metaPixelSnippet(true)) },
}).Parse(dashboardHTML))
