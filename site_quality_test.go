package main

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"
)

// Tests for the quality fixes of 2026-10-26: spelling and terminology, icons,
// the bilingual evidence pages, mobile-safe tables, copy buttons, print, and the
// CI gate script the docs show. Each one fails on the specific defect it names.

func englishPaths() []string {
	var out []string
	for _, p := range publicPagePaths {
		if !strings.HasPrefix(p, "/id/") {
			out = append(out, p)
		}
	}
	return out
}

func indonesianPaths() []string {
	var out []string
	for _, p := range publicPagePaths {
		if strings.HasPrefix(p, "/id/") {
			out = append(out, p)
		}
	}
	return out
}

// The English pages mixed "authorisation" and "authorization". One spelling, and it
// is British: the operator is in Singapore, which writes British English.
func TestSpellingIsBritish(t *testing.T) {
	american := regexp.MustCompile(`(?i)\b(authoriz\w*|organiz\w*|minimiz\w*|prioritiz\w*|analyz\w*|recogniz\w*|categoriz\w*|summariz\w*|center|centers|behavior\w*|colors?|licens\w*|defense|catalog)\b`)
	texts := map[string]string{}
	for _, path := range englishPaths() {
		texts[path] = visibleText(renderPublic(t, path))
	}
	texts["llms-full.txt"] = llmsText(true)
	for name, text := range texts {
		if m := american.FindString(text); m != "" {
			t.Errorf("%s uses the American spelling %q", name, m)
		}
	}
}

// The Indonesian copy read like a translation in places: "dewan" for a board of
// directors, "hanya-baca" for read-only, "pimpinan teknik" for engineering lead.
func TestIndonesianTermsAreConsistent(t *testing.T) {
	banned := []string{"dewan", "hanya-baca", "pimpinan teknik", "positif palsu", "jaminan", "(bahasa inggris)"}
	texts := map[string]string{}
	for _, path := range indonesianPaths() {
		texts[path] = strings.ToLower(visibleText(renderPublic(t, path)))
	}
	for name, text := range texts {
		for _, b := range banned {
			if strings.Contains(text, b) {
				t.Errorf("%s says %q; use the agreed term (see site_copy_id.go)", name, b)
			}
		}
	}
	// And the agreed terms are the ones used.
	ent := strings.ToLower(visibleText(renderPublic(t, "/id/enterprise")))
	for _, want := range []string{"direksi", "engineering lead", "read-only"} {
		if !strings.Contains(ent, want) {
			t.Errorf("/id/enterprise does not use %q", want)
		}
	}
}

// The brand is "Netsekurity"; only the User-Agent the scanner sends is "NetSekurity".
func TestBrandSpellingIsOneThing(t *testing.T) {
	odd := regexp.MustCompile(`NetSekurity|NETSEKURITY|netSekurity`)
	for _, path := range publicPagePaths {
		text := visibleText(renderPublic(t, path))
		text = strings.ReplaceAll(text, "NetSekurity Scanner", "")
		text = strings.ReplaceAll(text, "NetSekurity Verification", "")
		text = strings.ReplaceAll(text, "NETSEKURITY_API_TOKEN", "") // an environment variable name
		if m := odd.FindString(text); m != "" {
			t.Errorf("%s writes the brand as %q", path, m)
		}
	}
}

// ---- icons -----------------------------------------------------------------

var (
	reIconCall = regexp.MustCompile(`icon "([a-z0-9]+)"`)
	rePickCall = regexp.MustCompile(`pick "([a-z-]+)"`)
)

