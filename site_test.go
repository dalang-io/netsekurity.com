package main

import (
	"encoding/json"
	"html"
	"image/png"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
	"unicode/utf8"
)

// publicPagePaths is every public page the site serves, bilingual and not.
var publicPagePaths = []string{
	"/", "/id/", "/getting-started", "/id/panduan-memulai",
	"/enterprise", "/id/enterprise",
	"/trust", "/id/trust",
	"/external-assessment-vs-penetration-test", "/id/asesmen-eksternal-vs-uji-penetrasi",
	"/sample-report", "/docs", "/product-scope",
	"/id/contoh-laporan", "/id/dokumentasi", "/id/cakupan-produk",
}

func renderPublic(t *testing.T, path string) string {
	t.Helper()
	rr := httptest.NewRecorder()
	handleIndex(rr, httptest.NewRequest("GET", path, nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("%s: status %d", path, rr.Code)
	}
	return rr.Body.String()
}

var (
	reTitle  = regexp.MustCompile(`(?s)<title>(.*?)</title>`)
	reDesc   = regexp.MustCompile(`<meta name="description" content="([^"]*)"`)
	reCanon  = regexp.MustCompile(`<link rel="canonical" href="([^"]*)"`)
	reAlt    = regexp.MustCompile(`<link rel="alternate" hreflang="([^"]*)" href="([^"]*)"`)
	reLD     = regexp.MustCompile(`(?s)<script type="application/ld\+json"[^>]*>(.*?)</script>`)
	reH      = regexp.MustCompile(`<h([1-6])[\s>]`)
	reID     = regexp.MustCompile(`\sid="([^"]+)"`)
	reImg    = regexp.MustCompile(`<img\b[^>]*>`)
	reHref   = regexp.MustCompile(`href="(/[^"]*)"`)
	reScript = regexp.MustCompile(`(?s)<(script|style)\b.*?</(script|style)>`)
	reTag    = regexp.MustCompile(`<[^>]+>`)
	reToken  = regexp.MustCompile(`\{[A-Za-z_]+\}`)
)

// apiParams are path parameters in API documentation ("GET /api/v1/assessments/{id}"),
// which look like copy tokens and are not.
var apiParams = map[string]bool{"{id}": true, "{assessment_id}": true}

// unfilled lists the {tokens} in s that are neither filled prices nor API path parameters.
func unfilled(s string) []string {
	var out []string
	for _, tok := range reToken.FindAllString(s, -1) {
		if !apiParams[tok] {
			out = append(out, tok)
		}
	}
	return out
}

func firstMatch(re *regexp.Regexp, s string) string {
	m := re.FindStringSubmatch(s)
	if m == nil {
		return ""
	}
	return html.UnescapeString(m[1])
}

// visibleText is the page text a reader (or a model) sees: no scripts, styles or tags.
func visibleText(page string) string {
	page = reScript.ReplaceAllString(page, " ")
	return html.UnescapeString(strings.Join(strings.Fields(reTag.ReplaceAllString(page, " ")), " "))
}

// ---- copy completeness ------------------------------------------------------

type copyPair struct {
	name   string
	en, id any
}

func copyPairs() []copyPair {
	return []copyPair{
		{"common", commonEN(), commonID()},
		{"home", homeEN(), homeID()},
		{"enterprise", enterpriseEN(), enterpriseID()},
		{"trust", trustEN(), trustID()},
		{"compare", compareEN(), compareID()},
	}
}

func addressable(v any) reflect.Value {
	p := reflect.New(reflect.TypeOf(v))
	p.Elem().Set(reflect.ValueOf(v))
	return p.Elem()
}

// A section missing from one language, or an empty string, fails here. The two
// languages are the same structure by construction; this proves they are also
// the same size.
func TestSiteCopyIsComplete(t *testing.T) {
	optional := func(path string) bool {
		return strings.HasSuffix(path, ".Bullets (empty list)") || strings.HasSuffix(path, ".Code") ||
			strings.HasSuffix(path, ".Paragraphs (empty list)")
	}
	for _, p := range copyPairs() {
		for lang, v := range map[string]any{"en": p.en, "id": p.id} {
			var empty []string
			emptyStrings(addressable(v), p.name+"/"+lang, &empty)
			for _, e := range empty {
				if !optional(e) {
					t.Errorf("empty copy: %s", e)
				}
			}
		}
		se, si := map[string]int{}, map[string]int{}
		shapeOf(addressable(p.en), p.name, se)
		shapeOf(addressable(p.id), p.name, si)
		if !reflect.DeepEqual(se, si) {
			for k, n := range se {
				if si[k] != n {
					t.Errorf("%s: English has %d entries, Indonesian has %d", k, n, si[k])
				}
			}
		}
	}
}

// Only the tokens site.go fills may appear in copy; a typo would ship as literal text.
func TestSiteCopyTokensAreKnown(t *testing.T) {
	known := map[string]bool{"{single}": true, "{volume}": true, "{volumeCredits}": true, "{whitebox}": true, "{packs}": true, "{ci}": true}
	for k := range apiParams {
		known[k] = true
	}
	for _, p := range copyPairs() {
		for lang, v := range map[string]any{"en": p.en, "id": p.id} {
			val := addressable(v)
			walkStrings(val, func(s string) string {
				for _, tok := range reToken.FindAllString(s, -1) {
					if !known[tok] {
						t.Errorf("%s/%s: unknown token %s in %q", p.name, lang, tok, s)
					}
				}
				return s
			})
		}
	}
}

// ---- rendered pages ---------------------------------------------------------

func TestPublicPagesRenderWithSoundMetadata(t *testing.T) {
	titles := map[string]string{}
	descs := map[string]string{}
	for _, path := range publicPagePaths {
		page := renderPublic(t, path)
		title, desc, canon := firstMatch(reTitle, page), firstMatch(reDesc, page), firstMatch(reCanon, page)

		if n := utf8.RuneCountInString(title); n < 25 || n > 72 {
			t.Errorf("%s: title is %d characters (%q), want 25-72", path, n, title)
		}
		if n := utf8.RuneCountInString(desc); n < 110 || n > 170 {
			t.Errorf("%s: description is %d characters, want 110-170: %q", path, n, desc)
		}
		if canon != siteBase+path {
			t.Errorf("%s: canonical = %q, want %q", path, canon, siteBase+path)
		}
		if other, dup := titles[title]; dup {
			t.Errorf("%s and %s share the title %q", path, other, title)
		}
		titles[title] = path
		if other, dup := descs[desc]; dup {
			t.Errorf("%s and %s share a description", path, other)
		}
		descs[desc] = path

		for _, want := range []string{`<meta name="robots" content="index,follow,max-image-preview:large"`, `property="og:title"`, `property="og:image"`, `name="twitter:card"`, `<main id="main"`, `href="#main"`, `name="viewport"`} {
			if !strings.Contains(page, want) {
				t.Errorf("%s: missing %q", path, want)
			}
		}
		if u := unfilled(visibleText(page)); len(u) > 0 {
			t.Errorf("%s: an unfilled {token} is visible: %v", path, u)
		}

		// One h1, and no skipped heading levels.
		prev, h1 := 0, 0
		for _, m := range reH.FindAllStringSubmatch(page, -1) {
			lvl := int(m[1][0] - '0')
			if lvl == 1 {
				h1++
			}
			if prev != 0 && lvl > prev+1 {
				t.Errorf("%s: heading jumps from h%d to h%d", path, prev, lvl)
			}
			prev = lvl
		}
		if h1 != 1 {
			t.Errorf("%s: %d h1 elements, want exactly 1", path, h1)
		}

		seen := map[string]bool{}
		for _, m := range reID.FindAllStringSubmatch(page, -1) {
			if seen[m[1]] {
				t.Errorf("%s: duplicate id %q", path, m[1])
			}
			seen[m[1]] = true
		}
		for _, img := range reImg.FindAllString(page, -1) {
			if !strings.Contains(img, ` alt=`) {
				t.Errorf("%s: image without alt: %s", path, img)
			}
		}
		// The terminal look is gone from the whole public site.
		for _, banned := range []string{"scanlines", "bg-ink", "glow", "$ cat ", "./run"} {
			if strings.Contains(page, banned) {
				t.Errorf("%s still carries the old terminal styling: %q", path, banned)
			}
		}
	}
}

func TestLanguagesAreMarkedAndCrossLinked(t *testing.T) {
	for _, p := range sitePages {
		en, id := siteBase+p.Paths["en"], siteBase+p.Paths["id"]
		for code, path := range p.Paths {
			page := renderPublic(t, path)
			wantLang := map[string]string{"en": `<html lang="en">`, "id": `<html lang="id">`}[code]
			if !strings.Contains(page, wantLang) {
				t.Errorf("%s: missing %s", path, wantLang)
			}
			alts := map[string]string{}
			for _, m := range reAlt.FindAllStringSubmatch(page, -1) {
				alts[m[1]] = m[2]
			}
			if alts["en"] != en || alts["id"] != id || alts["x-default"] != en {
				t.Errorf("%s: hreflang set is %v, want en=%s id=%s x-default=%s", path, alts, en, id, en)
			}
			if !strings.Contains(page, `hreflang="`+map[string]string{"en": "id", "id": "en"}[code]+`"`) {
				t.Errorf("%s: no visible link to the other language", path)
			}
		}
	}
}

func jsonLD(t *testing.T, page string) map[string][]map[string]any {
	t.Helper()
	m := reLD.FindStringSubmatch(page)
	if m == nil {
		t.Fatal("no JSON-LD block")
	}
	var doc struct {
		Context string           `json:"@context"`
		Graph   []map[string]any `json:"@graph"`
	}
	if err := json.Unmarshal([]byte(m[1]), &doc); err != nil {
		t.Fatalf("JSON-LD is not valid JSON: %v", err)
	}
	if doc.Context != "https://schema.org" {
		t.Errorf("@context = %q", doc.Context)
	}
	out := map[string][]map[string]any{}
	for _, n := range doc.Graph {
		out[n["@type"].(string)] = append(out[n["@type"].(string)], n)
	}
	return out
}

func TestStructuredDataIsPresentAndTruthful(t *testing.T) {
	pr := loadSitePricing("en")
	for _, path := range publicPagePaths {
		g := jsonLD(t, renderPublic(t, path))
		for _, typ := range []string{"Organization", "WebSite", "WebPage"} {
			if len(g[typ]) != 1 {
				t.Errorf("%s: %d %s nodes, want 1", path, len(g[typ]), typ)
			}
		}
		org := g["Organization"][0]
		if org["legalName"] != "Dalang Pte. Ltd." {
			t.Errorf("%s: legalName = %v", path, org["legalName"])
		}
		if id, _ := org["identifier"].(map[string]any); id["value"] != "202622331M" {
			t.Errorf("%s: UEN = %v", path, org["identifier"])
		}
		// No source says which languages support answers in (virus.ai.id's rule).
		if strings.Contains(renderPublic(t, path), "availableLanguage") {
			t.Errorf("%s claims availableLanguage", path)
		}
		isHome := path == "/" || path == "/id/"
		if !isHome && len(g["BreadcrumbList"]) != 1 {
			t.Errorf("%s: missing BreadcrumbList", path)
		}
	}

	// The offers are the prices the checkout charges.
	g := jsonLD(t, renderPublic(t, "/"))
	svc := g["Service"]
	if len(svc) != 1 {
		t.Fatalf("home has %d Service nodes", len(svc))
	}
	cat := svc[0]["offers"].(map[string]any)["itemListElement"].([]any)
	if len(cat) != len(pr.Rows)+1 {
		t.Fatalf("offers = %d, want %d packs + whitebox", len(cat), len(pr.Rows))
	}
	for i, r := range pr.Rows {
		o := cat[i].(map[string]any)
		if o["price"] != r.PriceNum || o["priceCurrency"] != "USD" || o["name"] != r.Label {
			t.Errorf("offer %d = %v, want %s (%s) USD", i, o, r.Label, r.PriceNum)
		}
	}
	if len(g["Article"]) != 0 {
		t.Error("home must not be marked up as an Article")
	}
	if len(jsonLD(t, renderPublic(t, "/external-assessment-vs-penetration-test"))["Article"]) != 1 {
		t.Error("the guide is not marked up as an Article")
	}
}

// The visible FAQ and the FAQPage JSON-LD come from one slice. This proves they
// still agree after rendering: same questions, same order, same answer text.
func TestFAQMarkupMatchesVisibleFAQ(t *testing.T) {
	for _, p := range sitePages {
		for code, path := range p.Paths {
			v := buildSiteView(p, langByCode(code), false, loadSitePricing(code))
			faq := pageFAQ(v)
			page := renderPublic(t, path)
			g := jsonLD(t, page)
			if len(faq) == 0 {
				if len(g["FAQPage"]) != 0 {
					t.Errorf("%s: FAQPage JSON-LD without a visible FAQ", path)
				}
				continue
			}
			if len(g["FAQPage"]) != 1 {
				t.Errorf("%s: want one FAQPage node", path)
				continue
			}
			ents := g["FAQPage"][0]["mainEntity"].([]any)
			if len(ents) != len(faq) {
				t.Errorf("%s: JSON-LD has %d questions, page has %d", path, len(ents), len(faq))
				continue
			}
			// Inline tags become a space in visibleText ("tiers</a>," reads "tiers ,"),
			// so compare with all whitespace removed.
			squash := func(s string) string { return strings.Join(strings.Fields(s), "") }
			vis := squash(visibleText(page))
			for i, q := range faq {
				e := ents[i].(map[string]any)
				if e["name"] != q.Q {
					t.Errorf("%s: question %d is %q in JSON-LD, %q on the page", path, i, e["name"], q.Q)
				}
				if got := e["acceptedAnswer"].(map[string]any)["text"]; got != plainText(q.A) {
					t.Errorf("%s: answer %d differs between JSON-LD and page", path, i)
				}
				if !strings.Contains(vis, squash(plainText(q.A))) {
					t.Errorf("%s: answer %d is not visible on the page", path, i)
				}
			}
		}
	}
}

// Prices on the pages are the prices in credit_packages, not copy.
func TestPagesFollowTheDatabasePrices(t *testing.T) {
	db.Exec(`UPDATE credit_packages SET usd_price=60 WHERE id='starter'`)
	t.Cleanup(func() { db.Exec(`UPDATE credit_packages SET usd_price=50 WHERE id='starter'`) })

	for _, path := range []string{"/", "/id/", "/enterprise", "/external-assessment-vs-penetration-test"} {
		page := renderPublic(t, path)
		vis := visibleText(page)
		if !strings.Contains(vis, "$60") {
			t.Errorf("%s does not show the changed starter price", path)
		}
		if strings.Contains(vis, "From $50") || strings.Contains(vis, "Mulai $50") {
			t.Errorf("%s still shows the old price", path)
		}
	}
	rr := httptest.NewRecorder()
	handleLLMS(false)(rr, httptest.NewRequest("GET", "/llms.txt", nil))
	if !strings.Contains(rr.Body.String(), "| 1 assessment | 1 | $60 |") {
		t.Errorf("llms.txt does not follow the database price:\n%s", rr.Body.String())
	}
}

func TestPricingFrom(t *testing.T) {
	sp := pricingFrom(defaultPacks, "en")
	if sp.From != "$50" || sp.Volume != "$20" || sp.VolumeCredits != "50" || sp.Whitebox != "$10,000" {
		t.Errorf("pricing = %+v", sp)
	}
	if !strings.Contains(sp.PacksLine, "Standard $100 (3 credits)") || !strings.Contains(sp.PacksLine, "Starter $50 (1 credit)") || !strings.Contains(sp.PacksLine, "Enterprise $1,000 (50 credits)") {
		t.Errorf("packs line = %q", sp.PacksLine)
	}
	if len(sp.Rows) != 4 || sp.Rows[1].PerScan != "$33.33" || sp.Rows[3].Price != "$1,000" || sp.Rows[3].PriceNum != "1000" {
		t.Errorf("rows = %+v", sp.Rows)
	}
	// Labels say what the pack buys, never "Enterprise".
	if sp.Rows[0].Label != "1 assessment" || sp.Rows[3].Label != "50 assessments" {
		t.Errorf("labels = %q, %q", sp.Rows[0].Label, sp.Rows[3].Label)
	}
	// Only the pack with the lowest price per assessment is marked as such.
	for i, r := range sp.Rows {
		if r.Best != (i == 3) {
			t.Errorf("row %d Best = %v", i, r.Best)
		}
	}
	id := pricingFrom(defaultPacks, "id")
	if id.Whitebox != "US$10.000" || id.Rows[3].Price != "US$1.000" || id.Rows[1].PerScan != "US$33,33" || id.Rows[3].Label != "50 asesmen" || !strings.Contains(id.PacksLine, "Enterprise US$1.000 (50 kredit)") {
		t.Errorf("Indonesian pricing = %+v", id)
	}
}

// ---- claims -----------------------------------------------------------------

// Things the site must never say: certifications that belong to someone else,
// customers that are Dalang infrastructure customers, scan capabilities that do
// not run, and durations nobody measured. See CLAIMS.md.
var bannedClaims = []string{
	"iso 27001", "iso 9001", "soc 2", "soc2", "pci dss", "hipaa", "gdpr compliant",
	"trusted by", "our customers include", "telkom", "binus", "trustmedis", "bp2mi",
	"vibe", "pentest as a service", "ptaas", "destructive mode",
	"open-port", "open port", "cms/stack", "screenshots",
	"few minutes", "within minutes", "in seconds",
	"guarantee", "100%", "unlimited", "zero-day", "military",
}

func TestPublicPagesMakeNoUnsupportedClaims(t *testing.T) {
	texts := map[string]string{}
	for _, path := range publicPagePaths {
		texts[path] = strings.ToLower(visibleText(renderPublic(t, path)))
	}
	for _, full := range []bool{false, true} {
		rr := httptest.NewRecorder()
		handleLLMS(full)(rr, httptest.NewRequest("GET", "/llms.txt", nil))
		texts["llms full="+map[bool]string{false: "no", true: "yes"}[full]] = strings.ToLower(rr.Body.String())
	}
	for name, text := range texts {
		for _, banned := range bannedClaims {
			// "does not scan ports" and "Network port scanning" are the honest wording.
			if strings.Contains(text, banned) {
				t.Errorf("%s says %q", name, banned)
			}
		}
	}
}

// What a Board reader is told about the operator, certification and human
// review must be exactly what CLAIMS.md allows.
func TestEnterpriseClaimsAreTheAllowedOnes(t *testing.T) {
	ent := visibleText(renderPublic(t, "/enterprise"))
	for _, want := range []string{
		"Dalang Pte. Ltd.", "202622331M", "160 Robinson Road",
		"makes no claim of certification or independent audit",
	} {
		if !strings.Contains(ent, want) {
			t.Errorf("enterprise page missing %q", want)
		}
	}
	if !strings.Contains(visibleText(renderPublic(t, "/")), "No human reads a standard report before delivery") {
		t.Error("the home page must state that no human reads a standard report")
	}
	trust := visibleText(renderPublic(t, "/trust"))
	for _, want := range []string{"AS154547", "20 October 2026", "security@dalang.io", "within 2 business days", "The response body is discarded"} {
		if !strings.Contains(trust, want) {
			t.Errorf("trust page missing %q", want)
		}
	}
}

// ---- links, sitemap, routing ------------------------------------------------

func TestInternalLinksAndAnchorsResolve(t *testing.T) {
	routes := map[string]bool{
		"/login": true, "/dashboard": true, "/checkout": true, "/contact": true, "/logout": true,
		"/sitemap.xml": true, "/llms.txt": true, "/llms-full.txt": true, "/.well-known/security.txt": true,
	}
	for _, p := range publicPagePaths {
		routes[p] = true
	}
	pages := map[string]string{}
	for _, p := range publicPagePaths {
		pages[p] = renderPublic(t, p)
	}
	for from, page := range pages {
		for _, m := range reHref.FindAllStringSubmatch(page, -1) {
			href := html.UnescapeString(m[1])
			path, frag := href, ""
			if i := strings.Index(href, "#"); i >= 0 {
				path, frag = href[:i], href[i+1:]
			}
			if i := strings.Index(path, "?"); i >= 0 {
				path = path[:i]
			}
			if path == "" {
				path = from
			}
			if !routes[path] {
				if _, err := fs.Stat(staticFS, "static"+path); err != nil {
					t.Errorf("%s links to %q, which is neither a route nor a static file", from, href)
				}
				continue
			}
			if target, ok := pages[path]; ok && frag != "" && !strings.Contains(target, `id="`+frag+`"`) {
				t.Errorf("%s links to %s#%s, but that page has no such anchor", from, path, frag)
			}
		}
	}
}

func TestSitemapListsEveryPageWithAlternates(t *testing.T) {
	rr := httptest.NewRecorder()
	handleSitemap(rr, httptest.NewRequest("GET", "/sitemap.xml", nil))
	body := rr.Body.String()
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/xml") {
		t.Errorf("content type = %q", ct)
	}
	locs := regexp.MustCompile(`<loc>([^<]+)</loc>`).FindAllStringSubmatch(body, -1)
	var got []string
	for _, m := range locs {
		got = append(got, strings.TrimPrefix(m[1], siteBase))
	}
	sort.Strings(got)
	want := append([]string(nil), publicPagePaths...)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("sitemap urls = %v\nwant %v", got, want)
	}
	for _, must := range []string{`hreflang="en"`, `hreflang="id"`, `hreflang="x-default"`, "<lastmod>" + siteModified + "</lastmod>"} {
		if !strings.Contains(body, must) {
			t.Errorf("sitemap missing %s", must)
		}
	}
	if strings.Contains(body, "/dashboard") || strings.Contains(body, "/reports/") || strings.Contains(body, "/su") {
		t.Error("sitemap lists a private page")
	}
}

