package main

import (
	"fmt"
	"net/http"
	"strings"
)

// /llms.txt and /llms-full.txt are what language-model crawlers read. They are
// generated from the same copy structs and the same credit_packages prices as
// the pages, so an answer assembled from them cannot quote a stale price or a
// scope the pages no longer state. Both are English; the Indonesian pages are
// linked.

func llmsView() *siteView {
	home, _ := pageByKey("home")
	return buildSiteView(home, siteLangs[0], false, loadSitePricing())
}

func mdLine(b *strings.Builder, format string, a ...any) { fmt.Fprintf(b, format+"\n", a...) }

func llmsText(full bool) string {
	v := llmsView()
	h, c := v.Home, v.C
	ent, trust, cmp := func() (enterpriseCopy, trustCopy, compareCopy) {
		rep := v.Pricing.replacer()
		e, t, g := enterpriseEN(), trustEN(), compareEN()
		walkStrings(reflectValue(&e), rep.Replace)
		walkStrings(reflectValue(&t), rep.Replace)
		walkStrings(reflectValue(&g), rep.Replace)
		return e, t, g
	}()

	var b strings.Builder
	mdLine(&b, "# Netsekurity")
	mdLine(&b, "")
	mdLine(&b, "> %s", h.Description)
	mdLine(&b, "")
	mdLine(&b, "%s", h.Definition)
	mdLine(&b, "")
	mdLine(&b, "**This is an external security assessment, not a full penetration test.** It does not test authenticated areas, authorization (IDOR), business logic or injection flaws. See \"What is not covered\" below. Operator: Dalang Pte. Ltd., Singapore (UEN 202622331M).")
	mdLine(&b, "")
	mdLine(&b, "## Key pages")
	mdLine(&b, "")
	mdLine(&b, "- [Home](%s/): what it is, how it works, pricing, FAQ.", siteBase)
	mdLine(&b, "- [Enterprise](%s%s): answers for boards, engineering leads and procurement, including a procurement fact sheet.", siteBase, pathFor("enterprise", "en"))
	mdLine(&b, "- [Trust and security](%s%s): how scans stay read-only, what data is kept, who can see reports, vulnerability disclosure, operator facts.", siteBase, pathFor("trust", "en"))
	mdLine(&b, "- [External assessment vs penetration test](%s%s): the difference, what each finds and misses, how to choose, glossary.", siteBase, pathFor("compare", "en"))
	mdLine(&b, "- [Sample report](%s/sample-report): a worked example of the PDF report.", siteBase)
	mdLine(&b, "- [Product scope and methodology](%s/product-scope): the exact checks, exclusions and deliverables.", siteBase)
	mdLine(&b, "- [CI/CD docs and API](%s/docs): start assessments from a pipeline; examples for %d CI systems.", siteBase, len(ciExamples))
	mdLine(&b, "- Indonesian: [Beranda](%s%s), [Enterprise](%s%s), [Kepercayaan](%s%s), [Asesmen eksternal vs uji penetrasi](%s%s).",
		siteBase, pathFor("home", "id"), siteBase, pathFor("enterprise", "id"), siteBase, pathFor("trust", "id"), siteBase, pathFor("compare", "id"))
	if !full {
		mdLine(&b, "- [Full text for LLMs](%s/llms-full.txt): everything on this page and more, in one file.", siteBase)
	}
	mdLine(&b, "")
	mdLine(&b, "## Pricing (USD)")
	mdLine(&b, "")
	mdLine(&b, "1 credit = 1 external assessment of 1 verified domain. Credits do not expire; a failed scan is refunded.")
	mdLine(&b, "")
	mdLine(&b, "| Pack | Credits | Price | Per assessment |")
	mdLine(&b, "|------|---------|-------|----------------|")
	for _, r := range v.Pricing.Rows {
		mdLine(&b, "| %s | %s | %s | %s |", r.Name, r.Credits, r.Price, r.PerScan)
	}
	mdLine(&b, "")
	mdLine(&b, "Expert review add-on: priced on request. Whitebox assessment: %s per app or domain.", v.Pricing.Whitebox)
	mdLine(&b, "")
	mdLine(&b, "## What is included in every assessment")
	mdLine(&b, "")
	for _, k := range c.Coverage.In {
		mdLine(&b, "- **%s.** %s", k.Title, k.Body)
	}
	mdLine(&b, "")
	mdLine(&b, "## What is not covered")
	mdLine(&b, "")
	for _, k := range c.Coverage.Out {
		mdLine(&b, "- **%s.** %s", k.Title, k.Body)
	}
	mdLine(&b, "")
	mdLine(&b, "%s", c.Coverage.Note)
	mdLine(&b, "")
	mdLine(&b, "## Report contents")
	mdLine(&b, "")
	for _, k := range h.Report {
		mdLine(&b, "- **%s.** %s", k.Title, k.Body)
	}
	mdLine(&b, "")
	mdLine(&b, "## How it works")
	mdLine(&b, "")
	for i, s := range h.Steps {
		mdLine(&b, "%d. **%s.** %s", i+1, s.Title, s.Body)
	}
	mdLine(&b, "")
	mdLine(&b, "## Frequently asked questions")
	mdLine(&b, "")
	for _, q := range h.FAQ {
		mdLine(&b, "**%s**", q.Q)
		mdLine(&b, "")
		mdLine(&b, "%s", plainText(q.A))
		mdLine(&b, "")
	}

	if full {
		mdLine(&b, "## For boards and executives")
		mdLine(&b, "")
		mdLine(&b, "%s", ent.BoardIntro)
		mdLine(&b, "")
		for _, k := range ent.Board {
			mdLine(&b, "- **%s** %s", k.Title, k.Body)
		}
		mdLine(&b, "")
		mdLine(&b, "%s", ent.BoardCallout)
		mdLine(&b, "")
		mdLine(&b, "## For engineering leads")
		mdLine(&b, "")
		mdLine(&b, "%s", ent.EngIntro)
		mdLine(&b, "")
		for _, blk := range ent.Eng {
			mdLine(&b, "### %s", blk.Title)
			mdLine(&b, "")
			for _, p := range blk.Paragraphs {
				mdLine(&b, "%s", p)
				mdLine(&b, "")
			}
			for _, x := range blk.Bullets {
				mdLine(&b, "- %s", x)
			}
			if len(blk.Bullets) > 0 {
				mdLine(&b, "")
			}
			if blk.Code != "" {
				mdLine(&b, "```")
				mdLine(&b, "%s", blk.Code)
				mdLine(&b, "```")
				mdLine(&b, "")
			}
		}
		mdLine(&b, "## For procurement and risk")
		mdLine(&b, "")
		mdLine(&b, "%s", ent.ProcIntro)
		mdLine(&b, "")
		for _, k := range ent.ProcFacts {
			mdLine(&b, "- **%s:** %s", k.Title, k.Body)
		}
		mdLine(&b, "")
		mdLine(&b, "%s", ent.ProcNote)
		mdLine(&b, "")
		for _, q := range ent.FAQ {
			mdLine(&b, "**%s**", q.Q)
			mdLine(&b, "")
			mdLine(&b, "%s", plainText(q.A))
			mdLine(&b, "")
		}
		mdLine(&b, "## Escalation tiers")
		mdLine(&b, "")
		for _, t := range c.Tiers {
			mdLine(&b, "- **%s** (%s): %s %s", t.Name, t.Price, t.Audience, strings.Join(t.Points, "; ")+".")
		}
		mdLine(&b, "")
		mdLine(&b, "## Trust and security")
		mdLine(&b, "")
		for _, k := range trust.Resp {
			mdLine(&b, "- **%s.** %s", k.Title, k.Body)
		}
		mdLine(&b, "")
		mdLine(&b, "What is kept:")
		mdLine(&b, "")
		for _, d := range trust.Data {
			mdLine(&b, "- **%s:** %s", d.Item, d.Detail)
		}
		mdLine(&b, "")
		for _, k := range trust.Access {
			mdLine(&b, "- **%s.** %s", k.Title, k.Body)
		}
		mdLine(&b, "")
		mdLine(&b, "Vulnerability disclosure: %s %s", trust.DiscBody, strings.Join(trust.Disc, " "))
		mdLine(&b, "")
		for _, p := range trust.OpBody {
			mdLine(&b, "%s", p)
			mdLine(&b, "")
		}
		mdLine(&b, "%s", trust.OpSource)
		mdLine(&b, "")
		mdLine(&b, "%s", trust.LimitsBody)
		mdLine(&b, "")
		mdLine(&b, "## External assessment vs penetration test")
		mdLine(&b, "")
		mdLine(&b, "%s", cmp.Lead)
		mdLine(&b, "")
		for _, row := range cmp.Rows {
			mdLine(&b, "- **%s.** External assessment: %s Penetration test: %s", row.Label, row.Cells[0], row.Cells[1])
		}
		mdLine(&b, "")
		mdLine(&b, "### Terms")
		mdLine(&b, "")
		for _, g := range cmp.Gloss {
			mdLine(&b, "- **%s:** %s", g.Title, g.Body)
		}
		mdLine(&b, "")
		if md, err := staticFS.ReadFile("static/PRODUCT_SCOPE.md"); err == nil {
			mdLine(&b, "## Product scope (verbatim)")
			mdLine(&b, "")
			mdLine(&b, "%s", string(md))
			mdLine(&b, "")
		}
	}

	mdLine(&b, "## Operator and contact")
	mdLine(&b, "")
	mdLine(&b, "%s", c.Operator)
	mdLine(&b, "")
	mdLine(&b, "Sales: sales@dalang.io (replies within 1 business day). Security: security@dalang.io. Updated %s.", siteModified)
	return b.String()
}

func handleLLMS(full bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		w.Write([]byte(llmsText(full)))
	}
}
