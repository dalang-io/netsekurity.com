package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Partner API — the Dalang platform (api.dalang.io) reserves, pays for, and
// tracks a scan here. Netsekurity never touches a Dalang wallet, and Dalang
// never holds a Netsekurity credit: the platform debits its own wallet first
// and calls reserve/confirm, so a rejected scan costs nothing and a failure is
// refunded on the platform side.
//
// Authentication is a shared-secret HMAC over timestamp + nonce + body. The
// callers share a host with this process (api.dalang.io -> 127.0.0.1:8094), and
// Pingora also fronts this process, so an IP allowlist proves nothing; the HMAC
// and the replay ledger are the control.

const (
	partnerPlatform        = "dalang"
	partnerMaxSkew         = 5 * time.Minute
	partnerReserveTTL      = 15 * time.Minute
	partnerReserveTTLMax   = 60 * time.Minute
	partnerScanDeadline    = 6 * time.Hour
	partnerMaxAttempts     = 3
	partnerReportURLTTL    = 10 * time.Minute
	partnerProofTXT        = "txt"
	partnerProofCloudflare = "cloudflare"
	// Registrar/panel proof: a trusted platform (hosting.dalang.io) asserts it
	// sold or verified the domain at the registrar. Still an assertion, but a
	// stronger one than an arbitrary row.
	partnerProofRegistrar = "registrar"
)

func partnerSecret() string { return getenv("PARTNER_SECRET", "") }

func partnerID() string { return getenv("PARTNER_ID", partnerPlatform) }

func partnerJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func partnerErr(w http.ResponseWriter, status int, code, msg string) {
	partnerJSON(w, status, map[string]string{"error": code, "message": msg})
}

// partnerAuthOK verifies the HMAC envelope and consumes the nonce. It leaves
// the request body readable for the handler.
func partnerAuthOK(w http.ResponseWriter, r *http.Request) bool {
	secret := partnerSecret()
	if secret == "" {
		partnerErr(w, http.StatusServiceUnavailable, "PARTNER_API_DISABLED", "partner API is not configured")
		return false
	}
	id := r.Header.Get("X-Partner-Id")
	tsRaw := r.Header.Get("X-Timestamp")
	nonce := r.Header.Get("X-Nonce")
	sig := r.Header.Get("X-Signature")
	if id == "" || tsRaw == "" || nonce == "" || sig == "" {
		partnerErr(w, http.StatusUnauthorized, "PARTNER_AUTH_MISSING", "signed request headers are required")
		return false
	}
	if id != partnerID() {
		partnerErr(w, http.StatusUnauthorized, "PARTNER_UNKNOWN", "unknown partner")
		return false
	}
	ts, err := strconv.ParseInt(tsRaw, 10, 64)
	if err != nil {
		partnerErr(w, http.StatusUnauthorized, "PARTNER_AUTH_INVALID", "bad timestamp")
		return false
	}
	if skew := time.Since(time.Unix(ts, 0)); skew > partnerMaxSkew || skew < -partnerMaxSkew {
		partnerErr(w, http.StatusUnauthorized, "PARTNER_AUTH_STALE", "timestamp outside the allowed window")
		return false
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		partnerErr(w, http.StatusBadRequest, "PARTNER_BODY", "cannot read body")
		return false
	}
	r.Body = io.NopCloser(bytes.NewReader(body))

	want := partnerSignature(id, tsRaw, nonce, body)
	if !secureCompare(sig, want) {
		partnerErr(w, http.StatusUnauthorized, "PARTNER_AUTH_INVALID", "signature mismatch")
		return false
	}
	// Nonce is single-use: a captured request cannot be replayed inside the
	// timestamp window.
	if _, err := db.Exec(`INSERT INTO partner_nonces (nonce, seen_at) VALUES (?,?)`, nonce, now()); err != nil {
		partnerErr(w, http.StatusConflict, "PARTNER_REPLAY", "nonce already used")
		return false
	}
	db.Exec(`DELETE FROM partner_nonces WHERE seen_at < ?`, time.Now().UTC().Add(-10*time.Minute).Format(time.RFC3339))
	return true
}

func partnerSignature(id, ts, nonce string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(partnerSecret()))
	mac.Write([]byte(id + "\n" + ts + "\n" + nonce + "\n"))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

