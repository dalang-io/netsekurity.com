package main

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// htmxSHA256 is the hash of htmx 1.9.12 as published on npm and as unpkg served it
// when the site was switched to self-hosting. If the vendored file changes, this
// fails: replace it only with a deliberate, reviewed upgrade.
const htmxSHA256 = "449317ade7881e949510db614991e195c3a099c4c791c24dacec55f9f4a2a452"

// The six headers this product's own scanner checks for must be on every kind of
// response from this site, including errors, redirects, static files and the API.
func TestEveryResponseCarriesTheSecurityHeaders(t *testing.T) {
	h := securityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/redirect":
			http.Redirect(w, r, "/login", http.StatusFound)
		case "/missing":
			http.NotFound(w, r)
		case "/enterprise":
			handleIndex(w, r)
		default:
			w.Write([]byte("ok"))
		}
	}))
	for _, path := range []string{"/enterprise", "/redirect", "/missing", "/api/pentests/list", "/css/styles.css", "/fonts/x.woff2"} {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest("GET", path, nil))
		got := rr.Header()
		for _, want := range []string{"Strict-Transport-Security", "Content-Security-Policy", "X-Frame-Options", "X-Content-Type-Options", "Referrer-Policy", "Permissions-Policy", "Cross-Origin-Opener-Policy"} {
			if got.Get(want) == "" {
				t.Errorf("%s: missing %s", path, want)
			}
		}
		if got.Get("X-Frame-Options") != "DENY" || got.Get("X-Content-Type-Options") != "nosniff" {
			t.Errorf("%s: framing/sniffing headers = %q / %q", path, got.Get("X-Frame-Options"), got.Get("X-Content-Type-Options"))
		}
		if !strings.Contains(got.Get("Strict-Transport-Security"), "max-age=31536000") {
			t.Errorf("%s: HSTS = %q", path, got.Get("Strict-Transport-Security"))
		}
		if strings.Contains(got.Get("Strict-Transport-Security"), "preload") || strings.Contains(got.Get("Strict-Transport-Security"), "includeSubDomains") {
			t.Errorf("%s: HSTS must not preload or cover subdomains: %q", path, got.Get("Strict-Transport-Security"))
		}
	}
}

// Everything versioned caches for a year; nothing else does. The font and htmx are
// versioned by file name, the rest by the ?v=<content hash> every page appends.
func TestVersionedAssetsAreImmutable(t *testing.T) {
	h := securityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("x")) }))
	immutable := func(url string) bool {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest("GET", url, nil))
		cc := rr.Header().Get("Cache-Control")
		return strings.Contains(cc, "immutable") && strings.Contains(cc, "max-age=31536000")
	}
	for _, u := range []string{
		"/fonts/PlusJakartaSans-latin-v1.woff2", "/js/htmx-1.9.12.min.js",
		assetURL("/css/styles.css"), assetURL("/favicon.svg"), assetURL("/og.png"), assetURL("/netsekurity-sample-report-en.pdf"),
	} {
		if !immutable(u) {
			t.Errorf("%s should be immutable", u)
		}
	}
	// Unversioned, the same files are not: a bare URL could be a stale copy.
	for _, u := range []string{"/css/styles.css", "/favicon.svg", "/og.png", "/robots.txt"} {
		if immutable(u) {
			t.Errorf("%s has no version in its URL and must not be immutable", u)
		}
	}
	// A ?v= on something that is not a file must not make it cacheable.
	for _, u := range []string{"/dashboard?v=1", "/reports/20260101-00:00-example.com.pdf?v=1", "/enterprise?v=1", "/api/tokens?v=1"} {
		if immutable(u) {
			t.Errorf("%s is not a static file and must never be immutable", u)
		}
	}
}

// The font cannot carry a query string (the @font-face rule and the preload must name
// the same URL), so it is versioned by name, and its bytes are pinned here: change the
// font, rename the file (…-v2.woff2) and update the three places that name it.
const fontSHA256 = "cd8db90cd950e26bc8761f65d323588bd5cd112d326d6d322bc7c8ea86771215"