func TestRobotsAndSecurityTxt(t *testing.T) {
	robots, err := staticFS.ReadFile("static/robots.txt")
	if err != nil {
		t.Fatal(err)
	}
	rs := string(robots)
	for _, want := range []string{"Sitemap: https://netsekurity.com/sitemap.xml", "Disallow: /dashboard", "Disallow: /checkout", "Disallow: /reports/", "Disallow: /su", "User-agent: GPTBot", "User-agent: ClaudeBot", "User-agent: PerplexityBot"} {
		if !strings.Contains(rs, want) {
			t.Errorf("robots.txt missing %q", want)
		}
	}
	if strings.Contains(rs, "Disallow: /\n") {
		t.Error("robots.txt blocks the whole site")
	}
	sec, _ := staticFS.ReadFile("static/security.txt")
	if !strings.Contains(string(sec), "Policy: https://netsekurity.com/trust#disclosure") {
		t.Error("security.txt does not point at the disclosure policy")
	}
}

func TestCanonicalPathsRedirect(t *testing.T) {
	for from, to := range map[string]string{"/enterprise/": "/enterprise", "/id": "/id/", "/trust/": "/trust", "/id/enterprise/": "/id/enterprise"} {
		rr := httptest.NewRecorder()
		handleIndex(rr, httptest.NewRequest("GET", from, nil))
		if rr.Code != http.StatusMovedPermanently || rr.Header().Get("Location") != to {
			t.Errorf("%s -> %d %q, want 301 %s", from, rr.Code, rr.Header().Get("Location"), to)
		}
	}
	rr := httptest.NewRecorder()
	handleIndex(rr, httptest.NewRequest("GET", "/id/nope", nil))
	if rr.Code != http.StatusNotFound {
		t.Errorf("unknown page = %d, want 404", rr.Code)
	}
}

