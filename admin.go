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
	ID int64
	UserEmail, Domain, Status, TXT, CreatedAt string
}

type suPentest struct {
	ID, UserEmail, Domain, Status, ReportRef, CreatedAt string
}

type suData struct {
	Users    []suUser
	Trx      []suTrx
	Domains  []suDomain
	Pentests []suPentest
}

// reportsDir returns the directory where pentest PDF reports are stored.
func reportsDir() string {
	path := getenv("DATABASE_PATH", "./data/netsekurity.db")
	dir := filepath.Dir(path)
	return filepath.Join(dir, "reports")
}

// handleAdmin renders the super-admin dashboard.
func handleAdmin(w http.ResponseWriter, r *http.Request) {
	// Mark overdue pending invoices as expired (Xendit invoice_duration = 1 day).
	db.Exec(`UPDATE payments SET status='expired' WHERE status='pending' AND created_at < datetime('now','-1 day')`)

	data := suData{Users: []suUser{}, Trx: []suTrx{}, Domains: []suDomain{}, Pentests: []suPentest{}}

	rows, _ := db.Query(`SELECT u.id, u.email, COALESCE(u.name,''), COALESCE(u.role,'user'), COALESCE(cb.balance,0)
		FROM users u LEFT JOIN credit_balance cb ON cb.user_id=u.id ORDER BY u.created_at DESC`)
	for rows.Next() {
		var u suUser
		rows.Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.Credits)
		data.Users = append(data.Users, u)
	}
	rows.Close()

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

	prows, _ := db.Query(`SELECT p.id, COALESCE(u.email,''), COALESCE(d.domain,''), p.status, COALESCE(p.report_ref,''), p.created_at
		FROM pentests p LEFT JOIN users u ON u.id=p.user_id LEFT JOIN domains d ON d.id=p.domain_id
		ORDER BY p.created_at DESC LIMIT 100`)
	for prows.Next() {
		var p suPentest
		prows.Scan(&p.ID, &p.UserEmail, &p.Domain, &p.Status, &p.ReportRef, &p.CreatedAt)
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
	db.Exec(`UPDATE users SET role=? WHERE id=?`, role, uid)
	http.Redirect(w, r, "/su", http.StatusSeeOther)
}

// handleAdminPentest triggers a pentest for a verified domain, consuming 1 credit.
func handleAdminPentest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
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
		http.Error(w, "domain must be verified first", http.StatusBadRequest)
		return
	}
	// Consume one credit atomically.
	res, err := db.Exec(`UPDATE credit_balance SET balance = balance - 1, updated_at=? WHERE user_id=? AND balance >= 1`, now(), userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, "insufficient credits", http.StatusPaymentRequired)
		return
	}
	db.Exec(`INSERT INTO credit_transactions (user_id, type, amount, description, reference_id) VALUES (?,'spend',1,'Pentest scheduled',?)`, userID, fmt.Sprintf("pentest-%d", domainID))
	pid := "pt_" + randomHex(8)
	db.Exec(`INSERT INTO pentests (id, user_id, domain_id, status) VALUES (?,?,?,'queued')`, pid, userID, domainID)
	fmt.Fprintf(w, `<div class="rounded bg-emerald-500/15 px-3 py-2 text-xs text-emerald-300">Pentest scheduled (%s) — 1 credit used.</div>`, pid)
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

// handleReport serves a pentest report PDF.
func handleReport(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/reports/")
	if !strings.HasSuffix(strings.ToLower(name), ".pdf") {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, filepath.Join(reportsDir(), name))
}

var suTpl = template.Must(template.New("su").Parse(suHTML))

