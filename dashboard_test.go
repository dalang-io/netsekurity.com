package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// dashboardPages are the template names of every dashboard page, in sidebar
// order. The tests below walk all of them so a page cannot be added, or broken,
// without one failing.
var dashboardPages = []string{"dashboard", "dashboard_websites", "dashboard_scans", "dashboard_credits", "dashboard_automation"}

func sampleDashboardData() dashboardData {
	return dashboardData{
		Name: "Siti Rahma", Email: "siti@example.com", Balance: 3, IsAdmin: true,
		Packages:     []pkg{{ID: "starter", Name: "Starter", USD: 50, Credits: 1}, {ID: "standard", Name: "Standard", USD: 100, Credits: 3}},
		Transactions: []txn{{Type: "topup", Description: "Starter pack", Amount: 3, CreatedAt: "2026-08-18 10:00:00"}, {Type: "spend", Description: "Scan", Amount: 1, CreatedAt: "2026-08-18T11:00:00Z"}},
		Payments:     []pymt{{ExternalID: "NSK-1", Package: "Starter", Status: "paid", Credits: 1, AmountUSD: 50, CreatedAt: "2026-08-18 10:00:00"}},
		Domains:      []dom{{Domain: "ok.example.com", Status: "verified", TXT: "ns-verify-aaa"}, {Domain: "new.example.com", Status: "pending", TXT: "ns-verify-bbb"}},
		ScanCount:    2, VerifiedCount: 1,
	}
}

// Every page executes, carries the shell, and marks exactly its own sidebar
// entry as current.
func TestDashboardPagesRenderInTheShell(t *testing.T) {
	wantNav := map[string]string{
		"dashboard":            "/dashboard",
		"dashboard_websites":   "/dashboard/websites",
		"dashboard_scans":      "/dashboard/scans",
		"dashboard_credits":    "/dashboard/credits",
		"dashboard_automation": "/dashboard/automation",
	}
	for _, page := range dashboardPages {
		var b strings.Builder
		if err := tmpl.ExecuteTemplate(&b, page, sampleDashboardData()); err != nil {
			t.Fatalf("%s: %v", page, err)
		}
		out := b.String()
		for _, want := range []string{`id="nsk-sidebar"`, `id="flash"`, `id="credit-balance"`, `class="nsk-app`, "Plus+Jakarta+Sans", "siti@example.com", `href="/su"`} {
			if !strings.Contains(out, want) {
				t.Errorf("%s: missing %q", page, want)
			}
		}
		if n := strings.Count(out, `aria-current="page"`); n != 1 {
			t.Errorf("%s: %d sidebar entries marked current, want 1", page, n)
		}
		want := `<a href="` + wantNav[page] + `" class="nsk-nav-link" aria-current="page">`
		if !strings.Contains(out, want) {
			t.Errorf("%s: sidebar does not mark %s current", page, wantNav[page])
		}
		// The terminal look is gone from the customer app.
		for _, banned := range []string{"scanlines", "bg-ink", "$ balance", "$ topup", "# loading"} {
			if strings.Contains(out, banned) {
				t.Errorf("%s still carries the old terminal styling: %q", page, banned)
			}
		}
	}
}

// Only an admin is offered the /su door.
func TestDashboardHidesAdminLinkFromCustomers(t *testing.T) {
	d := sampleDashboardData()
	d.IsAdmin = false
	var b strings.Builder
	if err := tmpl.ExecuteTemplate(&b, "dashboard", d); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(b.String(), `href="/su"`) {
		t.Error("a customer is shown the Super Admin link")
	}
}