func TestPagesAreNeverCachedBecauseOfTheNonce(t *testing.T) {
	get := func(path string) *httptest.ResponseRecorder {
		rr := httptest.NewRecorder()
		securityHeaders(http.HandlerFunc(handleIndex)).ServeHTTP(rr, httptest.NewRequest("GET", path, nil))
		return rr
	}
	a, b := get("/enterprise"), get("/enterprise")
	if cc := a.Header().Get("Cache-Control"); !strings.Contains(cc, "private") || !strings.Contains(cc, "no-store") {
		t.Errorf("Cache-Control = %q, want private, no-store", cc)
	}
	if a.Header().Get("ETag") != "" {
		t.Error("an ETag would let a 304 pair a cached body with a new nonce")
	}
	if a.Header().Get("Vary") != "Cookie" || a.Header().Get("Content-Language") != "en" {
		t.Errorf("headers = %v", a.Header())
	}
	na, nb := nonceOf(t, a), nonceOf(t, b)
	if na == "" || na == nb {
		t.Errorf("the nonce must be fresh per response: %q then %q", na, nb)
	}
	if id := get("/id/"); id.Header().Get("Content-Language") != "id" {
		t.Errorf("Indonesian page Content-Language = %q", id.Header().Get("Content-Language"))
	}
}