// An unknown icon name renders nothing, silently. Every name a template asks for exists.
func TestEveryIconTheTemplatesUseExists(t *testing.T) {
	for _, src := range []string{siteLayoutHTML, siteSharedHTML, siteHomeHTML, siteEnterpriseHTML, siteTrustHTML, siteCompareHTML, siteSampleHTML, siteDocsHTML, siteScopeHTML} {
		for _, m := range reIconCall.FindAllStringSubmatch(src, -1) {
			if _, ok := iconPaths[m[1]]; !ok {
				t.Errorf("a template uses icon %q, which does not exist", m[1])
			}
		}
		for _, m := range rePickCall.FindAllStringSubmatch(src, -1) {
			if _, ok := iconSets[m[1]]; !ok {
				t.Errorf("a template picks from icon set %q, which does not exist", m[1])
			}
		}
	}
	for set, names := range iconSets {
		for _, n := range names {
			if _, ok := iconPaths[n]; !ok {
				t.Errorf("icon set %q lists %q, which does not exist", set, n)
			}
		}
	}
}

// Each row of cards has an icon per card, and no two cards in a row share one: the
// trust cards used to be four padlocks.
func TestIconSetsCoverTheCopy(t *testing.T) {
	for _, lang := range []string{"en", "id"} {
		h, tr := homeFor(lang), trustFor(lang)
		for set, n := range map[string]int{"home-trust": len(h.Trust), "audiences": len(h.Audiences), "trust-resp": len(tr.Resp)} {
			if len(iconSets[set]) < n {
				t.Errorf("%s: icon set %q has %d icons for %d cards", lang, set, len(iconSets[set]), n)
			}
			seen := map[string]bool{}
			for _, name := range iconSets[set][:n] {
				if seen[name] {
					t.Errorf("icon set %q repeats %q", set, name)
				}
				seen[name] = true
			}
		}
	}
}

// ---- illustrations -----------------------------------------------------------

// The isometric drawings are decoration: hidden from assistive technology, with
// the idea each one carries stated in text beside it.
func TestIllustrationsAreDecorativeAndAccompaniedByText(t *testing.T) {
	for _, path := range []string{"/", "/id/", "/trust", "/id/trust"} {
		page := renderPublic(t, path)
		if strings.Count(page, "<svg") < 4 {
			t.Errorf("%s: expected illustrations, found %d", path, strings.Count(page, "<svg"))
		}
		for _, m := range regexp.MustCompile(`<svg[^>]*>`).FindAllString(page, -1) {
			if !strings.Contains(m, `aria-hidden="true"`) {
				t.Errorf("%s: an svg is not aria-hidden: %.80s", path, m)
			}
			if strings.Contains(m, `role="img"`) {
				t.Errorf("%s: a decorative svg claims role=img", path)
			}
		}
	}
	home := visibleText(renderPublic(t, "/"))
	for _, want := range []string{"Scanned", "Not scanned: behind a login", "Not scanned: data and logic"} {
		if !strings.Contains(home, want) {
			t.Errorf("the scope drawing has no text legend entry %q", want)
		}
	}
	// The data-flow list is numbered 1-5 like the diagram.
	trust := renderPublic(t, "/trust")
	if n := len(trustEN().Flow); n != 5 || strings.Count(trust, `class="nsk-steps`) < 1 {
		t.Errorf("the data flow has %d steps and the page %d numbered lists", n, strings.Count(trust, `class="nsk-steps`))
	}
	for _, f := range []string{"site_svg.go"} {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		// Words baked into a drawing cannot be translated: only the digits 1-5 are allowed.
		for _, m := range regexp.MustCompile(`<text[^>]*>([^<]*)</text>`).FindAllStringSubmatch(string(src), -1) {
			if !regexp.MustCompile(`^[1-5]$`).MatchString(m[1]) {
				t.Errorf("%s bakes the text %q into a drawing", f, m[1])
			}
		}
	}
}

// ---- the evidence pages are bilingual ------------------------------------------