// A brand-new account sees exactly one current step, and it is the first.
func TestGettingStartedHasOneCurrentStep(t *testing.T) {
	cases := []struct {
		name    string
		d       dashboardData
		current int // 0 = none (all done)
		done    int
	}{
		{"new account", dashboardData{}, 1, 0},
		{"domain added", dashboardData{Domains: []dom{{Domain: "a.example.com", Status: "pending"}}}, 2, 1},
		{"domain verified", dashboardData{Domains: []dom{{Domain: "a.example.com", Status: "verified"}}, VerifiedCount: 1}, 3, 2},
		{"has credits", dashboardData{Domains: []dom{{Domain: "a.example.com", Status: "verified"}}, VerifiedCount: 1, Balance: 2}, 4, 3},
		{"has scanned", dashboardData{Domains: []dom{{Domain: "a.example.com", Status: "verified"}}, VerifiedCount: 1, ScanCount: 1}, 0, 4},
	}
	for _, tc := range cases {
		current := 0
		for _, s := range tc.d.Steps() {
			if s.Current {
				if current != 0 {
					t.Errorf("%s: more than one current step", tc.name)
				}
				current = s.N
			}
		}
		if current != tc.current {
			t.Errorf("%s: current step = %d, want %d", tc.name, current, tc.current)
		}
		if got := tc.d.StepsDone(); got != tc.done {
			t.Errorf("%s: steps done = %d, want %d", tc.name, got, tc.done)
		}
		if tc.d.AllDone() != (tc.current == 0) {
			t.Errorf("%s: AllDone = %v", tc.name, tc.d.AllDone())
		}
	}
}

// A new account's Home leads with the checklist; a finished one leads with the
// scan picker. Neither shows the other.
func TestHomeShowsChecklistOrScanPicker(t *testing.T) {
	render := func(d dashboardData) string {
		var b strings.Builder
		if err := tmpl.ExecuteTemplate(&b, "dashboard", d); err != nil {
			t.Fatal(err)
		}
		return b.String()
	}
	fresh := render(dashboardData{})
	if !strings.Contains(fresh, "Get started") || strings.Contains(fresh, `id="scan-picker"`) {
		t.Error("a new account should see the checklist and no scan picker")
	}
	if !strings.Contains(fresh, "Add a website") {
		t.Error("the first step's call to action is missing")
	}
	done := render(dashboardData{Balance: 2, ScanCount: 1, VerifiedCount: 1, Domains: []dom{{Domain: "a.example.com", Status: "verified"}}})
	if strings.Contains(done, "Get started") || !strings.Contains(done, `id="scan-picker"`) {
		t.Error("a finished account should see the scan picker and no checklist")
	}
}

// An unverified domain's card explains what to do; a verified one offers the scan.
func TestDomainCardsGuideTheCustomer(t *testing.T) {
	var b strings.Builder
	if err := tmpl.ExecuteTemplate(&b, "domainlist", sampleDashboardData()); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	for _, want := range []string{"Prove this website is yours", "ns-verify-bbb", "_netsekurity", "Check now", "Verified", "Remove"} {
		if !strings.Contains(out, want) {
			t.Errorf("domain list missing %q", want)
		}
	}
	// The verification token belongs to the unverified domain only.
	if strings.Contains(out, "ns-verify-aaa") {
		t.Error("a verified domain still shows its verification record")
	}
}

// A refused request is swapped into the flash area, not dropped on the floor.
func TestShellSwapsOurErrorAlerts(t *testing.T) {
	var b strings.Builder
	if err := tmpl.ExecuteTemplate(&b, "dashboard", dashboardData{}); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"htmx:beforeSwap", "htmx:responseError", "htmx:sendError", "nsk-alert"} {
		if !strings.Contains(b.String(), want) {
			t.Errorf("shell script missing %q", want)
		}
	}
}

// End to end through the handlers: auth gate, then a 200 with the right page.
func TestDashboardHandlersServeEachPage(t *testing.T) {
	const uid = "u_dashpages_test"
	db.Exec(`INSERT OR IGNORE INTO users (id, email, name) VALUES (?, 'pages@test.local', 'Pat Lee')`, uid)
	token, err := issueJWT(uid, "pages@test.local")
	if err != nil {
		t.Fatal(err)
	}
	handlers := map[string]struct {
		h    http.HandlerFunc
		want string
	}{
		"/dashboard":            {handleDashboard, "Hi, Pat"},
		"/dashboard/websites":   {handleDashboardWebsites, "Add a website"},
		"/dashboard/scans":      {handleDashboardScans, "Start a scan"},
		"/dashboard/credits":    {handleDashboardCredits, "Buy credits"},
		"/dashboard/automation": {handleDashboardAutomation, "Create an access token"},
	}
	for path, tc := range handlers {
		anon := httptest.NewRecorder()
		tc.h(anon, httptest.NewRequest("GET", path, nil))
		if anon.Code != http.StatusFound || anon.Header().Get("Location") != "/login" {
			t.Errorf("%s anonymous: %d -> %q, want a redirect to /login", path, anon.Code, anon.Header().Get("Location"))
		}

		r := httptest.NewRequest("GET", path, nil)
		r.AddCookie(&http.Cookie{Name: authCookie, Value: token})
		w := httptest.NewRecorder()
		tc.h(w, r)
		if w.Code != http.StatusOK {
			t.Errorf("%s: status %d", path, w.Code)
			continue
		}
		if !strings.Contains(w.Body.String(), tc.want) {
			t.Errorf("%s: missing %q", path, tc.want)
		}
		if cc := w.Header().Get("Cache-Control"); !strings.Contains(cc, "no-store") {
			t.Errorf("%s: Cache-Control = %q, a private page must not be cached", path, cc)
		}
	}
}

