package main

import "html/template"

// English copy for the public site. Tokens in braces are filled at render time:
// {single} the cheapest pack price, {volume} the best per-assessment price,
// {volumeCredits} the credits in that pack, {whitebox} the whitebox price,
// {packs} every pack on one line, {ci} the number of CI examples in the docs.
// See CLAIMS.md for the source of every statement.
//
// Spelling is British (authorisation, prioritised, centre): TestSpellingIsBritish
// fails on the American forms. The user agent "NetSekurity Scanner" keeps the
// capital S the scanner sends.

// apiExample is the request pair shown on the home and enterprise pages. It is a
// constant so the two pages cannot show different calls.
const apiExample = "# start an assessment\n" +
	"curl -s -X POST https://netsekurity.com/api/v1/assessments \\\n" +
	"  -H \"X-API-Token: $NETSEKURITY_API_TOKEN\" \\\n" +
	"  -d \"domain=app.example.com\"\n\n" +
	"# read its status and the count of findings\n" +
	"curl -s https://netsekurity.com/api/v1/assessments/$ASSESSMENT_ID \\\n" +
	"  -H \"X-API-Token: $NETSEKURITY_API_TOKEN\""

func commonEN() commonCopy {
	return commonCopy{
		Skip: "Skip to content", MenuOpen: "Open menu", MenuClose: "Close menu", LangLabel: "Language",
		NavHow: "How it works", NavEnterprise: "Enterprise", NavTrust: "Trust", NavPricing: "Pricing",
		NavSample: "Sample report", NavDocs: "Docs",
		LogIn: "Log in", Dashboard: "Dashboard", Start: "Start an assessment", Buy: "Buy",
		Tagline:    "Continuous external web security assessment.",
		Operator:   "Netsekurity is a product of Dalang Pte. Ltd. (UEN 202622331M), 160 Robinson Road, #14-04, Singapore 068914.",
		Rights:     "© 2026 Dalang Pte. Ltd.",
		ColProduct: "Product", ColEnterprise: "Enterprise", ColResources: "Resources",
		LinkScope: "Product scope", LinkGuide: "Assessment vs penetration test", LinkLLMS: "llms.txt",
		LinkSecurityTxt: "security.txt", LinkSitemap: "Sitemap", LinkDalang: "Dalang.io", LinkPrivacy: "Privacy policy", LinkTerms: "Terms of use", LinkTrustCenter: "Security and trust centre",
		LinkContact: "Contact sales", LinkProcurement: "Procurement fact sheet",
		Breadcrumb: "Home",
		CopyLabel:  "Copy", CopiedLabel: "Copied", PrintLabel: "Print or save as PDF",
		Form: formCopy{
			Title:              "Talk to us",
			Lead:               "Questions about scope, volume, purchase terms or the whitebox tier? Send a note, or email sales@dalang.io.",
			Name:               "Name (optional)",
			Email:              "Work email",
			Message:            "Message",
			MessagePlaceholder: "Which websites, how many, and what does your process need from a supplier?",
			Send:               "Send message",
			Reply:              "We reply within 1 business day.",
		},
		Coverage: coverageCopy{
			H2:       "Exactly what is checked, and what is not",
			Lead:     "The checks are a fixed list, the same for every customer. Nothing below is implied beyond its wording.",
			InTitle:  "Included in every assessment",
			OutTitle: "Not included",
			In: []kv{
				{"Attack-surface recon", "DNS resolution and passive subdomain listing from certificate-transparency and public sources. Subdomains are listed for your information; only the domain you verified is assessed."},
				{"Security headers", "HSTS, Content-Security-Policy, X-Frame-Options, X-Content-Type-Options, Referrer-Policy and Permissions-Policy, plus clickjacking protection."},
				{"Web fingerprint", "Server header, server-version disclosure and web application firewall detection, read from the response."},
				{"TLS certificate", "Validity, days until expiry, issuer and number of SAN entries. A failed handshake is reported as a finding."},
				{"Exposed files and paths", "13 probes, including .git, .env, backup archives, server-status, phpinfo, SVN metadata and admin or login paths."},
				{"Open redirects", "Nine common redirect parameters, tested for a redirect to an external address."},
			},
			Out: []kv{
				{"Anything behind a login", "Authenticated areas are not exercised."},
				{"Authorisation and IDOR", "Needs several identities; not tested."},
				{"Injection testing", "No SQLi, XSS, LFI/RFI or path-traversal payloads are sent. The scan finds exposure and misconfiguration, not exploitable input bugs."},
				{"Business logic", "Workflow, payment and race-condition testing."},
				{"Network port scanning", "The standard assessment does not scan ports."},
				{"Other classes", "SSRF, GraphQL, WebSocket, request smuggling, cache poisoning, source-code analysis and multi-stage exploit chains."},
			},
			Note: "A finding that does not appear means “not scanned for”, never “verified safe”. Requests are read-only: nothing is written, uploaded or exploited.",
		},
		TiersH2:   "Choose the assurance level that fits the risk",
		TiersLead: "The standard assessment is the base layer. Deeper work is a separate, priced engagement, not a hidden upsell.",
		Tiers: []tier{
			{Name: "Automated external assessment", Price: "From {single} per domain", Audience: "Keep checking what the public internet can see.",
				Points: []string{"35+ read-only checks", "Severity, CVSS v3.1 score, CWE and OWASP category per finding", "Start from the dashboard or from CI/CD", "Credits do not expire"}},
			{Name: "Expert review add-on", Price: "Priced on request", Audience: "Have an engineer confirm the findings.",
				Points: []string{"A security engineer re-tests every finding in one completed report", "False positives removed", "A fix plan written for your stack"}},
			{Name: "Whitebox assessment", Price: "{whitebox} per app or domain", Audience: "Test what only an authenticated, informed tester can reach.",
				Points: []string{"Source review and authenticated testing", "Business-logic and exploit-chain analysis", "A dedicated security engineer with agent-assisted analysis", "Scoped per app or domain"}},
		},
	}
}

