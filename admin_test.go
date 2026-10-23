package main

import (
	"bytes"
	"strings"
	"testing"
)

// The console must never leak the synthetic address behind a Dalang scan.
func TestAdminTemplateLabelsDalangScans(t *testing.T) {
	const external = "a3ca32c3-4eb8-4f37-ab03-90f80f016207"
	data := suData{
		CurrentID: "u_super_admin",
		Users: []suUser{{
			ID: "u_partner_x", Email: "dalang+" + external + "@partner.invalid", PartnerUser: external,
		}},
		Domains: []suDomain{{
			ID: 1, Domain: "test.virus.ai.id", Status: "verified", PartnerUser: external,
		}},
		Pentests: []suPentest{{
			ID: "pt_x", Domain: "test.virus.ai.id", Status: "completed", Mode: "standard",
			PartnerUser: external, Origin: "dalang", ExternalOrder: "sc_x",
		}},
	}
	var buf bytes.Buffer
	if err := suTpl.ExecuteTemplate(&buf, "su", data); err != nil {
		t.Fatalf("render: %v", err)
	}
	html := buf.String()
	if strings.Contains(html, "partner.invalid") {
		t.Error("synthetic partner address leaked into the console")
	}
	if !strings.Contains(html, ">dalang<") {
		t.Error("missing dalang badge")
	}
	if !strings.Contains(html, "a3ca32c3…") {
		t.Error("missing shortened external id")
	}
	if !strings.Contains(html, "— Dalang —") {
		t.Error("partner row still offers role controls")
	}
}

func TestCustomerCountExcludesDalangAccounts(t *testing.T) {
	const uid = "u_partner_count_test"
	db.Exec(`INSERT OR REPLACE INTO users (id,email) VALUES (?, ?)`, uid, "dalang+count@partner.invalid")
	db.Exec(`INSERT OR REPLACE INTO partner_accounts (platform, external_user_id, user_id) VALUES ('dalang','count-test',?)`, uid)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM partner_accounts WHERE user_id=?`, uid)
		db.Exec(`DELETE FROM users WHERE id=?`, uid)
	})

	var all int
	db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&all)
	got, err := countCustomerUsers()
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if got != all-1 {
		t.Errorf("customer count = %d, want %d (synthetic account excluded)", got, all-1)
	}
}
