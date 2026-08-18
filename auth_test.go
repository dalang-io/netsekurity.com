package main

import (
	"os"
	"testing"
)

func TestIsAdmin_SuperAdminEmail(t *testing.T) {
	os.Setenv("SUPER_ADMIN_EMAIL", "hans@dalang.io")
	if !isAdmin("hans@dalang.io") {
		t.Errorf("super admin email must be admin")
	}
	if isAdmin("other@dalang.io") {
		t.Errorf("unknown email must not be admin")
	}
}

func TestIsAdmin_RoleColumn(t *testing.T) {
	os.Setenv("SUPER_ADMIN_EMAIL", "hans@dalang.io")
	db.Exec(`INSERT OR IGNORE INTO users (id, email, role) VALUES ('u_adm_test', 'adm@test.local', 'admin')`)
	db.Exec(`INSERT OR IGNORE INTO users (id, email, role) VALUES ('u_usr_test', 'usr@test.local', 'user')`)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM users WHERE id IN ('u_adm_test','u_usr_test')`)
	})
	if !isAdmin("adm@test.local") {
		t.Errorf("user with role=admin must be admin")
	}
	if isAdmin("usr@test.local") {
		t.Errorf("user with role=user must not be admin")
	}
}

func TestInitDB_SeedsSuperAdmin(t *testing.T) {
	var role string
	err := db.QueryRow(`SELECT role FROM users WHERE email=?`, getenv("SUPER_ADMIN_EMAIL", "hans@dalang.io")).Scan(&role)
	if err != nil {
		t.Fatalf("super admin row missing: %v", err)
	}
	if role != "admin" {
		t.Errorf("super admin role = %q, want admin", role)
	}
}