// PartnerSignatureForTest exposes the canonical signature so tests (and the
// integration documentation) do not reimplement the scheme.
func partnerSignatureForTest(id, ts, nonce string, body []byte) string {
	return partnerSignature(id, ts, nonce, body)
}

func partnerReportSignature(scanID string, exp int64) string {
	mac := hmac.New(sha256.New, []byte(partnerSecret()))
	fmt.Fprintf(mac, "report\n%s\n%d", scanID, exp)
	return hex.EncodeToString(mac.Sum(nil))
}

// partnerSignatureForReport is the exported form used by docs/tests.
func partnerSignatureForReport(scanID string, exp int64) string {
	return partnerReportSignature(scanID, exp)
}

func partnerReportURL(scanID string) string {
	exp := time.Now().Add(partnerReportURLTTL).Unix()
	return "/api/partner/v1/scans/" + scanID + "/report?exp=" + strconv.FormatInt(exp, 10) +
		"&sig=" + partnerReportSignature(scanID, exp)
}

// partnerUser resolves the platform user to a local shadow account. Users.email
// and users.google_sub are UNIQUE, so a synthetic address keeps a later Google
// login with the real email from colliding with this account.
func partnerUser(externalID string) (string, error) {
	var uid string
	err := db.QueryRow(`SELECT user_id FROM partner_accounts WHERE platform=? AND external_user_id=?`,
		partnerPlatform, externalID).Scan(&uid)
	if err == nil {
		return uid, nil
	}
	email := partnerPlatform + "+" + sanitizeEmailLocal(externalID) + "@partner.invalid"
	uid = "u_partner_" + randomHex(12)
	if _, err := db.Exec(`INSERT OR IGNORE INTO users (id, email, name) VALUES (?,?,?)`,
		uid, email, "Dalang "+externalID); err != nil {
		return "", err
	}
	// A concurrent reserve may have inserted the same external user first.
	db.QueryRow(`SELECT id FROM users WHERE email=?`, email).Scan(&uid)
	if _, err := db.Exec(`INSERT OR IGNORE INTO partner_accounts (platform, external_user_id, user_id) VALUES (?,?,?)`,
		partnerPlatform, externalID, uid); err != nil {
		return "", err
	}
	db.QueryRow(`SELECT user_id FROM partner_accounts WHERE platform=? AND external_user_id=?`,
		partnerPlatform, externalID).Scan(&uid)
	return uid, nil
}

func sanitizeEmailLocal(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	if b.Len() == 0 {
		return "unknown"
	}
	return b.String()
}

// partnerTargetBlocked refuses platform-owned hosts: scanning them would expose
// the platform's own edge, and the Owner decided .svc.dalang.io is out of scope.
func partnerTargetBlocked(domain string) bool {
	for _, host := range []string{"dalang.io", "netsekurity.com"} {
		if domain == host || strings.HasSuffix(domain, "."+host) {
			return true
		}
	}
	return false
}

type partnerReserveRequest struct {
	ExternalOrderID string `json:"external_order_id"`
	ExternalUserID  string `json:"external_user_id"`
	Domain          string `json:"domain"`
	ProofType       string `json:"proof_type"`
	ProofRef        string `json:"proof_ref"`
	TTLSeconds      int    `json:"ttl_seconds"`
	IdempotencyKey  string `json:"idempotency_key"`
}

type partnerScanView struct {
	ScanID        string         `json:"scan_id"`
	Domain        string         `json:"domain,omitempty"`
	Status        string         `json:"status"`
	Attempts      int            `json:"attempts"`
	DeadlineAt    string         `json:"deadline_at,omitempty"`
	ReportReady   bool           `json:"report_ready"`
	ReportURL     string         `json:"report_url,omitempty"`
	FailureReason string         `json:"failure_reason,omitempty"`
	Findings      map[string]int `json:"findings,omitempty"`
}