func homeEN() homeCopy {
	return homeCopy{
		Title:       "Netsekurity: Continuous External Web Security Assessment",
		Description: "Continuous, read-only external web security assessment. CVSS-scored findings, evidence and a 30/60/90-day fix plan for each domain. From {single} per assessment.",
		Eyebrow:     "External web security assessment",
		H1:          "Know what attackers can see of your websites.",
		Lead:        "Netsekurity examines the websites you own from the outside, the way an attacker would: read-only and without logging in. Each assessment ends in a prioritised report: every finding scored for severity, with the evidence and a 30/60/90-day fix plan.",
		SubLead:     "Anyone on the internet can look for exposed files, missing security headers, expiring certificates and open redirects. Netsekurity looks the same way, with your permission.",
		CTAPrimary:  "See a sample report", CTASecondary: "Talk to us about enterprise use", CTATertiary: "Start with one domain · {single}",
		SignInNote: "Sign in with Google to start. You only pay when you buy credits.",
		Assurances: []string{
			"External, unauthenticated and read-only",
			"AI-triaged, with optional expert review",
			"Not a penetration test, and we say so up front",
		},
		Card: sampleCard{
			Label: "Sample report", Domain: "shop.example.com",
			Summary: "1 High · 2 Medium · 2 Low · 4 Info", Posture: "Overall posture: needs attention",
			Findings: []string{
				"F-01 · Environment file exposed at /.env — High · CVSS 7.5 · CWE-538",
				"F-02 · Missing security headers (4 of 6) — Medium · CVSS 5.3",
				"F-03 · Open redirect via ?next= — Medium · CVSS 5.4",
			},
			Note: "Fictional domain, real report structure.", Link: "Open the full sample report",
		},
		DefinitionH2: "What is Netsekurity?",
		Definition:   "Netsekurity is an external web security assessment service operated by Dalang Pte. Ltd. A scanner runs more than 35 read-only checks against a website you have proved you own: DNS and subdomains, security headers, TLS certificates, exposed files and open redirects. An AI analysis agent then removes duplicates, ranks the findings by severity and business impact, and writes the report. It is an automated assessment, not a penetration test.",
		ScopeLegend: []kv{
			{"Scanned", "The public surface: the pages, files, headers, certificate and redirects that anyone on the internet can request."},
			{"Not scanned: behind a login", "Signed-in areas are never exercised."},
			{"Not scanned: data and logic", "Databases, payments and workflow logic are out of reach of an external, read-only check."},
		},
		TypicalH3: "Examples of what it looks for",
		Typical: []string{
			"A .env or .git file that the web server hands to anyone who asks",
			"Security headers missing from responses",
			"A certificate that expires next month",
			"A redirect that lets a phishing link carry your domain",
		},
		AudiencesH2: "Built for the three people who sign off on web risk",
		Audiences: []audience{
			{Role: "Board and executives", Tagline: "Evidence, not assurances.",
				Bullets:   []string{"A dated, per-domain report you can file with your risk register.", "A published price per assessment: {single} for one, {volume} each in the {volumeCredits}-credit pack.", "A written boundary between what was checked and what was not."},
				LinkLabel: "For boards and executives", Anchor: "board"},
			{Role: "Engineering leads", Tagline: "Signal you can act on.",
				Bullets:   []string{"Every finding carries a CVSS v3.1 score, CWE, OWASP category, evidence and a fix.", "Start assessments from CI/CD with an API token, on every deploy.", "Plain HTTP GET requests only: no injection payloads, nothing written to your systems."},
				LinkLabel: "For engineering leads", Anchor: "engineering"},
			{Role: "Procurement and risk", Tagline: "Clear scope, clear terms.",
				Bullets:   []string{"A defined deliverable with published scope and exclusions.", "Prices in US dollars. Credits do not expire, and a failed scan is refunded.", "Ownership proven by a DNS record before any scan runs."},
				LinkLabel: "Procurement fact sheet", Anchor: "procurement"},
		},
		ReportH2:   "What a report contains",
		ReportLead: "Every report follows the same structure, in English, as a PDF for each domain.",
		ReportLink: "Read the sample report",
		Report: []kv{
			{"Executive summary", "Counts by severity and a plain statement of overall posture."},
			{"Risk register", "Every finding with severity, CVSS v3.1 score, CWE and OWASP category."},
			{"Findings and evidence", "What was seen, the request that showed it as proof of concept, the CWE and CVE references, the impact and the fix."},
			{"30/60/90-day plan", "What to do today, this month and next quarter."},
			{"Scope statement", "What the report does not cover, stated in every report."},
		},
		PlainH3: "Terms in plain words", PlainLink: "Full glossary",
		Plain: []kv{
			{"CVSS", "A score from 0 to 10 for how severe a vulnerability is. The higher it is, the sooner it needs attention."},
			{"CWE", "A public catalogue of types of software weakness. CWE-601, for example, is an open redirect."},
			{"OWASP Top 10", "A widely used list of the most critical web application risk categories. We use it to label findings, not to claim that every category was tested."},
			{"DNS TXT record", "A line of text in your domain's DNS settings. You add one so that we can confirm you control the domain."},
		},
		HowH2: "How it works",
		Steps: []kv{
			{"Prove ownership", "Add one DNS TXT record, _netsekurity.<your-domain>, containing a unique token. We check it before any scan can run."},
			{"Start the assessment", "Press a button in the dashboard, or call the API from your pipeline. One assessment uses one credit."},
			{"Scan and triage", "A scanner runs 35+ read-only checks. An AI analysis agent removes duplicates, ranks findings by severity and business impact, and writes the report. No human reads a standard report before delivery."},
			{"Receive the report", "A scan takes 10 minutes to 24 hours, depending on what it finds. The report is a PDF in English, downloadable from your dashboard. Only the account that ordered it can open it."},
		},
		BandH2: "Ready to see what is exposed on your own domain?", BandBody: "Start with one domain, or read the structure of a real report first.",
		CICDH2:   "Continuous, not annual",
		CICDLead: "A yearly report describes a site that has since been released many times. Add one API call to your pipeline and each release is assessed the day it ships.",
		CICD: []string{
			"Create an API token in the dashboard. It is shown once and expires after 7 to 90 days, as you choose.",
			"Call POST /api/v1/assessments with the X-API-Token header and the domain.",
			"Read the status and the count of findings by severity with GET /api/v1/assessments/{id}. The PDF stays in the dashboard.",
			"Works with any system that can send an HTTP request. Examples for GitHub Actions, GitLab CI, Jenkins and eight more are in the docs.",
		},
		CICDNote:    "Continuous means the public surface is re-checked on every release. It does not mean that logic behind a login is re-tested.",
		CICDLink:    "Read the CI/CD docs",
		PricingH2:   "Published pricing",
		PricingLead: "1 credit = 1 external assessment of 1 domain. All prices are in US dollars.",
		PricingNote: "Credits do not expire and a failed scan is refunded automatically.",
		PricingCols: []string{"Pack", "Credits", "Price", "Per assessment", "Order"},
		PricingBest: "Lowest price per assessment",
		PricingTalk: "Need many domains, a purchase order or a vendor questionnaire answered?", PricingTalkBtn: "Talk to us",
		TrustH2:   "Responsible by design",
		TrustLead: "A security product has to be safe to point at your production site. These are the controls that make it so.",
		TrustLink: "Read how we scan and what we keep",
		Trust: []kv{
			{"Ownership first", "No DNS record, no scan. Only domains you control can be assessed."},
			{"Read-only", "Plain GET requests, a DNS lookup and a TLS handshake. Nothing is written to your systems."},
			{"Data minimised", "Exposure probes record the status code and size, then discard the response body."},
			{"Restricted reports", "Served only to the signed-in account that ordered them, never cached, kept out of search engines."},
		},
		FAQH2: "Frequently asked questions",
		FAQ: []qa{
			{"What is Netsekurity?", `Netsekurity is an external web security assessment service operated by Dalang Pte. Ltd. It checks a website you own from the outside, read-only and without logging in, and delivers a prioritised PDF report with CVSS-scored findings, evidence and a 30/60/90-day fix plan.`},
			{"Is it a penetration test?", `<strong>No.</strong> It is an automated, external, unauthenticated, read-only assessment. It does not test anything behind a login, authorisation (IDOR), business logic or injection flaws. For those risks use the <a href="/enterprise#escalation">expert review or whitebox tiers</a>, or a manual penetration test. The <a href="/external-assessment-vs-penetration-test">full comparison</a> explains the difference.`},
			{"What does one assessment cover?", `More than 35 read-only checks on the domain you verified: DNS and passive subdomain listing, six security headers and clickjacking protection, server-version disclosure, WAF detection, TLS certificate validity and expiry, 13 exposed-file and admin-path probes, and nine open-redirect parameters. See the <a href="/product-scope">product scope</a> for the exact list.`},
			{"What does it not cover?", `Anything behind a login, authorisation and IDOR, business logic, injection testing (SQLi, XSS, LFI/RFI), network port scanning, SSRF, GraphQL, WebSocket, request smuggling, cache poisoning, source-code analysis and multi-stage exploit chains. A finding that does not appear means “not scanned for”, not “verified safe”.`},
			{"How is domain ownership verified?", `Each domain gets a unique token. You publish it as a DNS TXT record at <code>_netsekurity.&lt;your-domain&gt;</code>, and we check it against public resolvers and the domain's authoritative name servers. No record, no scan.`},
			{"How much does it cost?", `1 credit = 1 assessment of 1 domain. A single credit costs {single}; in the {volumeCredits}-credit pack each assessment costs {volume}. Prices are in US dollars and credits do not expire. Expert review is priced on request, and the whitebox tier is {whitebox} per app or domain.`},
			{"Is a human involved?", `Not in the standard assessment. An AI analysis agent removes duplicates, ranks the findings and writes the report, and no human reads it before delivery. If you want an engineer to re-test each finding, that is the expert review add-on.`},
			{"Does Netsekurity use AI, and what does it see?", `An AI analysis agent operates the assessment and sees the scanner's output: findings, status codes, header values and short evidence strings. Exposure probes record the status code and size and discard the response body. The model provider and its region are not named on this site yet: ask <a href="mailto:security@dalang.io">security@dalang.io</a> before you buy. The <a href="/trust#third-parties">third-party list</a> says what else touches your data.`},
			{"Can we run it from CI/CD?", `Yes. Create an API token in the dashboard and call <code>POST /api/v1/assessments</code> on each deploy, then read the result with <code>GET /api/v1/assessments/{id}</code>. It works with any system that can send an HTTP request; the <a href="/docs">docs</a> have examples for GitHub Actions, GitLab CI, Jenkins and eight more.`},
			{"Does it suit applications built with AI coding tools?", `Yes. An external assessment examines what a website exposes, not how the code was written, so it applies equally to hand-written code and code produced with AI coding assistants. Exposed environment files and missing security headers occur in both.`},
			{"How long does a scan take?", `A scan takes between 10 minutes and 24 hours, depending on what it finds. You can leave the page: the report appears in your dashboard when it is ready.`},
			{"Who is behind Netsekurity?", `Netsekurity is a product of Dalang Pte. Ltd., a Singapore company (UEN 202622331M). See the <a href="/trust#operator">operator facts</a> and the <a href="/enterprise#procurement">procurement fact sheet</a>.`},
		},
	}
}

