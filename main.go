package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

//go:embed static
var staticFS embed.FS

var staticSub fs.FS

// cssHash is the content hash of the built stylesheet (see assets.go): the CSS URL
// changes whenever the CSS does, so no cache ever serves a stale stylesheet.
var cssHash = assetHash("/css/styles.css")

func init() { _ = cssHash }

func main() {
	loadEnv(".env")
	checkSecrets()

	if err := initDB(); err != nil {
		log.Fatalf("db init: %v", err)
	}
	// Auto-credit poller: credits are added by checking Xendit periodically,
	// independent of the (optional) webhook.
	go startPaymentPolling()
	// Ensure a landing placeholder for login/register CTA.
	log.Printf("netsekurity ready (env: %s)", getenv("ENVIRONMENT", "dev"))
	if w := operatorCurrencyWarning(); w != "" {
		log.Printf("WARNING: currency mismatch — %s", w)
	}

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
	// RFC 9116 security.txt (Go's FileServer hides dotfiles, so serve explicitly).
	mux.HandleFunc("/.well-known/security.txt", handleSecurityTxt)
	// Generated from the page registry so it cannot drift from the pages.
	mux.HandleFunc("/sitemap.xml", handleSitemap)
	// Generated from the same copy and prices as the pages (site_llms.go).
	mux.HandleFunc("/llms.txt", handleLLMS(false))
	mux.HandleFunc("/llms-full.txt", handleLLMS(true))

	// Per-IP ceilings on everything that creates rows or calls a third party.
	// Generous for a person, tight for a loop.
	writeLimit := newRateLimiter(30, time.Minute)  // domains, scans, tokens
	readLimit := newRateLimiter(300, time.Minute)  // API result polling, per client address
	payLimit := newRateLimiter(10, 10*time.Minute) // Xendit invoices
	leadLimit := newRateLimiter(5, 10*time.Minute) // contact form
	authLimit := newRateLimiter(20, time.Minute)   // One Tap token verification

	// Auth
	mux.HandleFunc("/login", handleGoogleLogin)
	mux.HandleFunc("/auth/google", handleGoogleLogin)
	mux.HandleFunc("/auth/google/callback", handleGoogleCallback)
	mux.HandleFunc("/auth/google/onetap", rateLimit(authLimit, handleOneTap))
	mux.HandleFunc("/logout", handleLogout)

	// Dashboard (protected)
	mux.HandleFunc("/dashboard", requireAuth(handleDashboard))
	mux.HandleFunc("/dashboard/websites", requireAuth(handleDashboardWebsites))
	mux.HandleFunc("/dashboard/scans", requireAuth(handleDashboardScans))
	mux.HandleFunc("/dashboard/credits", requireAuth(handleDashboardCredits))
	mux.HandleFunc("/dashboard/automation", requireAuth(handleDashboardAutomation))
	mux.HandleFunc("/checkout", requireAuth(handleCheckout))

	// Super admin
	mux.HandleFunc("/su", requireAdmin(handleAdmin))
	mux.HandleFunc("/su/users/add", requireAdmin(handleAdminAddUser))
	mux.HandleFunc("/su/users/role", requireAdmin(handleAdminSetRole))
	mux.HandleFunc("/su/users/credit", requireAdmin(handleAdminAddCredit))
	mux.HandleFunc("/su/domains/pentest", requireAdmin(handleAdminPentest))
	mux.HandleFunc("/su/reports/upload", requireAdmin(handleAdminUploadReport))
	mux.HandleFunc("/su/contacts/handled", requireAdmin(handleAdminContactHandled))
	mux.HandleFunc("/reports/", requireAuth(handleReport))

	// Pentests (user scan + worker intake)
	mux.HandleFunc("/api/pentests/start", rateLimit(writeLimit, requireAuth(handleStartPentest)))
	mux.HandleFunc("/api/pentests/list", requireAuth(handlePentestList))
	mux.HandleFunc("/api/pentests/worker/claim", handleWorkerClaim)
	mux.HandleFunc("/api/pentests/worker/report", handleWorkerReport)
	mux.HandleFunc("/api/pentests/worker/fail", handleWorkerFail)

	// Partner API (Dalang platform). HMAC-authenticated; never touches credits.
	// The report route is a signed URL so the platform can hand it to a browser.
	mux.HandleFunc("/api/partner/v1/scans/reserve", handlePartnerReserve)
	mux.HandleFunc("/api/partner/v1/scans/", handlePartnerScan)

	// API tokens (user-managed for CI/CD)
	mux.HandleFunc("/api/tokens", requireAuth(handleTokensList))                                 // GET list
	mux.HandleFunc("/api/tokens/create", rateLimit(writeLimit, requireAuth(handleTokensCreate))) // POST create
	mux.HandleFunc("/api/tokens/", requireAuth(handleTokensDelete))                              // DELETE /api/tokens/{id}

	// Payments
	mux.HandleFunc("/api/topup", rateLimit(payLimit, handleTopUp))
	mux.HandleFunc("/webhook/xendit", handleXenditWebhook)

	// Domains
	mux.HandleFunc("/api/domains", rateLimit(writeLimit, handleAddDomain))
	mux.HandleFunc("/api/domains/verify", rateLimit(writeLimit, handleVerifyDomain))
	mux.HandleFunc("/api/domains/delete", rateLimit(writeLimit, handleDeleteDomain))

	// Marketing pages / HTMX fragments
	mux.HandleFunc("/contact", rateLimit(leadLimit, handleContact))

	// CI/CD integration (authenticated by X-API-Token)
	registerAPIv1(mux, writeLimit, readLimit)

	addr := ":" + port
	log.Printf("netsekurity.com listening on %s", addr)
	if err := http.ListenAndServe(addr, securityHeaders(mux)); err != nil {
		log.Fatal(err)
	}
}