func TestFontIsPinnedByItsVersionedName(t *testing.T) {
	b, err := staticFS.ReadFile("static/fonts/PlusJakartaSans-latin-v1.woff2")
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(b)
	if got := hex.EncodeToString(sum[:]); got != fontSHA256 {
		t.Errorf("the font changed but its name still says v1 (sha256 %s): rename it, so caches fetch the new file", got)
	}
}

// Every static file a page references is addressed by its content hash, so a release
// that changes a file changes its URL and nothing has to be purged from Cloudflare.
func TestEveryStaticAssetIsAddressedByItsContentHash(t *testing.T) {
	ref := regexp.MustCompile(`(?:href|src|content)="((?:https://netsekurity\.com)?/[^"]+\.(?:css|js|png|svg|pdf|woff2|ico|jpg|webp))(\?v=[0-9a-f]+)?"`)
	byName := regexp.MustCompile(`-v[0-9]+\.|[0-9]+\.[0-9]+\.[0-9]+\.min\.js`) // the font and htmx
	seen := 0
	for _, path := range publicPagePaths {
		for _, m := range ref.FindAllStringSubmatch(renderPublic(t, path), -1) {
			file := strings.TrimPrefix(m[1], "https://netsekurity.com")
			if !isStaticAsset(file) {
				continue
			}
			seen++
			if byName.MatchString(file) {
				continue
			}
			if m[2] != "?v="+assetHash(file) {
				t.Errorf("%s references %s as %q, want ?v=%s", path, file, m[1]+m[2], assetHash(file))
			}
		}
	}
	if seen < 30 {
		t.Errorf("only %d static references found across the public pages; the pattern is not matching", seen)
	}
	// The templates of the customer app and the admin panel go through the same helper:
	// no raw /favicon… or /css… URL may remain in any Go source.
	raw := regexp.MustCompile(`(?:href|src)="/(?:favicon|css/|og)[^"]*"`)
	files, _ := filepath.Glob("*.go")
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, _ := os.ReadFile(f)
		for _, m := range raw.FindAllString(string(src), -1) {
			if !strings.Contains(m, "{{") {
				t.Errorf("%s references %s without a content hash", f, m)
			}
		}
	}
	// assetHash tracks the file: a different file has a different hash, a missing one a placeholder.
	if assetHash("/favicon.svg") == assetHash("/og.png") || assetHash("/missing.png") != "0" {
		t.Error("assetHash does not follow the file content")
	}
}

var (
	reScriptOpen = regexp.MustCompile("<script[^>]*>")
	reInlineEv   = regexp.MustCompile(`\son[a-z]+="`)
	reStyleTag   = regexp.MustCompile(`<style[\s>]`)
	reStyleAttr  = regexp.MustCompile(`\sstyle="`)
	reScriptSrc  = regexp.MustCompile(`<script[^>]*\ssrc="([^"]+)"`)
)

// The strict policy only holds if the markup keeps its side: every script carries
// the nonce, nothing is inline, and the only third-party script is our analytics.
func TestPublicPagesObeyTheirStrictPolicy(t *testing.T) {
	for _, path := range publicPagePaths {
		rr := httptest.NewRecorder()
		securityHeaders(http.HandlerFunc(handleIndex)).ServeHTTP(rr, httptest.NewRequest("GET", path, nil))
		page, csp := rr.Body.String(), rr.Header().Get("Content-Security-Policy")
		nonce := nonceOf(t, rr)
		if nonce == "" {
			t.Errorf("%s: no nonce in %q", path, csp)
			continue
		}
		for _, must := range []string{"'strict-dynamic'", "object-src 'none'", "frame-ancestors 'none'", "base-uri 'self'", "form-action 'self'", "default-src 'self'", "font-src 'self'"} {
			if !strings.Contains(csp, must) {
				t.Errorf("%s: CSP lacks %s", path, must)
			}
		}
		for _, banned := range []string{"'unsafe-eval'", "script-src 'self'", "'unsafe-inline'; style", "script-src * ", "http:"} {
			if strings.Contains(csp, banned) {
				t.Errorf("%s: CSP contains %q: %s", path, banned, csp)
			}
		}
		// script-src must not allow unsafe-inline: only style-src-attr may.
		if m := regexp.MustCompile(`script-src[^;]*`).FindString(csp); strings.Contains(m, "unsafe-inline") {
			t.Errorf("%s: script-src allows unsafe-inline: %s", path, m)
		}
		for _, tag := range reScriptOpen.FindAllString(page, -1) {
			if !strings.Contains(tag, `nonce="`+nonce+`"`) {
				t.Errorf("%s: a script without the response's nonce: %s", path, tag)
			}
		}
		// Markup only: the prose and string literals inside <script> bodies (the consent
		// gate quotes the old onclick="…" handlers in a comment) are not attributes.
		markup := reScript.ReplaceAllString(page, "")
		if reInlineEv.MatchString(markup) {
			t.Errorf("%s: inline event handler attribute %v", path, reInlineEv.FindString(markup))
		}
		if reStyleTag.MatchString(page) || reStyleAttr.MatchString(markup) {
			t.Errorf("%s: inline <style> or style= attribute in the server markup", path)
		}
		for _, m := range reScriptSrc.FindAllStringSubmatch(page, -1) {
			if !strings.HasPrefix(m[1], "/") && !strings.HasPrefix(m[1], "https://umami.dalang.io/") {
				t.Errorf("%s: loads a script from %s", path, m[1])
			}
		}
		for _, banned := range []string{"googleapis.com", "gstatic.com", "unpkg.com", "cdn.jsdelivr", "cdnjs"} {
			if strings.Contains(page, banned) {
				t.Errorf("%s: loads from %s", path, banned)
			}
		}
	}
}