func enterpriseEN() enterpriseCopy {
	return enterpriseCopy{
		Title:       "Netsekurity for Enterprise: Board, Engineering, Procurement",
		Description: "How boards, engineering leads and procurement teams use Netsekurity: scope, evidence, CI/CD, pricing, data handling and escalation to expert review.",
		Eyebrow:     "Enterprise",
		H1:          "External assurance for boards, engineering leaders and procurement.",
		Lead:        "One assessment per domain, one report format, one published price. This page answers the questions each of the three audiences usually asks, and says plainly what Netsekurity does not do.",
		HeroCTA:     "Talk to sales",
		Jump:        []jump{{"Board", "board"}, {"Engineering", "engineering"}, {"Procurement", "procurement"}, {"Escalation", "escalation"}, {"Contact", "contact"}},
		BoardH2:     "For boards and executives: what the internet can see, and how you know",
		BoardIntro:  "A board does not need scanner output. It needs to know that the internet-facing estate is looked at regularly, by whom, with what result and at what cost. Netsekurity is built around those four answers.",
		Board: []kv{
			{"Is anyone checking?", "Every assessment is a dated report for one domain. Run on a schedule or on every release from CI/CD, the reports form a record of what was checked and when."},
			{"What did it find?", "The executive summary gives counts by severity and a plain statement of overall posture. The risk register lists every finding with severity, CVSS v3.1 score, CWE and OWASP category."},
			{"What does it cost?", "{single} for one assessment, and {volume} each in the {volumeCredits}-credit pack. Prices are published, in US dollars, and credits do not expire."},
			{"What does it not tell us?", "It is external, unauthenticated and read-only. It does not test anything behind a login, authorisation, business logic or injection flaws, and a finding that does not appear means “not scanned for”. For those risks, use the whitebox tier or a manual penetration test."},
		},
		BoardCallout: "We do not claim that an assessment makes a company secure or compliant. It closes one specific blind spot: what is exposed on the public web, today.",
		EngH2:        "For engineering leads: what runs, what comes back, how to wire it in",
		EngIntro:     "The scanner is deterministic and its behaviour is documented. What follows is what your WAF and your on-call engineer will see.",
		Eng: []engBlock{
			{Title: "What runs", Paragraphs: []string{"More than 35 read-only checks: the exact list is under “Exactly what is checked” further down this page, and on the product scope page. Subdomains are listed, not assessed: only the domain you verified is scanned, and each hostname you want assessed is its own domain and its own credit."}},
			{Title: "How it touches your systems", Paragraphs: []string{"Plain HTTP GET requests, a DNS lookup and a TLS handshake. Every HTTP request carries the User-Agent “NetSekurity Scanner”, so it can be recognised in your logs, and the exposure and redirect probes are spaced 0.3 seconds apart. No credentials, no injection payloads, nothing written or uploaded. Exposure probes record the HTTP status code and response size and discard the body; the exception is robots.txt, whose disallowed paths are summarised as evidence."}},
			{Title: "What comes back", Paragraphs: []string{"A PDF report in English for each assessment, downloadable from the dashboard. Each finding has a severity, a CVSS v3.1 score where one applies, CWE and CVE references, an OWASP category, the evidence as a proof of concept, the impact and a fix. The CI/CD API starts assessments and reports their status and the count of findings by severity; the PDF is downloaded from the dashboard."}},
			{Title: "Wiring it into CI/CD", Paragraphs: []string{"Create an API token in the dashboard. Tokens are shown once, stored only as a SHA-256 hash, and expire after 7 to 90 days that you choose. One assessment runs at a time per account. Works with any system that can send an HTTP request; the docs have GitHub Actions, GitLab CI and Jenkins examples."},
				Code: apiExample},
			{Title: "False positives", Paragraphs: []string{"A probe counts a path that answers 200, 301 or 302, so a single-page application or a login redirect can look like an exposure. The AI analysis agent de-duplicates and ranks findings, but it does not replace an engineer. Having an engineer re-test each finding is the expert review add-on."}},
		},
		ProcH2:    "For procurement and risk: scope, terms and vendor facts",
		ProcIntro: "Everything below is true today. Where a vendor process needs something we have not listed, tell us what it requires.",
		ProcFacts: []kv{
			{"Supplier", "Dalang Pte. Ltd., Singapore. UEN 202622331M. 160 Robinson Road, #14-04, Singapore 068914."},
			{"Product", "Netsekurity: an automated external web security assessment (blackbox, unauthenticated, read-only). Not a penetration test."},
			{"Unit of purchase", "1 credit = 1 external assessment of 1 verified domain."},
			{"Prices (USD)", "{packs}. Expert review: priced on request. Whitebox: {whitebox} per app or domain."},
			{"Payment", "Online, in US dollars, through Xendit. Netsekurity never sees or stores card details. Tax treatment and invoice details: ask sales@dalang.io before you buy."},
			{"Credits and refunds", "Credits do not expire. A scan that fails is refunded automatically."},
			{"Deliverable", "A PDF report in English for each domain, downloadable by the account that ordered it."},
			{"Ownership control", "Each domain is verified by a DNS TXT record before any scan. Domains you do not control cannot be assessed."},
			{"Accounts", "Sign-in is with Google. Accounts are individual today, and credits and reports belong to the account that bought them."},
			{"Third parties", "Google (sign-in), Xendit (payments), Cloudflare (network), Umami (analytics), Meta (only after consent) and the AI model provider. The trust page lists what each receives."},
			{"Policies", "Dalang publishes its privacy policy (dalang.io/privacy), terms of use (dalang.io/terms-of-use) and security and trust centre (dalang.io/security)."},
			{"Support and security", "sales@dalang.io, replies within 1 business day. Security issues: security@dalang.io. See the trust page for the disclosure policy."},
			{"Contract and paperwork", "Need a purchase order, specific contract terms or a vendor questionnaire answered? Email sales@dalang.io and tell us what your process requires."},
		},
		ProcNote:    "This page makes no claim of certification or independent audit for Netsekurity itself.",
		ProfileLead: "Company facts on this page come from the Dalang Company Profile 2026.", ProfileLink: "Open the Company Profile (PDF)",
		EscLink: "How the tiers compare with a penetration test",
		FAQH2:   "Enterprise questions",
		FAQ: []qa{
			{"Can we run assessments on a schedule?", `Netsekurity has no built-in scheduler. Call the API from your CI/CD pipeline on each release, or from any scheduler that can send an HTTP request.`},
			{"Can a pipeline fail a build on findings?", `Yes, from the pipeline: read the finding counts with <code>GET /api/v1/assessments/{id}</code> and fail the job on High or Critical. A scan takes 10 minutes to 24 hours, so start it on deploy and check from a later job. The <a href="/docs#gate">docs</a> have a working example.`},
			{"Is the report suitable as audit evidence?", `It is a dated record of what an external, read-only assessment found on a domain, with a stated scope. Whether that satisfies a particular control is for your auditor to decide, and it does not replace testing of authenticated areas.`},
			{"Can we assess many domains?", `Yes. Each domain is verified and assessed separately and uses one credit. Larger packs lower the price per assessment; for volume or purchase terms, <a href="#contact">talk to us</a>.`},
			{"Who can see our reports?", `The account that ordered the assessment, and the Netsekurity operators who administer the service. Reports are served only to signed-in users and are never cached. See <a href="/trust#access">who can see a report</a>.`},
			{"What if we need a person to review the findings?", `Order the expert review add-on: a security engineer re-tests every finding in one completed report, removes false positives and writes a fix plan for your stack. It is priced on request.`},
		},
	}
}

