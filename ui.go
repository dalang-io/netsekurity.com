package main

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"time"
)

// Presentation helpers shared by the customer app: the dashboard pages, the
// checkout page and the HTMX fragments they load (scans, domains, tokens).
// Nothing here touches the database or money — it only decides how a value is
// worded and drawn.

// fmtCredits renders a credit amount the way a person says it: "3", not "3.0".
// A fractional balance (a legacy 2-credit scan, a partial refund) keeps one
// decimal so the figure is never silently rounded.
func fmtCredits(v float64) string {
	if v == float64(int64(v)) {
		return strconv.FormatInt(int64(v), 10)
	}
	return strconv.FormatFloat(v, 'f', 1, 64)
}

// creditsLabel is fmtCredits plus the noun, with the singular handled: the old
// dashboard printed "1 credits".
func creditsLabel(v float64) string {
	s := fmtCredits(v)
	if s == "1" {
		return "1 credit"
	}
	return s + " credits"
}

// storedTimeLayouts are the two formats that coexist in the database: Go writes
// RFC3339, SQLite triggers and column defaults write datetime('now').
var storedTimeLayouts = []string{time.RFC3339, "2006-01-02 15:04:05"}

func parseStoredTime(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	for _, layout := range storedTimeLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), true
		}
	}
	return time.Time{}, false
}

// timeTag renders a stored timestamp as a <time> element. The text is a readable
// UTC fallback ("18 Aug 2026, 15:59 UTC"); the script in the page shell swaps it
// for the viewer's own timezone, so a customer in Jakarta is not asked to do
// arithmetic on an ISO string. An unparseable value is shown as-is, escaped.
func timeTag(s string) template.HTML {
	t, ok := parseStoredTime(s)
	if !ok {
		return template.HTML(template.HTMLEscapeString(s))
	}
	return template.HTML(fmt.Sprintf(`<time class="nsk-time" datetime="%s">%s</time>`,
		t.Format(time.RFC3339), t.Format("2 Jan 2006, 15:04")+" UTC"))
}

// firstName is what the greeting calls the customer. Google gives a full name;
// fall back to the part of the email before the @ rather than "undefined".
func firstName(name, email string) string {
	if f := strings.Fields(name); len(f) > 0 {
		return f[0]
	}
	if i := strings.Index(email, "@"); i > 0 {
		return email[:i]
	}
	return "there"
}

// initial is the single letter in the sidebar avatar.
func initial(name, email string) string {
	for _, s := range []string{name, email} {
		if r := []rune(strings.TrimSpace(s)); len(r) > 0 {
			return strings.ToUpper(string(r[0]))
		}
	}
	return "?"
}

// perScan is the price of one scan inside a package, so the packages can be
// compared without arithmetic. Derived from the stored price and credits only.
func perScan(usd, credits float64) string {
	if credits <= 0 {
		return ""
	}
	return "$" + trimUSD(float64(int64(usd/credits*100+0.5))/100)
}

// iconPaths are Lucide-style 24px stroke icons. Only the inner paths are stored.
var iconPaths = map[string]string{
	"home":    `<path d="m3 9 9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/>`,
	"globe":   `<circle cx="12" cy="12" r="10"/><path d="M2 12h20"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/>`,
	"shield":  `<path d="M20 13c0 5-3.5 7.5-7.66 8.95a1 1 0 0 1-.67-.01C7.5 20.5 4 18 4 13V6a1 1 0 0 1 1-1c2 0 4.5-1.2 6.24-2.72a1.17 1.17 0 0 1 1.52 0C14.51 3.81 17 5 19 5a1 1 0 0 1 1 1z"/><path d="m9 12 2 2 4-4"/>`,
	"card":    `<rect width="20" height="14" x="2" y="5" rx="2"/><line x1="2" x2="22" y1="10" y2="10"/>`,
	"zap":     `<path d="M4 14a1 1 0 0 1-.78-1.63l9.9-10.2a.5.5 0 0 1 .86.46l-1.92 6.02A1 1 0 0 0 13 10h7a1 1 0 0 1 .78 1.63l-9.9 10.2a.5.5 0 0 1-.86-.46l1.92-6.02A1 1 0 0 0 11 14z"/>`,
	"help":    `<circle cx="12" cy="12" r="10"/><path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/><path d="M12 17h.01"/>`,
	"file":    `<path d="M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7z"/><path d="M14 2v4a2 2 0 0 0 2 2h4"/><path d="M16 13H8"/><path d="M16 17H8"/><path d="M10 9H8"/>`,
	"logout":  `<path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" x2="9" y1="12" y2="12"/>`,
	"menu":    `<line x1="4" x2="20" y1="12" y2="12"/><line x1="4" x2="20" y1="6" y2="6"/><line x1="4" x2="20" y1="18" y2="18"/>`,
	"x":       `<path d="M18 6 6 18"/><path d="m6 6 12 12"/>`,
	"check":   `<path d="M20 6 9 17l-5-5"/>`,
	"copy":    `<rect width="14" height="14" x="8" y="8" rx="2" ry="2"/><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/>`,
	"lock":    `<rect width="18" height="11" x="3" y="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/>`,
	"arrow":   `<path d="M5 12h14"/><path d="m12 5 7 7-7 7"/>`,
	"scan":    `<path d="M3 7V5a2 2 0 0 1 2-2h2"/><path d="M17 3h2a2 2 0 0 1 2 2v2"/><path d="M21 17v2a2 2 0 0 1-2 2h-2"/><path d="M7 21H5a2 2 0 0 1-2-2v-2"/><path d="M7 12h10"/>`,
	"shield2": `<path d="M20 13c0 5-3.5 7.5-7.66 8.95a1 1 0 0 1-.67-.01C7.5 20.5 4 18 4 13V6a1 1 0 0 1 1-1c2 0 4.5-1.2 6.24-2.72a1.17 1.17 0 0 1 1.52 0C14.51 3.81 17 5 19 5a1 1 0 0 1 1 1z"/>`,
}

// icon returns an inline SVG. size is in pixels; an unknown name yields nothing
// rather than a broken tag.
func icon(name string, size int) template.HTML {
	p, ok := iconPaths[name]
	if !ok {
		return ""
	}
	return template.HTML(fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="%d" height="%d" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" class="shrink-0">%s</svg>`,
		size, size, p))
}

// alertHTML is the standard inline message for an HTMX result. kind is one of
// ok, warn, bad, info. body must already be safe HTML: callers escape any
// customer-supplied value before passing it in.
func alertHTML(kind, body string) string {
	role := "status"
	if kind == "bad" {
		role = "alert"
	}
	return `<div class="nsk-alert nsk-alert-` + kind + `" role="` + role + `">` + body + `</div>`
}