// The app pages keep inline handlers for now, so they get the baseline policy; it
// still pins origins and forbids framing, plugins and base-tag or form hijacking.
func TestAppPagesGetTheBaselinePolicy(t *testing.T) {
	rr := httptest.NewRecorder()
	securityHeaders(http.HandlerFunc(handleDashboard)).ServeHTTP(rr, httptest.NewRequest("GET", "/dashboard", nil))
	csp := rr.Header().Get("Content-Security-Policy")
	for _, must := range []string{"object-src 'none'", "frame-ancestors 'none'", "base-uri 'self'", "form-action 'self'", "default-src 'self'", "https://umami.dalang.io"} {
		if !strings.Contains(csp, must) {
			t.Errorf("baseline CSP lacks %s: %s", must, csp)
		}
	}
	if strings.Contains(csp, "'unsafe-eval'") || strings.Contains(csp, "script-src *") || strings.Contains(csp, "http:") {
		t.Errorf("baseline CSP is too loose: %s", csp)
	}
}

func TestWithNonce(t *testing.T) {
	got := withNonce(`<script>a()</script><script async src="x"></script><script nonce="keep">b()</script>`, "N1")
	want := `<script nonce="N1">a()</script><script nonce="N1" async src="x"></script><script nonce="keep">b()</script>`
	if got != want {
		t.Errorf("withNonce =\n%s\nwant\n%s", got, want)
	}
	a, b := newNonce(), newNonce()
	if a == "" || a == b || len(a) < 20 {
		t.Errorf("nonces %q, %q", a, b)
	}
	if strings.ContainsAny(a, "+/=&<>\"'") {
		t.Errorf("nonce %q has characters html/template would escape", a)
	}
}

// The consent gate's script gets the nonce, and the pixel it injects is given the same one.
func TestConsentGateWorksUnderANonce(t *testing.T) {
	page := renderPublic(t, "/")
	if !strings.Contains(page, "document.currentScript") || !strings.Contains(page, "setAttribute('nonce', NONCE)") {
		t.Error("the consent gate does not pass its nonce to the pixel script it injects")
	}
	if strings.Index(page, "insertBefore(el, document.body.firstChild)") < 0 {
		t.Error("the consent banner is not placed first in the keyboard order")
	}
}

// The vendored htmx is the audited release, byte for byte.
func TestVendoredHTMXIsThePinnedRelease(t *testing.T) {
	b, err := staticFS.ReadFile("static/js/htmx-1.9.12.min.js")
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(b)
	if got := hex.EncodeToString(sum[:]); got != htmxSHA256 {
		t.Errorf("static/js/htmx-1.9.12.min.js sha256 = %s, want %s", got, htmxSHA256)
	}
	if _, err := staticFS.ReadFile("static/fonts/PlusJakartaSans-latin-v1.woff2"); err != nil {
		t.Errorf("self-hosted font missing: %v", err)
	}
}
