package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMain sets up an isolated SQLite DB + JWT secret for all package tests.
func TestMain(m *testing.M) {
	os.Setenv("JWT_SECRET", "test-secret")
	dir, err := os.MkdirTemp("", "nsk-test-*")
	if err != nil {
		panic(err)
	}
	os.Setenv("DATABASE_PATH", filepath.Join(dir, "test.db"))
	if err := initDB(); err != nil {
		panic("initDB: " + err.Error())
	}
	code := m.Run()
	if db != nil {
		db.Close()
	}
	os.RemoveAll(dir)
	os.Exit(code)
}
