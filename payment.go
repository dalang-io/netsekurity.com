package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// handleTopUp creates a Xendit invoice for a credit package and returns the
// payment URL (used from the dashboard, HTMX friendly).
func handleTopUp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	claims, err := currentUser(r)
	if err != nil {
		w.Header().Set("HX-Redirect", "/login")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	packageID := r.FormValue("package_id")
	if packageID == "" {
		http.Error(w, "package_id required", http.StatusBadRequest)
		return
	}
	var name string
	var usd, credits float64
	if err := db.QueryRow(`SELECT name, usd_price, credits FROM credit_packages WHERE id=? AND is_active=1`, packageID).
		Scan(&name, &usd, &credits); err != nil {
		http.Error(w, "invalid package", http.StatusBadRequest)
		return
	}

	var email string
	db.QueryRow(`SELECT email FROM users WHERE id=?`, claims.Sub).Scan(&email)

	externalID := fmt.Sprintf("NSK-%s-%d", claims.Sub, time.Now().Unix())
	// Xendit charges in the account's configured currency. Most accounts only
	// support IDR (USD requires enabling with Xendit), so default to IDR and
	// convert via XENDIT_USD_RATE; set XENDIT_CURRENCY=USD once the account
	// supports it and the amount is charged in USD directly.
	currency := strings.ToUpper(getenv("XENDIT_CURRENCY", "IDR"))
	amount := topUpAmountUSD(usd, currency, parseFloat(getenv("XENDIT_USD_RATE", "16500")))

	payload, _ := json.Marshal(map[string]interface{}{
		"external_id":          externalID,
		"amount":               amount,
		"description":          fmt.Sprintf("Netsekurity credit top-up — %s (%v credits)", name, credits),
		"currency":             currency,
		"invoice_duration":     86400,
		"payer_email":          email,
		"success_redirect_url": getenv("XENDIT_SUCCESS_URL", "https://netsekurity.com/dashboard"),
		"failure_redirect_url": getenv("XENDIT_FAILURE_URL", "https://netsekurity.com/dashboard"),
	})

	auth := base64.StdEncoding.EncodeToString([]byte(getenv("XENDIT_SECRET_KEY", "") + ":"))
	req, _ := http.NewRequest("POST", "https://api.xendit.co/v2/invoices", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "payment provider unreachable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	var xr struct {
		ID         string `json:"id"`
		InvoiceURL string `json:"invoice_url"`
		ErrorMsg   string `json:"message"`
	}
	json.NewDecoder(resp.Body).Decode(&xr)
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		http.Error(w, "failed to create invoice: "+xr.ErrorMsg, http.StatusBadGateway)
		return
	}
	db.Exec(`INSERT INTO payments (user_id, external_id, xendit_invoice_id, package_id, amount_usd, credits, status, currency)
		VALUES (?,?,?,?,?,?,'pending',?)`, claims.Sub, externalID, xr.ID, packageID, usd, credits, currency)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<a href="%s" target="_blank" class="inline-block rounded-md bg-emerald-500 px-4 py-2 font-bold text-black">Pay now — %v credits</a>`, xr.InvoiceURL, credits)
}

// topUpAmountUSD returns the invoice amount in the configured currency. USD is
// charged as-is; IDR is converted at the configured rate (default 16500).
func topUpAmountUSD(usd float64, currency string, rate float64) int64 {
	if strings.EqualFold(currency, "USD") {
		return int64(usd)
	}
	if rate <= 0 {
		rate = 16500
	}
	return int64(usd * rate)
}

// handleXenditWebhook credits the user's balance when an invoice is paid.
// Idempotent: keyed on the unique external_id.
func handleXenditWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Callback token gate.
	if r.Header.Get("x-callback-token") != getenv("XENDIT_CALLBACK_KEY", "") {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var ev struct {
		ExternalID string `json:"external_id"`
		Status     string `json:"status"`
	}
	json.NewDecoder(r.Body).Decode(&ev)
	status := strings.ToUpper(ev.Status)
	if status != "PAID" && status != "SETTLED" {
		w.WriteHeader(http.StatusOK)
		return
	}
	credited, err := creditPaymentByExternalID(ev.ExternalID)
	if err != nil {
		log.Printf("webhook credit failed for %s: %v", ev.ExternalID, err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if credited {
		log.Printf("credited credits for invoice %s", ev.ExternalID)
	}
	w.WriteHeader(http.StatusOK)
}

// creditPaymentByExternalID atomically marks a pending payment as paid and
// credits the user's balance. It returns true when the credit was applied, or
// false if the payment was already processed (idempotent replay).
func creditPaymentByExternalID(externalID string) (bool, error) {
	tx, err := db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	// Mark the payment paid only if not already (atomic guard).
	res, err := tx.Exec(`UPDATE payments SET status='paid', paid_at=? WHERE external_id=? AND status='pending'`, now(), externalID)
	if err != nil {
		return false, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return false, nil // already processed — idempotent
	}

	var userID string
	var credits float64
	if err := tx.QueryRow(`SELECT user_id, credits FROM payments WHERE external_id=?`, externalID).Scan(&userID, &credits); err != nil {
		return false, err
	}

	// Atomically credit the balance (insert-or-create).
	if _, err := tx.Exec(`INSERT INTO credit_balance (user_id, balance, updated_at) VALUES (?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET balance = balance + excluded.balance, updated_at = excluded.updated_at`,
		userID, credits, now()); err != nil {
		return false, err
	}
	if _, err := tx.Exec(`INSERT INTO credit_transactions (user_id, type, amount, description, reference_id) VALUES (?, 'topup', ?, 'Credit top-up', ?)`,
		userID, credits, externalID); err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}