func TestEvidencePagesExistInBothLanguagesAndStayInTheirs(t *testing.T) {
	pairs := [][2]string{{"/sample-report", "/id/contoh-laporan"}, {"/docs", "/id/dokumentasi"}, {"/product-scope", "/id/cakupan-produk"}}
	for _, pr := range pairs {
		en, id := renderPublic(t, pr[0]), renderPublic(t, pr[1])
		if !strings.Contains(en, `<html lang="en">`) || !strings.Contains(id, `<html lang="id">`) {
			t.Errorf("%v: wrong html lang", pr)
		}
		for _, page := range []string{en, id} {
			alts := map[string]string{}
			for _, m := range reAlt.FindAllStringSubmatch(page, -1) {
				alts[m[1]] = m[2]
			}
			if alts["en"] != siteBase+pr[0] || alts["id"] != siteBase+pr[1] || alts["x-default"] != siteBase+pr[0] {
				t.Errorf("%v: hreflang set = %v", pr, alts)
			}
		}
	}
	// An Indonesian page links the Indonesian evidence pages, never the English ones
	// (the language switch, which exists to go to English, is the one exception).
	english := map[string]bool{"/sample-report": true, "/docs": true, "/product-scope": true}
	switchLink := regexp.MustCompile(`<a [^>]*hreflang="en"[^>]*>`)
	for _, path := range indonesianPaths() {
		page := switchLink.ReplaceAllString(renderPublic(t, path), "")
		for _, m := range reHref.FindAllStringSubmatch(page, -1) {
			href := m[1]
			if i := strings.IndexAny(href, "#?"); i >= 0 {
				href = href[:i]
			}
			if english[href] {
				t.Errorf("%s links the English page %s; the Indonesian twin exists", path, m[1])
			}
		}
	}
	// The same on the English side: no link to an Indonesian page except the switch.
	toID := regexp.MustCompile(`<a [^>]*hreflang="id"[^>]*>`)
	for _, path := range englishPaths() {
		page := toID.ReplaceAllString(renderPublic(t, path), "")
		for _, m := range reHref.FindAllStringSubmatch(page, -1) {
			if strings.HasPrefix(m[1], "/id/") {
				t.Errorf("%s links an Indonesian page: %s", path, m[1])
			}
		}
	}
}

