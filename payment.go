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
	// One source of truth with the display side (pricing.go): this used to read
	// XENDIT_CURRENCY itself, defaulting to IDR, so the invoice currency and the
	// currency the site reasoned about could drift apart.
	currency := chargeCurrency()
	amount := topUpAmountUSD(usd, currency, usdRate())

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
	// Capture the Meta Pixel cookies / IP / UA now: the settlement webhook is a
	// server-to-server call from Xendit and has no browser context to attribute.
	mu := metaUserFrom(r, email)
	db.Exec(`INSERT INTO payments (user_id, external_id, xendit_invoice_id, url, package_id, amount_usd, credits, status, currency, meta_fbp, meta_fbc, meta_ip, meta_ua)
		VALUES (?,?,?,?,?,?,?,'pending',?,?,?,?,?)`, claims.Sub, externalID, xr.ID, xr.InvoiceURL, packageID, usd, credits, currency, mu.FBP, mu.FBC, mu.IP, mu.UserAgent)

	// Redirect the browser straight to the Xendit payment link (HTMX sees
	// HX-Redirect and navigates to it). No extra "pay now" step.
	w.Header().Set("HX-Redirect", xr.InvoiceURL)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<meta http-equiv="refresh" content="0;url=%s"><div class="rounded bg-emerald-500/15 px-3 py-2 text-xs text-emerald-300">Redirecting to payment… %v credits</div>`, xr.InvoiceURL, credits)
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

// xenditInvoiceStatus returns the current status of an invoice on Xendit.
func xenditInvoiceStatus(invoiceID string) (string, error) {
	auth := base64.StdEncoding.EncodeToString([]byte(getenv("XENDIT_SECRET_KEY", "") + ":"))
	req, _ := http.NewRequest("GET", "https://api.xendit.co/v2/invoices/"+invoiceID, nil)
	req.Header.Set("Authorization", "Basic "+auth)
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var x struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&x); err != nil {
		return "", err
	}
	return strings.ToUpper(x.Status), nil
}

// syncPendingPayments polls Xendit for pending payments and credits any that are
// paid/settled (idempotent) — so credits are added automatically WITHOUT relying
// on the Xendit webhook. Also flips overdue pending invoices to expired.
func syncPendingPayments() {
	rows, err := db.Query(`SELECT external_id, xendit_invoice_id FROM payments
		WHERE status='pending' AND xendit_invoice_id IS NOT NULL AND xendit_invoice_id != ''`)
	if err != nil {
		log.Printf("sync payments: query: %v", err)
		return
	}
	defer rows.Close()
	var batch []struct{ ext, inv string }
	for rows.Next() {
		var ext, inv string
		if err := rows.Scan(&ext, &inv); err == nil {
			batch = append(batch, struct{ ext, inv string }{ext, inv})
		}
	}
	rows.Close()
	for _, it := range batch {
		st, err := xenditInvoiceStatus(it.inv)
		if err != nil {
			log.Printf("sync payments: check %s: %v", it.inv, err)
			continue
		}
		switch st {
		case "PAID", "SETTLED":
			credited, err := creditPaymentByExternalID(it.ext)
			if err != nil {
				log.Printf("sync payments: credit %s: %v", it.ext, err)
				continue
			}
			if credited {
				log.Printf("sync payments: credited %s (no webhook)", it.ext)
			}
		case "EXPIRED", "FAILED":
			db.Exec(`UPDATE payments SET status='expired' WHERE external_id=? AND status='pending'`, it.ext)
		}
	}
}

// startPaymentPolling runs syncPendingPayments on an interval (background).
func startPaymentPolling() {
	time.Sleep(15 * time.Second)
	ticker := time.NewTicker(30 * time.Second)
	for {
		syncPendingPayments()
		<-ticker.C
	}
}

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
	var amountUSD float64
	var mu metaUser
	if err := tx.QueryRow(`SELECT p.user_id, p.credits, COALESCE(p.amount_usd,0),
			COALESCE(u.email,''), COALESCE(p.meta_fbp,''), COALESCE(p.meta_fbc,''),
			COALESCE(p.meta_ip,''), COALESCE(p.meta_ua,'')
		FROM payments p LEFT JOIN users u ON u.id = p.user_id
		WHERE p.external_id=?`, externalID).
		Scan(&userID, &credits, &amountUSD, &mu.Email, &mu.FBP, &mu.FBC, &mu.IP, &mu.UserAgent); err != nil {
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
	// Meta Pixel: Purchase event (payment settled) — server-side with accurate
	// value + currency. Fired only once per external_id (idempotent guard above).
	// The currency is USD because the value reported is amount_usd; payments.currency
	// is the Xendit charge currency (IDR by default) and would misprice the event.
	// Sent from a goroutine so a slow Meta call never delays the webhook's 200.
	go sendMetaEvent(metaEvent{
		Name:      "Purchase",
		ID:        externalID,
		SourceURL: "https://netsekurity.com/dashboard",
		User:      mu,
		Custom: map[string]interface{}{
			"value":          amountUSD,
			"currency":       "USD",
			"content_name":   "Credit top-up",
			"content_type":   "product",
			"num_items":      1,
			"transaction_id": externalID,
		},
	})
	return true, nil
}
