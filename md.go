package main

import (
	"html"
	"html/template"
	"regexp"
	"strings"
)

// A deliberately small Markdown renderer for static/PRODUCT_SCOPE.md. It covers
// exactly what that document uses (ATX headings, paragraphs, bullet and numbered
// lists, tables, block quotes, horizontal rules, **bold**, *italic*, `code`,
// [links](url) and bare URLs) and nothing else. Every character of source text is
// HTML-escaped; only the tags this file writes are emitted.

var (
	mdHeading = regexp.MustCompile(`^(#{1,6})\s+(.*)$`)
	mdOrdered = regexp.MustCompile(`^\d+\.\s+(.*)$`)
	mdBullet  = regexp.MustCompile(`^[-*]\s+(.*)$`)
	mdRule    = regexp.MustCompile(`^(-{3,}|\*{3,})$`)
	mdTableSp = regexp.MustCompile(`^\|?\s*:?-{2,}:?\s*(\|\s*:?-{2,}:?\s*)*\|?$`)
	mdLink    = regexp.MustCompile(`\[([^\]]+)\]\((https?://[^)\s]+|/[^)\s]*)\)`)
	mdBare    = regexp.MustCompile(`(^|[\s(])(https?://[^\s<)]+)`)
	mdBold    = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	mdItalic  = regexp.MustCompile(`(^|[^*])\*([^*\s][^*]*)\*`)
	mdSlug    = regexp.MustCompile(`[^a-z0-9]+`)
)

func mdSlugify(s string) string {
	return strings.Trim(mdSlug.ReplaceAllString(strings.ToLower(s), "-"), "-")
}

// mdInline renders the inline syntax of one line of already-trusted source.
func mdInline(src string) string {
	// Code spans first, so their content is never re-interpreted.
	parts := strings.Split(src, "`")
	var b strings.Builder
	for i, part := range parts {
		if i%2 == 1 && i < len(parts)-1 {
			b.WriteString("<code>" + html.EscapeString(part) + "</code>")
			continue
		}
		if i%2 == 1 { // unmatched backtick: keep it literally
			b.WriteString("`" + mdText(part))
			continue
		}
		b.WriteString(mdText(part))
	}
	return b.String()
}

func mdText(s string) string {
	s = html.EscapeString(s)
	// Links: the target was matched on the escaped text, which is safe because
	// the pattern only admits http(s) and site-relative URLs.
	var links []string
	s = mdLink.ReplaceAllStringFunc(s, func(m string) string {
		sub := mdLink.FindStringSubmatch(m)
		links = append(links, `<a href="`+sub[2]+`">`+sub[1]+`</a>`)
		return "\x00L" + itoa(len(links)-1) + "\x00"
	})
	s = mdBare.ReplaceAllString(s, `$1<a href="$2">$2</a>`)
	s = mdBold.ReplaceAllString(s, "<strong>$1</strong>")
	s = mdItalic.ReplaceAllString(s, "$1<em>$2</em>")
	for i, l := range links {
		s = strings.Replace(s, "\x00L"+itoa(i)+"\x00", l, 1)
	}
	return s
}

func mdCells(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")
	var cells []string
	for _, c := range strings.Split(line, "|") {
		cells = append(cells, strings.TrimSpace(c))
	}
	return cells
}

// renderMarkdown converts src to HTML.
func renderMarkdown(src string) template.HTML {
	lines := strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n")
	var out strings.Builder
	i := 0
	for i < len(lines) {
		line := strings.TrimRight(lines[i], " \t")
		trim := strings.TrimSpace(line)
		switch {
		case trim == "":
			i++

		case mdRule.MatchString(trim):
			out.WriteString("<hr/>\n")
			i++

		case mdHeading.MatchString(trim):
			m := mdHeading.FindStringSubmatch(trim)
			lvl := len(m[1])
			text := strings.TrimSpace(m[2])
			plain := strings.NewReplacer("**", "", "`", "").Replace(text)
			out.WriteString("<h" + itoa(lvl) + ` id="` + mdSlugify(plain) + `">` + mdInline(text) + "</h" + itoa(lvl) + ">\n")
			i++

		case strings.HasPrefix(trim, ">"):
			var para []string
			for i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), ">") {
				para = append(para, strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[i]), ">")))
				i++
			}
			out.WriteString("<blockquote><p>" + mdInline(strings.Join(para, " ")) + "</p></blockquote>\n")

		case strings.HasPrefix(trim, "|") && i+1 < len(lines) && mdTableSp.MatchString(strings.TrimSpace(lines[i+1])):
			head := mdCells(trim)
			i += 2
			out.WriteString(`<div class="nsk-table-wrap"><table><thead><tr>`)
			for _, h := range head {
				out.WriteString("<th scope=\"col\">" + mdInline(h) + "</th>")
			}
			out.WriteString("</tr></thead><tbody>\n")
			for i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), "|") {
				out.WriteString("<tr>")
				for _, c := range mdCells(lines[i]) {
					out.WriteString("<td>" + mdInline(c) + "</td>")
				}
				out.WriteString("</tr>\n")
				i++
			}
			out.WriteString("</tbody></table></div>\n")

		case mdBullet.MatchString(trim) || mdOrdered.MatchString(trim):
			ordered := mdOrdered.MatchString(trim)
			tag := "ul"
			if ordered {
				tag = "ol"
			}
			out.WriteString("<" + tag + ">\n")
			for i < len(lines) {
				t := strings.TrimSpace(lines[i])
				var m []string
				if ordered {
					m = mdOrdered.FindStringSubmatch(t)
				} else {
					m = mdBullet.FindStringSubmatch(t)
				}
				if m == nil {
					break
				}
				item := m[1]
				i++
				// Continuation lines: indented, non-empty, not a new item.
				for i < len(lines) && strings.HasPrefix(lines[i], "  ") && strings.TrimSpace(lines[i]) != "" &&
					!mdBullet.MatchString(strings.TrimSpace(lines[i])) && !mdOrdered.MatchString(strings.TrimSpace(lines[i])) {
					item += " " + strings.TrimSpace(lines[i])
					i++
				}
				out.WriteString("<li>" + mdInline(item) + "</li>\n")
			}
			out.WriteString("</" + tag + ">\n")

		default:
			var para []string
			for i < len(lines) {
				t := strings.TrimSpace(lines[i])
				if t == "" || mdHeading.MatchString(t) || mdRule.MatchString(t) || strings.HasPrefix(t, ">") ||
					strings.HasPrefix(t, "|") || mdBullet.MatchString(t) || mdOrdered.MatchString(t) {
					break
				}
				para = append(para, t)
				i++
			}
			out.WriteString("<p>" + mdInline(strings.Join(para, " ")) + "</p>\n")
		}
	}
	return template.HTML(out.String())
}