// The product scope exists in two Markdown files, and they must say the same
// things: the same headings, tables and list items.
func TestScopeMarkdownTranslationsHaveTheSameShape(t *testing.T) {
	read := func(n string) string {
		b, err := staticFS.ReadFile(n)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	en, id := read("static/PRODUCT_SCOPE.md"), read("static/PRODUCT_SCOPE.id.md")
	shape := func(md string) map[string]int {
		c := map[string]int{}
		for _, line := range strings.Split(md, "\n") {
			switch {
			case strings.HasPrefix(line, "# "):
				c["h1"]++
			case strings.HasPrefix(line, "## "):
				c["h2"]++
			case strings.HasPrefix(line, "|") && !strings.Contains(line, "---"):
				c["table rows"]++
			case regexp.MustCompile(`^\d+\. `).MatchString(line):
				c["numbered"]++
			case strings.HasPrefix(line, "- "):
				c["bullets"]++
			case strings.HasPrefix(line, ">"):
				c["quote lines"]++
			case strings.TrimSpace(line) == "---":
				c["rules"]++
			}
		}
		return c
	}
	se, si := shape(en), shape(id)
	for k, n := range se {
		if si[k] != n {
			t.Errorf("product scope: English has %d %s, Indonesian %d", n, k, si[k])
		}
	}
	for _, md := range []string{en, id} {
		if strings.Contains(md, "$50") {
			t.Error("the product scope hard-codes a price; prices come from the database")
		}
	}
	if !strings.Contains(strings.ToLower(visibleText(renderPublic(t, "/id/cakupan-produk"))), "tidak dipindai") {
		t.Error("the Indonesian scope does not say what is not scanned")
	}
}

func TestDocsHaveOneNotePerCIExampleInBothLanguages(t *testing.T) {
	if len(docsEN().CINotes) != len(ciExamples) || len(docsID().CINotes) != len(ciExamples) {
		t.Fatalf("%d examples, %d English notes, %d Indonesian notes", len(ciExamples), len(docsEN().CINotes), len(docsID().CINotes))
	}
	for _, path := range []string{"/docs", "/id/dokumentasi"} {
		page := renderPublic(t, path)
		for _, e := range ciExamples {
			if !strings.Contains(page, e.Title) {
				t.Errorf("%s lacks the %s example", path, e.Title)
			}
		}
		if strings.Contains(page, "https://netsekurity.com/api/v1/pentests") || strings.Contains(strings.ToLower(visibleText(page)), "pentest-on") {
			t.Errorf("%s still teaches the old /pentests name in an example", path)
		}
	}
}

// "Not a penetration test" and an endpoint called /pentests contradicted each
// other. The public name is /assessments everywhere a customer reads it.
func TestPublicPagesNameTheEndpointAssessments(t *testing.T) {
	for _, path := range publicPagePaths {
		page := renderPublic(t, path)
		if strings.Contains(page, "POST /api/v1/pentests") || strings.Contains(page, "-X POST https://netsekurity.com/api/v1/pentests") {
			t.Errorf("%s shows the legacy endpoint as the way to start", path)
		}
	}
	for _, path := range []string{"/", "/id/", "/enterprise", "/id/enterprise", "/docs", "/id/dokumentasi"} {
		if !strings.Contains(renderPublic(t, path), "/api/v1/assessments") {
			t.Errorf("%s does not show /api/v1/assessments", path)
		}
	}
	// The one place the legacy name may be mentioned is the docs, saying it still works.
	if !strings.Contains(visibleText(renderPublic(t, "/docs")), "/api/v1/pentests, still works") {
		t.Error("the docs do not say the original name still works")
	}
}

// ---- language switch, menu, copy and print ------------------------------------

func TestLanguageSwitchHasAccessibleNames(t *testing.T) {
	for _, path := range publicPagePaths {
		page := renderPublic(t, path)
		for _, want := range []string{`aria-label="English (EN)"`, `aria-label="Bahasa Indonesia (ID)"`} {
			if !strings.Contains(page, want) {
				t.Errorf("%s: language switch lacks %s", path, want)
			}
		}
	}
}

func TestScriptsGiveTheMenuAnEscapeAndCodeACopyButton(t *testing.T) {
	page := renderPublic(t, "/")
	for _, want := range []string{"e.key === 'Escape'", "navigator.clipboard", "data-copy", "data-print"} {
		if !strings.Contains(page, want) {
			t.Errorf("layout script lacks %q", want)
		}
	}
	for _, path := range []string{"/docs", "/id/dokumentasi"} {
		page := renderPublic(t, path)
		want := len(ciExamples) + 4 // quickstart, start response, result, gate
		if n := strings.Count(page, `data-copy>`); n < want {
			t.Errorf("%s has %d copy buttons, want at least %d (one per code block)", path, n, want)
		}
		if !strings.Contains(page, `data-copied="`) {
			t.Errorf("%s: the copy button has no translated 'copied' label", path)
		}
	}
	// The labels are in the page's language.
	if id := renderPublic(t, "/id/dokumentasi"); !strings.Contains(id, ">Salin<") || !strings.Contains(id, `data-copied="Tersalin"`) {
		t.Error("Indonesian copy button is not translated")
	}
	// A code block wraps rather than widening the page.
	css, _ := staticFS.ReadFile("static/css/styles.css")
	for _, want := range []string{".nsk-pre", "overflow-wrap:anywhere", ".nsk-fit", ".nsk-stack", "nsk-print-procurement", "@page", "prefers-reduced-motion"} {
		if !strings.Contains(string(css), want) {
			t.Errorf("styles.css lacks %q (run npm run build:css)", want)
		}
	}
}

func TestProcurementFactSheetIsPrintable(t *testing.T) {
	for _, path := range []string{"/enterprise", "/id/enterprise"} {
		page := renderPublic(t, path)
		for _, want := range []string{`id="procurement"`, `data-print="procurement"`, "nsk-noprint"} {
			if !strings.Contains(page, want) {
				t.Errorf("%s: procurement fact sheet lacks %s", path, want)
			}
		}
	}
	for _, path := range []string{"/sample-report", "/id/contoh-laporan"} {
		if !strings.Contains(renderPublic(t, path), `data-print="page"`) {
			t.Errorf("%s has no print button", path)
		}
	}
}

// ---- tables that must not clip on a phone ----------------------------------------

// The Buy column of the pricing table sat off-screen at 390px, and the risk table
// of the sample report clipped. Neither may rely on horizontal scrolling now: the
// secondary columns fold into the first on small screens.
func TestWideTablesFoldInsteadOfScrolling(t *testing.T) {
	for _, path := range []string{"/", "/id/"} {
		page := renderPublic(t, path)
		i, j := strings.Index(page, `id="pricing"`), strings.Index(page, `id="tiers"`)
		pricing := page[i:j]
		if strings.Contains(pricing, "overflow-x-auto") || strings.Contains(pricing, "min-w-[") {
			t.Errorf("%s: the pricing table can still scroll sideways", path)
		}
		if strings.Count(pricing, "hidden") < 4 || !strings.Contains(pricing, "md:hidden") {
			t.Errorf("%s: the pricing table does not fold its secondary columns", path)
		}
		if !strings.Contains(pricing, `class="nsk-btn nsk-btn-sm"`) {
			t.Errorf("%s: no Buy button in the pricing table", path)
		}
	}
	for _, path := range []string{"/sample-report", "/id/contoh-laporan"} {
		page := renderPublic(t, path)
		i, j := strings.Index(page, `id="s3"`), strings.Index(page, `id="s4"`)
		risk := page[i:j]
		if strings.Contains(risk, "overflow-x-auto") || strings.Contains(risk, "min-w-[") || !strings.Contains(risk, "md:hidden") {
			t.Errorf("%s: the risk register can still clip on a phone", path)
		}
	}
	// Tables that turn into cards on a phone keep their table semantics.
	for _, path := range []string{"/external-assessment-vs-penetration-test", "/id/asesmen-eksternal-vs-uji-penetrasi", "/trust", "/id/trust"} {
		page := renderPublic(t, path)
		if !strings.Contains(page, `class="nsk-stack`) {
			t.Fatalf("%s: no stacking table", path)
		}
		for _, role := range []string{`role="table"`, `role="row"`, `role="columnheader"`, `role="rowheader"`, `role="cell"`, `data-label="`} {
			if !strings.Contains(page, role) {
				t.Errorf("%s: stacking table lacks %s", path, role)
			}
		}
	}
	// No table on any public page forces a minimum width.
	for _, path := range publicPagePaths {
		if strings.Contains(renderPublic(t, path), "min-w-[5") || strings.Contains(renderPublic(t, path), "min-w-[6") {
			t.Errorf("%s forces a wide table", path)
		}
	}
}

// Targets that are links on their own line are at least 40px tall.
func TestStandaloneLinksAreTapTargets(t *testing.T) {
	css, _ := staticFS.ReadFile("static/css/styles.css")
	for _, want := range []string{".nsk-textlink", ".nsk-footlink", "min-height:44px", "min-height:40px", "pointer:coarse"} {
		if !strings.Contains(string(css), want) {
			t.Errorf("styles.css lacks %q", want)
		}
	}
	for _, path := range []string{"/", "/enterprise", "/trust"} {
		page := renderPublic(t, path)
		if strings.Count(page, "nsk-footlink") < 12 {
			t.Errorf("%s: footer links are not tap targets", path)
		}
		// An arrow after a link's words is decoration: it sits in an aria-hidden span.
		hidden := len(regexp.MustCompile(`<span aria-hidden="true"> →</span>`).FindAllString(page, -1))
		if all := strings.Count(page, "→"); all != hidden {
			t.Errorf("%s: %d arrows, %d of them aria-hidden; the rest are read aloud", path, all, hidden)
		}
	}
	if strings.Contains(string(css), ".nsk-tag{") && regexp.MustCompile(`\.nsk-tag\{[^}]*font-size:10px`).Match(css) {
		t.Error("the tag is 10px text again")
	}
}

// ---- the home page tells a story ----------------------------------------------------

func TestHomeCallsToActionHaveAHierarchy(t *testing.T) {
	for _, lang := range []string{"en", "id"} {
		h := homeFor(lang)
		path := map[string]string{"en": "/", "id": "/id/"}[lang]
		vis := visibleText(renderPublic(t, path))
		hero := vis[:strings.Index(vis, h.AudiencesH2)]
		for _, want := range []string{h.CTAPrimary, h.CTASecondary, h.SignInNote} {
			want = strings.ReplaceAll(want, "{single}", "$50")
			if !strings.Contains(hero, want) {
				t.Errorf("%s: the first screen lacks %q", path, want)
			}
		}
		page := renderPublic(t, path)
		// One filled start button, with sample evidence as the secondary action.
		hs := page[strings.Index(page, "<main"):strings.Index(page, `id="audiences"`)]
		if n := strings.Count(hs, `class="nsk-btn"`); n != 1 {
			t.Errorf("%s: the hero has %d primary buttons, want 1", path, n)
		}
		if !strings.Contains(hs, `data-umami-event="netsekurity_hero_sample" class="nsk-btn nsk-btn-ghost"`) || !strings.Contains(hs, `data-umami-event="netsekurity_hero_cta" class="nsk-btn"`) {
			t.Errorf("%s: start must be primary and example results secondary", path)
		}
		// The lowest-price pack is marked, and only it.
		if strings.Count(page, "nsk-badge-ok") != 1 {
			t.Errorf("%s: %d 'lowest price' marks, want 1", path, strings.Count(page, "nsk-badge-ok"))
		}
	}
}

// The page used to repeat the same blocks. Each idea now lives where it is needed.
func TestSectionsAreNotRepeatedAcrossPages(t *testing.T) {
	cov := coverageCopy{}
	cov = commonEN().Coverage
	marker := cov.H2
	for path, want := range map[string]int{"/": 0, "/enterprise": 1, "/trust": 0} {
		if got := strings.Count(visibleText(renderPublic(t, path)), marker); got != want {
			t.Errorf("%s shows the coverage block %d times, want %d", path, got, want)
		}
	}
	tiersMarker := commonEN().TiersH2
	for path, want := range map[string]int{"/": 1, "/enterprise": 1, "/trust": 0} {
		if got := strings.Count(visibleText(renderPublic(t, path)), tiersMarker); got != want {
			t.Errorf("%s shows the tiers %d times, want %d", path, got, want)
		}
	}
	// On the enterprise page the tiers are one flat list, not three cards again.
	ent := renderPublic(t, "/enterprise")
	i, j := strings.Index(ent, `id="escalation"`), strings.Index(ent, `id="coverage"`)
	if strings.Contains(ent[i:j], "lg:grid-cols-3") {
		t.Error("the enterprise escalation section is the same three cards as the home page")
	}
}

// The trust page now names every outside party and says what each receives.
func TestTrustPageListsThirdPartiesAndTheAIQuestion(t *testing.T) {
	for _, lang := range []string{"en", "id"} {
		tr := trustFor(lang)
		path := map[string]string{"en": "/trust", "id": "/id/trust"}[lang]
		page := visibleText(renderPublic(t, path))
		for _, name := range []string{"Google", "Xendit", "Cloudflare", "Umami", "Meta"} {
			if !strings.Contains(page, name) {
				t.Errorf("%s does not list %s", path, name)
			}
		}
		if len(tr.Subs) != 6 || !strings.Contains(page, "security@dalang.io") {
			t.Errorf("%s: %d third parties", path, len(tr.Subs))
		}
		// The AI provider is not named: nothing in the repo says which it is.
		for _, banned := range []string{"OpenAI", "Anthropic", "Google Gemini", "Claude", "GPT"} {
			if strings.Contains(page, banned) {
				t.Errorf("%s names the AI provider %q; no source says which it is", path, banned)
			}
		}
	}
	for _, path := range []string{"/", "/enterprise", "/trust"} {
		if !strings.Contains(strings.ToLower(visibleText(renderPublic(t, path))), "ai") {
			t.Errorf("%s never says AI", path)
		}
	}
}

// ---- the CI gate the docs show actually works ----------------------------------------

// The docs print a shell script that starts an assessment, waits for it and fails on
// High or Critical. Run it, against a local server, in all four outcomes.
func TestDocsGateScriptRunsInAllFourOutcomes(t *testing.T) {
	stubControl(t, nil)
	for _, tool := range []string{"sh", "curl", "jq"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s is not installed", tool)
		}
	}
	script := docsEN().GateCode
	cases := []struct {
		name     string
		finish   func(id string) // what the scanner does after the first poll
		wantExit int
		wantOut  string
	}{
		{"clean", func(id string) {
			db.Exec(`UPDATE pentests SET status='completed', findings_reported=1, findings_critical=0, findings_high=0, findings_medium=2, findings_low=1, findings_info=3 WHERE id=?`, id)
		}, 0, "no High or Critical findings"},
		{"high", func(id string) {
			db.Exec(`UPDATE pentests SET status='completed', findings_reported=1, findings_critical=0, findings_high=1 WHERE id=?`, id)
		}, 1, `"high":1`},
		{"failed", func(id string) { db.Exec(`UPDATE pentests SET status='failed' WHERE id=?`, id) }, 2, "assessment failed"},
		{"never finishes", nil, 3, "still running after 30 minutes"},
	}
	for i, c := range cases {
		uid := fmt.Sprintf("u_gate_%d", i)
		token, mux := apiFixture(t, uid)
		db.Exec(`INSERT OR REPLACE INTO credit_balance (user_id, balance) VALUES (?, 5)`, uid)
		db.Exec(`INSERT INTO domains (user_id, domain, txt_verification_token, status) VALUES (?, 'app.example.com', 'x', 'verified')`, uid)

		polls := 0
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/v1/assessments/") {
				if polls++; polls == 2 && c.finish != nil {
					c.finish(strings.TrimPrefix(r.URL.Path, "/api/v1/assessments/"))
				}
			}
			mux.ServeHTTP(w, r)
		}))

		// The script is shown with the real host and a 60 second sleep; the test points it at the local server and does not wait.
		local := strings.ReplaceAll(script, "https://netsekurity.com", srv.URL)
		local = strings.ReplaceAll(local, "sleep 60", "sleep 0")
		cmd := exec.Command("sh", "-c", local)
		cmd.Env = append(os.Environ(), "NETSEKURITY_API_TOKEN="+token)
		done := make(chan struct{})
		var out []byte
		var err error
		go func() { out, err = cmd.CombinedOutput(); close(done) }()
		select {
		case <-done:
		case <-time.After(20 * time.Second):
			cmd.Process.Kill()
			t.Fatalf("%s: the script hung", c.name)
		}
		srv.Close()

		exit := 0
		if ee, ok := err.(*exec.ExitError); ok {
			exit = ee.ExitCode()
		} else if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if exit != c.wantExit || !strings.Contains(string(out), c.wantOut) {
			t.Errorf("%s: exit %d, output %q; want exit %d containing %q", c.name, exit, out, c.wantExit, c.wantOut)
		}
	}

	// A second start while one is running is refused, and the script says so instead of polling nothing.
	token, mux := apiFixture(t, "u_gate_busy")
	db.Exec(`INSERT OR REPLACE INTO credit_balance (user_id, balance) VALUES ('u_gate_busy', 5)`)
	db.Exec(`INSERT INTO domains (user_id, domain, txt_verification_token, status) VALUES ('u_gate_busy', 'app.example.com', 'x', 'verified')`)
	if rr := apiCall(mux, "POST", "/api/v1/assessments", token, "domain=app.example.com&authorised=true"); rr.Code != 200 {
		t.Fatalf("setup: %d %s", rr.Code, rr.Body)
	}
	srv := httptest.NewServer(mux)
	defer srv.Close()
	local := strings.ReplaceAll(strings.ReplaceAll(script, "https://netsekurity.com", srv.URL), "sleep 60", "sleep 0")
	cmd := exec.Command("sh", "-c", local)
	cmd.Env = append(os.Environ(), "NETSEKURITY_API_TOKEN="+token)
	out, err := cmd.CombinedOutput()
	if ee, ok := err.(*exec.ExitError); !ok || ee.ExitCode() != 1 || !strings.Contains(string(out), "could not start an assessment") {
		t.Errorf("busy account: %v %q, want exit 1 'could not start an assessment'", err, out)
	}

	// The Indonesian script differs only in the four messages.
	idScript := docsID().GateCode
	norm := func(s string) string {
		return regexp.MustCompile(`echo "[^"]*"; exit|echo "[^"]*"; exit`).ReplaceAllString(s, "")
	}
	_ = norm
	if strings.Count(idScript, "jq -e") != strings.Count(script, "jq -e") || strings.Count(idScript, "\n") != strings.Count(script, "\n") {
		t.Error("the Indonesian gate script is not the same script")
	}
}

