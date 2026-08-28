package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type suUser struct {
	ID, Email, Name, Role string
	Credits               float64
}

type suTrx struct {
	ExternalID, Package, Status, CreatedAt, PaidAt string
	Amount, Credits                                float64
}

type suDomain struct {
	ID                                        int64
	UserEmail, Domain, Status, TXT, CreatedAt string
}

type suPentest struct {
	ID, UserEmail, Domain, Status, Mode, ReportRef, CreatedAt string
}

// suStats are the aggregates the console needs at a glance. Without them the
// operator has to read every row to notice something like "no invoice has ever
// been paid", which is exactly what went unnoticed for six weeks.
type suStats struct {
	Users           int
	InvoicesPaid    int
	InvoicesExpired int
	InvoicesPending int
	RevenueUSD      float64
	CreditsOut      float64
	DomainsPending  int
	ScansActive     int
	UsersShown      int
}

type suData struct {
	Users     []suUser
	Trx       []suTrx
	Domains   []suDomain
	Pentests  []suPentest
	Credits   []suCredit
	Stats     suStats
	CurrentID string
}

// suCredit is one manual credit grant, shown as an audit trail beside the
// control that creates them.
type suCredit struct {
	Email, Amount, Description, CreatedAt string
}

// reportsDir returns the directory where pentest PDF reports are stored.
func reportsDir() string {
	path := getenv("DATABASE_PATH", "./data/netsekurity.db")
	dir := filepath.Dir(path)
	return filepath.Join(dir, "reports")
}

