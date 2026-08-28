package main

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestParseFindings(t *testing.T) {
	f, ok := parseFindings("critical=0,high=2,medium=7,low=3,info=9")
	if !ok {
		t.Fatal("well-formed summary rejected")
	}
	for k, want := range map[string]int{"critical": 0, "high": 2, "medium": 7, "low": 3, "info": 9} {
		if f[k] != want {
			t.Errorf("%s = %d, want %d", k, f[k], want)
		}
	}

	// Whitespace, case and unknown keys are tolerated.
	f, ok = parseFindings("  HIGH = 1 , bogus=5, low=2 ")
	if !ok || f["high"] != 1 || f["low"] != 2 {
		t.Errorf("lenient parse = %v ok=%v", f, ok)
	}
	if _, present := f["bogus"]; present {
		t.Error("unknown severity key was kept")
	}
}

// A missing breakdown must never be reported as a clean scan.
func TestParseFindingsRejectsUnusable(t *testing.T) {
	for _, in := range []string{"", "   ", "garbage", "high", "bogus=1", "high=-2", "high=abc"} {
		if f, ok := parseFindings(in); ok {
			t.Errorf("parseFindings(%q) = %v, want not-ok", in, f)
		}
	}
}

func TestSeverityChipsDistinguishCleanFromUnknown(t *testing.T) {
	// A scan that genuinely found nothing risk-bearing says so.
	clean := severityChips(map[string]int{"critical": 0, "high": 0, "medium": 0, "low": 0, "info": 4})
	if !strings.Contains(clean, "no findings") {
		t.Error("a zero-risk scan should say 'no findings'")
	}
	if !strings.Contains(clean, "4 info") {
		t.Error("info count should still show")
	}

	// Findings render most severe first, and zero rows are omitted.
	out := severityChips(map[string]int{"critical": 1, "high": 2, "medium": 0, "low": 3, "info": 0})
	if strings.Contains(out, "no findings") {
		t.Error("a scan with findings must not claim 'no findings'")
	}
	if strings.Contains(out, "medium") || strings.Contains(out, "info") {
		t.Error("zero-count severities should be omitted")
	}
	ci, hi, li := strings.Index(out, "1 critical"), strings.Index(out, "2 high"), strings.Index(out, "3 low")
	if ci < 0 || hi < 0 || li < 0 {
		t.Fatalf("missing chips: %s", out)
	}
	if !(ci < hi && hi < li) {
		t.Error("chips are not ordered most-severe first")
	}
}

// End to end: the worker's string reaches the rendered chips intact.
func TestFindingsRoundTrip(t *testing.T) {
	f, ok := parseFindings("critical=0,high=2,medium=7,low=3,info=9")
	if !ok {
		t.Fatal("parse failed")
	}
	out := severityChips(f)
	for _, want := range []string{"2 high", "7 medium", "3 low", "9 info"} {
		if !strings.Contains(out, want) {
			t.Errorf("chips missing %q: %s", want, out)
		}
	}
	if strings.Contains(out, "critical") {
		t.Error("zero critical should not render")
	}
}

// Full path: the worker POSTs a report plus its severity summary, and the
// customer's dashboard card renders the counts.
func TestWorkerReportToDashboardCard(t *testing.T) {
	const uid = "u_findings_test"
	env["BOT_AUTH_TOKEN"] = "test-bot-token"
	defer delete(env, "BOT_AUTH_TOKEN")

	db.Exec(`INSERT OR IGNORE INTO users (id, email, name) VALUES (?, 'findings@test.local', 'F')`, uid)
	db.Exec(`INSERT OR IGNORE INTO domains (id, user_id, domain, txt_verification_token, status)
		VALUES (901, ?, 'findings.example.com', 'tok', 'verified')`, uid)
	db.Exec(`DELETE FROM pentests WHERE id='pt_findings'`)
	db.Exec(`INSERT INTO pentests (id, user_id, domain_id, mode, status) VALUES ('pt_findings', ?, 901, 'standard', 'running')`, uid)

	// Upload as the worker does: PDF + findings field.
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	mw.WriteField("pentest_id", "pt_findings")
	mw.WriteField("findings", "critical=0,high=2,medium=7,low=3,info=9")
	fw, _ := mw.CreateFormFile("report", "20260828-10:00-findings.example.com.pdf")
	fw.Write([]byte("%PDF-1.4 test"))
	mw.Close()

	os.MkdirAll(reportsDir(), 0o755)
	req := httptest.NewRequest("POST", "/api/pentests/worker/report", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("X-Bot-Token", "test-bot-token")
	rec := httptest.NewRecorder()
	handleWorkerReport(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("worker report: %d %s", rec.Code, rec.Body.String())
	}

	var reported, high, med int
	db.QueryRow(`SELECT findings_reported, findings_high, findings_medium FROM pentests WHERE id='pt_findings'`).
		Scan(&reported, &high, &med)
	if reported != 1 || high != 2 || med != 7 {
		t.Fatalf("stored reported=%d high=%d medium=%d", reported, high, med)
	}

	// Now the customer's list fragment.
	token, err := issueJWT(uid, "findings@test.local")
	if err != nil {
		t.Fatal(err)
	}
	lreq := httptest.NewRequest("GET", "/api/pentests/list", nil)
	lreq.AddCookie(&http.Cookie{Name: authCookie, Value: token})
	lrec := httptest.NewRecorder()
	handlePentestList(lrec, lreq)
	out := lrec.Body.String()
	for _, want := range []string{"findings.example.com", "2 high", "7 medium", "3 low", "9 info"} {
		if !strings.Contains(out, want) {
			t.Errorf("card missing %q", want)
		}
	}
	if strings.Contains(out, "severity breakdown is in the report") {
		t.Error("card claims the breakdown is unavailable despite counts being stored")
	}

	// A scan whose worker sent nothing must not look like a clean result.
	db.Exec(`UPDATE pentests SET findings_reported=0 WHERE id='pt_findings'`)
	lrec2 := httptest.NewRecorder()
	handlePentestList(lrec2, lreq)
	out2 := lrec2.Body.String()
	if strings.Contains(out2, "no findings") {
		t.Error("an unreported breakdown renders as 'no findings'")
	}
	if !strings.Contains(out2, "severity breakdown is in the report") {
		t.Error("expected the fallback note when no counts were reported")
	}
}
