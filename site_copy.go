package main

import (
	"html/template"
	"reflect"
	"regexp"
	"strings"
)

// The public site (home, enterprise, trust, guide) is written once per language
// as plain Go values and rendered by one template per page. That is the point:
//
//   - the two languages cannot drift, because a section missing from one fails
//     to compile and an empty string fails TestSiteCopyIsComplete;
//   - the visible FAQ and the FAQPage JSON-LD are built from the same slice, so
//     they cannot disagree (virus.ai.id keeps two hand-edited copies and relies
//     on a README rule);
//   - every price is a {token} filled from credit_packages at request time, so
//     the page cannot quote a price the checkout does not charge.
//
// Facts come from CLAIMS.md. Nothing here may state a number, a certification or
// a customer that CLAIMS.md does not list.

// kv is a titled paragraph.
type kv struct{ Title, Body string }

// jump is one in-page navigation chip: its label in the page's language and the
// fragment it targets, so the template cannot pair the wrong two.
type jump struct{ Label, Anchor string }

// qa is one FAQ entry. A is trusted HTML written in this repository (strong,
// em, code, a); the JSON-LD copy is the same text with the tags removed.
type qa struct {
	Q string
	A template.HTML
}

type audience struct {
	Role, Tagline string
	Bullets       []string
	LinkLabel     string
	Anchor        string // fragment on the enterprise page
}

type tier struct {
	Name, Price, Audience string
	Points                []string
}

type sampleCard struct {
	Label, Domain, Summary, Posture string
	Findings                        []string
	Note, Link                      string
}

type formCopy struct {
	Title, Lead, Name, Email, Message, MessagePlaceholder, Send, Reply string
}

// coverageCopy is the single statement of what a standard assessment does and
// does not do. Home, enterprise and trust all render it, so scope can only be
// worded one way per language.
type coverageCopy struct {
	H2, Lead, InTitle, OutTitle string
	In, Out                     []kv
	Note                        string
}

type commonCopy struct {
	Skip, MenuOpen, MenuClose, LangLabel                                         string
	NavHow, NavEnterprise, NavTrust, NavPricing, NavSample, NavDocs              string
	LogIn, Dashboard, Start, Buy                                                 string
	Tagline, Operator, Rights                                                    string
	ColProduct, ColEnterprise, ColResources                                      string
	LinkStartGuide, LinkScope, LinkGuide, LinkLLMS, LinkSecurityTxt, LinkSitemap string
	LinkDalang, LinkContact, LinkProcurement                                     string
	LinkPrivacy, LinkTerms, LinkTrustCenter                                      string
	Breadcrumb                                                                   string
	CopyLabel, CopiedLabel, PrintLabel                                           string
	Form                                                                         formCopy
	Coverage                                                                     coverageCopy
	TiersH2, TiersLead                                                           string
	Tiers                                                                        []tier
}

type homeCopy struct {
	Title, Description                    string
	Eyebrow, H1, Lead, SubLead            string
	CTAPrimary, CTASecondary, CTATertiary string
	SignInNote                            string
	Assurances                            []string
	Card                                  sampleCard
	DefinitionH2, Definition              string
	ScopeLegend                           []kv // scanned, behind a login, data and logic
	TypicalH3                             string
	Typical                               []string
	AudiencesH2                           string
	Audiences                             []audience
	ReportH2, ReportLead, ReportLink      string
	Report                                []kv
	PlainH3, PlainLink                    string
	Plain                                 []kv
	HowH2                                 string
	Steps                                 []kv
	BandH2, BandBody                      string
	PricingH2, PricingLead, PricingNote   string
	PricingCols                           []string
	PricingBest                           string
	PricingTalk, PricingTalkBtn           string
	TrustH2, TrustLead, TrustLink         string
	Trust                                 []kv
	FAQH2                                 string
	FAQ                                   []qa
}

type engBlock struct {
	Title      string
	Paragraphs []string
	Bullets    []string
	Code       string
}

type enterpriseCopy struct {
	Title, Description, Eyebrow, H1, Lead string
	HeroCTA                               string
	Jump                                  []jump // board, engineering, procurement, escalation, contact
	BoardH2, BoardIntro                   string
	Board                                 []kv
	BoardCallout                          string
	EngH2, EngIntro                       string
	Eng                                   []engBlock
	ProcH2, ProcIntro                     string
	ProcFacts                             []kv
	ProcNote                              string
	ProfileLead, ProfileLink              string
	EscLink                               string
	FAQH2                                 string
	FAQ                                   []qa
}

type dataRow struct{ Item, Detail string }

// subRow is one row of the third-party list: who, for what, and what it receives.
type subRow struct{ Name, Purpose, Receives string }

type trustCopy struct {
	Title, Description, Eyebrow, H1, Lead string
	Jump                                  []jump // responsible, flow, data, third parties, access, disclosure, operator
	RespH2, RespIntro                     string
	Resp                                  []kv
	FlowH2, FlowLead                      string
	Flow                                  []kv // five steps, numbered to match the diagram
	DataH2, DataIntro                     string
	DataCols                              []string
	Data                                  []dataRow
	SubH2, SubIntro                       string
	SubCols                               []string
	Subs                                  []subRow
	AccessH2                              string
	Access                                []kv
	DiscH2, DiscBody                      string
	Disc                                  []string
	OpH2                                  string
	OpBody                                []string
	OpFacts                               []kv
	OpSource, PolicyLead                  string
	ProfileLead, ProfileLink              string
	LimitsH2, LimitsBody                  string
}

