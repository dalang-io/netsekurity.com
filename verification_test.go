package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"sync"
	"testing"
)

func stubControl(t *testing.T, result error) {
	t.Helper()
	old := checkDomainControl
	checkDomainControl = func(context.Context, string, string, string) error { return result }
	t.Cleanup(func() { checkDomainControl = old })
}

func TestVerificationRejectsSpecialNetworks(t *testing.T) {
	for _, raw := range []string{"127.0.0.1", "10.0.0.1", "172.16.0.1", "192.168.1.1", "169.254.169.254", "100.100.100.200", "0.0.0.0", "198.18.0.1", "192.0.2.5", "224.0.0.1", "255.255.255.255", "::1", "::ffff:127.0.0.1", "::ffff:10.1.1.1", "fc00::1", "fe80::1", "64:ff9b::7f00:1", "2002:7f00:1::", "2001:db8::1"} {
		if publicVerificationIP(netip.MustParseAddr(raw)) {
			t.Errorf("allowed %s", raw)
		}
	}
	for _, raw := range []string{"1.1.1.1", "8.8.8.8", "2606:4700:4700::1111"} {
		if !publicVerificationIP(netip.MustParseAddr(raw)) {
			t.Errorf("blocked %s", raw)
		}
	}
	for _, target := range []string{"127.0.0.1:443", "[::1]:443", "169.254.169.254:443", "1.1.1.1:80"} {
		if conn, err := verificationDial(context.Background(), "tcp", target); err == nil {
			conn.Close()
			t.Errorf("dial allowed %s", target)
		}
	}
}

func TestVerificationMetaLocationAndParsing(t *testing.T) {
	for _, tc := range []struct {
		body  string
		valid bool
	}{
		{`<html><head><meta name="netsekurity-verification" content="abc"></head><body></body></html>`, true},
		{`<HEAD><META content='abc' name='netsekurity-verification' /></HEAD>`, true},
		{`<head><meta name="netsekurity-verification" content="wrong"></head>`, false},
		{`<head></head><body><meta name="netsekurity-verification" content="abc"></body>`, false},
		{`<head><!-- <meta name="netsekurity-verification" content="abc"> --></head>`, false},
		{`<head><script>var tag='<meta name="netsekurity-verification" content="abc">';</script></head>`, false},
		{`<head><meta name="netsekurity-verification" content="wrong" content="abc"></head>`, false},
		{`<head><template><meta name="netsekurity-verification" content="abc"></template></head>`, false},
	} {
		if got := verificationMetaMatches(tc.body, "abc"); got != tc.valid {
			t.Errorf("%q = %v want %v", tc.body, got, tc.valid)
		}
	}
}

func TestHTTPSVerificationProof(t *testing.T) {
	for _, tc := range []struct {
		name, method, body string
		status             int
		valid              bool
	}{
		{"file", "file", "abc\n", 200, true},
		{"wrong account token", "file", "xyz", 200, false},
		{"token buried in HTML", "file", "<html>abc</html>", 200, false},
		{"too large", "file", strings.Repeat("x", 4097), 200, false},
		{"redirect", "file", "abc", 302, false},
		{"not found", "file", "abc", 404, false},
		{"meta", "meta", `<head><meta name="netsekurity-verification" content="abc"></head>`, 200, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				want := "/"
				if tc.method == "file" {
					want = verificationPath
				}
				if r.URL.Path != want {
					t.Errorf("path %s", r.URL.Path)
				}
				if r.Header.Get("User-Agent") != "NetSekurity Verification" {
					t.Error("missing verifier identity")
				}
				w.Header().Set("Location", "https://other.example.com/")
				w.WriteHeader(tc.status)
				fmt.Fprint(w, tc.body)
			}))
			defer srv.Close()
			client := srv.Client()
			client.CheckRedirect = verificationClient().CheckRedirect
			err := verifyHTTPControl(context.Background(), client, strings.TrimPrefix(srv.URL, "https://"), "abc", tc.method)
			if (err == nil) != tc.valid {
				t.Errorf("error=%v valid=%v", err, tc.valid)
			}
		})
	}
}

func TestVerificationRejectsUntrustedTLS(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "abc") }))
	defer srv.Close()
	// An ordinary client trusts the system roots, not the test server's self-signed
	// certificate. The production transport also leaves TLS verification enabled.
	if err := verifyHTTPControl(context.Background(), &http.Client{}, strings.TrimPrefix(srv.URL, "https://"), "abc", "file"); err == nil {
		t.Fatal("accepted an untrusted certificate")
	}
	transport := verificationClient().Transport.(*http.Transport)
	if transport.Proxy != nil || (transport.TLSClientConfig != nil && transport.TLSClientConfig.InsecureSkipVerify) {
		t.Fatal("verification bypasses destination or certificate checks")
	}
}

