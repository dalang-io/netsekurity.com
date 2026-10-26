package main

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"regexp"
	"strings"
)

// Security headers for every response.
//
// A security-assessment vendor is judged first on its own headers, and until
// 2026-10-26 this site sent none of the six checks its own scanner runs
// (HSTS, CSP, X-Frame-Options, X-Content-Type-Options, Referrer-Policy,
// Permissions-Policy) — finding F-02 of its own sample report. This file closes
// that.
//
// Two Content-Security-Policies:
//
//   - the public pages (home, enterprise, trust, guide, sample, docs, scope) get a
//     strict, nonce-based policy: no inline script runs unless it carries the
//     per-response nonce, and 'strict-dynamic' lets the consent gate load the Meta
//     pixel after the visitor accepts. Those templates contain no inline event
//     handlers, which is what makes this possible.
//   - everything else (dashboard, checkout, /su, fragments) still uses inline
//     handlers, so it gets a baseline policy that allows inline script but pins every
//     origin, forbids plugins, framing and base-tag/form hijacking. Moving those
//     pages to nonces is a follow-up: replace the onclick="…" attributes with
//     delegated listeners, then switch them to strictCSP.

const (
	hstsValue = "max-age=31536000" // no includeSubDomains / preload: those are hard to take back
	// permissionsPolicy turns off every powerful feature the site does not use.
	permissionsPolicy = "accelerometer=(), camera=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), payment=(), usb=()"

	// cspCommon is shared by both policies. connect-src and img-src list the only
	// third parties the site talks to: its own analytics, and Meta — which loads only
	// after the visitor accepts the cookie banner (consent.go).
	cspCommon = "default-src 'self'; " +
		"img-src 'self' data: https://www.facebook.com; " +
		"font-src 'self'; " +
		"connect-src 'self' https://umami.dalang.io https://www.facebook.com https://*.on.aws; " +
		"object-src 'none'; base-uri 'self'; frame-ancestors 'none'; form-action 'self'"

	baselineCSP = cspCommon + "; " +
		"script-src 'self' 'unsafe-inline' https://umami.dalang.io https://connect.facebook.net; " +
		"style-src 'self' 'unsafe-inline'"
)

// strictCSP is the policy for the public pages. style-src-attr allows only the
// style="…" attributes the consent banner writes; there is no inline <style>.
func strictCSP(nonce string) string {
	return cspCommon + "; " +
		"script-src 'nonce-" + nonce + "' 'strict-dynamic'; " +
		"style-src 'self'; style-src-attr 'unsafe-inline'"
}

// newNonce returns 128 random bits, base64url-encoded, for one response.
func newNonce() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand failing is not recoverable; an empty nonce makes the strict
		// policy block every script, which is the safe failure.
		return ""
	}
	// URL-safe alphabet: valid in a CSP nonce, and html/template never has to escape it.
	return base64.RawURLEncoding.EncodeToString(b)
}

var scriptOpenRE = regexp.MustCompile(`<script(\s[^>]*)?>`)

// withNonce adds nonce="…" to every <script> opening tag in html that lacks one.
// It is how the consent gate, the analytics tag and the ViewContent beacon —
// which arrive as strings — join the nonce-based policy.
func withNonce(html, nonce string) string {
	return scriptOpenRE.ReplaceAllStringFunc(html, func(tag string) string {
		if strings.Contains(tag, "nonce=") {
			return tag
		}
		return `<script nonce="` + nonce + `"` + strings.TrimPrefix(tag, "<script")
	})
}

// securityHeaders wraps the whole mux.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Strict-Transport-Security", hstsValue)
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Permissions-Policy", permissionsPolicy)
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		h.Set("Content-Security-Policy", baselineCSP) // public pages replace this with strictCSP
		// Fonts, scripts and the stylesheet are versioned by file name or by the ?v=
		// hash every page appends, so a year-long immutable cache is safe — and it
		// stops every visit re-downloading the stylesheet.
		// Immutable only for what is certain to be a versioned file: an embedded static
		// file that is versioned by name (/fonts, /js) or addressed with its content hash
		// (?v=, see assets.go). A handler (dashboard, reports, pages) is never a file, so
		// a stray ?v= on its URL cannot make an authenticated response cacheable.
		p := r.URL.Path
		if isStaticAsset(p) && (strings.HasPrefix(p, "/fonts/") || strings.HasPrefix(p, "/js/") || r.URL.Query().Get("v") != "") {
			h.Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		next.ServeHTTP(w, r)
	})
}