// handleAdmin renders the super-admin dashboard.
func handleAdmin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store, private")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	// Mark overdue pending invoices as expired (Xendit invoice_duration = 1 day).
	db.Exec(`UPDATE payments SET status='expired' WHERE status='pending' AND created_at < datetime('now','-1 day')`)

	data := suData{Users: []suUser{}, Trx: []suTrx{}, Domains: []suDomain{}, Pentests: []suPentest{}}

	// Identify the operator so the console can refuse to let them demote themselves.
	if claims, err := currentUser(r); err == nil {
		data.CurrentID = claims.Sub
	}

	st := &data.Stats
	db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&st.Users)
	db.QueryRow(`SELECT
			COALESCE(SUM(status='paid'),0), COALESCE(SUM(status='expired'),0), COALESCE(SUM(status='pending'),0),
			COALESCE(SUM(CASE WHEN status='paid' THEN amount_usd ELSE 0 END),0)
		FROM payments`).Scan(&st.InvoicesPaid, &st.InvoicesExpired, &st.InvoicesPending, &st.RevenueUSD)
	db.QueryRow(`SELECT COALESCE(SUM(balance),0) FROM credit_balance`).Scan(&st.CreditsOut)
	db.QueryRow(`SELECT COUNT(*) FROM domains WHERE status!='verified'`).Scan(&st.DomainsPending)
	db.QueryRow(`SELECT COUNT(*) FROM pentests WHERE status IN ('queued','running')`).Scan(&st.ScansActive)

	crows, _ := db.Query(`SELECT COALESCE(u.email,ct.user_id), ct.amount, COALESCE(ct.description,''), ct.created_at
		FROM credit_transactions ct LEFT JOIN users u ON u.id=ct.user_id
		WHERE ct.type='admin_credit' ORDER BY ct.id DESC LIMIT 25`)
	for crows.Next() {
		var c suCredit
		var amt float64
		crows.Scan(&c.Email, &amt, &c.Description, &c.CreatedAt)
		c.Amount = fmt.Sprintf("%.1f", amt)
		data.Credits = append(data.Credits, c)
	}
	crows.Close()

	// Bounded: this query used to load every user row on every page view.
	rows, _ := db.Query(`SELECT u.id, u.email, COALESCE(u.name,''), COALESCE(u.role,'user'), COALESCE(cb.balance,0)
		FROM users u LEFT JOIN credit_balance cb ON cb.user_id=u.id ORDER BY u.created_at DESC LIMIT 500`)
	for rows.Next() {
		var u suUser
		rows.Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.Credits)
		data.Users = append(data.Users, u)
	}
	rows.Close()
	data.Stats.UsersShown = len(data.Users)

	trows, _ := db.Query(`SELECT external_id, COALESCE(package_id,''), status, COALESCE(paid_at,''), amount_usd, credits, created_at
		FROM payments ORDER BY id DESC LIMIT 200`)
	for trows.Next() {
		var t suTrx
		trows.Scan(&t.ExternalID, &t.Package, &t.Status, &t.PaidAt, &t.Amount, &t.Credits, &t.CreatedAt)
		data.Trx = append(data.Trx, t)
	}
	trows.Close()

	drows, _ := db.Query(`SELECT d.id, COALESCE(u.email,''), d.domain, d.status, d.txt_verification_token, d.created_at
		FROM domains d LEFT JOIN users u ON u.id=d.user_id ORDER BY d.id DESC`)
	for drows.Next() {
		var d suDomain
		drows.Scan(&d.ID, &d.UserEmail, &d.Domain, &d.Status, &d.TXT, &d.CreatedAt)
		data.Domains = append(data.Domains, d)
	}
	drows.Close()

	prows, _ := db.Query(`SELECT p.id, COALESCE(u.email,''), COALESCE(d.domain,''), p.status, COALESCE(p.mode,'standard'), COALESCE(p.report_ref,''), p.created_at
		FROM pentests p LEFT JOIN users u ON u.id=p.user_id LEFT JOIN domains d ON d.id=p.domain_id
		ORDER BY p.created_at DESC LIMIT 100`)
	for prows.Next() {
		var p suPentest
		prows.Scan(&p.ID, &p.UserEmail, &p.Domain, &p.Status, &p.Mode, &p.ReportRef, &p.CreatedAt)
		data.Pentests = append(data.Pentests, p)
	}
	prows.Close()

	if err := suTpl.ExecuteTemplate(w, "su", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleAdminAddUser creates a user row by email (they link on Google login).
func handleAdminAddUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	if !strings.Contains(email, "@") {
		http.Error(w, "invalid email", http.StatusBadRequest)
		return
	}
	id := "u_" + randomHex(12)
	if _, err := db.Exec(`INSERT OR IGNORE INTO users (id, email, name, role) VALUES (?,?,?, 'user')`, id, email, email); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/su", http.StatusSeeOther)
}

// handleAdminSetRole updates a user's role (user/admin).
func handleAdminSetRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	uid := r.FormValue("user_id")
	role := r.FormValue("role")
	if role != "user" && role != "admin" {
		http.Error(w, "invalid role", http.StatusBadRequest)
		return
	}
	// An admin who demotes their own row is locked out permanently: only
	// SUPER_ADMIN_EMAIL is re-promoted at startup, so recovery means editing the
	// database by hand.
	if claims, err := currentUser(r); err == nil && claims.Sub == uid && role != "admin" {
		http.Error(w, "you cannot remove your own admin role", http.StatusBadRequest)
		return
	}
	db.Exec(`UPDATE users SET role=? WHERE id=?`, role, uid)
	http.Redirect(w, r, "/su", http.StatusSeeOther)
}

