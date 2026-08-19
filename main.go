package main

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
)

//go:embed static
var staticFS embed.FS

var staticSub fs.FS

// cssHash is a short sha256 of the built stylesheet, used to cache-bust the CSS
// URL on every deploy (?v=<hash>) so Cloudflare/browsers never serve stale CSS.
var cssHash = func() string {
	b, err := staticFS.ReadFile("static/css/styles.css")
	if err != nil {
		return "0"
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])[:12]
}()

func init() { _ = cssHash }

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
	mux.HandleFunc("/su/users/credit", requireAdmin(handleAdminAddCredit))
	mux.HandleFunc("/su/domains/pentest", requireAdmin(handleAdminPentest))
	mux.HandleFunc("/su/reports/upload", requireAdmin(handleAdminUploadReport))
	mux.HandleFunc("/reports/", requireAuth(handleReport))

	// Pentests (user scan + worker intake)
	mux.HandleFunc("/api/pentests/start", requireAuth(handleStartPentest))
	mux.HandleFunc("/api/pentests/list", requireAuth(handlePentestList))
	mux.HandleFunc("/api/pentests/worker/claim", handleWorkerClaim)
	mux.HandleFunc("/api/pentests/worker/report", handleWorkerReport)

	// API tokens (user-managed for CI/CD)
	mux.HandleFunc("/api/tokens", requireAuth(handleTokensList))          // GET list
	mux.HandleFunc("/api/tokens/create", requireAuth(handleTokensCreate)) // POST create
	mux.HandleFunc("/api/tokens/", requireAuth(handleTokensDelete))       // DELETE /api/tokens/{id}

	// Payments
	mux.HandleFunc("/api/topup", handleTopUp)
	mux.HandleFunc("/webhook/xendit", handleXenditWebhook)

	// Domains
	mux.HandleFunc("/api/domains", handleAddDomain)
	mux.HandleFunc("/api/domains/verify", handleVerifyDomain)
	mux.HandleFunc("/api/domains/delete", handleDeleteDomain)

	// Marketing / HTMX fragments
	mux.HandleFunc("/contact", handleContact)
	mux.HandleFunc("/docs", handleDocs)
	mux.HandleFunc("/api/txt", handleTXT)
	mux.HandleFunc("/api/verify", handleVerify)
	mux.HandleFunc("/api/faq", handleFAQ)

	// CI/CD integration (authenticated by X-API-Token)
	mux.HandleFunc("/api/v1/pentests", handleAPICreatePentest)

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
	s = strings.ReplaceAll(s, "__CSS_HASH__", cssHash)
	// Shared stack coverage section (real SVG logos).
	s = strings.ReplaceAll(s, "__STACK_BLOCK__", string(renderStack()))
	// Shared header component (landing + docs use the same renderHeader).
	landingNav := []hdrLink{
		{Href: "#how", Text: "ls how"},
		{Href: "#features", Text: "cat features"},
		{Href: "#stack", Text: "cat stack"},
		{Href: "#cicd", Text: "pip install cicd"},
		{Href: "/docs", Text: "man docs"},
		{Href: "#pricing", Text: "cat pricing"},
		{Href: "#verify", Text: "verify -d"},
		{Href: "#faq", Text: "man faq"},
	}
	// Auth-aware nav: logged-in shows dashboard, anonymous shows login.
	_, auErr := currentUser(r)
	hdrNav := `<a href="/dashboard" class="whitespace-nowrap rounded border border-emerald-400 bg-emerald-500/10 px-3 py-1.5 text-xs font-bold text-emerald-300 hover:bg-emerald-500/20 glow">./dashboard<span class="cursor"></span></a>`
	loginNav := `<a href="/login" class="whitespace-nowrap rounded border border-emerald-400 bg-emerald-500/10 px-3 py-1.5 text-xs font-bold text-emerald-300 hover:bg-emerald-500/20 glow">login</a>`
	if auErr == nil {
		s = strings.ReplaceAll(s, "__HEADER_BLOCK__", string(renderHeader(landingNav, template.HTML(hdrNav), "#top"))+headerMobileJS)
		s = strings.ReplaceAll(s, "__FOOTER_AUTH_NAV__", `<a href="/dashboard" class="hover:text-emerald-300">dashboard</a>`)
	} else {
		s = strings.ReplaceAll(s, "__HEADER_BLOCK__", string(renderHeader(landingNav, template.HTML(loginNav), "#top"))+headerMobileJS)
		s = strings.ReplaceAll(s, "__FOOTER_AUTH_NAV__", `<a href="/login" class="hover:text-emerald-300">login</a>`)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Write([]byte(s))
}
