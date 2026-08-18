package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
)

//go:embed static
var staticFS embed.FS

var staticSub fs.FS

func main() {
	loadEnv(".env")

	if err := initDB(); err != nil {
		log.Fatalf("db init: %v", err)
	}
	// Ensure a landing placeholder for login/register CTA.
	log.Printf("netsekurity ready (env: %s)", getenv("ENVIRONMENT", "dev"))

	port := getenv("PORT", "8094")
	if v := os.Getenv("PORT"); v != "" {
		port = v
	}

	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatalf("embed fs: %v", err)
	}
	staticSub = sub

	mux := http.NewServeMux()
	// Landing page + static assets. "/" is served through a handler so the
	// Google client ID can be injected; all other paths fall through to the
	// embedded FileServer (css, og.png, favicon, ...).
	mux.HandleFunc("/", handleIndex)

	// Auth
	mux.HandleFunc("/login", handleGoogleLogin)
	mux.HandleFunc("/auth/google", handleGoogleLogin)
	mux.HandleFunc("/auth/google/callback", handleGoogleCallback)
	mux.HandleFunc("/auth/google/onetap", handleOneTap)
	mux.HandleFunc("/logout", handleLogout)

	// Dashboard (protected)
	mux.HandleFunc("/dashboard", requireAuth(handleDashboard))

	// Super admin
	mux.HandleFunc("/su", requireAdmin(handleAdmin))
	mux.HandleFunc("/su/users/add", requireAdmin(handleAdminAddUser))
	mux.HandleFunc("/su/users/role", requireAdmin(handleAdminSetRole))
	mux.HandleFunc("/su/domains/pentest", requireAdmin(handleAdminPentest))
	mux.HandleFunc("/su/reports/upload", requireAdmin(handleAdminUploadReport))
	mux.HandleFunc("/reports/", requireAdmin(handleReport))

	// Payments
	mux.HandleFunc("/api/topup", handleTopUp)
	mux.HandleFunc("/webhook/xendit", handleXenditWebhook)

	// Domains
	mux.HandleFunc("/api/domains", handleAddDomain)
	mux.HandleFunc("/api/domains/verify", handleVerifyDomain)
	mux.HandleFunc("/api/domains/delete", handleDeleteDomain)

	// Marketing / HTMX fragments
	mux.HandleFunc("/contact", handleContact)
	mux.HandleFunc("/api/txt", handleTXT)
	mux.HandleFunc("/api/verify", handleVerify)
	mux.HandleFunc("/api/faq", handleFAQ)

	addr := ":" + port
	log.Printf("netsekurity.com listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

// handleIndex serves the embedded landing page (injecting the Google client ID
// into the One Tap script), and falls through to the FileServer for other
// static assets (css, og.png, favicon, ...).
func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		if staticSub == nil {
			staticSub, _ = fs.Sub(staticFS, "static")
		}
		http.FileServer(http.FS(staticSub)).ServeHTTP(w, r)
		return
	}
	b, err := staticFS.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "index missing", http.StatusInternalServerError)
		return
	}
	s := strings.ReplaceAll(string(b), "__GOOGLE_CLIENT_ID__", getenv("GOOGLE_CLIENT_ID", ""))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(s))
}
