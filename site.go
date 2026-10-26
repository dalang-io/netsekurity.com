package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

const (
	siteBase = "https://netsekurity.com"
	// siteModified is the dateModified of the public pages and the sitemap's
	// lastmod. Bump it whenever the copy changes (virus.ai.id keeps the same rule).
	siteModified = "2026-10-26"
	// whiteboxUSD is the whitebox tier price from static/PRODUCT_SCOPE.md.
	whiteboxUSD = 10000
	// umamiSite is the cookieless analytics site id (see consent.go).
	umamiSite = "a64a56c2-0b87-41a8-a254-bca52c48e2c1"
)

type siteLang struct {
	Code, HTMLLang, OGLocale, Label, Prefix string
	// Name is what a screen reader announces for the language switch: the visible
	// text is only "EN" or "ID", which says nothing to someone who cannot see the page.
	Name string
}

var siteLangs = []siteLang{
	{"en", "en", "en_US", "English", "", "English (EN)"},
	{"id", "id", "id_ID", "Bahasa Indonesia", "/id", "Bahasa Indonesia (ID)"},
}

func langByCode(code string) siteLang {
	for _, l := range siteLangs {
		if l.Code == code {
			return l
		}
	}
	return siteLangs[0]
}

// sitePage is one logical page, with a path per language.
type sitePage struct {
	Key      string
	Paths    map[string]string
	Priority string
	Change   string
	HTMX     bool // the page carries the contact form, which posts with HTMX
	Schema   string
}

var sitePages = []sitePage{
	{Key: "home", Paths: map[string]string{"en": "/", "id": "/id/"}, Priority: "1.0", Change: "weekly", HTMX: true, Schema: "service"},
	{Key: "enterprise", Paths: map[string]string{"en": "/enterprise", "id": "/id/enterprise"}, Priority: "0.9", Change: "monthly", HTMX: true},
	{Key: "trust", Paths: map[string]string{"en": "/trust", "id": "/id/trust"}, Priority: "0.8", Change: "monthly"},
	{Key: "compare", Paths: map[string]string{"en": "/external-assessment-vs-penetration-test", "id": "/id/asesmen-eksternal-vs-uji-penetrasi"}, Priority: "0.8", Change: "monthly", Schema: "article"},
	// The evidence pages. They used to be English-only, which flipped an Indonesian
	// reader's language mid-journey the moment they followed a link to proof.
	{Key: "sample", Paths: map[string]string{"en": "/sample-report", "id": "/id/contoh-laporan"}, Priority: "0.9", Change: "monthly"},
	{Key: "scope", Paths: map[string]string{"en": "/product-scope", "id": "/id/cakupan-produk"}, Priority: "0.8", Change: "monthly"},
	{Key: "start", Paths: map[string]string{"en": "/getting-started", "id": "/id/panduan-memulai"}, Priority: "0.8", Change: "monthly"},
	{Key: "docs", Paths: map[string]string{"en": "/docs", "id": "/id/dokumentasi"}, Priority: "0.7", Change: "weekly"},
}

func pageByKey(key string) (sitePage, bool) {
	for _, p := range sitePages {
		if p.Key == key {
			return p, true
		}
	}
	return sitePage{}, false
}

// lookupSitePage resolves a request path to a page and language.
func lookupSitePage(path string) (sitePage, siteLang, bool) {
	for _, p := range sitePages {
		for code, pp := range p.Paths {
			if pp == path {
				return p, langByCode(code), true
			}
		}
	}
	return sitePage{}, siteLang{}, false
}

// canonicalSitePath returns the canonical form when path is a page path with a
// stray trailing slash (or "/id"), so one page has one URL.
func canonicalSitePath(path string) string {
	if path == "/id" {
		return "/id/"
	}
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		if _, _, ok := lookupSitePage(strings.TrimRight(path, "/")); ok {
			return strings.TrimRight(path, "/")
		}
	}
	return ""
}

// ---- pricing ---------------------------------------------------------------

type packRow struct {
	ID, Name, Label, Credits, Price, PriceNum, PerScan string
	Best                                               bool // the lowest price per assessment
}

type sitePricing struct {
	Lang          string
	Rows          []packRow
	From          string // cheapest pack, e.g. "$50"
	Volume        string // best per-assessment price, e.g. "$20"
	VolumeCredits string // credits in that pack, e.g. "50"
	Whitebox      string
	PacksLine     string
}