func trustEN() trustCopy {
	return trustCopy{
		Title:       "Trust and Security: How Netsekurity Scans and Handles Data",
		Description: "How Netsekurity verifies ownership, keeps scans read-only, minimises data, restricts reports and accepts vulnerability disclosures. Operated by Dalang Pte. Ltd.",
		Eyebrow:     "Trust",
		H1:          "How Netsekurity scans, what it keeps and who can see it.",
		Lead:        "Statements on this page describe how the service works today. Each can be checked against the product scope and the sample report.",
		Jump: []jump{{"Scanning", "responsible"}, {"Data flow", "flow"}, {"What we keep", "data"}, {"Third parties", "third-parties"},
			{"Who can see a report", "access"}, {"Disclosure", "disclosure"}, {"Operator", "operator"}},
		RespH2:    "Responsible scanning",
		RespIntro: "A security tool is dangerous if it can be pointed at anyone. These controls stop that.",
		Resp: []kv{
			{"Ownership first", "Before any scan, you publish a unique token as a DNS TXT record at _netsekurity.<your-domain>. We check it against public resolvers and the domain's authoritative name servers. No record, no scan."},
			{"Read-only", "Requests are plain HTTP GETs, a DNS lookup and a TLS handshake. The standard assessment sends no injection payloads and writes nothing to your systems."},
			{"A fixed list of checks", "The checks are the list on the product scope page, the same for every customer. One assessment runs at a time per account."},
			{"Recognisable", "Every HTTP request carries the User-Agent “NetSekurity Scanner”, so it can be recognised in your logs."},
			{"Paced", "Exposure and redirect probes are spaced 0.3 seconds apart."},
			{"Failures are refunded", "If a scan cannot finish it is closed and the credit is returned automatically."},
		},
		FlowH2:   "Where your data goes",
		FlowLead: "Five steps from your website to your report. The numbers match the diagram.",
		Flow: []kv{
			{"Your website", "The scanner sends read-only requests to the domain you verified. Subdomains it finds are listed, not scanned."},
			{"The scanner", "Runs the fixed list of checks. For exposure probes it records the status code and size and discards the response body."},
			{"The AI analysis agent", "Operates the assessment, removes duplicates, ranks the findings and writes the report. No human reads a standard report before delivery."},
			{"The report", "A PDF in English, stored for the account that ordered it. Served only to that account and to Netsekurity operators, never cached, kept out of search engines."},
			{"You", "You download the report from your dashboard. The third parties that handle data on the way are listed below."},
		},
		DataH2:    "What we keep",
		DataIntro: "We keep what the service needs to run and to show you results, and no more.",
		DataCols:  []string{"Item", "What is kept"},
		Data: []dataRow{
			{"Account", "Your name and email address from Google sign-in."},
			{"Domains", "The domains you add and their verification tokens."},
			{"Assessments", "Findings with short evidence strings (status codes, header values, sizes) and the report PDF."},
			{"Exposure probes", "The HTTP status code and response size. The response body is discarded."},
			{"AI analysis", "The scanner's output is processed by an AI analysis agent to produce the report. The model provider is listed under third parties."},
			{"Payments", "Invoice records: pack, amount and status. Card details are handled by Xendit and never reach us."},
			{"Contact messages", "What you write, your email address and the IP address it came from, used for spam control."},
			{"Analytics", "Umami, which is cookieless. The Meta (Facebook) pixel loads only if you accept it in the cookie banner."},
		},
		SubH2:    "Third parties that handle data",
		SubIntro: "These are the outside services involved. If one you need is missing from this list, or you need a detail it does not give, ask security@dalang.io.",
		SubCols:  []string{"Provider", "What it is used for", "What it receives"},
		Subs: []subRow{
			{"Google", "Sign-in", "Your name and email address reach us from Google; Google learns that you signed in to Netsekurity."},
			{"Xendit", "Card and bank payments", "The payment details you enter on Xendit's page. We receive the invoice status, never card details."},
			{"Cloudflare", "Network proxy in front of netsekurity.com", "The requests made to this site, as any reverse proxy sees them."},
			{"Umami (umami.dalang.io)", "Cookieless page-view analytics", "Page views. No cookie is set."},
			{"Meta", "Advertising measurement, only if you accept the cookie banner", "Nothing, until you accept. After that, page-view and sign-up events."},
			{"AI model provider", "The analysis agent that writes the report", "The scanner's output for your assessment. The provider and its region are not published on this site; ask security@dalang.io before you buy."},
		},
		AccessH2: "Who can see a report",
		Access: []kv{
			{"Reports", "Only the account that ordered the assessment and the Netsekurity operators who administer the service. Reports are served only to signed-in users, are never cached by browsers or CDNs, and are excluded from search engines."},
			{"Sessions", "A signed cookie valid for 24 hours, marked HttpOnly, Secure and SameSite=Lax."},
			{"API tokens", "Shown once, stored only as a SHA-256 hash, and expiring after 7 to 90 days that you choose."},
		},
		DiscH2:   "Reporting a vulnerability in our own systems",
		DiscBody: "If you find a security issue in netsekurity.com or dalang.io, email security@dalang.io.",
		Disc: []string{
			"We acknowledge within 2 business days.",
			"We do not pursue good-faith researchers.",
			"We credit fixes on request.",
			"Please do not test against other customers' domains.",
		},
		OpH2: "Who operates Netsekurity",
		OpBody: []string{
			"Netsekurity is a product of Dalang Pte. Ltd., a company registered in Singapore.",
			"The Dalang group (Dalang Pte. Ltd. in Singapore, PT Mashanz Sumber Terbuka and PT Komponen Pusat Data in Indonesia) operates its own network, AS154547, which anyone can check at bgp.he.net/AS154547.",
		},
		OpFacts: []kv{
			{"Legal entity", "Dalang Pte. Ltd., UEN 202622331M"},
			{"Address", "160 Robinson Road, #14-04, Singapore 068914"},
			{"Sales", "sales@dalang.io"},
			{"Security", "security@dalang.io"},
			{"Group network", "AS154547 (public record)"},
		},
		OpSource:    "Entity and network facts are from the Dalang Company Profile 2026, as of 20 October 2026.",
		ProfileLead: "The Company Profile is Dalang's own document about the group.", ProfileLink: "Open the Company Profile (PDF)",
		PolicyLead: "Dalang publishes its policies at dalang.io.",
		LimitsH2:   "What this page does not claim",
		LimitsBody: "This page makes no claim of certification or independent audit for Netsekurity itself, and it does not say that an assessment makes a site secure. It states how the service behaves. If something you need is missing, ask us.",
	}
}

