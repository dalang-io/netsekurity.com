package main

import (
	"html/template"
	"net/http"
)

type dashboardData struct {
	Name         string
	Email        string
	Balance      float64
	IsAdmin      bool
	Packages     []pkg
	Transactions []txn
	Domains      []dom
}

type pkg struct {
	ID, Name string
	USD      float64
	Credits  float64
}

type txn struct {
	Type, Description, CreatedAt string
	Amount                       float64
}

type dom struct {
	Domain, Status, TXT string
}

// handleDashboard renders the authenticated dashboard.
func handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store, private")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	claims, err := currentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	var name, email string
	var balance float64
	db.QueryRow(`SELECT name, email FROM users WHERE id=?`, claims.Sub).Scan(&name, &email)
	db.QueryRow(`SELECT balance FROM credit_balance WHERE user_id=?`, claims.Sub).Scan(&balance)

	data := dashboardData{Name: name, Email: email, Balance: balance, IsAdmin: isAdmin(email)}

	rows, _ := db.Query(`SELECT id, name, usd_price, credits FROM credit_packages WHERE is_active=1 ORDER BY usd_price`)
	for rows.Next() {
		var p pkg
		rows.Scan(&p.ID, &p.Name, &p.USD, &p.Credits)
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

	drows, _ := db.Query(`SELECT domain, status, txt_verification_token FROM domains WHERE user_id=? ORDER BY created_at DESC`, claims.Sub)
	for drows.Next() {
		var d dom
		drows.Scan(&d.Domain, &d.Status, &d.TXT)
		data.Domains = append(data.Domains, d)
	}
	drows.Close()

	tmpl.ExecuteTemplate(w, "dashboard", data)
}