// defaultPacks mirror the seeds in db.go; used only when the table is empty
// (a fresh database), so the page is never blank.
var defaultPacks = []pkg{
	{ID: "starter", Name: "Starter", USD: 50, Credits: 1},
	{ID: "standard", Name: "Standard", USD: 100, Credits: 3},
	{ID: "professional", Name: "Professional", USD: 500, Credits: 20},
	{ID: "enterprise", Name: "Enterprise", USD: 1000, Credits: 50},
}

func loadSitePricing(lang string) sitePricing {
	var packs []pkg
	if rows, err := db.Query(`SELECT id, name, usd_price, credits FROM credit_packages WHERE is_active=1 ORDER BY usd_price`); err == nil {
		for rows.Next() {
			var p pkg
			rows.Scan(&p.ID, &p.Name, &p.USD, &p.Credits)
			packs = append(packs, p)
		}
		rows.Close()
	}
	if len(packs) == 0 {
		packs = defaultPacks
	}
	return pricingFrom(packs, lang)
}

// pricingFrom formats the packs for one language. PriceNum is the plain number the
// structured data needs; every visible price goes through formatUSD.
func pricingFrom(packs []pkg, lang string) sitePricing {
	sp := sitePricing{Lang: lang, Whitebox: formatUSD(whiteboxUSD, lang)}
	best := -1.0
	bestIdx := 0
	from := packs[0].USD
	var line []string
	for i, p := range packs {
		if p.USD < from {
			from = p.USD
		}
		per := 0.0
		if p.Credits > 0 {
			per = p.USD / p.Credits
		}
		if best < 0 || per < best {
			best, bestIdx = per, i
		}
		sp.Rows = append(sp.Rows, packRow{
			ID: p.ID, Name: p.Name, Label: assessmentsLabel(p.Credits, lang), Credits: fmtCredits(p.Credits),
			Price: formatUSD(p.USD, lang), PriceNum: trimUSD(p.USD), PerScan: perScanLang(p.USD, p.Credits, lang),
		})
		line = append(line, fmt.Sprintf("%s %s (%s)", p.Name, formatUSD(p.USD, lang), creditsLabelLang(p.Credits, lang)))
	}
	// Only a real difference earns the "lowest price" mark.
	if len(packs) > 1 && best < packs[0].USD/maxf(packs[0].Credits, 1) {
		sp.Rows[bestIdx].Best = true
	}
	sp.From = formatUSD(from, lang)
	sp.Volume = perScanLang(packs[bestIdx].USD, packs[bestIdx].Credits, lang)
	sp.VolumeCredits = fmtCredits(packs[bestIdx].Credits)
	sp.PacksLine = strings.Join(line, "; ")
	return sp
}

