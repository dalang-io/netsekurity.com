package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
)

// The customer dashboard is a handful of task-sized pages inside one shell
// (sidebar + top bar, the same layout as dalang.io/dashboard):
//
//	/dashboard             Home        — what to do next, balance, latest scans
//	/dashboard/websites    Websites    — add a domain, prove it is yours, scan it
//	/dashboard/scans       Scans       — start a scan, read the results
//	/dashboard/credits     Credits     — packages, payments, credit activity
//	/dashboard/automation  Automation  — API tokens for CI/CD (developers)
//
// Templates live in dashboard_html.go. Every page is rendered from the same
// dashboardData; loadDashboardData only fills what the page in question uses.

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

	// Home only.
	ScanCount     int   // every scan the account has ever started
	VerifiedCount int   // domains whose ownership is confirmed
	PendingPay    *pymt // an unpaid invoice that can still be finished
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

// pymt is an Xendit invoice shown in the payment history (incl. pending/expired).
type pymt struct {
	ExternalID, Package, Status, CreatedAt, URL string
	AmountUSD                                   float64
	Credits                                     float64
}

type dom struct {
	Domain, Status, TXT string
}

// step is one line of the Home page's getting-started checklist.
type step struct {
	N             int
	Title, Body   string
	CTA, Href     string
	Done, Current bool
}

func (d dashboardData) FirstName() string { return firstName(d.Name, d.Email) }
func (d dashboardData) Initial() string   { return initial(d.Name, d.Email) }

// Verified is the subset of domains a scan can be started for.
func (d dashboardData) Verified() []dom {
	var out []dom
	for _, x := range d.Domains {
		if x.Status == "verified" {
			out = append(out, x)
		}
	}
	return out
}

// Steps is the path from a new account to a first report. A step is done when
// the account's data says so, and exactly one — the first unfinished — is
// current, so the page always has a single obvious next action.
func (d dashboardData) Steps() []step {
	steps := []step{
		{N: 1, Title: "Add your website",
			Body: "Type the address of the website you want to check, like www.mybusiness.com.",
			CTA:  "Add a website", Href: "/dashboard/websites", Done: len(d.Domains) > 0},
		{N: 2, Title: "Prove it's yours",
			Body: "We only scan websites you own. You add one small record in your domain settings — we show you exactly how.",
			CTA:  "Verify my website", Href: "/dashboard/websites", Done: d.VerifiedCount > 0},
		{N: 3, Title: "Get a credit",
			Body: "Each scan uses 1 credit. Credits never expire.",
			CTA:  "Choose a credit pack", Href: "/dashboard/credits", Done: d.Balance > 0 || d.ScanCount > 0},
		{N: 4, Title: "Run your first scan",
			Body: "A scan takes 10 minutes to 24 hours, depending on what it finds. You get a PDF that explains what was found and what to fix.",
			CTA:  "Start a scan", Href: "/dashboard/scans", Done: d.ScanCount > 0},
	}
	for i := range steps {
		if !steps[i].Done {
			steps[i].Current = true
			break
		}
	}
	return steps
}

func (d dashboardData) StepsDone() int {
	n := 0
	for _, s := range d.Steps() {
		if s.Done {
			n++
		}
	}
	return n
}

func (d dashboardData) AllDone() bool { return d.StepsDone() == len(d.Steps()) }