func verificationRequest(t *testing.T, uid, method, path, body string, handler http.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if uid != "" {
		tok, err := issueJWT(uid, uid+"@test.local")
		if err != nil {
			t.Fatal(err)
		}
		r.AddCookie(&http.Cookie{Name: "nsk_session", Value: tok})
	}
	rr := httptest.NewRecorder()
	handler(rr, r)
	return rr
}

func TestVerificationMethodsPersistAndDownloadIsAccountBound(t *testing.T) {
	const uid = "u_verify_options"
	apiFixture(t, uid)
	db.Exec(`INSERT INTO domains (user_id,domain,txt_verification_token,status) VALUES (?,'verify.example.com','proof-for-owner','pending')`, uid)
	old := checkDomainControl
	t.Cleanup(func() { checkDomainControl = old })
	for _, method := range []string{"txt", "file", "meta"} {
		checkDomainControl = func(_ context.Context, domain, token, gotMethod string) error {
			if domain != "verify.example.com" || token != "proof-for-owner" || gotMethod != method {
				t.Errorf("wrong proof %s %s %s", domain, token, gotMethod)
			}
			return nil
		}
		rr := verificationRequest(t, uid, "POST", "/api/domains/verify", "domain=verify.example.com&method="+method, handleVerifyDomain)
		if rr.Code != 200 {
			t.Fatalf("%s: %d %s", method, rr.Code, rr.Body)
		}
		var status, stored string
		db.QueryRow(`SELECT status,verified_by FROM domains WHERE user_id=?`, uid).Scan(&status, &stored)
		if status != "verified" || stored != method {
			t.Errorf("%s %s", status, stored)
		}
	}
	checkDomainControl = func(context.Context, string, string, string) error { return errors.New("wrong proof") }
	if rr := verificationRequest(t, uid, "POST", "/api/domains/verify", "domain=verify.example.com&method=file", handleVerifyDomain); rr.Code != 422 {
		t.Errorf("wrong token: %d", rr.Code)
	}
	if rr := verificationRequest(t, uid, "POST", "/api/domains/verify", "domain=verify.example.com&method=email", handleVerifyDomain); rr.Code != 400 {
		t.Errorf("invalid method: %d", rr.Code)
	}
	for _, who := range []string{uid, "u_someone_else", ""} {
		rr := verificationRequest(t, who, "GET", "/api/domains/verification-file?domain=verify.example.com", "", handleVerificationFile)
		if who == uid {
			if rr.Code != 200 || rr.Body.String() != "proof-for-owner\n" || !strings.Contains(rr.Header().Get("Content-Disposition"), "attachment") {
				t.Errorf("download: %d %s", rr.Code, rr.Body)
			}
		} else if rr.Code == 200 || strings.Contains(rr.Body.String(), "proof-for-owner") {
			t.Error("proof leaked")
		}
	}
	if rr := verificationRequest(t, "u_someone_else", "POST", "/api/domains/verify", "domain=verify.example.com&method=meta", handleVerifyDomain); rr.Code != 404 {
		t.Errorf("other account verify: %d", rr.Code)
	}
}