func maxf(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func (sp sitePricing) replacer() *strings.Replacer {
	return strings.NewReplacer(
		"{single}", sp.From, "{volume}", sp.Volume, "{volumeCredits}", sp.VolumeCredits,
		"{whitebox}", sp.Whitebox, "{packs}", sp.PacksLine,
		"{ci}", strconv.Itoa(len(ciExamples)),
	)
}

// ---- view model ------------------------------------------------------------

type siteAlt struct{ Lang, Href, OG string }

type siteNav struct{ Key, Label, Href, Event string }

type siteLangLink struct {
	Code, Label, Name, Href string
	Current                 bool
}

type siteView struct {
	Lang        siteLang
	Key         string
	Path        string
	Title       string
	Description string
	Canonical   string
	OGImage     string
	Robots      string
	Alternates  []siteAlt
	JSONLD      template.JS
	C           commonCopy
	Pricing     sitePricing
	Nav         []siteNav
	LangLinks   []siteLangLink
	AuthHref    string
	AuthLabel   string
	StartHref   string
	CSSHash     string
	FBPixel     template.HTML
	Analytics   template.HTML
	ViewContent template.HTML
	HTMX        bool
	Modified    string
	ScansProof  template.HTML // "N assessments delivered", only above scansProofMin
	Nonce       string        // per-response CSP nonce; every inline script carries it
	Home        *homeCopy
	Ent         *enterpriseCopy
	Trust       *trustCopy
	Cmp         *compareCopy
	Sample      *sampleCopy
	Docs        *docsCopy
	Doc         template.HTML // the product scope, rendered from its Markdown
	SamplePDF   string        // the printable copy of the sample report, in this page's language
	Examples    []ciView      // the CI examples, with notes in this page's language
	APIExample  string
	DocsStart   string
	Crumb       string
	HomeHref    string
	EntHref     string
	TrustHref   string
	GuideHref   string
	ContactHref string
	SampleHref  string
	DocsHref    string
	SetupHref   string
	ScopeHref   string
	ProfileHref string // Dalang's Company Profile PDF, in this page's language
}

// ciView is one CI example as the docs page shows it.
type ciView struct {
	Title, Code string
	Note        template.HTML
}

func pathFor(key, lang string) string {
	p, _ := pageByKey(key)
	return p.Paths[lang]
}

// ogImageFor is the share card for a language. The version query is the file's content
// hash, so a crawler that cached the old card fetches the new one exactly when the card
// changes (tools/og/render.sh), and not on every unrelated release.
func ogImageFor(lang siteLang) string {
	name := "/og.png"
	if lang.Code == "id" {
		name = "/og-id.png"
	}
	return siteBase + assetURL(name)
}

// baseView builds the parts every page shares: head metadata, navigation,
// language links and the auth-aware buttons. paths is the page's path per language.
func baseView(key, path string, lang siteLang, paths map[string]string, htmx, authed bool, pr sitePricing) *siteView {
	rep := pr.replacer()
	expand := func(s string) string { return rep.Replace(s) }

	c := commonFor(lang.Code)
	walkStrings(reflect.ValueOf(&c).Elem(), expand)

	nonce := newNonce()
	v := &siteView{
		Lang: lang, Key: key, Path: path, C: c, Pricing: pr, Nonce: nonce,
		Robots: "index,follow,max-image-preview:large", OGImage: ogImageFor(lang),
		CSSHash: cssHash, HTMX: htmx, Modified: siteModified,
		FBPixel:   template.HTML(withNonce(consentGate(true), nonce)),
		Analytics: template.HTML(withNonce(analyticsSnippet(), nonce)),
		HomeHref:  pathFor("home", lang.Code), EntHref: pathFor("enterprise", lang.Code),
		TrustHref: pathFor("trust", lang.Code), GuideHref: pathFor("compare", lang.Code),
		SampleHref: pathFor("sample", lang.Code), DocsHref: pathFor("docs", lang.Code), ScopeHref: pathFor("scope", lang.Code),
		ProfileHref: "https://dalang.io/docs/dalang-company-profile-" + lang.Code + ".pdf",
		APIExample:  apiExample, DocsStart: docsStartExample,
	}
	v.SetupHref = pathFor("start", lang.Code)
	v.ContactHref = v.EntHref + "#contact"
	v.Canonical = siteBase + path

	for _, l := range siteLangs {
		v.Alternates = append(v.Alternates, siteAlt{Lang: l.HTMLLang, Href: siteBase + paths[l.Code], OG: l.OGLocale})
		v.LangLinks = append(v.LangLinks, siteLangLink{Code: l.Code, Label: strings.ToUpper(l.Code), Name: l.Name, Href: paths[l.Code], Current: l.Code == lang.Code})
	}
	v.Alternates = append(v.Alternates, siteAlt{Lang: "x-default", Href: siteBase + paths["en"]})

	home := v.HomeHref
	v.Nav = []siteNav{
		{"how", c.NavHow, home + "#how", "netsekurity_nav_cara_kerja"},
		{"sample", c.NavSample, v.SampleHref, "netsekurity_nav_sample"},
		{"pricing", c.NavPricing, home + "#pricing", "netsekurity_nav_harga"},
		{"enterprise", c.NavEnterprise, v.EntHref, "netsekurity_nav_enterprise"},
	}
	if authed {
		v.AuthHref, v.AuthLabel, v.StartHref = "/dashboard", c.Dashboard, "/dashboard"
	} else {
		v.AuthHref, v.AuthLabel, v.StartHref = "/login", c.LogIn, "/login"
	}
	return v
}

// buildSiteView assembles everything a bilingual page template needs.
func buildSiteView(p sitePage, lang siteLang, authed bool, pr sitePricing) *siteView {
	v := baseView(p.Key, p.Paths[lang.Code], lang, p.Paths, p.HTMX, authed, pr)
	rep := pr.replacer()
	expand := func(s string) string { return rep.Replace(s) }
	c := v.C

	switch p.Key {
	case "home":
		h := homeFor(lang.Code)
		walkStrings(reflect.ValueOf(&h).Elem(), expand)
		v.Home, v.Title, v.Description = &h, h.Title, h.Description
		v.ScansProof = template.HTML(renderScansProof())
		v.ViewContent = template.HTML(withNonce(metaTrackJS("ViewContent", `{"content_name":"Pricing","content_type":"product","currency":"USD","value":0.0}`), v.Nonce))
	case "enterprise":
		e := enterpriseFor(lang.Code)
		walkStrings(reflect.ValueOf(&e).Elem(), expand)
		v.Ent, v.Title, v.Description, v.Crumb = &e, e.Title, e.Description, c.NavEnterprise
	case "trust":
		t := trustFor(lang.Code)
		walkStrings(reflect.ValueOf(&t).Elem(), expand)
		v.Trust, v.Title, v.Description, v.Crumb = &t, t.Title, t.Description, c.NavTrust
	case "compare":
		cm := compareFor(lang.Code)
		walkStrings(reflect.ValueOf(&cm).Elem(), expand)
		v.Cmp, v.Title, v.Description, v.Crumb = &cm, cm.Title, cm.Description, c.LinkGuide
	case "sample":
		sm := sampleFor(lang.Code)
		walkStrings(reflect.ValueOf(&sm).Elem(), expand)
		v.Sample, v.Title, v.Description, v.Crumb = &sm, sm.Title, sm.Description, c.NavSample
		v.SamplePDF = assetURL(samplePDFPath(lang.Code))
	case "docs":
		d := docsFor(lang.Code)
		walkStrings(reflect.ValueOf(&d).Elem(), expand)
		v.Docs, v.Title, v.Description, v.Crumb = &d, d.Title, d.Description, c.NavDocs
		for i, e := range ciExamples {
			v.Examples = append(v.Examples, ciView{Title: e.Title, Code: e.Code, Note: d.CINotes[i]})
		}
	case "start":
		v.Crumb = c.LinkStartGuide
		name := "static/GETTING_STARTED.md"
		v.Title = "Netsekurity: Start Your First Website Security Check"
		v.Description = "Prepare your website, choose DNS, file or tag verification, start a check and share the results with your developer. A practical guide for website owners."
		if lang.Code == "id" {
			name = "static/GETTING_STARTED.id.md"
			v.Title = "Netsekurity: Panduan Memulai Pemeriksaan Website"
			v.Description = "Siapkan website, pilih verifikasi DNS, file atau tag, mulai pemeriksaan, dan berikan hasil kepada developer. Panduan praktis untuk pemilik website."
		}
		if md, err := staticFS.ReadFile(name); err == nil {
			v.Doc = renderMarkdown(string(md))
		}
	case "scope":
		sc := scopeFor(lang.Code)
		v.Title, v.Description, v.Crumb = sc.Title, sc.Description, sc.Crumb
		v.Doc = scopeDocument(lang.Code)
	}
	v.JSONLD = template.JS(siteJSONLD(p, lang, v))
	return v
}

func commonFor(code string) commonCopy {
	if code == "id" {
		return commonID()
	}
	return commonEN()
}

func homeFor(code string) homeCopy {
	if code == "id" {
		return homeID()
	}
	return homeEN()
}

func enterpriseFor(code string) enterpriseCopy {
	if code == "id" {
		return enterpriseID()
	}
	return enterpriseEN()
}

func trustFor(code string) trustCopy {
	if code == "id" {
		return trustID()
	}
	return trustEN()
}

func compareFor(code string) compareCopy {
	if code == "id" {
		return compareID()
	}
	return compareEN()
}

func sampleFor(code string) sampleCopy {
	if code == "id" {
		return sampleID()
	}
	return sampleEN()
}

func docsFor(code string) docsCopy {
	if code == "id" {
		return docsID()
	}
	return docsEN()
}

// samplePDFPath is the printable copy of the sample report. The files are rendered
// from the page by tools/pdf/render.sh and carry no price; TestSamplePDFsMatchTheirSource
// fails when the copy they were made from has changed.
func samplePDFPath(code string) string { return "/netsekurity-sample-report-" + code + ".pdf" }

func scopeFor(code string) scopeCopy {
	if code == "id" {
		return scopeID()
	}
	return scopeEN()
}

// scopeDocument renders the product scope from its Markdown source, one file per
// language (static/PRODUCT_SCOPE.md and static/PRODUCT_SCOPE.id.md). A missing file
// yields an empty document rather than a 500; the tests fail on that.
func scopeDocument(code string) template.HTML {
	name := "static/PRODUCT_SCOPE.md"
	if code == "id" {
		name = "static/PRODUCT_SCOPE.id.md"
	}
	md, err := staticFS.ReadFile(name)
	if err != nil {
		return ""
	}
	return renderMarkdown(string(md))
}

// ---- structured data -------------------------------------------------------

func orgNode() map[string]any {
	return map[string]any{
		"@type": "Organization", "@id": siteBase + "/#dalang",
		"name": "Dalang", "legalName": "Dalang Pte. Ltd.", "url": "https://dalang.io/",
		"identifier": map[string]any{"@type": "PropertyValue", "propertyID": "UEN", "value": "202622331M"},
		"address": map[string]any{
			"@type": "PostalAddress", "streetAddress": "160 Robinson Road, #14-04",
			"addressLocality": "Singapore", "postalCode": "068914", "addressCountry": "SG",
		},
		"email": "sales@dalang.io",
		// No availableLanguage: no source says which languages support answers in.
		"contactPoint": []any{
			map[string]any{"@type": "ContactPoint", "contactType": "sales", "email": "sales@dalang.io"},
			map[string]any{"@type": "ContactPoint", "contactType": "security", "email": "security@dalang.io"},
		},
	}
}

func faqNode(v *siteView, items []qa) map[string]any {
	var ents []any
	for _, q := range items {
		ents = append(ents, map[string]any{
			"@type": "Question", "name": q.Q,
			"acceptedAnswer": map[string]any{"@type": "Answer", "text": plainText(q.A)},
		})
	}
	return map[string]any{"@type": "FAQPage", "@id": v.Canonical + "#faq", "mainEntity": ents, "inLanguage": v.Lang.HTMLLang}
}

func pageFAQ(v *siteView) []qa {
	switch {
	case v.Home != nil:
		return v.Home.FAQ
	case v.Ent != nil:
		return v.Ent.FAQ
	case v.Cmp != nil:
		return v.Cmp.FAQ
	}
	return nil
}

func siteJSONLD(p sitePage, lang siteLang, v *siteView) string {
	org := map[string]any{"@id": siteBase + "/#dalang"}
	var inLangs []string
	for _, l := range siteLangs {
		inLangs = append(inLangs, l.HTMLLang)
	}
	graph := []any{
		orgNode(),
		map[string]any{
			"@type": "WebSite", "@id": siteBase + "/#website", "url": siteBase + "/",
			"name": "Netsekurity", "inLanguage": inLangs, "publisher": org,
		},
	}

	page := map[string]any{
		"@type": "WebPage", "@id": v.Canonical + "#webpage", "url": v.Canonical,
		"name": v.Title, "description": v.Description, "inLanguage": lang.HTMLLang,
		"isPartOf": map[string]any{"@id": siteBase + "/#website"}, "dateModified": siteModified,
	}
	if p.Key != "home" {
		page["breadcrumb"] = map[string]any{"@id": v.Canonical + "#breadcrumb"}
	}
	graph = append(graph, page)

	if p.Key != "home" {
		graph = append(graph, map[string]any{
			"@type": "BreadcrumbList", "@id": v.Canonical + "#breadcrumb",
			"itemListElement": []any{
				map[string]any{"@type": "ListItem", "position": 1, "name": v.C.Breadcrumb, "item": siteBase + v.HomeHref},
				map[string]any{"@type": "ListItem", "position": 2, "name": v.Crumb, "item": v.Canonical},
			},
		})
	}

	if p.Schema == "service" {
		var offers []any
		for _, r := range v.Pricing.Rows {
			offers = append(offers, map[string]any{
				"@type": "Offer", "name": r.Label,
				"price": r.PriceNum, "priceCurrency": "USD",
				"url": siteBase + pathFor("home", lang.Code) + "#pricing",
			})
		}
		offers = append(offers, map[string]any{
			"@type": "Offer", "name": "Whitebox assessment (per app or domain)",
			"price": fmt.Sprint(whiteboxUSD), "priceCurrency": "USD",
			"url": siteBase + pathFor("home", lang.Code) + "#tiers",
		})
		graph = append(graph, map[string]any{
			"@type": "Service", "@id": siteBase + "/#service",
			"name": "Netsekurity external web security assessment", "serviceType": "External web security assessment",
			"description": "Read-only, unauthenticated external assessment of a website the customer has proved it owns, delivered as a PDF report with CVSS-scored findings, evidence and a 30/60/90-day remediation plan.",
			"provider":    org, "url": siteBase + "/", "areaServed": "Worldwide",
			"offers": map[string]any{"@type": "OfferCatalog", "name": "Assessment credit packs and deeper tiers", "itemListElement": offers},
		})
	}
	if p.Schema == "article" {
		graph = append(graph, map[string]any{
			"@type": "Article", "@id": v.Canonical + "#article", "headline": v.Cmp.H1,
			"description": v.Description, "inLanguage": lang.HTMLLang,
			"datePublished": siteModified, "dateModified": siteModified,
			"author": org, "publisher": org, "mainEntityOfPage": map[string]any{"@id": v.Canonical + "#webpage"},
		})
	}
	if faq := pageFAQ(v); len(faq) > 0 {
		graph = append(graph, faqNode(v, faq))
	}

	b, err := json.Marshal(map[string]any{"@context": "https://schema.org", "@graph": graph})
	if err != nil {
		return "{}"
	}
	return string(b)
}

// ---- rendering -------------------------------------------------------------

var siteTemplates = map[string]*template.Template{}

var siteFuncs = template.FuncMap{
	"icon":  icon,
	"pick":  pickIcon,
	"asset": assetURL,
	"join":  strings.Join,
	"inc":   func(i int) int { return i + 1 },
	// codeArgs is the data the "codebox" template takes: the code and the two words
	// on its copy button, in the page's language.
	"codeArgs": func(c commonCopy, code string) map[string]string {
		return map[string]string{"Code": code, "Copy": c.CopyLabel, "Copied": c.CopiedLabel}
	},
}

func init() {
	bodies := map[string]string{
		"home": siteHomeHTML, "enterprise": siteEnterpriseHTML, "trust": siteTrustHTML, "compare": siteCompareHTML,
		"sample": siteSampleHTML, "docs": siteDocsHTML, "scope": siteScopeHTML, "start": siteScopeHTML,
	}
	for key, body := range bodies {
		siteTemplates[key] = template.Must(template.New(key).Funcs(siteFuncs).Parse(siteLayoutHTML + siteSharedHTML + siteSVGHTML + body))
	}
}

func renderSitePageHTML(p sitePage, lang siteLang, authed bool, pr sitePricing) ([]byte, *siteView, error) {
	v := buildSiteView(p, lang, authed, pr)
	var buf bytes.Buffer
	if err := siteTemplates[p.Key].ExecuteTemplate(&buf, "layout", v); err != nil {
		return nil, v, err
	}
	return buf.Bytes(), v, nil
}

func serveSitePage(w http.ResponseWriter, r *http.Request, p sitePage, lang siteLang) {
	_, authErr := currentUser(r)
	body, v, err := renderSitePageHTML(p, lang, authErr == nil, loadSitePricing(lang.Code))
	if err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
		return
	}
	writeSiteHTML(w, r, body, lang.HTMLLang, v.Nonce)
}

