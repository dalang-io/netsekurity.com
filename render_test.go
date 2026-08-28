package main

import (
	"strings"
	"testing"
)

// The dashboard and /su templates are only reachable behind auth, so parse
// success (template.Must) is not enough — these execute them with real-shaped
// data to catch runtime errors like a bad comparison or a missing field.
func TestDashboardTemplateExecutes(t *testing.T) {
	for _, tc := range []struct {
		name string
		data dashboardData
	}{
		{"empty", dashboardData{}},
		{"zero balance", dashboardData{
			Name: "Test", Email: "t@example.com", Balance: 0,
			RateNote: localRateNote(),
			Packages: []pkg{{ID: "starter", Name: "Starter", USD: 50, Credits: 1, Local: localPrice(50)}},
			Domains:  []dom{{Domain: "example.com", Status: "verified", TXT: "ns-verify-abc"}},
			Expired:  []pymt{{ExternalID: "NSK-1", Package: "starter", Status: "expired", Credits: 1, AmountUSD: 50}},
		}},
		{"funded", dashboardData{
			Balance: 20, IsAdmin: true,
			Domains:      []dom{{Domain: "a.example.com", Status: "verified"}, {Domain: "b.example.com", Status: "pending"}},
			Payments:     []pymt{{ExternalID: "NSK-2", Package: "pro", Status: "pending", URL: "https://pay", Credits: 20, AmountUSD: 500}},
			Transactions: []txn{{Type: "topup", Description: "d", Amount: 3}},
		}},
	} {
		var b strings.Builder
		if err := tmpl.ExecuteTemplate(&b, "dashboard", tc.data); err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if !strings.Contains(b.String(), "credit-balance") {
			t.Errorf("%s: balance element missing", tc.name)
		}
	}
}

// A zero balance must not render an enabled scan button — the whole point of the
// disabled state is that clicking it can only produce a 402.
func TestDashboardHidesScanAtZeroBalance(t *testing.T) {
	data := dashboardData{Balance: 0, Domains: []dom{{Domain: "example.com", Status: "verified", TXT: "x"}}}
	var b strings.Builder
	if err := tmpl.ExecuteTemplate(&b, "domainlist", data); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	if strings.Contains(out, `hx-post="/api/pentests/start"`) {
		t.Error("scan button is live at zero balance")
	}
	if !strings.Contains(out, "disabled") {
		t.Error("expected a disabled scan button")
	}

	data.Balance = 5
	var b2 strings.Builder
	if err := tmpl.ExecuteTemplate(&b2, "domainlist", data); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b2.String(), `hx-post="/api/pentests/start"`) {
		t.Error("scan button missing with credits available")
	}
}

func TestSuTemplateExecutes(t *testing.T) {
	data := suData{
		CurrentID: "u_me",
		Stats:     suStats{Users: 33, InvoicesPaid: 0, InvoicesExpired: 28, RevenueUSD: 0, CreditsOut: 21.5},
		Users: []suUser{
			{ID: "u_me", Email: "me@example.com", Name: "Me", Role: "admin", Credits: 5},
			{ID: "u_other", Email: "other@example.com", Name: "Other", Role: "user"},
		},
		Trx:      []suTrx{{ExternalID: "NSK-1", Package: "starter", Status: "expired", Amount: 50, Credits: 1}},
		Domains:  []suDomain{{ID: 1, UserEmail: "other@example.com", Domain: "example.com", Status: "verified"}},
		Pentests: []suPentest{{ID: "pt_1", UserEmail: "other@example.com", Domain: "example.com", Status: "completed", Mode: "standard"}},
		Credits:  []suCredit{{Email: "other@example.com", Amount: "5.0", Description: "Admin credit", CreatedAt: "2026-08-28"}},
	}
	var b strings.Builder
	if err := suTpl.ExecuteTemplate(&b, "su", data); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	// The operator's own row must not offer a role control.
	if strings.Count(out, `action="/su/users/role"`) != 1 {
		t.Errorf("expected exactly one role form (not the current admin's own row)")
	}
	if !strings.Contains(out, "— you —") {
		t.Error("current admin's row is not marked")
	}
	for _, want := range []string{"invoices paid", "su-destructive", "manual credit grants", "su-search"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q", want)
		}
	}
}

func TestLocalPriceFormatsCharge(t *testing.T) {
	env["XENDIT_CURRENCY"] = "IDR"
	env["XENDIT_USD_RATE"] = "16500"
	defer func() { delete(env, "XENDIT_CURRENCY"); delete(env, "XENDIT_USD_RATE") }()
	if got := localPrice(50); got != "Rp 825.000" {
		t.Errorf("localPrice(50) = %q, want %q", got, "Rp 825.000")
	}
	if got := localPrice(1000); got != "Rp 16.500.000" {
		t.Errorf("localPrice(1000) = %q, want %q", got, "Rp 16.500.000")
	}
	env["XENDIT_CURRENCY"] = "USD"
	if got := localPrice(50); got != "" {
		t.Errorf("localPrice with USD charge currency = %q, want empty", got)
	}
}

func TestNormalizeDomainStripsURLs(t *testing.T) {
	cases := map[string]string{
		"https://paroki.wensputra.my.id/": "paroki.wensputra.my.id",
		"  HTTP://App.Example.COM:8443/x": "app.example.com",
		"example.com.":                    "example.com",
		"app.example.com":                 "app.example.com",
	}
	for in, want := range cases {
		if got := normalizeDomain(in); got != want {
			t.Errorf("normalizeDomain(%q) = %q, want %q", in, got, want)
		}
	}
	// The pasted URL used to pass validation and then never verify.
	if isValidDomain("https://paroki.wensputra.my.id/") {
		t.Error("a raw URL still passes isValidDomain")
	}
	if !isValidDomain(normalizeDomain("https://paroki.wensputra.my.id/")) {
		t.Error("normalized hostname should be valid")
	}
}