// checkSecrets refuses to start in production without the secrets that make
// auth and the worker channel meaningful. jwt.go falls back to a well-known
// dev secret when JWT_SECRET is empty; in production that lets anyone mint an
// admin session.
func checkSecrets() {
	prod := getenv("ENVIRONMENT", "development") == "production"
	missing := []string{}
	for _, k := range []string{"JWT_SECRET", "BOT_AUTH_TOKEN", "XENDIT_CALLBACK_KEY"} {
		if getenv(k, "") == "" {
			missing = append(missing, k)
		}
	}
	if len(missing) == 0 {
		return
	}
	if prod {
		log.Fatalf("refusing to start: ENVIRONMENT=production but %s not set", strings.Join(missing, ", "))
	}
	log.Printf("WARNING: %s not set — using insecure defaults (fine for development only)", strings.Join(missing, ", "))
}

// handleSecurityTxt serves the RFC 9116 security.txt file.
func handleSecurityTxt(w http.ResponseWriter, r *http.Request) {
	b, err := staticFS.ReadFile("static/security.txt")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Write(b)
}

// handleIndex serves the public site pages (see site.go: home, enterprise,
// trust and the guide, each in English and Indonesian) and falls through to the
// FileServer for static assets (css, og.png, favicon, llms.txt ...).
func handleIndex(w http.ResponseWriter, r *http.Request) {
	if p, lang, ok := lookupSitePage(r.URL.Path); ok {
		serveSitePage(w, r, p, lang)
		return
	}
	// One page, one URL: "/enterprise/" and "/id" redirect to their canonical form.
	if canon := canonicalSitePath(r.URL.Path); canon != "" {
		http.Redirect(w, r, canon, http.StatusMovedPermanently)
		return
	}
	if staticSub == nil {
		staticSub, _ = fs.Sub(staticFS, "static")
	}
	http.FileServer(http.FS(staticSub)).ServeHTTP(w, r)
}
