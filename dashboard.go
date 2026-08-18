package main

import (
	"html/template"
	"net/http"
)

type dashboardData struct {
	Name        string
	Email       string
	Balance     float64
	Packages    []pkg
	Transactions []txn
	Domains     []dom
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
	Domain, Status string
}

// handleDashboard renders the authenticated dashboard.
func handleDashboard(w http.ResponseWriter, r *http.Request) {
	claims, err := currentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	var name, email string
	var balance float64
	db.QueryRow(`SELECT name, email FROM users WHERE id=?`, claims.Sub).Scan(&name, &email)
	db.QueryRow(`SELECT balance FROM credit_balance WHERE user_id=?`, claims.Sub).Scan(&balance)

	data := dashboardData{Name: name, Email: email, Balance: balance}

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

	drows, _ := db.Query(`SELECT domain, status FROM domains WHERE user_id=? ORDER BY created_at DESC`, claims.Sub)
	for drows.Next() {
		var d dom
		drows.Scan(&d.Domain, &d.Status)
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
<title>Dashboard — netsekurity</title>
<link rel="stylesheet" href="/css/styles.css"/>
<script src="https://unpkg.com/htmx.org@1.9.12/dist/htmx.min.js" defer></script>
</head>
<body class="bg-ink text-gray-200 min-h-screen">
<header class="border-b border-white/10 bg-ink/80">
  <div class="mx-auto flex max-w-6xl items-center justify-between px-4 py-3">
    <a href="/" class="font-mono text-base font-bold text-white"><span class="text-emerald-400">net</span>sekurity</a>
    <div class="flex items-center gap-3 text-xs">
      <span class="hidden text-gray-500 sm:inline">{{.Name}} · {{.Email}}</span>
      <a href="/logout" class="text-gray-500 hover:text-white">Logout</a>
    </div>
  </div>
</header>
<main class="mx-auto max-w-6xl px-4 py-5 sm:px-6">
  <div class="grid gap-4 lg:grid-cols-2">
    <!-- Left: balance + top up -->
    <div class="space-y-4">
      <section class="rounded-lg border border-emerald-500/40 bg-panel px-4 py-3">
        <div class="flex items-baseline justify-between">
          <h1 class="text-sm font-bold text-white">Credits</h1>
          <span class="font-mono text-2xl font-bold text-emerald-400">{{printf "%.1f" .Balance}}</span>
        </div>
        <p class="text-[11px] text-gray-500">1 credit = 1 pentest · 1 domain</p>
      </section>

      <section class="rounded-lg border border-white/10 bg-panel p-3">
        <h2 class="text-sm font-bold text-white">Top up</h2>
        <div class="mt-2 grid grid-cols-2 gap-2">
          {{range .Packages}}
          <form hx-post="/api/topup" hx-target="#topup-result" hx-swap="innerHTML" class="rounded-md border border-white/10 bg-ink p-3">
            <input type="hidden" name="package_id" value="{{.ID}}"/>
            <div class="flex items-baseline justify-between">
              <span class="text-sm font-bold text-white">{{.Name}}</span>
              <span class="font-mono text-sm text-emerald-400">${{printf "%.0f" .USD}}</span>
            </div>
            <div class="text-[11px] text-gray-500">{{printf "%.0f" .Credits}} credits</div>
            <button class="mt-2 w-full rounded bg-emerald-500 px-2 py-1.5 text-xs font-bold text-black hover:bg-emerald-400">Buy</button>
          </form>
          {{end}}
        </div>
        <div id="topup-result" class="mt-2"></div>
      </section>

      <section class="rounded-lg border border-white/10 bg-panel p-3">
        <h2 class="text-sm font-bold text-white">Transactions</h2>
        {{if .Transactions}}
        <ul class="mt-2 divide-y divide-white/10 text-xs">
          {{range .Transactions}}
          <li class="flex items-center justify-between py-1.5">
            <span class="{{if eq .Type "topup"}}text-emerald-300{{else}}text-yellow-300{{end}}">{{.Type}}</span>
            <span class="text-gray-500">{{.Description}}</span>
            <span class="font-mono {{if eq .Type "topup"}}text-emerald-400{{else}}text-gray-300{{end}}">{{if eq .Type "topup"}}+{{end}}{{printf "%.1f" .Amount}}</span>
          </li>
          {{end}}
        </ul>
        {{else}}
        <p class="mt-2 text-xs text-gray-500">No transactions yet.</p>
        {{end}}
      </section>
    </div>

    <!-- Right: domains -->
    <section class="rounded-lg border border-white/10 bg-panel p-3 h-fit">
      <h2 class="text-sm font-bold text-white">Domains</h2>
      <form class="mt-2 flex gap-2" hx-post="/api/domains" hx-target="#domain-result" hx-swap="innerHTML">
        <input name="domain" required placeholder="example.com" class="flex-1 rounded border border-white/10 bg-ink px-2 py-1.5 text-xs text-white focus:border-emerald-400 focus:outline-none"/>
        <button class="rounded bg-cyan-500 px-3 py-1.5 text-xs font-bold text-black hover:bg-cyan-400">Add</button>
      </form>
      <div id="domain-result" class="mt-2"></div>
      {{if .Domains}}
      <ul class="mt-2 space-y-1.5">
        {{range .Domains}}
        <li class="flex items-center justify-between gap-2 rounded border border-white/10 bg-ink px-2.5 py-1.5 text-xs">
          <span class="font-mono truncate">{{.Domain}}</span>
          <span class="flex items-center gap-1.5">
            {{if ne .Status "verified"}}
            <button hx-post="/api/domains/verify" hx-vals='{"domain":"{{.Domain}}"}'
              hx-target="#domain-result" hx-swap="innerHTML"
              hx-on::after-request="if(event.detail.successful) setTimeout(()=>location.reload(),500)"
              class="rounded bg-emerald-500/20 px-2 py-0.5 text-[11px] font-bold text-emerald-300 hover:bg-emerald-500/30">Verify</button>
            <button hx-post="/api/domains/delete" hx-vals='{"domain":"{{.Domain}}"}'
              hx-target="#domain-result" hx-swap="innerHTML"
              hx-on::after-request="if(event.detail.successful) setTimeout(()=>location.reload(),500)"
              class="rounded bg-red-500/15 px-2 py-0.5 text-[11px] font-bold text-red-300 hover:bg-red-500/30">Delete</button>
            {{else}}
            <span class="rounded-full bg-emerald-500/20 px-2 py-0.5 text-[11px] font-bold text-emerald-300">verified</span>
            {{end}}
          </span>
        </li>
        {{end}}
      </ul>
      {{else}}
      <p class="mt-2 text-xs text-gray-500">No domains yet. Add one to start a pentest.</p>
      {{end}}
    </section>
  </div>
</main>
</body>
</html>{{end}}`

var tmpl = template.Must(template.New("dashboard").Parse(dashboardHTML))