// loadDomains reads one customer's domains, newest first.
func loadDomains(userID string) []dom {
	rows, err := db.Query(`SELECT domain, status, txt_verification_token FROM domains WHERE user_id=? ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []dom
	for rows.Next() {
		var d dom
		rows.Scan(&d.Domain, &d.Status, &d.TXT)
		out = append(out, d)
	}
	return out
}

// renderDomainListOOB writes the domain list, and the scan picker that mirrors
// it, as HTMX out-of-band swaps. Handlers that change a domain append it to
// their response so the page refreshes in place — the dashboard used to force a
// full location.reload() instead, which threw away the very message the user
// had just been shown. A target that is not on the current page is skipped by
// HTMX, so this is safe to send from every page.
func renderDomainListOOB(w io.Writer, userID string) {
	var d dashboardData
	db.QueryRow(`SELECT balance FROM credit_balance WHERE user_id=?`, userID).Scan(&d.Balance)
	d.Domains = loadDomains(userID)
	for _, part := range []struct{ tpl, id string }{
		{"domainlist", "domain-list"},
		{"scanpicker", "scan-picker"},
	} {
		var b strings.Builder
		if err := tmpl.ExecuteTemplate(&b, part.tpl, d); err != nil {
			log.Printf("%s render error: %v", part.tpl, err)
			continue
		}
		io.WriteString(w, strings.Replace(b.String(),
			`<div id="`+part.id+`">`, `<div id="`+part.id+`" hx-swap-oob="true">`, 1))
	}
}

// renderBalanceOOB writes the credit balance as an HTMX out-of-band swap so an
// action that spends credits updates the figure without a page reload.
func renderBalanceOOB(w io.Writer, userID string) {
	var balance float64
	db.QueryRow(`SELECT balance FROM credit_balance WHERE user_id=?`, userID).Scan(&balance)
	fmt.Fprintf(w, `<span id="credit-balance" hx-swap-oob="true" class="font-extrabold">%s</span>`, fmtCredits(balance))
}

func handleDashboard(w http.ResponseWriter, r *http.Request) { serveDashboardPage(w, r, "dashboard") }
func handleDashboardWebsites(w http.ResponseWriter, r *http.Request) {
	serveDashboardPage(w, r, "dashboard_websites")
}
func handleDashboardScans(w http.ResponseWriter, r *http.Request) {
	serveDashboardPage(w, r, "dashboard_scans")
}
func handleDashboardCredits(w http.ResponseWriter, r *http.Request) {
	serveDashboardPage(w, r, "dashboard_credits")
}
func handleDashboardAutomation(w http.ResponseWriter, r *http.Request) {
	serveDashboardPage(w, r, "dashboard_automation")
}

// serveDashboardPage renders one named dashboard template for the signed-in user.
func serveDashboardPage(w http.ResponseWriter, r *http.Request, page string) {
	w.Header().Set("Cache-Control", "no-store, private")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	claims, err := currentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	// Sync any newly-paid payments so credits appear immediately (no webhook
	// needed). This calls Xendit once per pending invoice, so it runs only where
	// a customer returning from payment lands or reads the balance — not on every
	// click through the sidebar.
	if page == "dashboard" || page == "dashboard_credits" {
		syncPendingPayments()
	}
	data := loadDashboardData(claims.Sub, page)
	if err := tmpl.ExecuteTemplate(w, page, data); err != nil {
		log.Printf("%s render error: %v", page, err)
	}
}

// loadDashboardData reads the account header (name, balance) plus whatever the
// given page shows.
func loadDashboardData(userID, page string) dashboardData {
	var d dashboardData
	db.QueryRow(`SELECT name, email FROM users WHERE id=?`, userID).Scan(&d.Name, &d.Email)
	db.QueryRow(`SELECT balance FROM credit_balance WHERE user_id=?`, userID).Scan(&d.Balance)
	d.IsAdmin = isAdmin(d.Email)

	switch page {
	case "dashboard":
		d.Domains = loadDomains(userID)
		d.VerifiedCount = len(d.Verified())
		db.QueryRow(`SELECT COUNT(*) FROM pentests WHERE user_id=?`, userID).Scan(&d.ScanCount)
		d.PendingPay = loadPendingPayment(userID)
	case "dashboard_websites", "dashboard_scans":
		d.Domains = loadDomains(userID)
	case "dashboard_credits":
		loadCreditsData(&d, userID)
	}
	return d
}

// loadPendingPayment returns the newest invoice that is still payable, so Home
// can offer to finish it instead of leaving the customer to hunt for it.
func loadPendingPayment(userID string) *pymt {
	var p pymt
	err := db.QueryRow(`SELECT external_id, COALESCE((SELECT name FROM credit_packages cp WHERE cp.id=payments.package_id), COALESCE(package_id,'')),
			status, amount_usd, credits, created_at, COALESCE(url,'')
		FROM payments WHERE user_id=? AND status='pending' AND COALESCE(url,'') != '' AND created_at >= datetime('now','-1 day')
		ORDER BY id DESC LIMIT 1`, userID).
		Scan(&p.ExternalID, &p.Package, &p.Status, &p.AmountUSD, &p.Credits, &p.CreatedAt, &p.URL)
	if err != nil {
		return nil
	}
	return &p
}

func loadCreditsData(d *dashboardData, userID string) {
	rows, _ := db.Query(`SELECT id, name, usd_price, credits FROM credit_packages WHERE is_active=1 ORDER BY usd_price`)
	for rows.Next() {
		var p pkg
		rows.Scan(&p.ID, &p.Name, &p.USD, &p.Credits)
		d.Packages = append(d.Packages, p)
	}
	rows.Close()

	trows, _ := db.Query(`SELECT type, amount, description, created_at FROM credit_transactions WHERE user_id=? ORDER BY created_at DESC LIMIT 20`, userID)
	for trows.Next() {
		var t txn
		trows.Scan(&t.Type, &t.Amount, &t.Description, &t.CreatedAt)
		d.Transactions = append(d.Transactions, t)
	}
	trows.Close()

	// Mark overdue pending invoices as expired, then load the user's payment history
	// (pending/paid/expired) so the page reflects real invoice state.
	db.Exec(`UPDATE payments SET status='expired' WHERE user_id=? AND status='pending' AND created_at < datetime('now','-1 day')`, userID)
	ptrows, _ := db.Query(`SELECT external_id, COALESCE((SELECT name FROM credit_packages cp WHERE cp.id=payments.package_id), COALESCE(package_id,'')),
			status, COALESCE(paid_at,''), amount_usd, credits, created_at, COALESCE(url,'')
		FROM payments WHERE user_id=? ORDER BY id DESC LIMIT 20`, userID)
	for ptrows.Next() {
		var p pymt
		var paidSrc string
		_ = ptrows.Scan(&p.ExternalID, &p.Package, &p.Status, &paidSrc, &p.AmountUSD, &p.Credits, &p.CreatedAt, &p.URL)
		if p.Status == "paid" && paidSrc != "" {
			p.CreatedAt = paidSrc
		}
		// Expired invoices are noise on every page load — collapse them so the
		// history leads with what the customer can still act on.
		if p.Status == "expired" {
			d.Expired = append(d.Expired, p)
		} else {
			d.Payments = append(d.Payments, p)
		}
	}
	ptrows.Close()
}

// navItem is one sidebar entry. Key matches the active page passed to shell().
type navItem struct {
	Key, Href, Label, Icon string
}

var dashboardNav = []navItem{
	{"home", "/dashboard", "Home", "home"},
	{"sites", "/dashboard/websites", "My websites", "globe"},
	{"scans", "/dashboard/scans", "Security scans", "shield"},
	{"credits", "/dashboard/credits", "Credits & payments", "card"},
	{"auto", "/dashboard/automation", "Automation", "zap"},
}

// shellData is what the page shell needs: the page's own data plus which sidebar
// entry is current and what the browser tab should say.
type shellData struct {
	dashboardData
	Active, Title string
}

// txnLabel names a credit movement in words a customer uses.
func txnLabel(kind string) string {
	switch kind {
	case "topup":
		return "Credits added"
	case "spend":
		return "Scan"
	case "admin_credit":
		return "Credit adjustment"
	}
	return kind
}

// txnAmount signs a credit movement: scans take credits away, top-ups and
// adjustments add them (an adjustment can also be negative).
func txnAmount(kind string, amount float64) string {
	neg := kind == "spend" || amount < 0
	if amount < 0 {
		amount = -amount
	}
	if neg {
		return "−" + fmtCredits(amount)
	}
	return "+" + fmtCredits(amount)
}

var tmpl = template.Must(template.New("dashboard").Funcs(template.FuncMap{
	"cssHash": func() string { return cssHash },
	"asset":   assetURL,
	// template.HTML, not string: html/template escapes a plain string in an
	// element context and the pixel would render as literal &lt;script&gt; text.
	"fbpixel": func() template.HTML { return template.HTML(consentGate(true)) },

	"shell": func(d dashboardData, active, title string) shellData {
		return shellData{dashboardData: d, Active: active, Title: title}
	},
	"nav":          func() []navItem { return dashboardNav },
	"icon":         icon,
	"credits":      fmtCredits,
	"creditsLabel": creditsLabel,
	"usd":          func(v float64) string { return formatUSD(v, "en") },
	"perScan":      perScan,
	"timeTag":      timeTag,
	"txnLabel":     txnLabel,
	"txnAmount":    txnAmount,
}).Parse(dashboardHTML))