// handleAdminPentest triggers a pentest for a verified domain. Supports
// mode=standard (1 credit) and mode=destructive (2 credits).
func handleAdminPentest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	mode := strings.ToLower(strings.TrimSpace(r.FormValue("mode")))
	if mode == "" {
		mode = "standard"
	}
	if mode != "standard" && mode != "destructive" {
		http.Error(w, "invalid mode (standard|destructive)", http.StatusBadRequest)
		return
	}
	cost := 1.0
	if mode == "destructive" {
		cost = 2.0
	}
	var domainID int64
	var userID, status string
	fmt.Sscanf(r.FormValue("domain_id"), "%d", &domainID)
	if domainID == 0 {
		http.Error(w, "domain_id required", http.StatusBadRequest)
		return
	}
	err := db.QueryRow(`SELECT user_id, status FROM domains WHERE id=?`, domainID).Scan(&userID, &status)
	if err != nil {
		http.Error(w, "domain not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if status != "verified" {
		suErr(w, "Domain must be verified first.")
		return
	}
	// Consume credits atomically.
	res, err := db.Exec(`UPDATE credit_balance SET balance = balance - ?, updated_at=? WHERE user_id=? AND balance >= ?`, cost, now(), userID, cost)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		suErr(w, fmt.Sprintf("Insufficient credits — this scan costs %.0f.", cost))
		return
	}
	desc := "Pentest scheduled"
	if mode == "destructive" {
		desc = "Destructive pentest scheduled"
	}
	db.Exec(`INSERT INTO credit_transactions (user_id, type, amount, description, reference_id) VALUES (?,'spend',?,?,?)`, userID, cost, desc, fmt.Sprintf("pentest-%d", domainID))
	pid := "pt_" + randomHex(8)
	db.Exec(`INSERT INTO pentests (id, user_id, domain_id, mode, status) VALUES (?,?,?,?,'queued')`, pid, userID, domainID, mode)
	if mode == "destructive" {
		fmt.Fprintf(w, `<div class="rounded bg-red-500/15 px-3 py-2 text-xs text-red-300">Destructive pentest scheduled (%s) — 2 credits used.</div>`, pid)
	} else {
		fmt.Fprintf(w, `<div class="rounded bg-emerald-500/15 px-3 py-2 text-xs text-emerald-300">Pentest scheduled (%s) — 1 credit used.</div>`, pid)
	}
}

