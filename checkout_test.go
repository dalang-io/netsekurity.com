package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTrimUSD(t *testing.T) {
	for in, want := range map[float64]string{50: "50", 1000: "1000", 49.5: "49.5", 0: "0"} {
		if got := trimUSD(in); got != want {
			t.Errorf("trimUSD(%v) = %q, want %q", in, got, want)
		}
	}
}

// The footnote is the whole point of the page: the customer must learn the
// charge currency here, not on the payment provider's page.
func TestCheckoutFootnoteStatesTheChargeCurrency(t *testing.T) {
	delete(env, "XENDIT_CURRENCY")

	data := checkoutData{
		Email: "buyer@example.com", Balance: 0,
		Package: pkg{ID: "starter", Name: "Starter", USD: 50, Credits: 1},
		Total:   "$50", Mismatch: currencyMismatch(),
	}
	data.ChargeNote = "All prices are in US dollars. You will be charged $50 USD."

	var b strings.Builder
	if err := checkoutTpl.ExecuteTemplate(&b, "checkout", data); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	for _, want := range []string{"Starter", "$50", "USD", "buyer@example.com", "proceed to payment"} {
		if !strings.Contains(out, want) {
			t.Errorf("checkout page missing %q", want)
		}
	}
	// Charging USD means no other currency may appear anywhere.
	for _, banned := range []string{"Rp ", "IDR", "16.500"} {
		if strings.Contains(out, banned) {
			t.Errorf("checkout leaks %q while charging USD", banned)
		}
	}
}

// End to end through the handler, for both currency configurations.
func TestCheckoutHandler(t *testing.T) {
	const uid = "u_checkout_test"
	db.Exec(`INSERT OR IGNORE INTO users (id, email, name) VALUES (?, 'checkout@test.local', 'C')`, uid)
	token, err := issueJWT(uid, "checkout@test.local")
	if err != nil {
		t.Fatal(err)
	}
	get := func(q string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("GET", "/checkout"+q, nil)
		r.AddCookie(&http.Cookie{Name: authCookie, Value: token})
		w := httptest.NewRecorder()
		handleCheckout(w, r)
		return w
	}

	// Charging USD: reassurance, no converted amount.
	delete(env, "XENDIT_CURRENCY")
	w := get("?package=starter")
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	out := w.Body.String()
	if !strings.Contains(out, "All prices are in US dollars") {
		t.Error("missing the USD footnote")
	}
	for _, banned := range []string{"Rp ", "IDR"} {
		if strings.Contains(out, banned) {
			t.Errorf("leaks %q while charging USD", banned)
		}
	}

	// Charging something else: the customer is told the other amount HERE.
	env["XENDIT_CURRENCY"] = "IDR"
	env["XENDIT_USD_RATE"] = "16500"
	defer func() { delete(env, "XENDIT_CURRENCY"); delete(env, "XENDIT_USD_RATE") }()
	out = get("?package=starter").Body.String()
	if !strings.Contains(out, "825.000 IDR") {
		t.Error("a non-USD charge currency must state the converted amount on the checkout page")
	}
	if !strings.Contains(out, "Note on currency") {
		t.Error("expected the mismatch to be called out")
	}

	// Unknown or missing package falls back to the dashboard, never a blank page.
	for _, q := range []string{"", "?package=", "?package=does-not-exist"} {
		if code := get(q).Code; code != http.StatusFound {
			t.Errorf("checkout%q = %d, want a redirect", q, code)
		}
	}
}
