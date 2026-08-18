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


var suTpl = template.Must(template.New("su").Parse(suHTML))

const suHTML = `{{define "su"}}<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
<title>super admin — netsekurity</title>
<meta name="robots" content="noindex, nofollow"/>
<link rel="stylesheet" href="/css/styles.css"/>
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
<main class="mx-auto max-w-6xl px-4 py-5 sm:px-6 space-y-4">

  <!-- Users -->
  <section class="rounded border border-emerald-500/30 bg-[#04060c]">
    <div class="flex items-center justify-between border-b border-emerald-500/20 px-3 py-2">
      <div class="flex items-center gap-1.5">
        <span class="h-2.5 w-2.5 rounded-full bg-emerald-500/70"></span>
        <span class="font-mono text-[11px] text-gray-500">$ users</span>
      </div>
      <form class="flex gap-2" method="post" action="/su/users/add">
        <input name="email" required placeholder="new@email.com" class="rounded border border-white/15 bg-ink px-2 py-1 font-mono text-xs text-white focus:border-emerald-400 focus:outline-none"/>
        <button class="rounded border border-cyan-400 bg-cyan-500/10 px-2 py-1 font-mono text-xs font-bold text-cyan-300 hover:bg-cyan-500/20">add</button>
      </form>
    </div>
    <div class="overflow-x-auto p-3">
      <table class="w-full font-mono text-xs">
        <thead><tr class="text-left text-gray-500"><th class="py-1">email</th><th>name</th><th>role</th><th>credits</th><th>set role</th><th>add credit</th></tr></thead>
        <tbody>
        {{range .Users}}
        <tr class="border-t border-white/10">
          <td class="py-1 text-white">{{.Email}}</td>
          <td>{{.Name}}</td>
          <td>{{if eq .Role "admin"}}<span class="text-yellow-300">admin</span>{{else}}<span class="text-gray-400">user</span>{{end}}</td>
          <td class="text-emerald-300">{{printf "%.1f" .Credits}}</td>
          <td>
            <form method="post" action="/su/users/role" class="flex gap-1">
              <input type="hidden" name="user_id" value="{{.ID}}"/>
              <select name="role" class="rounded border border-white/15 bg-ink px-1 py-0.5 font-mono text-[11px] text-white">
                <option value="user" {{if eq .Role "user"}}selected{{end}}>user</option>
                <option value="admin" {{if eq .Role "admin"}}selected{{end}}>admin</option>
              </select>
              <button class="rounded border border-emerald-400/60 px-2 py-0.5 text-[10px] font-bold text-emerald-300 hover:bg-emerald-500/15">set</button>
            </form>
          </td>
          <td>
            <form method="post" action="/su/users/credit" class="flex gap-1">
              <input type="hidden" name="user_id" value="{{.ID}}"/>
              <input name="amount" type="number" step="0.5" min="0.5" placeholder="amt" class="w-16 rounded border border-white/15 bg-ink px-1 py-0.5 font-mono text-[11px] text-white"/>
              <button class="rounded border border-cyan-400/60 px-2 py-0.5 text-[10px] font-bold text-cyan-300 hover:bg-cyan-500/15">+credit</button>
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
        <thead><tr class="text-left text-gray-500">
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
        <thead><tr class="text-left text-gray-500"><th class="py-1">user</th><th>domain</th><th>status</th><th>pentest</th></tr></thead>
        <tbody>
        {{range .Domains}}
        <tr class="border-t border-white/10">
          <td class="py-1 text-gray-300">{{.UserEmail}}</td>
          <td class="text-white">{{.Domain}}</td>
          <td>{{if eq .Status "verified"}}<span class="text-emerald-300">[verified]</span>{{else}}<span class="text-yellow-300">[{{.Status}}]</span>{{end}}</td>
          <td>
            {{if eq .Status "verified"}}
            <button hx-post="/su/domains/pentest" hx-vals='{"domain_id":"{{.ID}}"}' hx-target="closest td" hx-swap="innerHTML"
              class="rounded border border-emerald-400 bg-emerald-500/10 px-2 py-0.5 text-[10px] font-bold text-emerald-300 hover:bg-emerald-500/20">run</button>
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
        <thead><tr class="text-left text-gray-500">
          <th class="py-1">id</th><th>user</th><th>domain</th><th>status</th><th>report</th><th>upload pdf</th>
        </tr></thead>
        <tbody>
        {{range .Pentests}}
        <tr class="border-t border-white/10">
          <td class="py-1 text-gray-300">{{.ID}}</td>
          <td>{{.UserEmail}}</td>
          <td class="text-white">{{.Domain}}</td>
          <td>{{.Status}}</td>
          <td>{{if .ReportRef}}<a class="text-emerald-300 underline" href="/reports/{{.ReportRef}}">view</a>{{else}}<span class="text-gray-600">—</span>{{end}}</td>
          <td>
            <form method="post" action="/su/reports/upload" enctype="multipart/form-data" class="flex gap-1">
              <input type="hidden" name="pentest_id" value="{{.ID}}"/>
              <input type="file" name="file" accept=".pdf" class="font-mono text-[10px] text-gray-500"/>
              <button class="rounded border border-cyan-400/60 px-2 py-0.5 text-[10px] font-bold text-cyan-300 hover:bg-cyan-500/15">up</button>
            </form>
          </td>
        </tr>
        {{end}}
        </tbody>
      </table>
    </div>
  </section>
</main>
</body>
</html>{{end}}`

var _ = time.Now