func compareEN() compareCopy {
	return compareCopy{
		Title:       "External Assessment vs Penetration Test: The Difference",
		Description: "An external web security assessment is automated and repeatable; a penetration test is a manual attempt to break in. What each finds, what each misses, how to choose.",
		Eyebrow:     "Guide",
		H1:          "External web security assessment vs penetration test: what is the difference?",
		Lead:        "An external web security assessment is an automated, repeatable check of what a website exposes to the public internet. A penetration test is a manual engagement in which a person tries to break into an application, usually with credentials, to show what an attacker could actually do. They answer different questions and work best together.",
		Updated:     "Last updated 26 October 2026.",
		TableH2:     "The two approaches side by side",
		Cols:        []string{"Aspect", "External assessment", "Penetration test"},
		Rows: []cmpRow{
			{"Question it answers", []string{"What does this site expose to the public internet right now?", "What could an attacker actually do to this application?"}},
			{"Who or what performs it", []string{"Automated scanner with AI triage; an engineer's review is optional.", "A security engineer, usually assisted by tools."}},
			{"Access level", []string{"External and unauthenticated: no credentials.", "Often authenticated, sometimes with source code (whitebox)."}},
			{"Typical output", []string{"A prioritised list of exposures and misconfigurations with evidence and fixes.", "Validated exploit paths, including business-logic and authorisation flaws."}},
			{"Cadence it suits", []string{"Every release, or on a regular schedule.", "Before a major launch, after major change, or on a periodic basis."}},
			{"What it can miss", []string{"Anything behind a login, authorisation, business logic, injection flaws.", "Whatever was out of the agreed scope or time."}},
			{"Netsekurity offer", []string{"Automated external assessment, from {single} per domain.", "Whitebox assessment, {whitebox} per app or domain; or a manual penetration test from another provider."}},
		},
		GoodH2:   "What an external assessment is good at",
		CannotH2: "What it cannot tell you",
		ChooseH2: "How to choose",
		Good: []string{
			"Exposed secrets and files such as .env, .git and backup archives",
			"Missing security headers and weak clickjacking protection",
			"Expiring or failing TLS certificates",
			"Open redirects that make phishing links look legitimate",
			"Forgotten hosts, through passive subdomain listing",
			"Regressions: the same checks on every release show what changed",
		},
		Cannot: []string{
			"Whether a logged-in user can see another user's data (authorisation, IDOR)",
			"Flaws in business logic, payments and workflows",
			"Exploitable input bugs such as SQL injection and cross-site scripting",
			"Anything on a host other than the one you verified",
			"Whether a site is “secure”: a clean report means “not found by these checks”",
		},
		Choose: []kv{
			{"You need to know what is exposed today, and keep checking", "Use the automated external assessment, and run it from CI/CD."},
			{"You need a person to confirm findings and remove false positives", "Add expert review to a completed report."},
			{"You have login areas, payments or sensitive data", "Commission a whitebox assessment or a manual penetration test, and keep the external assessment running between engagements."},
		},
		GlossH2: "Terms explained",
		Gloss: []kv{
			{"Attack surface", "Everything about a system that can be reached from outside: websites, subdomains, files, services."},
			{"External (blackbox) testing", "Testing from the public internet with no inside knowledge, as an attacker without access would."},
			{"Unauthenticated", "Without logging in. Nothing behind a login is exercised."},
			{"Whitebox", "Testing with inside knowledge, such as source code and test accounts."},
			{"CVSS", "The Common Vulnerability Scoring System: a 0 to 10 score for how severe a vulnerability is."},
			{"CWE", "Common Weakness Enumeration: a public catalogue of types of software weakness, such as CWE-601 for open redirect."},
			{"OWASP Top 10", "A widely used list of the most critical web application risk categories. Mapping a finding to it categorises the finding; it does not mean every category was tested."},
			{"DNS TXT record", "A text entry in a domain's DNS settings. Netsekurity uses one to confirm that you control the domain."},
			{"False positive", "A reported issue that is not a real problem."},
		},
		FAQH2: "Common questions",
		FAQ: []qa{
			{"Is an external assessment enough for compliance?", `It can supply evidence about what a site exposes, but compliance schemes generally ask for more than an external scan. Check what your auditor requires, and treat the assessment as one input.`},
			{"How often should an external assessment run?", `As often as the site changes. With CI/CD integration it can run on every release; without it, on a regular schedule that matches how often you deploy.`},
			{"Does a clean report mean the site is secure?", `No. It means these checks found nothing. Areas outside the scope, such as authenticated functions and business logic, were not examined.`},
		},
		CTA: "See a sample report", CTALink: "/sample-report",
	}
}