func partnerView(scanID string) (partnerScanView, error) {
	var v partnerScanView
	var reportRef, deadline, failure string
	var reported int
	var crit, high, med, low, info int
	err := db.QueryRow(`SELECT p.id, COALESCE(d.domain,''), p.status, p.attempts, COALESCE(p.deadline_at,''),
			COALESCE(p.report_ref,''), COALESCE(p.failure_reason,''), p.findings_reported,
			p.findings_critical, p.findings_high, p.findings_medium, p.findings_low, p.findings_info
		FROM pentests p LEFT JOIN domains d ON d.id=p.domain_id
		WHERE p.id=? AND p.billing_source='partner'`, scanID).
		Scan(&v.ScanID, &v.Domain, &v.Status, &v.Attempts, &deadline, &reportRef, &failure,
			&reported, &crit, &high, &med, &low, &info)
	if err != nil {
		return v, err
	}
	v.DeadlineAt = deadline
	v.FailureReason = failure
	v.ReportReady = v.Status == "completed" && reportRef != ""
	if v.ReportReady {
		v.ReportURL = partnerReportURL(scanID)
	}
	if reported == 1 {
		v.Findings = map[string]int{"critical": crit, "high": high, "medium": med, "low": low, "info": info}
	}
	return v, nil
}

// POST /api/partner/v1/scans/reserve
func handlePartnerReserve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		partnerErr(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "POST only")
		return
	}
	if !partnerAuthOK(w, r) {
		return
	}
	var req partnerReserveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		partnerErr(w, http.StatusBadRequest, "PARTNER_BODY_INVALID", "invalid JSON")
		return
	}
	req.ExternalOrderID = strings.TrimSpace(req.ExternalOrderID)
	req.ExternalUserID = strings.TrimSpace(req.ExternalUserID)
	req.ProofType = strings.ToLower(strings.TrimSpace(req.ProofType))
	req.IdempotencyKey = strings.TrimSpace(req.IdempotencyKey)
	if req.ExternalOrderID == "" || req.ExternalUserID == "" {
		partnerErr(w, http.StatusBadRequest, "PARTNER_INPUT", "external_order_id and external_user_id are required")
		return
	}
	if req.IdempotencyKey == "" {
		partnerErr(w, http.StatusBadRequest, "PARTNER_IDEMPOTENCY_REQUIRED", "idempotency_key is required")
		return
	}
	// Replay: same key returns the same scan, but the *current* state — a retry
	// after the scan progressed must not be told it is still reserved.
	var cached string
	if err := db.QueryRow(`SELECT response_json FROM partner_requests WHERE idempotency_key=?`,
		req.IdempotencyKey).Scan(&cached); err == nil {
		var cv partnerScanView
		if json.Unmarshal([]byte(cached), &cv) == nil && cv.ScanID != "" {
			if fresh, err := partnerView(cv.ScanID); err == nil {
				partnerJSON(w, http.StatusOK, fresh)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, cached)
		return
	}
	// A retried order (new idempotency key) must still map to one scan.
	if existing, ok := partnerScanByOrder(req.ExternalOrderID); ok {
		partnerJSON(w, http.StatusOK, existing)
		return
	}

	domain := normalizeDomain(req.Domain)
	if !isValidDomain(domain) {
		partnerErr(w, http.StatusBadRequest, "PARTNER_DOMAIN_INVALID", "domain is not a valid hostname")
		return
	}
	if partnerTargetBlocked(domain) {
		partnerErr(w, http.StatusBadRequest, "TARGET_NOT_ALLOWED", "platform-owned hostnames cannot be scanned")
		return
	}
	if req.ProofType != partnerProofTXT && req.ProofType != partnerProofCloudflare && req.ProofType != partnerProofRegistrar {
		partnerErr(w, http.StatusBadRequest, "PARTNER_PROOF_INVALID", "proof_type must be txt or cloudflare")
		return
	}

	userID, err := partnerUser(req.ExternalUserID)
	if err != nil {
		partnerErr(w, http.StatusInternalServerError, "PARTNER_ACCOUNT", "cannot map the platform user")
		return
	}

	var domainID int64
	switch req.ProofType {
	case partnerProofTXT:
		// Only a TXT the customer actually published counts here.
		err = db.QueryRow(`SELECT id FROM domains WHERE user_id=? AND domain=? AND status='verified'`,
			userID, domain).Scan(&domainID)
		if err != nil {
			partnerErr(w, http.StatusConflict, "DOMAIN_NOT_VERIFIED",
				"add the _netsekurity TXT record and verify the domain first")
			return
		}
	case partnerProofCloudflare, partnerProofRegistrar:
		// The platform holds a Cloudflare-activated hostname, or sold the domain
		// at the registrar; the HMAC proves the assertion came from the platform.
		verifiedBy := "partner"
		if req.ProofType == partnerProofRegistrar {
			verifiedBy = "registrar"
		}
		tok := "partner-" + req.ExternalOrderID
		if _, err := db.Exec(`INSERT INTO domains (user_id, domain, txt_verification_token, status, verified_at, verified_by, verified_ref)
			VALUES (?,?,?,'verified',?,?,?)
			ON CONFLICT(user_id, domain) DO UPDATE SET status='verified', verified_at=excluded.verified_at,
				verified_by=excluded.verified_by, verified_ref=excluded.verified_ref`,
			userID, domain, tok, now(), verifiedBy, req.ProofRef); err != nil {
			partnerErr(w, http.StatusInternalServerError, "PARTNER_DOMAIN", "cannot record the domain")
			return
		}
		db.QueryRow(`SELECT id FROM domains WHERE user_id=? AND domain=?`, userID, domain).Scan(&domainID)
	}

	var inflight int
	db.QueryRow(`SELECT COUNT(*) FROM pentests WHERE user_id=? AND status IN ('reserved','queued','running')`,
		userID).Scan(&inflight)
	if inflight > 0 {
		partnerErr(w, http.StatusConflict, "SCAN_IN_PROGRESS", "this user already has a scan in flight")
		return
	}

	ttl := time.Duration(req.TTLSeconds) * time.Second
	if ttl <= 0 {
		ttl = partnerReserveTTL
	}
	if ttl > partnerReserveTTLMax {
		ttl = partnerReserveTTLMax
	}
	if ttl < 5*time.Minute {
		ttl = 5 * time.Minute
	}

	pid := "pt_" + randomHex(8)
	deadline := time.Now().UTC().Add(ttl).Format(time.RFC3339)
	if _, err := db.Exec(`INSERT INTO pentests
		(id, user_id, domain_id, mode, status, origin, billing_source, external_order_id, attempts, deadline_at)
		VALUES (?,?,?,'standard','reserved','dalang','partner',?,0,?)`,
		pid, userID, domainID, req.ExternalOrderID, deadline); err != nil {
		partnerErr(w, http.StatusInternalServerError, "PARTNER_SCAN", "cannot create the scan")
		return
	}

	view, err := partnerView(pid)
	if err != nil {
		partnerErr(w, http.StatusInternalServerError, "PARTNER_SCAN", "scan created but cannot be read")
		return
	}
	body, _ := json.Marshal(view)
	db.Exec(`INSERT OR IGNORE INTO partner_requests (idempotency_key, platform, response_json) VALUES (?,?,?)`,
		req.IdempotencyKey, partnerPlatform, string(body))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(body)
}

