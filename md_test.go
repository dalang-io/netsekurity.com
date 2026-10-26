package main

import (
	"strings"
	"testing"
)

func TestMarkdownBlocks(t *testing.T) {
	src := "# Title here\n\nA paragraph with **bold**, *italic* and `code <b>`.\nSecond line joins.\n\n" +
		"- one\n- two\n  continued\n\n1. first\n2. second\n\n> quoted **text**\n\n---\n\n" +
		"| A | B |\n|---|---|\n| x | **y** |\n"
	out := string(renderMarkdown(src))
	for _, want := range []string{
		`<h1 id="title-here">Title here</h1>`,
		"<strong>bold</strong>", "<em>italic</em>", "<code>code &lt;b&gt;</code>",
		"Second line joins.",
		"<ul>", "<li>one</li>", "<li>two continued</li>",
		"<ol>", "<li>first</li>", "<li>second</li>",
		"<blockquote><p>quoted <strong>text</strong></p></blockquote>",
		"<hr/>",
		`<th scope="col">A</th>`, "<td><strong>y</strong></td>", "<table>",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

// Source text is never trusted as markup: only tags this file writes come out.
func TestMarkdownEscapesEverything(t *testing.T) {
	out := string(renderMarkdown("# <script>alert(1)</script>\n\n<img src=x onerror=alert(1)> and [x](javascript:alert(1)) and [ok](https://netsekurity.com/a?b=1&c=2)\n\n| <i> |\n|---|\n| <u> |"))
	for _, bad := range []string{"<script", "<img", "<i>", "<u>", `href="javascript`} {
		if strings.Contains(out, bad) {
			t.Errorf("unescaped %q in:\n%s", bad, out)
		}
	}
	if !strings.Contains(out, `<a href="https://netsekurity.com/a?b=1&amp;c=2">ok</a>`) {
		t.Errorf("a safe link was not rendered:\n%s", out)
	}
}

func TestMarkdownLinksAndBareURLs(t *testing.T) {
	out := string(renderMarkdown("See [the scope](/product-scope) or https://netsekurity.com/sample-report today."))
	if !strings.Contains(out, `<a href="/product-scope">the scope</a>`) || !strings.Contains(out, `<a href="https://netsekurity.com/sample-report">https://netsekurity.com/sample-report</a>`) {
		t.Errorf("links = %s", out)
	}
}

// The real document renders into the structure the scope page depends on.
func TestProductScopeDocumentRenders(t *testing.T) {
	md, err := staticFS.ReadFile("static/PRODUCT_SCOPE.md")
	if err != nil {
		t.Fatal(err)
	}
	out := string(renderMarkdown(string(md)))
	for _, want := range []string{"<h1", "<h2", "<table", "<ol>", "<ul>", "<blockquote>", "Network port scanning", "security@dalang.io"} {
		if !strings.Contains(out, want) {
			t.Errorf("document missing %q", want)
		}
	}
	if strings.Contains(out, "**") || strings.Contains(out, "\n- ") {
		t.Error("raw markdown syntax leaked into the HTML")
	}
	// The out-of-scope list is numbered 1..9 with no stray "2a.".
	if strings.Contains(out, "2a.") {
		t.Error("the out-of-scope list still has a 2a. item")
	}
}