func sampleEN() sampleCopy {
	return sampleCopy{
		Title:       "Sample Security Assessment Report: What One Credit Delivers",
		Description: "A worked example of a Netsekurity external web security assessment report: document control, findings summary, risk register, evidence and a 30/60/90-day plan.",
		Eyebrow:     "Sample · fictional domain, real report structure",
		H1:          "What one credit delivers",
		Lead:        "Every finding below comes from a check the scanner really performs. The domain and the evidence are invented. Your report is a PDF in English with this structure, downloadable from the dashboard; the full PDF also contains the methodology and an evidence appendix.",
		CTA:         "Get this for your domain · {single}", CTAAll: "All packages", PrintLabel: "Print or save as PDF", PDFLabel: "Download the sample as PDF",
		DocH2: "1 · Document control",
		Doc: []kv{
			{"Target", "shop.example.com"},
			{"Assessment type", "External · unauthenticated · read-only"},
			{"Ownership", "Verified via TXT _netsekurity"},
			{"Completed", "2026-09-01 09:14 UTC"},
			{"Prepared by", "Netsekurity.com Security Agent"},
			{"Classification", "Confidential: owner and operators only"},
		},
		SumH2:   "2 · Findings summary",
		Badges:  []sevBadge{{"1 High", "bad"}, {"2 Medium", "warn"}, {"2 Low", "blue"}, {"4 Info", "muted"}},
		Posture: "Overall posture: needs attention.", PostureBody: "One file exposure can leak credentials today; the remaining items are hardening.",
		RiskH2:   "3 · Risk register",
		RiskCols: []string{"#", "Finding", "Severity", "CVSS", "CWE", "OWASP"},
		Risks: []riskRow{
			{"F-01", "Environment file exposed at /.env", "bad", "High", "7.5", "CWE-538", "A05 · A02"},
			{"F-02", "Missing security headers (4 of 6)", "warn", "Medium", "5.3", "CWE-1021", "A05"},
			{"F-03", "Open redirect via ?next=", "warn", "Medium", "5.4", "CWE-601", "A01"},
			{"F-04", "Server version disclosed (nginx/1.18.0)", "blue", "Low", "3.7", "CWE-200", "A05"},
			{"F-05", "TLS certificate expires in 41 days", "blue", "Low", "3.7", "CWE-295", "A02"},
			{"I-01…04", "DNS resolution · 7 subdomains · WAF detected · robots.txt disallow entries", "muted", "Info", "—", "—", "—"},
		},
		FindH2: "4 · Findings and analysis",
		Findings: []finding{
			{"F-01 · Environment file exposed at /.env", "High · CVSS 7.5", "bad", []frow{
				{"What we saw", "The path /.env returned HTTP 200 with a 1.2 KB body. Files with this name normally hold database passwords, API keys and signing secrets. We did not keep the contents.", false},
				{"Evidence", "GET /.env -&gt; 200 1247", true},
				{"Impact", "Anyone can read your production secrets.", false},
				{"Fix", `Block the file at the web server (nginx: <span class="nsk-mono">location ~ /\.env { deny all; }</span>), move it outside the web root, then rotate every secret it contains: assume it has already been read.`, false},
			}},
			{"F-02 · Missing security headers", "Medium · CVSS 5.3", "warn", []frow{
				{"What we saw", "The response lacks strict-transport-security, content-security-policy, referrer-policy and permissions-policy. x-frame-options and x-content-type-options are present.", false},
				{"Evidence", "missing=strict-transport-security, content-security-policy, referrer-policy, permissions-policy", true},
				{"Fix", `Add the four headers at the proxy or framework. Start HSTS at max-age=300, raise it to a year once nothing breaks. A CSP of <span class="nsk-mono">default-src 'self'</span> plus your CDN hosts is a safe first version.`, false},
			}},
			{"F-03 · Open redirect via ?next=", "Medium · CVSS 5.4", "warn", []frow{
				{"What we saw", "Requesting /?next=https://evil.example answered with a 302 to that external address.", false},
				{"Evidence", "GET /?next=https://evil.example -&gt; 302 https://evil.example", true},
				{"Impact", "Phishing links can carry your real domain and still land on an attacker's page.", false},
				{"Fix", "Only accept a relative path that starts with a single /, or check the target against an allowlist of your own hosts.", false},
			}},
		},
		MoreSummary: "F-04, F-05 and the four informational items",
		More: []template.HTML{
			`<strong class="font-extrabold">F-04 · Server version disclosed.</strong> <span class="nsk-mono">server: nginx/1.18.0</span> tells an attacker exactly which CVEs to try. Set <span class="nsk-mono">server_tokens off;</span>.`,
			`<strong class="font-extrabold">F-05 · Certificate expires in 41 days.</strong> Issuer Let's Encrypt, 3 SAN entries. Confirm auto-renewal is running; an expired certificate takes the site offline.`,
			`<strong class="font-extrabold">I-01 · DNS.</strong> Apex resolves to 203.0.113.10. <strong class="font-extrabold">I-02 · Subdomains.</strong> 7 hosts found via certificate transparency, including staging.shop.example.com: check it is meant to be public. <strong class="font-extrabold">I-03 · WAF.</strong> Cloudflare fingerprint present. <strong class="font-extrabold">I-04 · robots.txt.</strong> Disallows /admin/, which confirms the path exists.`,
		},
		PlanH2: "5 · Remediation plan (30 / 60 / 90)",
		Plan: []kv{
			{"Today", "Block /.env. Rotate every secret in it."},
			{"This month", "Fix the redirect. Add HSTS, CSP, Referrer-Policy and Permissions-Policy."},
			{"Next quarter", "Hide the server version. Decide whether staging. should be public. Re-scan (1 credit) to confirm."},
		},
		LimitsH2:   "6 · What this report does not cover",
		LimitsBody: `Nothing behind a login, no authorisation (IDOR) or business-logic testing, and no injection payloads (SQLi, XSS, LFI). Those need the whitebox tier. A finding that does not appear means “not scanned for”, not “verified safe”. The full list is on the <a class="font-bold underline" href="/product-scope">product scope</a> page.`,
	}
}