// writeSiteHTML sends a rendered page. It deliberately sends no ETag and forbids
// caching: every response carries its own CSP nonce, so a 304 would pair a cached
// body (old nonce) with a fresh header (new nonce) and block every script, and a
// shared cache that replayed one response would reuse a nonce across visitors. The
// body also varies with sign-in state (the header button), hence Vary.
func writeSiteHTML(w http.ResponseWriter, r *http.Request, body []byte, htmlLang, nonce string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("Vary", "Cookie")
	w.Header().Set("Content-Language", htmlLang)
	w.Header().Set("Content-Security-Policy", strictCSP(nonce))
	w.Write(body)
}

// handleSitemap generates the sitemap from the page registry, so a page cannot
// be added or renamed without the sitemap following.
func handleSitemap(w http.ResponseWriter, r *http.Request) {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml">` + "\n")
	for _, p := range sitePages {
		codes := make([]string, 0, len(p.Paths))
		for c := range p.Paths {
			codes = append(codes, c)
		}
		sort.Strings(codes)
		for _, code := range codes {
			l := langByCode(code)
			fmt.Fprintf(&b, "  <url>\n    <loc>%s%s</loc>\n    <lastmod>%s</lastmod>\n    <changefreq>%s</changefreq>\n    <priority>%s</priority>\n",
				siteBase, p.Paths[code], siteModified, p.Change, p.Priority)
			for _, alt := range siteLangs {
				fmt.Fprintf(&b, `    <xhtml:link rel="alternate" hreflang="%s" href="%s%s"/>`+"\n", alt.HTMLLang, siteBase, p.Paths[alt.Code])
			}
			fmt.Fprintf(&b, `    <xhtml:link rel="alternate" hreflang="x-default" href="%s%s"/>`+"\n  </url>\n", siteBase, p.Paths["en"])
			_ = l
		}
	}
	b.WriteString("</urlset>\n")
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Write([]byte(b.String()))
}