func TestStartRequiresPermissionAndCurrentProofWithoutSpendingOnFailure(t *testing.T) {
	const uid = "u_verify_start"
	token, mux := apiFixture(t, uid)
	db.Exec(`INSERT INTO credit_balance (user_id,balance) VALUES (?,3)`, uid)
	db.Exec(`INSERT INTO domains (user_id,domain,txt_verification_token,status,verified_by) VALUES (?,'start.example.com','abc','verified','file')`, uid)
	stubControl(t, errors.New("proof removed"))
	for _, api := range []bool{false, true} {
		call := func(body string) *httptest.ResponseRecorder {
			if api {
				return apiCall(mux, "POST", "/api/v1/assessments", token, body)
			}
			return verificationRequest(t, uid, "POST", "/api/pentests/start", body, handleStartPentest)
		}
		if rr := call("domain=start.example.com"); rr.Code != 400 {
			t.Errorf("no permission: %d", rr.Code)
		}
		db.Exec(`UPDATE domains SET status='verified' WHERE user_id=?`, uid)
		if rr := call("domain=start.example.com&authorised=true"); rr.Code != 409 {
			t.Errorf("proof removed: %d", rr.Code)
		}
		var balance float64
		var count int
		db.QueryRow(`SELECT balance FROM credit_balance WHERE user_id=?`, uid).Scan(&balance)
		db.QueryRow(`SELECT COUNT(*) FROM pentests WHERE user_id=?`, uid).Scan(&count)
		if balance != 3 || count != 0 {
			t.Errorf("failed start spent credits: %v jobs=%d", balance, count)
		}
	}
	checkDomainControl = func(context.Context, string, string, string) error { return nil }
	db.Exec(`UPDATE domains SET status='verified' WHERE user_id=?`, uid)
	r := httptest.NewRequest("POST", "/api/v1/assessments", strings.NewReader(`{"domain":"start.example.com","authorised":true}`))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-API-Token", token)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, r)
	if rr.Code != 200 {
		t.Fatalf("valid JSON start: %d %s", rr.Code, rr.Body)
	}
	var auth string
	db.QueryRow(`SELECT authorized_at FROM pentests WHERE user_id=?`, uid).Scan(&auth)
	if auth == "" {
		t.Error("permission not recorded")
	}
}

func TestConcurrentStartsChargeOnce(t *testing.T) {
	const uid = "u_verify_concurrent"
	apiFixture(t, uid)
	stubControl(t, nil)
	db.Exec(`INSERT INTO credit_balance (user_id,balance) VALUES (?,3)`, uid)
	db.Exec(`INSERT INTO domains (user_id,domain,txt_verification_token,status) VALUES (?,'concurrent.example.com','abc','verified')`, uid)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			queueVerifiedAssessment(context.Background(), uid, "concurrent.example.com", "standard", 1)
		}()
	}
	wg.Wait()
	var balance float64
	var jobs, ledger int
	db.QueryRow(`SELECT balance FROM credit_balance WHERE user_id=?`, uid).Scan(&balance)
	db.QueryRow(`SELECT COUNT(*) FROM pentests WHERE user_id=?`, uid).Scan(&jobs)
	db.QueryRow(`SELECT COUNT(*) FROM credit_transactions WHERE user_id=?`, uid).Scan(&ledger)
	if balance != 2 || jobs != 1 || ledger != 1 {
		t.Errorf("balance=%v jobs=%d ledger=%d", balance, jobs, ledger)
	}
}

func TestWorkerDoesNotReceiveTargetAfterProofIsRemoved(t *testing.T) {
	const uid = "u_worker_recheck"
	apiFixture(t, uid)
	stubControl(t, nil)
	t.Setenv("BOT_AUTH_TOKEN", "worker-recheck-test")
	db.Exec(`INSERT INTO credit_balance(user_id,balance) VALUES (?,2)`, uid)
	db.Exec(`INSERT INTO domains(user_id,domain,txt_verification_token,status) VALUES (?,'worker.example.com','abc','verified')`, uid)
	id, err := queueVerifiedAssessment(context.Background(), uid, "worker.example.com", "standard", 1)
	if err != nil {
		t.Fatal(err)
	}
	checkDomainControl = func(context.Context, string, string, string) error { return errors.New("proof removed") }
	r := httptest.NewRequest("POST", "/api/pentests/worker/claim", nil)
	r.Header.Set("X-Bot-Token", "worker-recheck-test")
	rr := httptest.NewRecorder()
	handleWorkerClaim(rr, r)
	if rr.Code != 200 || !strings.Contains(rr.Body.String(), `"empty":true`) || strings.Contains(rr.Body.String(), "worker.example.com") {
		t.Fatalf("target handed to worker: %d %s", rr.Code, rr.Body)
	}
	if err := failUnverifiedQueuedAssessment(context.Background(), id, uid, "standard"); err != nil {
		t.Fatal(err)
	}
	var balance float64
	var status string
	db.QueryRow(`SELECT balance FROM credit_balance WHERE user_id=?`, uid).Scan(&balance)
	db.QueryRow(`SELECT status FROM pentests WHERE id=?`, id).Scan(&status)
	if balance != 2 || status != "failed" {
		t.Errorf("balance=%v status=%s; expected one refund", balance, status)
	}
}

func TestVerificationRejectsMalformedHostnames(t *testing.T) {
	for _, host := range []string{"127.0.0.1", "[::1]", "evil.com\\@127.0.0.1", "bad.com\n", "éxample.com", "a.<script>", "a.%00"} {
		if isValidDomain(host) {
			t.Errorf("accepted %q", host)
		}
	}
}
