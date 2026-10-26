package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"strings"
	"sync"
)

// Cache-busting for everything the pages load from /static.
//
// A page references an asset as "/favicon.svg?v=<12 hex of the file's sha256>". The
// hash is the file's content, so it changes exactly when the file does, with no
// release step to remember: a deploy that changes the stylesheet, the favicon, an OG
// card or a sample PDF changes the URL, and Cloudflare and every browser fetch the new
// one. A deploy that changes none of them leaves the URLs, and their year-long cache,
// alone. Nothing needs purging.
//
// The two things that cannot carry a query string are the web font (the @font-face
// rule in the stylesheet and the <link rel="preload"> must name the same URL) and the
// vendored htmx: those are versioned by file name instead (…-latin-v1.woff2,
// htmx-1.9.12.min.js) and their sha256 is pinned by a test, so changing the bytes
// without renaming the file fails the build.

var assetHashes sync.Map // path -> hash

// assetHash is a short sha256 of an embedded static file, or "0" if there is none.
func assetHash(path string) string {
	if v, ok := assetHashes.Load(path); ok {
		return v.(string)
	}
	b, err := staticFS.ReadFile("static" + path)
	h := "0"
	if err == nil {
		sum := sha256.Sum256(b)
		h = hex.EncodeToString(sum[:])[:12]
	}
	assetHashes.Store(path, h)
	return h
}

// assetURL is path with its content hash: "/favicon.svg?v=3f9a1c0b7d2e".
func assetURL(path string) string { return path + "?v=" + assetHash(path) }

// isStaticAsset reports whether path names a file embedded in the binary. Pages,
// the dashboard and reports are handlers, never files, so they can never be marked
// immutable by a stray ?v= on their URL.
func isStaticAsset(path string) bool {
	if strings.Contains(path, "..") {
		return false
	}
	st, err := fs.Stat(staticFS, "static"+path)
	return err == nil && !st.IsDir()
}