// The Home page offers to finish an unpaid invoice instead of hiding it.
func TestHomeOffersToFinishAnUnpaidInvoice(t *testing.T) {
	d := dashboardData{PendingPay: &pymt{Package: "Starter", AmountUSD: 50, URL: "https://pay.example/inv"}}
	var b strings.Builder
	if err := tmpl.ExecuteTemplate(&b, "dashboard", d); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	if !strings.Contains(out, "Finish payment") || !strings.Contains(out, "https://pay.example/inv") {
		t.Error("an unpaid invoice is not offered for payment")
	}
}

func TestPentestListLimit(t *testing.T) {
	const uid = "u_listlimit_test"
	db.Exec(`INSERT OR IGNORE INTO users (id, email, name) VALUES (?, 'limit@test.local', 'L')`, uid)
	db.Exec(`INSERT OR IGNORE INTO domains (id, user_id, domain, txt_verification_token, status) VALUES (911, ?, 'limit.example.com', 't', 'verified')`, uid)
	db.Exec(`DELETE FROM pentests WHERE user_id=?`, uid)
	// 'completed', not 'queued': the package shares one test DB, and a queued row
	// would be claimed by the worker and partner tests that run after this one.
	for _, id := range []string{"pt_l1", "pt_l2", "pt_l3", "pt_l4"} {
		db.Exec(`INSERT INTO pentests (id, user_id, domain_id, mode, status) VALUES (?, ?, 911, 'standard', 'completed')`, id, uid)
	}
	t.Cleanup(func() { db.Exec(`DELETE FROM pentests WHERE user_id=?`, uid) })
	tok, _ := issueJWT(uid, "limit@test.local")
	get := func(q string) string {
		r := httptest.NewRequest("GET", "/api/pentests/list"+q, nil)
		r.AddCookie(&http.Cookie{Name: authCookie, Value: tok})
		w := httptest.NewRecorder()
		handlePentestList(w, r)
		return w.Body.String()
	}
	if n := strings.Count(get("?limit=2"), "<li "); n != 2 {
		t.Errorf("limit=2 rendered %d scans", n)
	}
	if n := strings.Count(get(""), "<li "); n != 4 {
		t.Errorf("no limit rendered %d scans, want all 4", n)
	}
	// A hostile or nonsense limit falls back to the default rather than erroring.
	for _, q := range []string{"?limit=0", "?limit=-3", "?limit=abc", "?limit=9999"} {
		if n := strings.Count(get(q), "<li "); n != 4 {
			t.Errorf("%s rendered %d scans, want the default (4 exist)", q, n)
		}
	}
}