// suErr writes an error as a styled fragment. These handlers target a table cell,
// so http.Error's bare plain text used to land unstyled in the middle of a table.
func suErr(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<span class="rounded bg-red-500/15 px-2 py-0.5 text-[11px] text-red-300">%s</span>`,
		template.HTMLEscapeString(msg))
}

// handleAdminAddCredit adds credit to a user's balance (super admin).
func handleAdminAddCredit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	uid := r.FormValue("user_id")
	amount := parseFloat(r.FormValue("amount"))
	if uid == "" || amount <= 0 {
		http.Error(w, "user_id and positive amount required", http.StatusBadRequest)
		return
	}
	var exists string
	if err := db.QueryRow(`SELECT id FROM users WHERE id=?`, uid).Scan(&exists); err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	// Atomically credit the balance (insert-or-create) + log a transaction.
	if _, err := db.Exec(`INSERT INTO credit_balance (user_id, balance, updated_at) VALUES (?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET balance = balance + excluded.balance, updated_at = excluded.updated_at`,
		uid, amount, now()); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	db.Exec(`INSERT INTO credit_transactions (user_id, type, amount, description, reference_id) VALUES (?, 'admin_credit', ?, 'Admin credit', ?)`,
		uid, amount, "admin-"+now())
	http.Redirect(w, r, "/su", http.StatusSeeOther)
}

// handleAdminUploadReport uploads a pentest result PDF.
func handleAdminUploadReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	pid := r.FormValue("pentest_id")
	if pid == "" {
		http.Error(w, "pentest_id required", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file required", http.StatusBadRequest)
		return
	}
	defer file.Close()
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".pdf") {
		http.Error(w, "only PDF allowed", http.StatusBadRequest)
		return
	}
	name := pid + ".pdf"
	dst, err := os.Create(filepath.Join(reportsDir(), name))
	if err != nil {
		http.Error(w, "cannot write report", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "cannot write report", http.StatusInternalServerError)
		return
	}
	db.Exec(`UPDATE pentests SET report_ref=?, status='completed' WHERE id=?`, name, pid)
	http.Redirect(w, r, "/su", http.StatusSeeOther)
}

// handleReport serves a pentest report PDF to the owning user or an admin.
func handleReport(w http.ResponseWriter, r *http.Request) {
	// Never let Cloudflare (or browsers) cache reports — otherwise a cached
	// copy can be served publicly without hitting the origin auth check.
	w.Header().Set("Cache-Control", "no-store, private")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	name := strings.TrimPrefix(r.URL.Path, "/reports/")
	if !strings.HasSuffix(strings.ToLower(name), ".pdf") {
		http.NotFound(w, r)
		return
	}
	// Find the pentest whose report_ref matches; resolve its owner.
	var ownerID string
	err := db.QueryRow(`SELECT p.user_id FROM pentests p WHERE p.report_ref=?`, name).Scan(&ownerID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	// Allow if authenticated and either the owner or an admin.
	claims, err := currentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if claims.Sub != ownerID {
		var email string
		db.QueryRow(`SELECT email FROM users WHERE id=?`, claims.Sub).Scan(&email)
		if !isAdmin(email) {
			http.NotFound(w, r)
			return
		}
	}
	http.ServeFile(w, r, filepath.Join(reportsDir(), name))
}

var suTpl = template.Must(template.New("su").Funcs(template.FuncMap{"cssHash": func() string { return cssHash }}).Parse(suHTML))

const suHTML = `{{define "su"}}<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
<title>super admin — netsekurity</title>
<meta name="robots" content="noindex, nofollow"/>
<link rel="stylesheet" href="/css/styles.css?v={{cssHash}}"/>
<script src="https://unpkg.com/htmx.org@1.9.12/dist/htmx.min.js" defer></script>
</head>
<body class="scanlines bg-ink text-gray-300 min-h-screen">
<header class="border-b border-yellow-500/30 bg-ink/85">
  <div class="mx-auto flex max-w-6xl items-center justify-between px-4 py-3">
    <a href="/" class="font-mono text-base font-bold text-white"><span class="text-emerald-400">net</span>sekurity<span class="text-emerald-500">.com</span> <span class="text-yellow-300">/su</span></a>
    <div class="flex items-center gap-3 font-mono text-xs">
      <a href="/dashboard" class="prompt text-gray-500 hover:text-white">dashboard</a>
      <a href="/logout" class="prompt text-gray-500 hover:text-white">logout</a>
    </div>
  </div>
</header>
<main id="main" class="mx-auto max-w-6xl px-4 py-5 sm:px-6 space-y-4">

  <!-- At-a-glance totals: the reason to open this page -->
  <section class="grid grid-cols-2 gap-2 sm:grid-cols-3 lg:grid-cols-6">
    <div class="rounded border border-white/10 bg-[#04060c] px-3 py-2">
      <div class="font-mono text-[10px] uppercase tracking-wide text-gray-500">users</div>
      <div class="font-mono text-xl font-bold text-white">{{.Stats.Users}}</div>
    </div>
    <div class="rounded border {{if .Stats.InvoicesPaid}}border-emerald-500/30{{else}}border-red-500/40{{end}} bg-[#04060c] px-3 py-2">
      <div class="font-mono text-[10px] uppercase tracking-wide text-gray-500">invoices paid</div>
      <div class="font-mono text-xl font-bold {{if .Stats.InvoicesPaid}}text-emerald-300{{else}}text-red-300{{end}}">{{.Stats.InvoicesPaid}}</div>
      <div class="font-mono text-[10px] text-gray-600">{{.Stats.InvoicesPending}} pending · {{.Stats.InvoicesExpired}} expired</div>
    </div>
    <div class="rounded border border-white/10 bg-[#04060c] px-3 py-2">
      <div class="font-mono text-[10px] uppercase tracking-wide text-gray-500">revenue</div>
      <div class="font-mono text-xl font-bold text-emerald-300">${{printf "%.0f" .Stats.RevenueUSD}}</div>
    </div>
    <div class="rounded border border-white/10 bg-[#04060c] px-3 py-2">
      <div class="font-mono text-[10px] uppercase tracking-wide text-gray-500">credits out</div>
      <div class="font-mono text-xl font-bold text-yellow-300">{{printf "%.1f" .Stats.CreditsOut}}</div>
    </div>
    <div class="rounded border border-white/10 bg-[#04060c] px-3 py-2">
      <div class="font-mono text-[10px] uppercase tracking-wide text-gray-500">domains pending</div>
      <div class="font-mono text-xl font-bold text-cyan-300">{{.Stats.DomainsPending}}</div>
    </div>
    <div class="rounded border border-white/10 bg-[#04060c] px-3 py-2">
      <div class="font-mono text-[10px] uppercase tracking-wide text-gray-500">scans active</div>
      <div class="font-mono text-xl font-bold text-cyan-300">{{.Stats.ScansActive}}</div>
    </div>
  </section>

  <div class="flex items-center gap-2">
    <input id="su-search" type="search" placeholder="filter every table — email, domain, invoice…" aria-label="Filter rows"
      oninput="suFilter(this.value)"
      class="w-full rounded border border-white/15 bg-ink px-3 py-2 font-mono text-xs text-white focus:border-emerald-400 focus:outline-none"/>
    <span id="su-search-count" class="whitespace-nowrap font-mono text-[11px] text-gray-500"></span>
  </div>

  <!-- Users -->
  <section class="rounded border border-emerald-500/30 bg-[#04060c]">
    <div class="flex items-center justify-between border-b border-emerald-500/20 px-3 py-2">
      <div class="flex items-center gap-1.5">
        <span class="h-2.5 w-2.5 rounded-full bg-emerald-500/70"></span>
        <span class="font-mono text-[11px] text-gray-500">$ users</span>
      </div>
      <form class="flex gap-2" method="post" action="/su/users/add">
        <input name="email" required placeholder="new@email.com" aria-label="New user email" class="rounded border border-white/15 bg-ink px-2 py-1 font-mono text-xs text-white focus:border-emerald-400 focus:outline-none"/>
        <button class="rounded border border-cyan-400 bg-cyan-500/10 px-2 py-1 font-mono text-xs font-bold text-cyan-300 hover:bg-cyan-500/20">add</button>
      </form>
    </div>
    <div class="overflow-x-auto p-3">
      <table class="w-full font-mono text-xs">
        <thead class="sticky top-0 z-10 bg-[#04060c]"><tr class="text-left text-gray-500"><th class="py-1">email</th><th>name</th><th>role</th><th>credits</th><th>set role</th><th>add credit</th></tr></thead>
        <tbody>
        {{range .Users}}
        <tr class="border-t border-white/10">
          <td class="py-1 text-white">{{.Email}}</td>
          <td>{{.Name}}</td>
          <td>{{if eq .Role "admin"}}<span class="text-yellow-300">admin</span>{{else}}<span class="text-gray-400">user</span>{{end}}</td>
          <td class="text-emerald-300">{{printf "%.1f" .Credits}}</td>
          <td>
            {{if eq .ID $.CurrentID}}
            <span class="font-mono text-[11px] text-gray-600" title="You cannot change your own role.">— you —</span>
            {{else}}
            <form method="post" action="/su/users/role" class="flex gap-1"
              onsubmit="return confirm('Set {{.Email}} to ' + this.role.value + '?')">
              <input type="hidden" name="user_id" value="{{.ID}}"/>
              <select name="role" aria-label="Role" class="rounded border border-white/15 bg-ink px-1 py-0.5 font-mono text-[11px] text-white">
                <option value="user" {{if eq .Role "user"}}selected{{end}}>user</option>
                <option value="admin" {{if eq .Role "admin"}}selected{{end}}>admin</option>
              </select>
              <button class="rounded border border-emerald-400/60 px-2 py-0.5 text-[11px] font-bold text-emerald-300 hover:bg-emerald-500/15">set</button>
            </form>
            {{end}}
          </td>
          <td>
            <form method="post" action="/su/users/credit" class="flex gap-1"
              onsubmit="return confirm('Grant ' + this.amount.value + ' credit(s) to {{.Email}}?\n\nCredits are the product currency. This is immediate and has no undo.')">
              <input type="hidden" name="user_id" value="{{.ID}}"/>
              <input name="amount" type="number" step="0.5" min="0.5" required placeholder="amt" aria-label="Credit amount" class="w-16 rounded border border-white/15 bg-ink px-1 py-0.5 font-mono text-[11px] text-white"/>
              <button class="rounded border border-cyan-400/60 px-2 py-0.5 text-[11px] font-bold text-cyan-300 hover:bg-cyan-500/15">+credit</button>
            </form>
          </td>
        </tr>
        {{end}}
        </tbody>
      </table>
    </div>
  </section>

  <!-- Transactions -->
  <section class="rounded border border-white/10 bg-[#04060c]">
    <div class="flex items-center gap-1.5 border-b border-white/10 px-3 py-2">
      <span class="h-2.5 w-2.5 rounded-full bg-red-500/70"></span>
      <span class="font-mono text-[11px] text-gray-500">$ transactions</span>
    </div>
    <div class="overflow-x-auto p-3">
      <table class="w-full font-mono text-xs">
        <thead class="sticky top-0 z-10 bg-[#04060c]"><tr class="text-left text-gray-500">
          <th class="py-1">external_id</th><th>package</th><th>amount</th><th>credits</th><th>status</th><th>created</th><th>paid</th>
        </tr></thead>
        <tbody>
        {{range .Trx}}
        <tr class="border-t border-white/10">
          <td class="py-1 text-gray-300">{{.ExternalID}}</td>
          <td>{{.Package}}</td>
          <td>${{printf "%.0f" .Amount}}</td>
          <td>{{printf "%.0f" .Credits}}</td>
          <td>
            {{if eq .Status "paid"}}<span class="text-emerald-300">[paid]</span>
            {{else if eq .Status "expired"}}<span class="text-red-300">[expired]</span>
            {{else}}<span class="text-yellow-300">[{{.Status}}]</span>{{end}}
          </td>
          <td class="text-gray-500">{{.CreatedAt}}</td>
          <td class="text-gray-500">{{.PaidAt}}</td>
        </tr>
        {{end}}
        </tbody>
      </table>
    </div>
  </section>

  <!-- Domains -->
  <section class="rounded border border-cyan-500/30 bg-[#04060c]">
    <div class="flex items-center gap-1.5 border-b border-cyan-500/20 px-3 py-2">
      <span class="h-2.5 w-2.5 rounded-full bg-cyan-500/70"></span>
      <span class="font-mono text-[11px] text-gray-500">$ websites (domains)</span>
    </div>
    <div class="overflow-x-auto p-3">
      <table class="w-full font-mono text-xs">
        <thead class="sticky top-0 z-10 bg-[#04060c]"><tr class="text-left text-gray-500"><th class="py-1">user</th><th>domain</th><th>status</th><th>pentest</th></tr></thead>
        <tbody>
        {{range .Domains}}
        <tr class="border-t border-white/10">
          <td class="py-1 text-gray-300">{{.UserEmail}}</td>
          <td class="text-white">{{.Domain}}</td>
          <td>{{if eq .Status "verified"}}<span class="text-emerald-300">[verified]</span>{{else}}<span class="text-yellow-300">[{{.Status}}]</span>{{end}}</td>
          <td data-pentest-cell="{{.ID}}">
            {{if eq .Status "verified"}}
            <div class="flex gap-1">
            <button hx-post="/su/domains/pentest" hx-vals='{"domain_id":"{{.ID}}","mode":"standard"}' hx-target="closest td" hx-swap="innerHTML"
              title="Standard scan — 1 credit (read-only)"
              class="rounded border border-emerald-400 bg-emerald-500/10 px-2 py-0.5 text-[11px] font-bold text-emerald-300 hover:bg-emerald-500/20">run</button>
            <button type="button" onclick="suOpenDestructive('{{.ID}}','{{.Domain}}','{{.UserEmail}}')"
              title="Destructive scan — 2 of this customer's credits (exploit/RCE/webshell/takeover)."
              class="rounded border border-red-400/70 bg-red-500/10 px-2 py-0.5 text-[11px] font-bold text-red-300 hover:bg-red-500/20">run -x</button>
            </div>
            {{else}}<span class="text-gray-600">—</span>{{end}}
          </td>
        </tr>
        {{end}}
        </tbody>
      </table>
    </div>
  </section>

  <!-- Pentests / reports -->
  <section class="rounded border border-white/10 bg-[#04060c]">
    <div class="flex items-center gap-1.5 border-b border-white/10 px-3 py-2">
      <span class="h-2.5 w-2.5 rounded-full bg-yellow-500/70"></span>
      <span class="font-mono text-[11px] text-gray-500">$ pentests &amp; reports</span>
    </div>
    <div class="overflow-x-auto p-3">
      <table class="w-full font-mono text-xs">
        <thead class="sticky top-0 z-10 bg-[#04060c]"><tr class="text-left text-gray-500">
          <th class="py-1">id</th><th>user</th><th>domain</th><th>mode</th><th>status</th><th>report</th><th>upload pdf</th>
        </tr></thead>
        <tbody>
        {{range .Pentests}}
        <tr class="border-t border-white/10">
          <td class="py-1 text-gray-300">{{.ID}}</td>
          <td>{{.UserEmail}}</td>
          <td class="text-white">{{.Domain}}</td>
          <td>{{if eq .Mode "destructive"}}<span class="text-red-300 font-bold">destructive</span>{{else}}<span class="text-cyan-300">{{.Mode}}</span>{{end}}</td>
          <td>{{.Status}}</td>
          <td>{{if .ReportRef}}<a class="text-emerald-300 underline" href="/reports/{{.ReportRef}}">view</a>{{else}}<span class="text-gray-600">—</span>{{end}}</td>
          <td>
            <details>
              <summary class="cursor-pointer font-mono text-[11px] text-cyan-400 hover:underline">upload</summary>
              <form method="post" action="/su/reports/upload" enctype="multipart/form-data" class="mt-1 flex gap-1">
                <input type="hidden" name="pentest_id" value="{{.ID}}"/>
                <input type="file" name="file" accept=".pdf" required aria-label="Report PDF" class="font-mono text-[11px] text-gray-500"/>
                <button class="rounded border border-cyan-400/60 px-2 py-0.5 text-[11px] font-bold text-cyan-300 hover:bg-cyan-500/15">up</button>
              </form>
            </details>
          </td>
        </tr>
        {{end}}
        </tbody>
      </table>
    </div>
  </section>
  <!-- Manual credit grants: the audit trail for the +credit control above -->
  {{if .Credits}}
  <section class="rounded border border-white/10 bg-[#04060c]">
    <div class="flex items-center gap-1.5 border-b border-white/10 px-3 py-2">
      <span class="h-2.5 w-2.5 rounded-full bg-cyan-500/70"></span>
      <span class="font-mono text-[11px] text-gray-500">$ manual credit grants</span>
    </div>
    <div class="overflow-x-auto p-3">
      <table class="w-full font-mono text-xs">
        <thead class="sticky top-0 z-10 bg-[#04060c]"><tr class="text-left text-gray-500"><th class="py-1">user</th><th>amount</th><th>note</th><th>when</th></tr></thead>
        <tbody>
        {{range .Credits}}
        <tr class="border-t border-white/10">
          <td class="py-1 text-gray-300">{{.Email}}</td>
          <td class="text-yellow-300">+{{.Amount}}</td>
          <td class="text-gray-500">{{.Description}}</td>
          <td class="text-gray-500">{{.CreatedAt}}</td>
        </tr>
        {{end}}
        </tbody>
      </table>
    </div>
  </section>
  {{end}}
</main>

<!-- Destructive scan confirm — same typed phrase as the customer dashboard -->
<div id="su-destructive" class="fixed inset-0 z-[100] hidden items-center justify-center bg-black/70 p-4">
  <div class="w-full max-w-lg rounded-lg border border-red-500/40 bg-[#0a0d16] p-5">
    <div class="mb-2 font-mono text-sm font-bold text-red-300">⚠ DESTRUCTIVE PENTEST — CUSTOMER DOMAIN</div>
    <div class="space-y-2 font-mono text-xs leading-relaxed text-gray-300">
      <p>Active exploitation (<b>RCE, webshell upload, takeover attempts</b>) against a domain you do not own. It can modify, corrupt or take down the target.</p>
      <p id="su-destructive-target" class="text-yellow-300"></p>
      <label class="mt-2 block">
        <span class="text-gray-400">Type <b class="text-red-300">AGREE AND PROCEED</b> to confirm:</span>
        <input id="su-destructive-agree" type="text" autocomplete="off"
          class="mt-1 w-full rounded border border-red-500/40 bg-black px-3 py-2 font-mono text-sm text-white focus:border-red-400 focus:outline-none"
          placeholder="AGREE AND PROCEED"/>
      </label>
    </div>
    <div class="mt-4 flex items-center justify-end gap-2">
      <button type="button" onclick="suCloseDestructive()" class="rounded border border-white/20 px-3 py-1.5 font-mono text-xs text-gray-400 hover:bg-white/5">cancel</button>
      <button type="button" id="su-destructive-go" onclick="suRunDestructive()" disabled
        class="rounded border border-red-500 bg-red-600/20 px-3 py-1.5 font-mono text-xs font-bold text-red-200 hover:bg-red-600/30 disabled:cursor-not-allowed disabled:opacity-40">run destructive →</button>
    </div>
  </div>
</div>

<script>
// Client-side filter across every table on the page.
function suFilter(q) {
  q = (q || '').trim().toLowerCase();
  var shown = 0, total = 0;
  document.querySelectorAll('main table tbody tr').forEach(function (tr) {
    total++;
    var hit = !q || tr.textContent.toLowerCase().indexOf(q) !== -1;
    tr.style.display = hit ? '' : 'none';
    if (hit) shown++;
  });
  document.getElementById('su-search-count').textContent = q ? (shown + ' / ' + total + ' rows') : '';
}

var suTargetID = null;
function suOpenDestructive(id, domain, email) {
  suTargetID = id;
  document.getElementById('su-destructive-target').textContent =
    'target: ' + domain + ' · owner: ' + email + ' · spends 2 of their credits';
  document.getElementById('su-destructive-agree').value = '';
  document.getElementById('su-destructive-go').disabled = true;
  var m = document.getElementById('su-destructive');
  m.classList.remove('hidden'); m.classList.add('flex');
}
function suCloseDestructive() {
  var m = document.getElementById('su-destructive');
  m.classList.add('hidden'); m.classList.remove('flex');
  suTargetID = null;
}
document.getElementById('su-destructive-agree').addEventListener('input', function () {
  document.getElementById('su-destructive-go').disabled = (this.value.trim() !== 'AGREE AND PROCEED');
});
function suRunDestructive() {
  if (!suTargetID) return;
  var btn = document.getElementById('su-destructive-go');
  btn.disabled = true; btn.textContent = 'running…';
  var xhr = new XMLHttpRequest();
  xhr.open('POST', '/su/domains/pentest');
  xhr.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
  xhr.onload = function () {
    var cell = document.querySelector('[data-pentest-cell="' + suTargetID + '"]');
    if (cell) cell.innerHTML = xhr.responseText;
    suCloseDestructive();
    btn.textContent = 'run destructive →';
  };
  xhr.send('domain_id=' + encodeURIComponent(suTargetID) + '&mode=destructive');
}
</script>
</body>
</html>{{end}}`

var _ = time.Now
