package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

const testPartnerSecret = "partner-test-secret-test-secret"

func withPartner(t *testing.T) {
	t.Helper()
	setTestEnv(t, "PARTNER_SECRET", testPartnerSecret)
	setTestEnv(t, "PARTNER_ID", "dalang")
	setTestEnv(t, "BOT_AUTH_TOKEN", "test-bot-token")
}

func partnerSignedRequest(method, path string, body []byte, ts, nonce string) *http.Request {
	sig := partnerSignatureForTest("dalang", ts, nonce, body)
	r := httptest.NewRequest(method, path, bytes.NewReader(body))
	r.Header.Set("X-Partner-Id", "dalang")
	r.Header.Set("X-Timestamp", ts)
	r.Header.Set("X-Nonce", nonce)
	r.Header.Set("X-Signature", sig)
	r.Header.Set("Content-Type", "application/json")
	return r
}

func partnerRequest(method, path string, body []byte) *http.Request {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	return partnerSignedRequest(method, path, body, ts, randomHex(16))
}

func reserveBody(order, user, domain, proof, key string) []byte {
	b, _ := json.Marshal(map[string]interface{}{
		"external_order_id": order,
		"external_user_id":  user,
		"domain":            domain,
		"proof_type":        proof,
		"idempotency_key":   key,
	})
	return b
}

func decodePartnerView(t *testing.T, body []byte) partnerScanView {
	t.Helper()
	var v partnerScanView
	if err := json.Unmarshal(body, &v); err != nil {
		t.Fatalf("decode view: %v (%s)", err, body)
	}
	return v
}

func TestPartnerAuthEnvelope(t *testing.T) {
	withPartner(t)
	body := reserveBody("ord-auth", "ext-auth", "auth.example.test", "cloudflare", "key-auth")

	// No signature at all.
	rr := httptest.NewRecorder()
	handlePartnerReserve(rr, httptest.NewRequest("POST", "/api/partner/v1/scans/reserve", bytes.NewReader(body)))
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("unsigned request = %d, want 401", rr.Code)
	}

	// Stale timestamp.
	ts := strconv.FormatInt(time.Now().Add(-time.Hour).Unix(), 10)
	rr = httptest.NewRecorder()
	handlePartnerReserve(rr, partnerSignedRequest("POST", "/api/partner/v1/scans/reserve", body, ts, randomHex(8)))
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("stale timestamp = %d, want 401", rr.Code)
	}

	// Wrong signature (valid shape, different body).
	now := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := randomHex(8)
	req := partnerSignedRequest("POST", "/api/partner/v1/scans/reserve", body, now, nonce)
	req.Header.Set("X-Signature", partnerSignatureForTest("dalang", now, nonce, []byte("other")))
	rr = httptest.NewRecorder()
	handlePartnerReserve(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("bad signature = %d, want 401", rr.Code)
	}

	// Valid request, then an exact replay of the same nonce.
	rr = httptest.NewRecorder()
	handlePartnerReserve(rr, partnerSignedRequest("POST", "/api/partner/v1/scans/reserve", body, now, nonce))
	if rr.Code != http.StatusCreated {
		t.Fatalf("valid reserve = %d: %s", rr.Code, rr.Body.String())
	}
	rr = httptest.NewRecorder()
	handlePartnerReserve(rr, partnerSignedRequest("POST", "/api/partner/v1/scans/reserve", body, now, nonce))
	if rr.Code != http.StatusConflict {
		t.Errorf("replayed nonce = %d, want 409", rr.Code)
	}
}

func TestPartnerReserveProofAndTargets(t *testing.T) {
	withPartner(t)

	// TXT proof requires a verified domain the platform does not assert.
	rr := httptest.NewRecorder()
	handlePartnerReserve(rr, partnerRequest("POST", "/api/partner/v1/scans/reserve",
		reserveBody("ord-txt", "ext-txt", "unverified.example.test", "txt", "key-txt")))
	if rr.Code != http.StatusConflict || !strings.Contains(rr.Body.String(), "DOMAIN_NOT_VERIFIED") {
		t.Errorf("txt proof without verification = %d %s, want 409 DOMAIN_NOT_VERIFIED", rr.Code, rr.Body.String())
	}

	// Platform-owned hostnames stay out of scope.
	rr = httptest.NewRecorder()
	handlePartnerReserve(rr, partnerRequest("POST", "/api/partner/v1/scans/reserve",
		reserveBody("ord-svc", "ext-svc", "app.svc.dalang.io", "cloudflare", "key-svc")))
	if rr.Code != http.StatusBadRequest || !strings.Contains(rr.Body.String(), "TARGET_NOT_ALLOWED") {
		t.Errorf("platform host = %d %s, want 400 TARGET_NOT_ALLOWED", rr.Code, rr.Body.String())
	}

	// Invalid hostname.
	rr = httptest.NewRecorder()
	handlePartnerReserve(rr, partnerRequest("POST", "/api/partner/v1/scans/reserve",
		reserveBody("ord-bad", "ext-bad", "not a host", "cloudflare", "key-bad")))
	if rr.Code != http.StatusBadRequest {
		t.Errorf("invalid host = %d, want 400", rr.Code)
	}
}