func partnerScanByOrder(orderID string) (partnerScanView, bool) {
	var id string
	if err := db.QueryRow(`SELECT id FROM pentests WHERE billing_source='partner' AND external_order_id=?`,
		orderID).Scan(&id); err != nil {
		return partnerScanView{}, false
	}
	v, err := partnerView(id)
	return v, err == nil
}

// handlePartnerScan routes /api/partner/v1/scans/{id}[/confirm|/release|/report].
func handlePartnerScan(w http.ResponseWriter, r *http.Request) {
	rest := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/partner/v1/scans/"), "/")
	parts := strings.Split(rest, "/")
	if len(parts) == 2 && parts[1] == "report" {
		handlePartnerReport(w, r, parts[0])
		return
	}
	if !partnerAuthOK(w, r) {
		return
	}
	if len(parts) == 0 || parts[0] == "" {
		partnerErr(w, http.StatusNotFound, "NOT_FOUND", "scan id required")
		return
	}
	id := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}
	switch action {
	case "":
		if r.Method != http.MethodGet {
			partnerErr(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET only")
			return
		}
		view, err := partnerView(id)
		if err != nil {
			partnerErr(w, http.StatusNotFound, "SCAN_NOT_FOUND", "scan not found")
			return
		}
		partnerJSON(w, http.StatusOK, view)
	case "confirm":
		if r.Method != http.MethodPost {
			partnerErr(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "POST only")
			return
		}
		partnerConfirm(w, id)
	case "release":
		if r.Method != http.MethodPost {
			partnerErr(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "POST only")
			return
		}
		partnerRelease(w, id)
	default:
		partnerErr(w, http.StatusNotFound, "NOT_FOUND", "unknown action")
	}
}