// ---- the printable sample report -------------------------------------------------

func sampleSources() string {
	var b strings.Builder
	for _, lang := range []string{"en", "id"} {
		sum := sha256.Sum256([]byte(fmt.Sprintf("%#v", sampleFor(lang))))
		fmt.Fprintf(&b, "%s %x\n", lang, sum[:8])
	}
	return b.String()
}

// The sample report is also offered as a PDF. The PDF is rendered from the page, so
// it goes stale when the copy changes: this fails until tools/pdf/render.sh has been
// run again. (A template or print-CSS change is not detected; the script's header
// says to re-render then too.)
func TestSamplePDFsMatchTheirSource(t *testing.T) {
	if os.Getenv("WRITE_SAMPLE_SOURCES") == "1" {
		if err := os.WriteFile("static/sample-report.sources", []byte(sampleSources()), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	got, err := os.ReadFile("static/sample-report.sources")
	if err != nil {
		t.Fatalf("static/sample-report.sources is missing: run tools/pdf/render.sh (%v)", err)
	}
	if string(got) != sampleSources() {
		t.Error("the sample report copy has changed since the PDFs were rendered: run tools/pdf/render.sh and commit static/netsekurity-sample-report-*.pdf")
	}
	for _, lang := range []string{"en", "id"} {
		b, err := staticFS.ReadFile("static" + samplePDFPath(lang))
		if err != nil {
			t.Fatalf("%s missing: %v", samplePDFPath(lang), err)
		}
		if !strings.HasPrefix(string(b), "%PDF-") || len(b) < 20_000 {
			t.Errorf("%s is not a plausible PDF (%d bytes)", samplePDFPath(lang), len(b))
		}
		// The PDF is free of prices: nothing in it can go stale with credit_packages.
		if strings.Contains(string(b), "$50") {
			t.Errorf("%s mentions a price", samplePDFPath(lang))
		}
		page := renderPublic(t, map[string]string{"en": "/sample-report", "id": "/id/contoh-laporan"}[lang])
		if !strings.Contains(page, `href="`+samplePDFPath(lang)+`?v=`+assetHash(samplePDFPath(lang))+`"`) {
			t.Errorf("the %s sample page does not link its PDF", lang)
		}
	}
}

// The consent question is asked in the language of the page: an Indonesian page with an
// English cookie banner was one of the places the site flipped language on the visitor.
func TestConsentBannerSpeaksIndonesian(t *testing.T) {
	gate := consentGate(false)
	for _, want := range []string{"documentElement.lang", "Cookie &amp; pelacakan", "Tolak", "Terima", "persetujuan cookie dan pelacakan", "Cookies &amp; tracking", "Decline", "Accept"} {
		if !strings.Contains(gate, want) {
			t.Errorf("consent banner lacks %q", want)
		}
	}
}

// site_svg.go is generated. Hand edits (or a generator change nobody re-ran) are caught here.
func TestSVGFileIsGeneratedFromTheTool(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 is not installed")
	}
	if out, err := exec.Command("python3", "-I", "tools/svg/iso.py", "--check").CombinedOutput(); err != nil {
		t.Errorf("site_svg.go differs from what tools/svg/iso.py generates: run `python3 tools/svg/iso.py` (%v %s)", err, out)
	}
}
