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
  <div class="mx-auto flex max-w-5xl items-center justify-between px-4 py-4 sm:px-6">
    <a href="/" class="font-mono text-lg font-bold text-white"><span class="text-emerald-400">net</span>sekurity</a>
    <div class="flex items-center gap-4 text-sm">
      <span class="text-gray-400">{{.Name}} · {{.Email}}</span>
      <a href="/logout" class="text-gray-400 hover:text-white">Logout</a>
    </div>
  </div>
</header>
<main class="mx-auto max-w-5xl px-4 py-10 sm:px-6 space-y-10">
  <!-- Balance -->
  <section class="rounded-xl border border-emerald-500/40 bg-panel p-6">
    <h1 class="text-2xl font-bold text-white">Credits balance</h1>
    <div class="mt-2 font-mono text-4xl font-bold text-emerald-400">{{printf "%.1f" .Balance}} <span class="text-lg text-gray-400">credits</span></div>
    <p class="mt-1 text-sm text-gray-500">1 credit = 1 pentest · 1 domain</p>
  </section>

  <!-- Top up -->
  <section class="rounded-xl border border-white/10 bg-panel p-6">
    <h2 class="text-lg font-bold text-white">Top up credits</h2>
    <div class="mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
      {{range .Packages}}
      <form hx-post="/api/topup" hx-target="#topup-result" hx-swap="innerHTML" class="rounded-lg border border-white/10 bg-ink p-4">
        <input type="hidden" name="package_id" value="{{.ID}}"/>
        <div class="font-bold text-white">{{.Name}}</div>
        <div class="mt-1 font-mono text-xl text-emerald-400">${{printf "%.0f" .USD}}</div>
        <div class="text-sm text-gray-500">{{printf "%.0f" .Credits}} credits</div>
        <button class="mt-3 w-full rounded-md bg-emerald-500 px-3 py-2 text-sm font-bold text-black hover:bg-emerald-400">Buy</button>
      </form>
      {{end}}
    </div>
    <div id="topup-result" class="mt-4"></div>
  </section>

  <!-- Domains -->
  <section class="rounded-xl border border-white/10 bg-panel p-6">
    <h2 class="text-lg font-bold text-white">Domains</h2>
    <form class="mt-4 flex gap-2" hx-post="/api/domains" hx-target="#domain-result" hx-swap="innerHTML">
      <input name="domain" required placeholder="example.com" class="flex-1 rounded-md border border-white/10 bg-ink px-3 py-2 text-white focus:border-emerald-400 focus:outline-none"/>
      <button class="rounded-md bg-cyan-500 px-4 py-2 text-sm font-bold text-black hover:bg-cyan-400">Add domain</button>
    </form>
    <div id="domain-result" class="mt-4"></div>
    {{if .Domains}}
    <ul class="mt-4 space-y-2">
      {{range .Domains}}
      <li class="flex items-center justify-between rounded-md border border-white/10 bg-ink px-4 py-2 text-sm">
        <span class="font-mono">{{.Domain}}</span>
        <span class="rounded-full px-2 py-0.5 text-xs {{if eq .Status "verified"}}bg-emerald-500/20 text-emerald-300{{else}}bg-yellow-500/20 text-yellow-300{{end}}">{{.Status}}</span>
      </li>
      {{end}}
    </ul>
    {{end}}
  </section>

  <!-- Transactions -->
  <section class="rounded-xl border border-white/10 bg-panel p-6">
    <h2 class="text-lg font-bold text-white">Transaction history</h2>
    {{if .Transactions}}
    <ul class="mt-4 divide-y divide-white/10">
      {{range .Transactions}}
      <li class="flex items-center justify-between py-2 text-sm">
        <span class="{{if eq .Type "topup"}}text-emerald-300{{else}}text-yellow-300{{end}}">{{.Type}}</span>
        <span class="text-gray-400">{{.Description}}</span>
        <span class="font-mono {{if eq .Type "topup"}}text-emerald-400{{else}}text-gray-300{{end}}">{{if eq .Type "topup"}}+{{end}}{{printf "%.1f" .Amount}}</span>
      </li>
      {{end}}
    </ul>
    {{else}}
    <p class="mt-4 text-sm text-gray-500">No transactions yet. Top up credits to get started.</p>
    {{end}}
  </section>
</main>
</body>
</html>{{end}}`

var tmpl = template.Must(template.New("dashboard").Parse(dashboardHTML))
