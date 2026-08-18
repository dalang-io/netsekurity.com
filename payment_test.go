package main

import (
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
