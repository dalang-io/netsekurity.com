package main

import (
	"fmt"
	"html/template"
)

// analyticsSnippet returns the page-view analytics tag for public pages. The
// Meta Pixel only sees ad traffic it can attribute; without a general analytics
// tool nothing records where organic and referral visitors come from, so there
// was no way to tell which channel brings buyers.
//
// Configured by environment, nothing is invented: set PLAUSIBLE_DOMAIN
// (privacy-friendly, no cookie banner needed) or GA_MEASUREMENT_ID. Both empty
// means no tag at all.
func analyticsSnippet() string {
	if d := getenv("PLAUSIBLE_DOMAIN", ""); d != "" {
		return fmt.Sprintf(`<script defer data-domain="%s" src="https://plausible.io/js/script.outbound-links.js"></script>`,
			template.HTMLEscapeString(d))
	}
	if id := getenv("GA_MEASUREMENT_ID", ""); id != "" {
		id = template.JSEscapeString(id)
		return fmt.Sprintf(`<script async src="https://www.googletagmanager.com/gtag/js?id=%s"></script>
<script>window.dataLayer=window.dataLayer||[];function gtag(){dataLayer.push(arguments);}gtag('js',new Date());gtag('config','%s');</script>`, id, id)
	}
	return ""
}
