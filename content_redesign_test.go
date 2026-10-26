package main

import (
	"context"
	"strings"
	"testing"
)

func TestHomeFollowsTheOwnersDecisionPath(t *testing.T) {
	for _, lang := range []string{"en", "id"} {
		path := pathFor("home", lang)
		page := renderPublic(t, path)
		previous := -1
		for _, id := range []string{"audiences", "report", "how", "what", "trust", "pricing", "tiers", "faq"} {
			pos := strings.Index(page, `id="`+id+`"`)
			if pos < 0 || pos <= previous {
				t.Errorf("%s: missing/out-of-order section %s", path, id)
			}
			previous = pos
		}
		main := strings.Index(page, "<main")
		hero := page[main:strings.Index(page, `id="audiences"`)]
		for _, jargon := range []string{"CVSS", "CWE", "OWASP", "whitebox", "agent harness"} {
			if strings.Contains(hero, jargon) {
				t.Errorf("%s: hero requires technical vocabulary %s", path, jargon)
			}
		}
		if strings.Contains(page, "curl -") {
			t.Errorf("%s: API code still dominates the homepage", path)
		}
		p, _ := pageByKey("home")
		v := buildSiteView(p, langByCode(lang), false, loadSitePricing(lang))
		if len(v.Nav) != 4 {
			t.Fatalf("%s has %d main menu entries", lang, len(v.Nav))
		}
		for i, key := range []string{"how", "sample", "pricing", "enterprise"} {
			if v.Nav[i].Key != key {
				t.Errorf("%s menu %d = %s", lang, i, v.Nav[i].Key)
			}
		}
		if !strings.Contains(page, pathFor("start", lang)) {
			t.Errorf("%s has no getting-started guide", path)
		}
	}
}

func TestGettingStartedAndVerificationClaimsMatchCapabilities(t *testing.T) {
	for _, lang := range []string{"en", "id"} {
		page := visibleText(renderPublic(t, pathFor("start", lang)))
		for _, word := range []string{"DNS", "meta tag", "/.well-known/netsekurity-verification.txt", "HTTPS", "PDF"} {
			if !strings.Contains(page, word) {
				t.Errorf("%s guide lacks %s", lang, word)
			}
		}
		for _, key := range []string{"home", "trust", "enterprise"} {
			page := visibleText(renderPublic(t, pathFor(key, lang)))
			if !strings.Contains(page, "meta tag") {
				t.Errorf("%s/%s lacks verification choices", lang, key)
			}
			for _, stale := range []string{"No record, no scan", "No DNS record", "Tanpa record DNS", "Tanpa record, tanpa"} {
				if strings.Contains(page, stale) {
					t.Errorf("%s/%s retains %q", lang, key, stale)
				}
			}
		}
	}
}

func TestResultGuidanceDistinguishesMissingAndZero(t *testing.T) {
	missingEN, missingID := resultNextStep(0, 0, 0, 0, 0)
	zeroEN, zeroID := resultNextStep(1, 0, 0, 0, 0)
	if missingEN == zeroEN || missingID == zeroID {
		t.Fatal("unknown and zero results have the same guidance")
	}
	for _, tc := range []struct {
		reported, critical, high, medium, low int
		en, id                                string
	}{
		{0, 0, 0, 0, 0, "unavailable", "belum tersedia"},
		{1, 1, 0, 0, 0, "promptly", "segera"},
		{1, 0, 1, 0, 0, "promptly", "segera"},
		{1, 0, 0, 1, 0, "schedule repairs", "jadwalkan"},
		{1, 0, 0, 0, 1, "schedule repairs", "jadwalkan"},
		{1, 0, 0, 0, 0, "does not establish", "tidak membuktikan"},
	} {
		en, id := resultNextStep(tc.reported, tc.critical, tc.high, tc.medium, tc.low)
		if !strings.Contains(en, tc.en) || !strings.Contains(id, tc.id) {
			t.Errorf("unexpected guidance: %s / %s", en, id)
		}
	}
}

func TestQueueFailureRollsBackCreditAndLedger(t *testing.T) {
	const uid = "u_queue_rollback"
	apiFixture(t, uid)
	stubControl(t, nil)
	db.Exec(`INSERT INTO credit_balance (user_id,balance) VALUES (?,2)`, uid)
	db.Exec(`INSERT INTO domains (user_id,domain,txt_verification_token,status) VALUES (?,'rollback.example.com','abc','verified')`, uid)
	_, err := db.Exec(`CREATE TRIGGER reject_test_queue BEFORE INSERT ON pentests WHEN NEW.user_id='u_queue_rollback' BEGIN SELECT RAISE(FAIL,'simulated queue failure'); END`)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Exec(`DROP TRIGGER IF EXISTS reject_test_queue`) })
	if _, err := queueVerifiedAssessment(context.Background(), uid, "rollback.example.com", "standard", 1); err == nil {
		t.Fatal("failed queue reported success")
	}
	var balance float64
	var ledger int
	db.QueryRow(`SELECT balance FROM credit_balance WHERE user_id=?`, uid).Scan(&balance)
	db.QueryRow(`SELECT COUNT(*) FROM credit_transactions WHERE user_id=?`, uid).Scan(&ledger)
	if balance != 2 || ledger != 0 {
		t.Errorf("queue failure spent credits: balance=%v ledger=%d", balance, ledger)
	}
}