func docsEN() docsCopy {
	return docsCopy{
		Title:       "CI/CD Integration and API Reference | Netsekurity",
		Description: "Start a Netsekurity assessment from your pipeline, read its status and gate a release on the findings: quickstart, HTTP API reference and examples for {ci} CI systems.",
		Eyebrow:     "Docs",
		H1:          "CI/CD integration and API reference",
		Lead:        "Start an external web security assessment from your pipeline on every deploy, then read its status and finding counts. Two HTTP requests, authenticated with an API token.",
		Jump:        []jump{{"Quickstart", "quickstart"}, {"API reference", "api"}, {"Gate a release", "gate"}, {"CI examples", "ci"}, {"Security notes", "notes"}},
		QuickH2:     "Quickstart",
		Quick: []template.HTML{
			`<strong>Add and verify your domain.</strong> Dashboard → My websites → add the domain → add the TXT record → check. The domain must show as <em>Verified</em>.`,
			`<strong>Create an API token.</strong> Dashboard → Automation → choose how long it is valid (7, 14, 30, 60 or 90 days) → create. Copy it now: it is shown only once.`,
			`<strong>Store it as a secret</strong> in your CI, for example <code>NETSEKURITY_API_TOKEN</code>.`,
			`<strong>Call the endpoint</strong> on deploy. Netsekurity queues the assessment and runs it; the report appears in your dashboard, and the status is available from the API.`,
		},
		QuickNote: "1 credit per assessment. External and read-only, so safe to point at production.",
		APIH2:     "HTTP API reference",
		StartH3:   "POST /api/v1/assessments",
		StartLead: "Start an assessment. The original name, /api/v1/pentests, still works and behaves identically.",
		ParamCols: []string{"Parameter", "Type", "Description"},
		Params: []paramRow{
			{"X-API-Token", "header", "Required. Your API token."},
			{"domain", "form or JSON", "Required. A verified domain owned by the token's account."},
			{"mode", "form or JSON", `Optional. <code class="nsk-mono">standard</code> is the only value: 1 credit, read-only.`},
		},
		StartResp: "# 200 OK\n" +
			"{\"assessment_id\":\"pt_1f2e…\",\"pentest_id\":\"pt_1f2e…\",\"domain\":\"app.example.com\",\"mode\":\"standard\",\"status\":\"queued\"}\n" +
			"# 401 missing, invalid or expired token\n" +
			"# 400 domain not verified for this account\n" +
			"# 402 insufficient credits\n" +
			"# 409 an assessment is already queued or running (one per account at a time)\n" +
			"# 429 rate limited",
		ResultH3:   "GET /api/v1/assessments/{assessment_id}",
		ResultLead: "Read the status of an assessment you started, and the count of findings by severity once it is done.",
		ResultResp: "# 200 OK, while the scan runs\n" +
			"{\"assessment_id\":\"pt_1f2e…\",\"domain\":\"app.example.com\",\"status\":\"running\",\"started_at\":\"2026-10-26T09:00:12Z\",\"completed_at\":null,\"findings\":null}\n" +
			"# 200 OK, once it has finished and the report carries a severity breakdown\n" +
			"{\"assessment_id\":\"pt_1f2e…\",\"domain\":\"app.example.com\",\"status\":\"completed\",\"started_at\":\"2026-10-26T09:00:12Z\",\"completed_at\":\"2026-10-26T09:14:40Z\",\"findings\":{\"critical\":0,\"high\":1,\"medium\":2,\"low\":2,\"info\":4}}\n" +
			"# 401 missing, invalid or expired token\n" +
			"# 404 no such assessment for this token's account",
		ResultParam: "The assessment_id returned when you started it. Only your own account's assessments are visible: any other id answers 404.",
		ResultNote:  `<code>findings</code> is <code>null</code> until the report has been delivered with a severity breakdown. <code>null</code> does not mean there were no findings. The PDF is not served here: download it from the signed-in dashboard.`,
		Statuses: []kv{
			{"queued", "Accepted and waiting for a scanner."},
			{"running", "A scanner is working on it. A scan takes 10 minutes to 24 hours, depending on what it finds."},
			{"completed", "The report is ready in the dashboard."},
			{"failed", "The scan could not finish. The credit is refunded automatically."},
		},
		GateH3:   "Gate a release on the findings",
		GateLead: "This job starts an assessment, waits up to 30 minutes for it, and fails when it finds a High or Critical issue, when the scan fails, when the status cannot be read, or when it is still running at the end of the wait.",
		GateCode: gateScript("no High or Critical findings", "could not start an assessment", "assessment failed", "assessment still running after 30 minutes", "could not read the assessment status"),
		GateNote: "A scan takes 10 minutes to 24 hours, so a job that waits will often wait a long time. Usually it is better to start the assessment on deploy and run this check from a later job.",
		CIH2:     "CI examples",
		CILead:   "Any system that can send an HTTP request will work. These start an assessment; add the check above where you need the result.",
		CINotes: []template.HTML{
			`Add <code>NETSEKURITY_API_TOKEN</code> under Settings → Secrets and variables → Actions.`,
			`Add <code>NETSEKURITY_API_TOKEN</code> under Settings → CI/CD → Variables (masked).`,
			`Create a <em>Secret text</em> credential named <code>nsk-api-token</code>.`,
			`Add <code>NETSEKURITY_API_TOKEN</code> under Project → Settings → Environment Variables.`,
			`Add <code>NETSEKURITY_API_TOKEN</code> under Pipelines → Library → Variable group (keep it secret).`,
			`Add <code>NETSEKURITY_API_TOKEN</code> under Repository → Settings → Pipelines → Repository variables (secured).`,
			`Add the encrypted secret with <code>travis encrypt NETSEKURITY_API_TOKEN=xxx</code>.`,
			`Define <code>env.NSK_API_TOKEN</code> in the build configuration (Parameters, Password type).`,
			`Add <code>NETSEKURITY_API_TOKEN</code> under Pipeline → Settings → Environment variables (secret).`,
			`Set <code>NETSEKURITY_API_TOKEN</code> in the pipeline environment as a secret.`,
			`Set <code>NETSEKURITY_API_TOKEN</code> in the pipeline environment as a secret.`,
		},
		NotesH2: "Configuration and security notes",
		Notes: []template.HTML{
			`<strong>Tokens expire.</strong> Choose an expiry that matches your deploy cadence, and rotate by creating a new token and deleting the old one.`,
			`<strong>Store tokens as CI secrets.</strong> Never commit them to source control. A token is stored only as a SHA-256 hash on our side, and it is never accepted in a URL.`,
			`<strong>One token per pipeline</strong> is fine. Delete any token from the dashboard at any time.`,
			`<strong>One assessment at a time.</strong> Only one queued or running assessment per account; the API answers 409 if one is already active.`,
			`<strong>Credits.</strong> One assessment is one credit. An insufficient balance returns HTTP 402.`,
			`<strong>Polling.</strong> The status endpoint is limited per client address. Poll once a minute, not in a tight loop.`,
			`<strong>Scope.</strong> External, unauthenticated and read-only. See the <a href="/product-scope">product scope</a> for exactly what is and is not checked.`,
		},
	}
}