func partnerConfirm(w http.ResponseWriter, id string) {
	var status string
	if err := db.QueryRow(`SELECT status FROM pentests WHERE id=? AND billing_source='partner'`, id).
		Scan(&status); err != nil {
		partnerErr(w, http.StatusNotFound, "SCAN_NOT_FOUND", "scan not found")
		return
	}
	switch status {
	case "reserved":
		nowS := now()
		deadline := time.Now().UTC().Add(partnerScanDeadline).Format(time.RFC3339)
		if _, err := db.Exec(`UPDATE pentests SET status='queued', queued_at=?, deadline_at=?
			WHERE id=? AND status='reserved'`, nowS, deadline, id); err != nil {
			partnerErr(w, http.StatusInternalServerError, "PARTNER_SCAN", "cannot confirm the scan")
			return
		}
	case "queued", "running", "completed", "failed":
		// Idempotent: a retried confirm reports the current state. A released
		// (cancelled) reservation is not confirmable — that order was abandoned.
	default:
		partnerErr(w, http.StatusConflict, "SCAN_STATE", "scan is not in a confirmable state")
		return
	}
	view, err := partnerView(id)
	if err != nil {
		partnerErr(w, http.StatusInternalServerError, "PARTNER_SCAN", "cannot read the scan")
		return
	}
	partnerJSON(w, http.StatusOK, view)
}

func partnerRelease(w http.ResponseWriter, id string) {
	var status string
	if err := db.QueryRow(`SELECT status FROM pentests WHERE id=? AND billing_source='partner'`, id).
		Scan(&status); err != nil {
		partnerErr(w, http.StatusNotFound, "SCAN_NOT_FOUND", "scan not found")
		return
	}
	switch status {
	case "reserved":
		if _, err := db.Exec(`UPDATE pentests SET status='cancelled', completed_at=?, failure_reason='released by platform'
			WHERE id=? AND status='reserved'`, now(), id); err != nil {
			partnerErr(w, http.StatusInternalServerError, "PARTNER_SCAN", "cannot release the scan")
			return
		}
	case "cancelled":
		// idempotent
	case "queued", "running":
		partnerErr(w, http.StatusConflict, "SCAN_STATE", "scan is already paid; use the compensation path")
		return
	default:
		partnerErr(w, http.StatusConflict, "SCAN_STATE", "scan cannot be released")
		return
	}
	view, err := partnerView(id)
	if err != nil {
		partnerErr(w, http.StatusInternalServerError, "PARTNER_SCAN", "cannot read the scan")
		return
	}
	partnerJSON(w, http.StatusOK, view)
}

// GET /api/partner/v1/scans/{id}/report?exp=&sig=
// Signed URL: api.dalang.io passes this URL to the customer, so it must work
// without the HMAC headers while still being unforgeable and short-lived.
func handlePartnerReport(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodGet {
		partnerErr(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET only")
		return
	}
	if partnerSecret() == "" {
		partnerErr(w, http.StatusServiceUnavailable, "PARTNER_API_DISABLED", "partner API is not configured")
		return
	}
	exp, err := strconv.ParseInt(r.URL.Query().Get("exp"), 10, 64)
	if err != nil || time.Now().Unix() > exp {
		partnerErr(w, http.StatusForbidden, "REPORT_LINK_EXPIRED", "report link expired")
		return
	}
	if !secureCompare(r.URL.Query().Get("sig"), partnerReportSignature(id, exp)) {
		partnerErr(w, http.StatusForbidden, "REPORT_LINK_INVALID", "invalid report link")
		return
	}
	var ref string
	var status string
	if err := db.QueryRow(`SELECT COALESCE(report_ref,''), status FROM pentests WHERE id=? AND billing_source='partner'`, id).
		Scan(&ref, &status); err != nil || status != "completed" || ref == "" {
		partnerErr(w, http.StatusNotFound, "REPORT_NOT_READY", "report not ready")
		return
	}
	if !reportNameOK.MatchString(ref) {
		partnerErr(w, http.StatusInternalServerError, "REPORT_INVALID", "report reference is invalid")
		return
	}
	w.Header().Set("Cache-Control", "no-store, private")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	http.ServeFile(w, r, filepath.Join(reportsDir(), ref))
}