// Creating a token shows the secret once with a copy button, and tells the list
// on the page to refresh; the list then shows the token by name, never the secret.
func TestTokenCreateAndListFragments(t *testing.T) {
	const uid = "u_tokens_ui_test"
	db.Exec(`INSERT OR IGNORE INTO users (id, email, name) VALUES (?, 'tokensui@test.local', 'T')`, uid)
	db.Exec(`DELETE FROM api_tokens WHERE user_id=?`, uid)
	t.Cleanup(func() { db.Exec(`DELETE FROM api_tokens WHERE user_id=?`, uid) })
	tok, _ := issueJWT(uid, "tokensui@test.local")
	auth := func(r *http.Request) *http.Request {
		r.AddCookie(&http.Cookie{Name: authCookie, Value: tok})
		return r
	}

	form := strings.NewReader("name=Deploy+%3Cb%3Ebot%3C%2Fb%3E&expiry_days=30")
	req := auth(httptest.NewRequest("POST", "/api/tokens/create", form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	handleTokensCreate(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}
	if w.Header().Get("HX-Trigger") != "tokens-changed" {
		t.Errorf("create did not ask the list to refresh: HX-Trigger=%q", w.Header().Get("HX-Trigger"))
	}
	body := w.Body.String()
	for _, want := range []string{"nsk_live_", "shown only once", "nskCopy(this)", "X-API-Token"} {
		if !strings.Contains(body, want) {
			t.Errorf("create response missing %q", want)
		}
	}
	if strings.Contains(body, "<b>bot</b>") {
		t.Error("a token name was echoed unescaped")
	}
	// The retired "generate another" button posted to a GET-only route.
	if strings.Contains(body, `hx-post="/api/tokens"`) {
		t.Error("the broken generate-another button is back")
	}

	lw := httptest.NewRecorder()
	handleTokensList(lw, auth(httptest.NewRequest("GET", "/api/tokens", nil)))
	list := lw.Body.String()
	if !strings.Contains(list, "Deploy &lt;b&gt;bot&lt;/b&gt;") || !strings.Contains(list, "Active") || !strings.Contains(list, "Delete") {
		t.Errorf("token list = %s", list)
	}
	empty := httptest.NewRecorder()
	otherTok, _ := issueJWT("u_nobody_tokens", "nobody@test.local")
	r := httptest.NewRequest("GET", "/api/tokens", nil)
	r.AddCookie(&http.Cookie{Name: authCookie, Value: otherTok})
	handleTokensList(empty, r)
	if !strings.Contains(empty.Body.String(), "No tokens yet") {
		t.Error("empty state missing")
	}
}

// Adding a website tidies a pasted URL, reports it in plain words, and refreshes
// both the domain list and the scan picker out of band.
func TestAddDomainFragments(t *testing.T) {
	const uid = "u_adddomain_ui_test"
	db.Exec(`INSERT OR IGNORE INTO users (id, email, name) VALUES (?, 'adddomain@test.local', 'A')`, uid)
	db.Exec(`DELETE FROM domains WHERE user_id=?`, uid)
	t.Cleanup(func() { db.Exec(`DELETE FROM domains WHERE user_id=?`, uid) })
	tok, _ := issueJWT(uid, "adddomain@test.local")
	post := func(v string) *httptest.ResponseRecorder {
		req := httptest.NewRequest("POST", "/api/domains", strings.NewReader("domain="+v))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: authCookie, Value: tok})
		w := httptest.NewRecorder()
		handleAddDomain(w, req)
		return w
	}

	w := post("https%3A%2F%2FWWW.Toko-Siti.com%2Fproduk%3Fx%3D1")
	out := w.Body.String()
	if w.Code != http.StatusOK || !strings.Contains(out, "www.toko-siti.com") || !strings.Contains(out, "added") {
		t.Errorf("add: %d %s", w.Code, out)
	}
	for _, id := range []string{`id="domain-list" hx-swap-oob="true"`, `id="scan-picker" hx-swap-oob="true"`} {
		if !strings.Contains(out, id) {
			t.Errorf("add response does not refresh %s", id)
		}
	}
	if !strings.Contains(out, "Prove this website is yours") {
		t.Error("the new card does not carry the verification steps")
	}

	again := post("www.toko-siti.com").Body.String()
	if !strings.Contains(again, "already in your list") {
		t.Errorf("a duplicate should say so: %s", again)
	}

	bad := post("hello+world")
	if bad.Code != http.StatusBadRequest || !strings.Contains(bad.Body.String(), "nsk-alert") {
		t.Errorf("bad address: %d %s", bad.Code, bad.Body.String())
	}
}

// The empty state must not send someone who already verified a website back to
// "add and verify a website" — it only says what will appear.
func TestPentestListEmptyStateDoesNotMisdirect(t *testing.T) {
	const uid = "u_emptylist_test"
	db.Exec(`INSERT OR IGNORE INTO users (id, email, name) VALUES (?, 'emptylist@test.local', 'E')`, uid)
	tok, _ := issueJWT(uid, "emptylist@test.local")
	r := httptest.NewRequest("GET", "/api/pentests/list", nil)
	r.AddCookie(&http.Cookie{Name: authCookie, Value: tok})
	w := httptest.NewRecorder()
	handlePentestList(w, r)
	out := w.Body.String()
	if !strings.Contains(out, "No scans yet") || strings.Contains(out, "verify a website") {
		t.Errorf("empty state = %s", out)
	}
}