const suHTML = `{{define "su"}}<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
<title>Super Admin — netsekurity</title>
<link rel="stylesheet" href="/css/styles.css"/>
<script src="https://unpkg.com/htmx.org@1.9.12/dist/htmx.min.js" defer></script>
</head>
<body class="bg-ink text-gray-200 min-h-screen">
<header class="border-b border-white/10 bg-ink/80">
  <div class="mx-auto flex max-w-6xl items-center justify-between px-4 py-3">
    <a href="/" class="font-mono text-base font-bold text-white"><span class="text-emerald-400">net</span>sekurity <span class="text-xs text-yellow-300">/ su</span></a>
    <div class="flex items-center gap-3 text-xs">
      <a href="/dashboard" class="text-gray-500 hover:text-white">Dashboard</a>
      <a href="/logout" class="text-gray-500 hover:text-white">Logout</a>
    </div>
  </div>
</header>
<main class="mx-auto max-w-6xl px-4 py-5 sm:px-6 space-y-5">
  <!-- Users -->
  <section class="rounded-lg border border-white/10 bg-panel p-3">
    <div class="flex items-center justify-between">
      <h2 class="text-sm font-bold text-white">Users</h2>
      <form class="flex gap-2" method="post" action="/su/users/add">
        <input name="email" required placeholder="new@email.com" class="rounded border border-white/10 bg-ink px-2 py-1 text-xs text-white"/>
        <button class="rounded bg-cyan-500 px-3 py-1 text-xs font-bold text-black">Add user</button>
      </form>
    </div>
    <table class="mt-2 w-full text-xs">
      <thead><tr class="text-left text-gray-500">
        <th class="py-1">Email</th><th>Name</th><th>Role</th><th>Credits</th><th>Action</th>
      </tr></thead>
      <tbody>
      {{range .Users}}
      <tr class="border-t border-white/10">
        <td class="py-1 font-mono">{{.Email}}</td>
        <td>{{.Name}}</td>
        <td>{{.Role}}</td>
        <td class="font-mono">{{printf "%.1f" .Credits}}</td>
        <td>
          <form method="post" action="/su/users/role" class="flex gap-1">
            <input type="hidden" name="user_id" value="{{.ID}}"/>
            <select name="role" class="rounded border border-white/10 bg-ink px-1 py-0.5 text-[11px] text-white">
              <option value="user" {{if eq .Role "user"}}selected{{end}}>user</option>
              <option value="admin" {{if eq .Role "admin"}}selected{{end}}>admin</option>
            </select>
            <button class="rounded bg-emerald-500/20 px-2 py-0.5 font-bold text-emerald-300">Save</button>
          </form>
        </td>
      </tr>
      {{end}}
      </tbody>
    </table>
  </section>

  <!-- Transactions -->
  <section class="rounded-lg border border-white/10 bg-panel p-3">
    <h2 class="text-sm font-bold text-white">Transactions</h2>
    <table class="mt-2 w-full text-xs">
      <thead><tr class="text-left text-gray-500">
        <th class="py-1">External ID</th><th>Package</th><th>Amount</th><th>Credits</th><th>Status</th><th>Created</th><th>Paid</th>
      </tr></thead>
      <tbody>
      {{range .Trx}}
      <tr class="border-t border-white/10">
        <td class="py-1 font-mono">{{.ExternalID}}</td>
        <td>{{.Package}}</td>
        <td class="font-mono">${{printf "%.0f" .Amount}}</td>
        <td class="font-mono">{{printf "%.0f" .Credits}}</td>
        <td><span class="rounded px-1.5 py-0.5 {{if eq .Status "paid"}}bg-emerald-500/20 text-emerald-300{{else if eq .Status "expired"}}bg-red-500/20 text-red-300{{else}}bg-yellow-500/20 text-yellow-300{{end}}">{{.Status}}</span></td>
        <td class="text-gray-500">{{.CreatedAt}}</td>
        <td class="text-gray-500">{{.PaidAt}}</td>
      </tr>
      {{end}}
      </tbody>
    </table>
  </section>

  <!-- Domains + pentest -->
  <section class="rounded-lg border border-white/10 bg-panel p-3">
    <h2 class="text-sm font-bold text-white">Websites (domains)</h2>
    <table class="mt-2 w-full text-xs">
      <thead><tr class="text-left text-gray-500">
        <th class="py-1">User</th><th>Domain</th><th>Status</th><th>Pentest</th>
      </tr></thead>
      <tbody>
      {{range .Domains}}
      <tr class="border-t border-white/10">
        <td class="py-1 font-mono">{{.UserEmail}}</td>
        <td class="font-mono">{{.Domain}}</td>
        <td><span class="rounded px-1.5 py-0.5 {{if eq .Status "verified"}}bg-emerald-500/20 text-emerald-300{{else}}bg-yellow-500/20 text-yellow-300{{end}}">{{.Status}}</span></td>
        <td>
          {{if eq .Status "verified"}}
          <button hx-post="/su/domains/pentest" hx-vals='{"domain_id":"{{.ID}}"}' hx-target="closest td" hx-swap="innerHTML"
            class="rounded bg-emerald-500 px-2 py-0.5 font-bold text-black">Pentest</button>
          {{else}}<span class="text-gray-600">—</span>{{end}}
        </td>
      </tr>
      {{end}}
      </tbody>
    </table>
  </section>

  <!-- Pentests / reports -->
  <section class="rounded-lg border border-white/10 bg-panel p-3">
    <h2 class="text-sm font-bold text-white">Pentests &amp; reports</h2>
    <table class="mt-2 w-full text-xs">
      <thead><tr class="text-left text-gray-500">
        <th class="py-1">ID</th><th>User</th><th>Domain</th><th>Status</th><th>Report PDF</th><th>Upload</th>
      </tr></thead>
      <tbody>
      {{range .Pentests}}
      <tr class="border-t border-white/10">
        <td class="py-1 font-mono">{{.ID}}</td>
        <td class="font-mono">{{.UserEmail}}</td>
        <td class="font-mono">{{.Domain}}</td>
        <td>{{.Status}}</td>
        <td>
          {{if .ReportRef}}<a class="text-emerald-300 underline" href="/reports/{{.ReportRef}}">view PDF</a>{{else}}<span class="text-gray-600">—</span>{{end}}
        </td>
        <td>
          <form method="post" action="/su/reports/upload" enctype="multipart/form-data" class="flex gap-1">
            <input type="hidden" name="pentest_id" value="{{.ID}}"/>
            <input type="file" name="file" accept=".pdf" class="text-[11px] text-gray-500"/>
            <button class="rounded bg-cyan-500/20 px-2 py-0.5 font-bold text-cyan-300">Upload</button>
          </form>
        </td>
      </tr>
      {{end}}
      </tbody>
    </table>
  </section>
</main>
</body>
</html>{{end}}`

var _ = time.Now