func scopeEN() scopeCopy {
	return scopeCopy{
		Title:       "Product Scope and Methodology: What Netsekurity Checks",
		Description: "Exactly what a Netsekurity external web security assessment checks, what it does not, how findings are scored, and the deliverables, with the whitebox escalation tier.",
		Crumb:       "Product scope",
	}
}

// gateScript is the CI gate shown in the docs, the same in both languages apart
// from the five messages it prints. TestDocsGateScriptRunsInAllFourOutcomes runs it
// against a local server, so what the docs show is what works. Exit codes: 0 clean,
// 1 High or Critical found (or could not start), 2 the scan failed, 3 still running
// when the wait ended, 4 the status could not be read.
func gateScript(ok, cannotStart, failed, stillRunning, cannotRead string) string {
	return "set -eu\n" +
		"API=https://netsekurity.com/api/v1/assessments\n" +
		"AUTH=\"X-API-Token: $NETSEKURITY_API_TOKEN\"\n\n" +
		"start=$(curl -fsS -X POST \"$API\" -H \"$AUTH\" -d \"domain=app.example.com\") \\\n" +
		"  || { echo \"" + cannotStart + "\"; exit 1; }\n" +
		"id=$(echo \"$start\" | jq -r .assessment_id)\n\n" +
		"for i in $(seq 1 30); do\n" +
		"  r=$(curl -fsS \"$API/$id\" -H \"$AUTH\") || { echo \"" + cannotRead + "\"; exit 4; }\n" +
		"  case $(echo \"$r\" | jq -r .status) in\n" +
		"    completed)\n" +
		"      echo \"$r\" | jq -e '.findings != null and .findings.critical == 0 and .findings.high == 0' >/dev/null \\\n" +
		"        && { echo \"" + ok + "\"; exit 0; }\n" +
		"      echo \"$r\"; exit 1 ;;\n" +
		"    failed) echo \"" + failed + "\"; exit 2 ;;\n" +
		"  esac\n" +
		"  sleep 60\n" +
		"done\n" +
		"echo \"" + stillRunning + "\"; exit 3"
}