const dashboardHTML = `{{define "dashboard"}}<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
<title>dashboard — netsekurity</title>
<meta name="robots" content="noindex, nofollow"/>
<link rel="stylesheet" href="/css/styles.css?v={{cssHash}}"/>
<link rel="icon" type="image/svg+xml" href="/favicon.svg" />
<script src="https://unpkg.com/htmx.org@1.9.12/dist/htmx.min.js" defer></script>
</head>
<body class="scanlines bg-ink text-gray-300 min-h-screen">
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
<main id="main" class="mx-auto max-w-6xl px-4 py-5 sm:px-6">
  <div class="grid gap-4 lg:grid-cols-2">
    <!-- Left: balance + topup + transactions -->
    <div class="space-y-4">
      <section class="rounded border border-emerald-500/30 bg-[#04060c]">
        <div class="flex items-center gap-1.5 border-b border-emerald-500/20 px-3 py-2">
          <span class="h-2.5 w-2.5 rounded-full bg-emerald-500/70"></span>
          <span class="ml-1 font-mono text-[11px] text-gray-500">$ balance</span>
        </div>
        <div class="px-4 py-3">
          <div class="flex items-baseline justify-between">
            <span class="font-mono text-xs text-gray-400">credits</span>
            <span class="font-mono text-3xl font-bold text-emerald-300 glow">{{printf "%.1f" .Balance}}</span>
          </div>
          <p class="mt-1 font-mono text-[11px] text-gray-600"># 1 credit = 1 pentest · 1 domain</p>
        </div>
      </section>

      <section class="rounded border border-white/10 bg-[#04060c]">
        <div class="flex items-center gap-1.5 border-b border-white/10 px-3 py-2">
          <span class="h-2.5 w-2.5 rounded-full bg-red-500/70"></span>
          <span class="ml-1 font-mono text-[11px] text-gray-500">$ topup</span>
        </div>
        <div class="p-3">
          <div class="grid grid-cols-2 gap-2">
            {{range .Packages}}
            <form hx-post="/api/topup" hx-target="#topup-result" hx-swap="innerHTML" class="rounded border border-white/15 bg-ink p-3 hover:border-emerald-500/40">
              <input type="hidden" name="package_id" value="{{.ID}}"/>
              <div class="flex items-baseline justify-between">
                <span class="font-mono text-sm text-white">{{.Name}}</span>
                <span class="font-mono text-sm text-emerald-400">${{printf "%.0f" .USD}}</span>
              </div>
              <div class="font-mono text-[11px] text-gray-500">{{printf "%.0f" .Credits}} credits</div>
              <button class="mt-2 w-full rounded border border-emerald-400 bg-emerald-500/10 px-2 py-1.5 font-mono text-xs font-bold text-emerald-300 hover:bg-emerald-500/20">buy</button>
            </form>
            {{end}}
          </div>
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
          <table class="w-full font-mono text-xs">
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
          {{else}}
          <p class="font-mono text-xs text-gray-600"># no transactions yet</p>
          {{end}}
        </div>
      </section>
    </div>

    <!-- Right: domains -->
    <section class="rounded border border-cyan-500/30 bg-[#04060c] h-fit">
      <div class="flex items-center gap-1.5 border-b border-cyan-500/20 px-3 py-2">
        <span class="h-2.5 w-2.5 rounded-full bg-cyan-500/70"></span>
        <span class="ml-1 font-mono text-[11px] text-gray-500">$ domains</span>
      </div>
      <div class="p-3">
        <form class="flex gap-2" hx-post="/api/domains" hx-target="#domain-result" hx-swap="innerHTML">
          <input name="domain" required placeholder="target.example.com" aria-label="Domain to verify" class="flex-1 rounded border border-white/15 bg-ink px-2 py-1.5 font-mono text-xs text-white focus:border-cyan-400 focus:outline-none"/>
          <button class="rounded border border-cyan-400 bg-cyan-500/10 px-3 py-1.5 font-mono text-xs font-bold text-cyan-300 hover:bg-cyan-500/20">add</button>
        </form>
        <div id="domain-result" class="mt-2"></div>
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
                  hx-on::after-request="if(event.detail.successful) setTimeout(()=>location.reload(),500)"
                  class="rounded border border-emerald-400/60 px-2 py-0.5 text-[11px] font-bold text-emerald-300 hover:bg-emerald-500/15">verify</button>
                <button hx-post="/api/domains/delete" hx-vals='{"domain":"{{.Domain}}"}'
                  hx-target="#domain-result" hx-swap="innerHTML"
                  hx-on::after-request="if(event.detail.successful) setTimeout(()=>location.reload(),500)"
                  class="rounded border border-red-400/50 px-2 py-0.5 text-[11px] font-bold text-red-300 hover:bg-red-500/15">del</button>
                {{else}}
                <span class="rounded bg-emerald-500/20 px-2 py-0.5 text-[11px] font-bold text-emerald-300">[verified]</span>
                <button hx-post="/api/pentests/start" hx-vals='{"domain":"{{.Domain}}"}'
                  hx-target="#pentest-result" hx-swap="innerHTML"
                  hx-on::after-request="if(event.detail.successful) setTimeout(()=>location.reload(),800)"
                  class="rounded border border-cyan-400/60 px-2 py-0.5 text-[11px] font-bold text-cyan-300 hover:bg-cyan-500/15">scan</button>
                {{end}}
              </span>
            </div>
            <div class="mt-1 flex items-center gap-1 text-[11px] text-gray-500">
              <span class="text-cyan-400">_netsekurity</span>
              <span class="text-gray-600">TXT</span>
              <span class="truncate">{{.TXT}}</span>
              <button class="shrink-0 text-cyan-400 hover:underline"
                onclick="navigator.clipboard.writeText('{{.TXT}}');this.textContent='copied';setTimeout(()=>this.textContent='copy',1200)">copy</button>
            </div>
          </li>
          {{end}}
        </ul>
        {{else}}
        <p class="mt-2 font-mono text-xs text-gray-600"># no domains — add one to start a pentest</p>
        {{end}}
      </div>
    </section>
  </div>
  <div id="pentest-result" class="mt-3"></div>
  <div id="pentest-list">
    <div hx-get="/api/pentests/list" hx-trigger="load, every 10s" hx-swap="innerHTML"></div>
  </div>
</main>
</body>
</html>{{end}}`

var tmpl = template.Must(template.New("dashboard").Funcs(template.FuncMap{"cssHash": func() string { return cssHash }}).Parse(dashboardHTML))
