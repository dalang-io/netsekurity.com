package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB

// initDB opens the SQLite database and applies the schema (idempotent).
func initDB() error {
	path := getenv("DATABASE_PATH", "./data/netsekurity.db")
	if i := strings.LastIndex(path, "/"); i > 0 {
		if err := os.MkdirAll(path[:i], 0o755); err != nil {
			return fmt.Errorf("mkdir db dir: %w", err)
		}
	}
	var err error
	db, err = sql.Open("sqlite", path)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at TEXT)`,
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			name TEXT,
			picture TEXT,
			google_sub TEXT UNIQUE,
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS credit_balance (
			user_id TEXT PRIMARY KEY,
			balance REAL NOT NULL DEFAULT 0,
			updated_at TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS credit_transactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT NOT NULL,
			type TEXT NOT NULL,
			amount REAL NOT NULL,
			description TEXT,
			reference_id TEXT,
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS credit_packages (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			usd_price REAL NOT NULL,
			credits REAL NOT NULL,
			is_active INTEGER NOT NULL DEFAULT 1
		)`,
		`CREATE TABLE IF NOT EXISTS payments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT NOT NULL,
			external_id TEXT UNIQUE NOT NULL,
			xendit_invoice_id TEXT,
			url TEXT,
			package_id TEXT,
			amount_usd REAL NOT NULL,
			credits REAL NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			currency TEXT NOT NULL DEFAULT 'IDR',
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			paid_at TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS domains (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT NOT NULL,
			domain TEXT NOT NULL,
			txt_verification_token TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			verified_at TEXT,
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			UNIQUE(user_id, domain)
		)`,
		`CREATE TABLE IF NOT EXISTS pentests (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			domain_id INTEGER,
			mode TEXT NOT NULL DEFAULT 'standard',
			status TEXT NOT NULL DEFAULT 'queued',
			report_ref TEXT,
			started_at TEXT,
			completed_at TEXT,
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}

	// Seed credit packages (idempotent).
	packages := []struct {
		id, name     string
		usd, credits float64
	}{
		{"starter", "Starter", 50, 1},
		{"standard", "Standard", 100, 3},
		{"professional", "Professional", 500, 20},
		{"enterprise", "Enterprise", 1000, 50},
	}
	for _, p := range packages {
		db.Exec(`INSERT OR IGNORE INTO credit_packages (id, name, usd_price, credits, is_active) VALUES (?,?,?,?,1)`,
			p.id, p.name, p.usd, p.credits)
	}

	// Add pentests.mode if the column is missing (existing installs).
	if !columnExists("pentests", "mode") {
		if _, err := db.Exec(`ALTER TABLE pentests ADD COLUMN mode TEXT NOT NULL DEFAULT 'standard'`); err != nil {
			return fmt.Errorf("add pentests.mode: %w", err)
		}
	}

	// Add payments.url if the column is missing (existing installs).
	if !columnExists("payments", "url") {
		if _, err := db.Exec(`ALTER TABLE payments ADD COLUMN url TEXT`); err != nil {
			return fmt.Errorf("add payments.url: %w", err)
		}
	}

	// Add users.role (admin / user) if the column is missing.
	if !columnExists("users", "role") {
		if _, err := db.Exec(`ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'user'`); err != nil {
			return fmt.Errorf("add users.role: %w", err)
		}
	}

	// Make sure the super admin exists (linked lazily on Google login) and is admin.
	sa := getenv("SUPER_ADMIN_EMAIL", "hans@dalang.io")
	db.Exec(`INSERT OR IGNORE INTO users (id, email, name, role) VALUES (?,?,?, 'admin')`, "u_super_admin", sa, "Super Admin")
	db.Exec(`UPDATE users SET role='admin' WHERE email=?`, sa)

	// Ensure the pentest report directory exists (next to the DB file).
	if i := strings.LastIndex(path, "/"); i > 0 {
		os.MkdirAll(path[:i]+"/reports", 0o755)
	}
	return nil
}

// columnExists reports whether a table has a given column.
func columnExists(table, column string) bool {
	rows, err := db.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		return false
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt interface{}
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err == nil && name == column {
			return true
		}
	}
	return false
}

func now() string { return time.Now().UTC().Format(time.RFC3339) }

func logErr(tag string, err error) {
	if err != nil {
		log.Printf("%s: %v", tag, err)
	}
}