// A signed-in visitor is sent to the dashboard, an anonymous one to login.
func TestHeaderButtonFollowsSignIn(t *testing.T) {
	anon := renderPublic(t, "/enterprise")
	if !strings.Contains(anon, `href="/login" data-umami-event="netsekurity_header_masuk"`) {
		t.Error("anonymous visitor is not offered login")
	}
	const uid = "u_site_signedin"
	db.Exec(`INSERT OR IGNORE INTO users (id, email, name) VALUES (?, 'signedin@test.local', 'S')`, uid)
	tok, _ := issueJWT(uid, "signedin@test.local")
	req := httptest.NewRequest("GET", "/enterprise", nil)
	req.AddCookie(&http.Cookie{Name: authCookie, Value: tok})
	rr := httptest.NewRecorder()
	handleIndex(rr, req)
	if !strings.Contains(rr.Body.String(), `href="/dashboard" data-umami-event="netsekurity_header_masuk"`) {
		t.Error("signed-in visitor is not offered the dashboard")
	}
}

// ---- the contact form ---------------------------------------------------------

func TestContactFormAnswersInTheVisitorsLanguage(t *testing.T) {
	post := func(body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("POST", "/contact", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handleContact(rr, r)
		return rr
	}
	if out := post("email=lead-id%40example.com&message=halo&source=enterprise&lang=id"); !strings.Contains(out.Body.String(), "Pesan diterima") {
		t.Errorf("Indonesian ok = %s", out.Body.String())
	}
	if out := post("email=lead-en%40example.com&message=hello&source=enterprise&lang=en"); !strings.Contains(out.Body.String(), "Message received") {
		t.Errorf("English ok = %s", out.Body.String())
	}
	bad := post("email=nope&message=&lang=id")
	if bad.Code != http.StatusBadRequest || !strings.Contains(bad.Body.String(), "alamat email yang valid") {
		t.Errorf("Indonesian error = %d %s", bad.Code, bad.Body.String())
	}
	// The form on every page that carries it posts the language and the source.
	for _, path := range []string{"/", "/id/", "/enterprise", "/id/enterprise"} {
		page := renderPublic(t, path)
		for _, want := range []string{`hx-post="/contact"`, `name="lang"`, `name="source"`, `name="website"`} {
			if !strings.Contains(page, want) {
				t.Errorf("%s: contact form missing %s", path, want)
			}
		}
	}
}

// ---- analytics hooks the funnel reports depend on -----------------------------

func TestAnalyticsEventsSurvive(t *testing.T) {
	home := renderPublic(t, "/")
	for _, ev := range []string{"netsekurity_hero_cta", "netsekurity_header_masuk", "netsekurity_nav_harga", "netsekurity_nav_docs", "netsekurity_kontak_form_kirim", "netsekurity_pricing_buy", "netsekurity_footer_masuk"} {
		if !strings.Contains(home, `data-umami-event="`+ev+`"`) {
			t.Errorf("home lost the %s event", ev)
		}
	}
	if !strings.Contains(home, "umami.dalang.io/script.js") {
		t.Error("the cookieless analytics script is gone")
	}
	if !strings.Contains(home, "nskTrack('ViewContent'") {
		t.Error("home no longer reports ViewContent through the consent gate")
	}
	if !strings.Contains(home, `name="facebook-domain-verification"`) {
		t.Error("the Meta domain verification tag is gone")
	}
}

// ---- llms.txt ------------------------------------------------------------------

func TestLLMSFilesAreGeneratedAndCoherent(t *testing.T) {
	short := llmsText(false)
	full := llmsText(true)
	for name, txt := range map[string]string{"llms.txt": short, "llms-full.txt": full} {
		for _, want := range []string{"# Netsekurity", "not a full penetration test", "Dalang Pte. Ltd.", "| 1 assessment | 1 | $50 | $50 |", "| 50 assessments | 50 | $1,000 | $20 |", "$10,000", "sales@dalang.io", "security@dalang.io", "What is not covered", "Network port scanning"} {
			if !strings.Contains(txt, want) {
				t.Errorf("%s missing %q", name, want)
			}
		}
		if u := unfilled(txt); len(u) > 0 {
			t.Errorf("%s has an unfilled token %v", name, u)
		}
	}
	for _, want := range []string{"## For procurement and risk", "## Trust and security", "## Product scope (verbatim)", "Network port scanning"} {
		if !strings.Contains(full, want) {
			t.Errorf("llms-full.txt missing %q", want)
		}
	}
	if len(full) <= len(short) {
		t.Error("llms-full.txt is not fuller than llms.txt")
	}
	if !strings.Contains(short, "/id/") {
		t.Error("llms.txt does not link the Indonesian pages")
	}
}

// ---- product scope --------------------------------------------------------------

func TestProductScopeIsSemanticHTMLAndAccurate(t *testing.T) {
	page := renderPublic(t, "/product-scope")
	for _, want := range []string{"<h2 id=", "<table", "<ol>", "<blockquote>", `class="nsk-doc`} {
		if !strings.Contains(page, want) {
			t.Errorf("scope page missing %q", want)
		}
	}
	if strings.Contains(page, "<pre") {
		t.Error("scope page is preformatted text again")
	}
	text := strings.ToLower(visibleText(page))
	// The scanner's non-deep run does not scan ports or identify CMS stacks.
	if !strings.Contains(text, "network port scanning") {
		t.Error("scope does not list port scanning as not included")
	}
	if !strings.Contains(text, "subdomains are listed") {
		t.Error("scope does not say subdomains are listed, not assessed")
	}
}

// The share cards exist, are 1200x630 PNGs, and each language links its own.
func TestOpenGraphImages(t *testing.T) {
	for name, want := range map[string]string{"/og.png": "en", "/og-id.png": "id"} {
		f, err := staticFS.Open("static" + name)
		if err != nil {
			t.Fatalf("%s missing: %v", name, err)
		}
		cfg, err := png.DecodeConfig(f)
		f.Close()
		if err != nil || cfg.Width != 1200 || cfg.Height != 630 {
			t.Errorf("%s: %v %dx%d, want a 1200x630 PNG", name, err, cfg.Width, cfg.Height)
		}
		path := map[string]string{"en": "/enterprise", "id": "/id/enterprise"}[want]
		page := renderPublic(t, path)
		if !strings.Contains(page, `property="og:image" content="`+siteBase+name+"?v="+assetHash(name)+`"`) {
			t.Errorf("%s does not use %s", path, name)
		}
		if !strings.Contains(page, `property="og:image:alt"`) || !strings.Contains(page, `name="twitter:image:alt"`) {
			t.Errorf("%s: the share image has no alt text", path)
		}
	}
}

// The pages say the scanner identifies itself as "NetSekurity Scanner" and paces
// its probes. Both statements are checked against the scanner source, so the
// claim and the code cannot drift apart (and a Chrome string cannot come back).
func TestScannerUserAgentMatchesThePublicClaim(t *testing.T) {
	src, err := os.ReadFile("nsec_scan.py")
	if err != nil {
		t.Fatal(err)
	}
	m := regexp.MustCompile(`(?m)^UA = "([^"]+)"`).FindSubmatch(src)
	if m == nil || string(m[1]) != "NetSekurity Scanner" {
		t.Fatalf("nsec_scan.py UA = %q, want \"NetSekurity Scanner\"", m)
	}
	if regexp.MustCompile(`(?i)mozilla|chrome/|safari/`).Match(src) {
		t.Error("nsec_scan.py still contains a browser User-Agent string")
	}
	if n := strings.Count(string(src), `-A '{UA}'`); n < 4 {
		t.Errorf("only %d curl calls pass the User-Agent; every HTTP request must", n)
	}
	if n := strings.Count(string(src), "time.sleep(0.3)"); n < 2 {
		t.Errorf("found %d 0.3s pauses; the exposure and redirect loops must both pace themselves", n)
	}
	for _, path := range []string{"/trust", "/id/trust", "/enterprise", "/id/enterprise"} {
		if !strings.Contains(visibleText(renderPublic(t, path)), "NetSekurity Scanner") {
			t.Errorf("%s does not state the scanner's User-Agent", path)
		}
	}
}

// Dalang's privacy policy, terms of use and security and trust center are linked
// from every public page, and the three destinations are exactly the ones the
// owner named.
func TestPolicyLinksAreOnEveryPublicPage(t *testing.T) {
	for _, path := range publicPagePaths {
		page := renderPublic(t, path)
		for _, want := range []string{`href="https://dalang.io/privacy"`, `href="https://dalang.io/terms-of-use"`, `href="https://dalang.io/security"`} {
			if !strings.Contains(page, want) {
				t.Errorf("%s: footer missing %s", path, want)
			}
		}
	}
	for _, path := range []string{"/trust", "/id/trust"} {
		if strings.Count(renderPublic(t, path), `href="https://dalang.io/privacy"`) < 2 {
			t.Errorf("%s: the operator section does not repeat the policy links", path)
		}
	}
}

var reNonce = regexp.MustCompile(`script-src 'nonce-([^']+)'`)

func nonceOf(t *testing.T, rr *httptest.ResponseRecorder) string {
	t.Helper()
	m := reNonce.FindStringSubmatch(rr.Header().Get("Content-Security-Policy"))
	if m == nil {
		return ""
	}
	return m[1]
}

// Money is written the way the reader writes it, and consistently: the site once
// printed "$1000" next to "$10,000". Thousands get a separator in every language.
func TestMoneyIsFormattedConsistently(t *testing.T) {
	for _, tc := range []struct {
		v    float64
		lang string
		want string
	}{
		{50, "en", "$50"}, {1000, "en", "$1,000"}, {10000, "en", "$10,000"}, {33.33, "en", "$33.33"}, {49.5, "en", "$49.50"}, {0, "en", "$0"},
		{50, "id", "US$50"}, {1000, "id", "US$1.000"}, {10000, "id", "US$10.000"}, {33.33, "id", "US$33,33"},
	} {
		if got := formatUSD(tc.v, tc.lang); got != tc.want {
			t.Errorf("formatUSD(%v,%q) = %q, want %q", tc.v, tc.lang, got, tc.want)
		}
	}
	bare := regexp.MustCompile(`\$\d{4,}`) // a dollar amount of 1000 or more with no separator
	for _, path := range publicPagePaths {
		page := visibleText(renderPublic(t, path))
		if m := bare.FindString(page); m != "" {
			t.Errorf("%s prints %q without a thousands separator", path, m)
		}
	}
	if !strings.Contains(visibleText(renderPublic(t, "/")), "$1,000") || !strings.Contains(visibleText(renderPublic(t, "/")), "$10,000") {
		t.Error("English home does not show $1,000 and $10,000")
	}
	if id := visibleText(renderPublic(t, "/id/")); !strings.Contains(id, "US$1.000") || !strings.Contains(id, "US$10.000") || strings.Contains(id, "$1,000") {
		t.Error("Indonesian home does not use Indonesian money formatting")
	}
}