type cmpRow struct {
	Label string
	Cells []string
}

type compareCopy struct {
	Title, Description, Eyebrow, H1, Lead string
	Updated                               string
	TableH2                               string
	Cols                                  []string
	Rows                                  []cmpRow
	GoodH2, CannotH2, ChooseH2            string
	Good, Cannot                          []string
	Choose                                []kv
	GlossH2                               string
	Gloss                                 []kv
	FAQH2                                 string
	FAQ                                   []qa
	CTA, CTALink                          string
}

// reflectValue returns the addressable struct a pointer points to, for walkStrings.
func reflectValue(p any) reflect.Value { return reflect.ValueOf(p).Elem() }

// tagRE strips the few tags allowed in an FAQ answer.
var tagRE = regexp.MustCompile(`<[^>]+>`)

// plainText turns FAQ answer HTML into the text Google reads from JSON-LD.
func plainText(h template.HTML) string {
	s := tagRE.ReplaceAllString(string(h), "")
	r := strings.NewReplacer("&amp;", "&", "&lt;", "<", "&gt;", ">", "&quot;", `"`, "&#39;", "'", "&nbsp;", " ")
	return strings.Join(strings.Fields(r.Replace(s)), " ")
}

// walkStrings applies f to every string reachable from v (strings, slices,
// arrays, structs, template.HTML). v must be addressable.
func walkStrings(v reflect.Value, f func(string) string) {
	switch v.Kind() {
	case reflect.String:
		if v.CanSet() {
			v.SetString(f(v.String()))
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			walkStrings(v.Index(i), f)
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			walkStrings(v.Field(i), f)
		}
	case reflect.Ptr, reflect.Interface:
		if !v.IsNil() {
			walkStrings(v.Elem(), f)
		}
	}
}

// emptyStrings lists the path of every empty string in v, for the completeness test.
func emptyStrings(v reflect.Value, path string, out *[]string) {
	switch v.Kind() {
	case reflect.String:
		if strings.TrimSpace(v.String()) == "" {
			*out = append(*out, path)
		}
	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			*out = append(*out, path+" (empty list)")
		}
		for i := 0; i < v.Len(); i++ {
			emptyStrings(v.Index(i), path+"["+itoa(i)+"]", out)
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			emptyStrings(v.Field(i), path+"."+v.Type().Field(i).Name, out)
		}
	}
}

// shapeOf describes list lengths, so two languages can be compared structurally.
func shapeOf(v reflect.Value, path string, out map[string]int) {
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		out[path] = v.Len()
		for i := 0; i < v.Len(); i++ {
			shapeOf(v.Index(i), path+"[]", out)
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			shapeOf(v.Field(i), path+"."+v.Type().Field(i).Name, out)
		}
	}
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var b [20]byte
	n := len(b)
	for i > 0 {
		n--
		b[n] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		n--
		b[n] = '-'
	}
	return string(b[n:])
}

// ---- evidence pages: sample report, docs, product scope ---------------------
//
// These used to be English-only HTML in Go strings. They are registry pages now,
// one copy struct per language, so an Indonesian reader who follows a link from
// an Indonesian page stays in Indonesian. Code, evidence strings and API names
// are the same in both languages and live in the template or in shared data.

type sevBadge struct{ Label, Class string }

// riskRow is one line of the sample risk register. Class is the badge colour
// (bad, warn, blue, muted), which is the same in every language.
type riskRow struct{ ID, Finding, Class, Severity, CVSS, CWE, OWASP string }

// frow is one labelled line of a sample finding. Mono marks evidence strings.
type frow struct {
	Label string
	Body  template.HTML
	Mono  bool
}

type finding struct {
	Title, Badge, Class string
	Rows                []frow
}

type sampleCopy struct {
	Title, Description, Eyebrow, H1, Lead string
	CTA, CTAAll, PrintLabel, PDFLabel     string
	DocH2                                 string
	Doc                                   []kv
	SumH2                                 string
	Badges                                []sevBadge
	Posture, PostureBody                  string
	RiskH2                                string
	RiskCols                              []string
	Risks                                 []riskRow
	FindH2                                string
	Findings                              []finding
	MoreSummary                           string
	More                                  []template.HTML
	PlanH2                                string
	Plan                                  []kv
	LimitsH2                              string
	LimitsBody                            template.HTML
}

type paramRow struct {
	Name, Type string
	Desc       template.HTML
}

type docsCopy struct {
	Title, Description, Eyebrow, H1, Lead string
	Jump                                  []jump
	QuickH2                               string
	Quick                                 []template.HTML
	QuickNote                             string
	APIH2                                 string
	StartH3, StartLead                    string
	ParamCols                             []string
	Params                                []paramRow
	StartResp                             string
	ResultH3, ResultLead                  string
	ResultParam, ResultResp               string
	ResultNote                            template.HTML
	Statuses                              []kv
	GateH3, GateLead                      string
	GateCode                              string
	GateNote                              string
	CIH2, CILead                          string
	CINotes                               []template.HTML // one per entry of ciExamples
	NotesH2                               string
	Notes                                 []template.HTML
}

type scopeCopy struct {
	Title, Description, Crumb string
}