func TestPartnerReserveConfirmClaimAndIdempotency(t *testing.T) {
	withPartner(t)
	order := "ord-life-" + randomHex(6)
	user := "ext-life-" + randomHex(6)
	key := "key-life-" + randomHex(6)

	rr := httptest.NewRecorder()
	handlePartnerReserve(rr, partnerRequest("POST", "/api/partner/v1/scans/reserve",
		reserveBody(order, user, "shop.example.test", "cloudflare", key)))
	if rr.Code != http.StatusCreated {
		t.Fatalf("reserve = %d: %s", rr.Code, rr.Body.String())
	}
	view := decodePartnerView(t, rr.Body.Bytes())
	if view.Status != "reserved" || view.ScanID == "" || view.ReportReady {
		t.Fatalf("reserve view = %+v", view)
	}

	// A reservation is not claimable: the platform has not paid yet.
	rr = httptest.NewRecorder()
	handleWorkerClaim(rr, botRequest("/api/pentests/worker/claim"))
	if !strings.Contains(rr.Body.String(), `"empty":true`) && !strings.Contains(rr.Body.String(), `"empty": true`) {
		t.Errorf("reserved scan was claimed: %s", rr.Body.String())
	}

	// Confirm moves it into the queue and starts the 6h clock.
	rr = httptest.NewRecorder()
	handlePartnerScan(rr, partnerRequest("POST", "/api/partner/v1/scans/"+view.ScanID+"/confirm", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("confirm = %d: %s", rr.Code, rr.Body.String())
	}
	confirmView := decodePartnerView(t, rr.Body.Bytes())
	if confirmView.Status != "queued" || confirmView.DeadlineAt == "" {
		t.Errorf("confirm view = %+v", confirmView)
	}

	rr = httptest.NewRecorder()
	handleWorkerClaim(rr, botRequest("/api/pentests/worker/claim"))
	var claim struct {
		PentestID string `json:"pentest_id"`
		Empty     bool   `json:"empty"`
	}
	json.Unmarshal(rr.Body.Bytes(), &claim)
	if claim.Empty || claim.PentestID != view.ScanID {
		t.Fatalf("claim = %s", rr.Body.String())
	}
	var attempts int
	db.QueryRow(`SELECT attempts FROM pentests WHERE id=?`, view.ScanID).Scan(&attempts)
	if attempts != 1 {
		t.Errorf("attempts after first claim = %d, want 1", attempts)
	}

	// Same idempotency key returns the original answer.
	rr = httptest.NewRecorder()
	handlePartnerReserve(rr, partnerRequest("POST", "/api/partner/v1/scans/reserve",
		reserveBody(order, user, "shop.example.test", "cloudflare", key)))
	if rr.Code != http.StatusOK {
		t.Fatalf("replay by key = %d: %s", rr.Code, rr.Body.String())
	}
	if replay := decodePartnerView(t, rr.Body.Bytes()); replay.ScanID != view.ScanID {
		t.Errorf("replay by key = %s, want %s", replay.ScanID, view.ScanID)
	} else {
		var current string
		db.QueryRow(`SELECT status FROM pentests WHERE id=?`, view.ScanID).Scan(&current)
		if replay.Status != current {
			t.Errorf("replay by key reported %s, want the current state %s", replay.Status, current)
		}
	}

	// A new key for the same order still maps to one scan.
	rr = httptest.NewRecorder()
	handlePartnerReserve(rr, partnerRequest("POST", "/api/partner/v1/scans/reserve",
		reserveBody(order, user, "shop.example.test", "cloudflare", key+"-2")))
	if rr.Code != http.StatusOK {
		t.Fatalf("replay by order = %d: %s", rr.Code, rr.Body.String())
	}
	if replay := decodePartnerView(t, rr.Body.Bytes()); replay.ScanID != view.ScanID {
		t.Errorf("replay by order = %s, want %s", replay.ScanID, view.ScanID)
	}

	// Releasing a scan that is already queued must be refused.
	rr = httptest.NewRecorder()
	handlePartnerScan(rr, partnerRequest("POST", "/api/partner/v1/scans/"+view.ScanID+"/release", nil))
	if rr.Code != http.StatusConflict {
		t.Errorf("release queued = %d, want 409", rr.Code)
	}
}

func TestPartnerReleaseReservation(t *testing.T) {
	withPartner(t)
	order := "ord-rel-" + randomHex(6)
	user := "ext-rel-" + randomHex(6)

	rr := httptest.NewRecorder()
	handlePartnerReserve(rr, partnerRequest("POST", "/api/partner/v1/scans/reserve",
		reserveBody(order, user, "rel.example.test", "cloudflare", "key-rel-"+order)))
	if rr.Code != http.StatusCreated {
		t.Fatalf("reserve = %d: %s", rr.Code, rr.Body.String())
	}
	view := decodePartnerView(t, rr.Body.Bytes())

	rr = httptest.NewRecorder()
	handlePartnerScan(rr, partnerRequest("POST", "/api/partner/v1/scans/"+view.ScanID+"/release", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("release = %d: %s", rr.Code, rr.Body.String())
	}
	if released := decodePartnerView(t, rr.Body.Bytes()); released.Status != "cancelled" {
		t.Errorf("release status = %s", released.Status)
	}

	rr = httptest.NewRecorder()
	handlePartnerScan(rr, partnerRequest("POST", "/api/partner/v1/scans/"+view.ScanID+"/confirm", nil))
	if rr.Code != http.StatusConflict {
		t.Errorf("confirm cancelled = %d, want 409", rr.Code)
	}
}

func TestPartnerWorkerFailDoesNotRefundCredits(t *testing.T) {
	withPartner(t)
	order := "ord-fail-" + randomHex(6)
	user := "ext-fail-" + randomHex(6)

	rr := httptest.NewRecorder()
	handlePartnerReserve(rr, partnerRequest("POST", "/api/partner/v1/scans/reserve",
		reserveBody(order, user, "fail.example.test", "cloudflare", "key-fail-"+order)))
	if rr.Code != http.StatusCreated {
		t.Fatalf("reserve = %d: %s", rr.Code, rr.Body.String())
	}
	view := decodePartnerView(t, rr.Body.Bytes())
	handlePartnerScan(httptest.NewRecorder(), partnerRequest("POST", "/api/partner/v1/scans/"+view.ScanID+"/confirm", nil))
	handleWorkerClaim(httptest.NewRecorder(), botRequest("/api/pentests/worker/claim"))

	var userID string
	db.QueryRow(`SELECT user_id FROM pentests WHERE id=?`, view.ScanID).Scan(&userID)

	body := strings.NewReader("pentest_id=" + view.ScanID + "&reason=target%20unreachable")
	req := httptest.NewRequest("POST", "/api/pentests/worker/fail", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Bot-Token", "test-bot-token")
	rr = httptest.NewRecorder()
	handleWorkerFail(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("worker fail = %d: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `"refunded":0`) {
		t.Errorf("partner failure refunded a credit: %s", rr.Body.String())
	}
	var txns int
	db.QueryRow(`SELECT COUNT(*) FROM credit_transactions WHERE user_id=?`, userID).Scan(&txns)
	if txns != 0 {
		t.Errorf("partner scan wrote %d credit transactions, want 0", txns)
	}

	// Control: a self-serve scan still gets its credit back.
	db.Exec(`INSERT OR IGNORE INTO users (id,email) VALUES ('u_self_fail','self-fail@example.test')`)
	db.Exec(`INSERT OR IGNORE INTO credit_balance (user_id,balance) VALUES ('u_self_fail',0)`)
	db.Exec(`INSERT INTO pentests (id,user_id,domain_id,mode,status,billing_source) VALUES ('pt_self_fail','u_self_fail',NULL,'standard','running','credits')`)
	req = httptest.NewRequest("POST", "/api/pentests/worker/fail",
		strings.NewReader("pentest_id=pt_self_fail&reason=boom"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Bot-Token", "test-bot-token")
	rr = httptest.NewRecorder()
	handleWorkerFail(rr, req)
	if !strings.Contains(rr.Body.String(), `"refunded":1`) {
		t.Errorf("self-serve failure did not refund: %s", rr.Body.String())
	}
}

func TestPartnerDeadlineAndAttemptsClose(t *testing.T) {
	withPartner(t)

	// Deadline exceeded while queued.
	order := "ord-dead-" + randomHex(6)
	user := "ext-dead-" + randomHex(6)
	rr := httptest.NewRecorder()
	handlePartnerReserve(rr, partnerRequest("POST", "/api/partner/v1/scans/reserve",
		reserveBody(order, user, "deadline.example.test", "cloudflare", "key-dead-"+order)))
	view := decodePartnerView(t, rr.Body.Bytes())
	handlePartnerScan(httptest.NewRecorder(), partnerRequest("POST", "/api/partner/v1/scans/"+view.ScanID+"/confirm", nil))
	db.Exec(`UPDATE pentests SET deadline_at=? WHERE id=?`, time.Now().Add(-time.Minute).UTC().Format(time.RFC3339), view.ScanID)
	sweepPartnerScans()
	var status, reason string
	db.QueryRow(`SELECT status, COALESCE(failure_reason,'') FROM pentests WHERE id=?`, view.ScanID).Scan(&status, &reason)
	if status != "failed" || reason != "deadline exceeded" {
		t.Errorf("deadline sweep = %s/%s, want failed/deadline exceeded", status, reason)
	}

	// Three claims is the limit: a stuck worker must not loop forever.
	order2 := "ord-att-" + randomHex(6)
	user2 := "ext-att-" + randomHex(6)
	rr = httptest.NewRecorder()
	handlePartnerReserve(rr, partnerRequest("POST", "/api/partner/v1/scans/reserve",
		reserveBody(order2, user2, "attempts.example.test", "cloudflare", "key-att-"+order2)))
	view2 := decodePartnerView(t, rr.Body.Bytes())
	handlePartnerScan(httptest.NewRecorder(), partnerRequest("POST", "/api/partner/v1/scans/"+view2.ScanID+"/confirm", nil))
	db.Exec(`UPDATE pentests SET status='running', attempts=?, started_at=? WHERE id=?`,
		partnerMaxAttempts, time.Now().Add(-2*time.Hour).UTC().Format(time.RFC3339), view2.ScanID)
	requeueStaleRunning()
	db.QueryRow(`SELECT status, COALESCE(failure_reason,'') FROM pentests WHERE id=?`, view2.ScanID).Scan(&status, &reason)
	if status != "failed" || reason != "attempts exhausted" {
		t.Errorf("attempt cap = %s/%s, want failed/attempts exhausted", status, reason)
	}
}

func TestPartnerReportSignedURL(t *testing.T) {
	withPartner(t)
	userID, err := partnerUser("ext-report-" + randomHex(6))
	if err != nil {
		t.Fatalf("partner user: %v", err)
	}
	const ref = "20260101-00:00-example.com.pdf"
	db.Exec(`INSERT INTO pentests (id,user_id,domain_id,mode,status,billing_source,report_ref,deadline_at)
		VALUES ('pt_report_test',? ,NULL,'standard','completed','partner',?,?)`,
		userID, ref, time.Now().Add(time.Hour).UTC().Format(time.RFC3339))
	os.MkdirAll(reportsDir(), 0o755)
	path := filepath.Join(reportsDir(), ref)
	if err := os.WriteFile(path, []byte("%PDF-1.4 test"), 0o644); err != nil {
		t.Fatalf("write report: %v", err)
	}
	t.Cleanup(func() { os.Remove(path) })

	exp := time.Now().Add(5 * time.Minute).Unix()
	signed := fmt.Sprintf("/api/partner/v1/scans/pt_report_test/report?exp=%d&sig=%s",
		exp, partnerSignatureForReport("pt_report_test", exp))
	rr := httptest.NewRecorder()
	handlePartnerScan(rr, httptest.NewRequest("GET", signed, nil))
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "%PDF") {
		t.Errorf("signed report = %d %q", rr.Code, rr.Body.String())
	}

	bad := fmt.Sprintf("/api/partner/v1/scans/pt_report_test/report?exp=%d&sig=deadbeef", exp)
	rr = httptest.NewRecorder()
	handlePartnerScan(rr, httptest.NewRequest("GET", bad, nil))
	if rr.Code != http.StatusForbidden {
		t.Errorf("bad signature = %d, want 403", rr.Code)
	}

	expired := fmt.Sprintf("/api/partner/v1/scans/pt_report_test/report?exp=%d&sig=%s",
		time.Now().Add(-time.Minute).Unix(), partnerSignatureForReport("pt_report_test", time.Now().Add(-time.Minute).Unix()))
	rr = httptest.NewRecorder()
	handlePartnerScan(rr, httptest.NewRequest("GET", expired, nil))
	if rr.Code != http.StatusForbidden {
		t.Errorf("expired link = %d, want 403", rr.Code)
	}
}

func botRequest(path string) *http.Request {
	r := httptest.NewRequest("POST", path, nil)
	r.Header.Set("X-Bot-Token", "test-bot-token")
	return r
}
