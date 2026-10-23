package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTopUpAmountUSD_IDRConversion(t *testing.T) {
	got := topUpAmountUSD(1000, "IDR", 16500)
	if got != 16500000 {
		t.Errorf("1000 USD @16500 = %d, want 16500000", got)
	}
	// zero rate falls back to default 16500
	if got := topUpAmountUSD(50, "IDR", 0); got != 825000 {
		t.Errorf("50 USD @default = %d, want 825000", got)
	}
}

func TestTopUpAmountUSD_ChargesUSDAsIs(t *testing.T) {
	if got := topUpAmountUSD(1000, "USD", 16500); got != 1000 {
		t.Errorf("USD amount = %d, want 1000", got)
	}
	if got := topUpAmountUSD(50, "usd", 0); got != 50 {
		t.Errorf("lowercase usd amount = %d, want 50", got)
	}
}

func TestCreditPayment_Idempotent(t *testing.T) {
	const uid = "u_topup_test"
	const ext = "EXT-IDEMPOTENT"
	db.Exec(`INSERT OR IGNORE INTO users (id, email, name) VALUES (?, 'topup@test.local', 'Topup')`, uid)
	db.Exec(`INSERT OR IGNORE INTO credit_balance (user_id, balance) VALUES (?, 0)`, uid)
	db.Exec(`INSERT INTO payments (user_id, external_id, package_id, amount_usd, credits, status) VALUES (?,?,'starter',50,1,'pending')`, uid, ext)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM credit_transactions WHERE user_id=?`, uid)
		db.Exec(`DELETE FROM payments WHERE external_id=?`, ext)
		db.Exec(`DELETE FROM credit_balance WHERE user_id=?`, uid)
		db.Exec(`DELETE FROM users WHERE id=?`, uid)
	})

	first, err := creditPaymentByExternalID(ext)
	if err != nil {
		t.Fatalf("first credit: %v", err)
	}
	if !first {
		t.Errorf("expected first call to credit (got false)")
	}

	// Replay (webhook redelivery) must NOT double-credit.
	second, err := creditPaymentByExternalID(ext)
	if err != nil {
		t.Fatalf("second credit: %v", err)
	}
	if second {
		t.Errorf("expected second call to be a no-op")
	}

	var balance float64
	db.QueryRow(`SELECT balance FROM credit_balance WHERE user_id=?`, uid).Scan(&balance)
	if balance != 1 {
		t.Errorf("balance = %v, want 1 (must not double-credit)", balance)
	}

	var txns int
	db.QueryRow(`SELECT COUNT(*) FROM credit_transactions WHERE user_id=? AND type='topup'`, uid).Scan(&txns)
	if txns != 1 {
		t.Errorf("topup transactions = %d, want 1", txns)
	}

	var status string
	db.QueryRow(`SELECT status FROM payments WHERE external_id=?`, ext).Scan(&status)
	if status != "paid" {
		t.Errorf("payment status = %q, want paid", status)
	}
}

// stubXendit points the invoice client at a local stub for the duration of the
// test. Production never sets XENDIT_API_BASE_URL and talks to api.xendit.co.
func stubXendit(t *testing.T, h http.HandlerFunc) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	setTestEnv(t, "XENDIT_API_BASE_URL", srv.URL)
}

func setTestEnv(t *testing.T, key, value string) {
	t.Helper()
	old, had := env[key]
	env[key] = value
	t.Cleanup(func() {
		if had {
			env[key] = old
		} else {
			delete(env, key)
		}
	})
}

func topUpRequest(t *testing.T, userID, email string) *http.Request {
	t.Helper()
	token, err := issueJWT(userID, email)
	if err != nil {
		t.Fatalf("issue jwt: %v", err)
	}
	r := httptest.NewRequest("POST", "/api/topup", strings.NewReader("package_id=starter"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.AddCookie(&http.Cookie{Name: authCookie, Value: token})
	return r
}

func seedTopUpUser(t *testing.T, id, email string) {
	t.Helper()
	if _, err := db.Exec(`INSERT OR REPLACE INTO users (id, email) VALUES (?, ?)`, id, email); err != nil {
		t.Fatalf("seed user: %v", err)
	}
}

// The invoice must be raised in the currency the checkout says it will charge,
// at the configured rate — the bug that made every top-up fail was a payload
// the provider rejected, and a silent amount/currency drift is the same class.
func TestTopUpInvoicesInTheChargedCurrency(t *testing.T) {
	const (
		userID = "u_topup_idr"
		email  = "topup-idr@example.test"
	)
	seedTopUpUser(t, userID, email)
	setTestEnv(t, "XENDIT_CURRENCY", "IDR")
	setTestEnv(t, "XENDIT_USD_RATE", "16500")
	setTestEnv(t, "XENDIT_SECRET_KEY", "xnd_development_test")

	var got struct {
		ExternalID  string `json:"external_id"`
		Amount      int64  `json:"amount"`
		Currency    string `json:"currency"`
		Description string `json:"description"`
	}
	var authHeader string
	stubXendit(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v2/invoices" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		authHeader = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"inv_stub_1","invoice_url":"https://checkout.xendit.co/web/inv_stub_1"}`))
	})

	rr := httptest.NewRecorder()
	handleTopUp(rr, topUpRequest(t, userID, email))

	if rr.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rr.Code, rr.Body.String())
	}
	if got.Currency != "IDR" {
		t.Errorf("invoice currency = %q, want IDR", got.Currency)
	}
	if got.Amount != 825000 {
		t.Errorf("invoice amount = %d, want 825000 ($50 at 16500)", got.Amount)
	}
	if !strings.HasPrefix(got.ExternalID, "NSK-"+userID+"-") {
		t.Errorf("external_id = %q, want NSK-<user>-<ts>", got.ExternalID)
	}
	if authHeader == "" {
		t.Error("invoice request sent no Authorization header")
	}
	if redirect := rr.Header().Get("HX-Redirect"); redirect != "https://checkout.xendit.co/web/inv_stub_1" {
		t.Errorf("HX-Redirect = %q", redirect)
	}

	var currency, status, invoiceID string
	var amountUSD float64
	if err := db.QueryRow(`SELECT currency, status, xendit_invoice_id, amount_usd FROM payments WHERE external_id=?`, got.ExternalID).
		Scan(&currency, &status, &invoiceID, &amountUSD); err != nil {
		t.Fatalf("payment row for %s: %v", got.ExternalID, err)
	}
	if currency != "IDR" || status != "pending" || invoiceID != "inv_stub_1" || amountUSD != 50 {
		t.Errorf("payment row = %s/%s/%s/%v, want IDR/pending/inv_stub_1/50", currency, status, invoiceID, amountUSD)
	}
}

// A provider that refuses the currency must surface a clear message and leave no
// payment row behind — this is the failure that silently ended every sale.
func TestTopUpSurfacesUnsupportedCurrency(t *testing.T) {
	const (
		userID = "u_topup_unsupported"
		email  = "topup-unsupported@example.test"
	)
	seedTopUpUser(t, userID, email)
	setTestEnv(t, "XENDIT_CURRENCY", "USD")
	setTestEnv(t, "XENDIT_SECRET_KEY", "xnd_development_test")

	stubXendit(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error_code":"UNSUPPORTED_CURRENCY","message":"currency USD is not configured"}`))
	})

	rr := httptest.NewRecorder()
	handleTopUp(rr, topUpRequest(t, userID, email))

	if rr.Code != http.StatusBadGateway {
		t.Fatalf("status %d, want 502", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "cannot bill in USD") {
		t.Errorf("body does not explain the currency refusal: %s", rr.Body.String())
	}
	var rows int
	db.QueryRow(`SELECT COUNT(*) FROM payments WHERE user_id=?`, userID).Scan(&rows)
	if rows != 0 {
		t.Errorf("%d payment rows written for a rejected invoice, want 0", rows)
	}
}
