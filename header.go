package main

import (
	"html/template"
	"strings"
)

// hdrLink is one navigation link in the shared header component.
type hdrLink struct {
	Href string
	Text string
	// NewTab opens in a new tab.
	NewTab bool
}

// renderHeader renders the shared site header (brand + desktop nav + mobile burger
// drawer + optional auth nav). Both the landing page (index) and /docs use this one
// component so the header never drifts between pages.
//
//   - nav:      desktop navigation links (hidden below md, shown as a hamburger drawer on mobile)
//   - authNav:  optional auth links (e.g. login or dashboard button) shown in the desktop row
//     and inside the mobile drawer.
//   - brandURL: where the logo links to.
func renderHeader(nav []hdrLink, authNav template.HTML, brandURL string) template.HTML {
	var desktop strings.Builder
	for _, l := range nav {
		target := ""
		if l.NewTab {
			target = ` target="_blank" rel="noopener"`
		}
		desktop.WriteString(`<a href="` + template.HTMLEscapeString(l.Href) + `"` + target + ` class="prompt whitespace-nowrap hover:text-emerald-300">` + template.HTMLEscapeString(l.Text) + `</a>`)
	}

	var mobile strings.Builder
	for _, l := range nav {
		target := ""
		if l.NewTab {
			target = ` target="_blank" rel="noopener"`
		}
		mobile.WriteString(`<a href="` + template.HTMLEscapeString(l.Href) + `"` + target + ` class="prompt whitespace-nowrap rounded px-2 py-2 text-gray-300 hover:bg-white/5 hover:text-emerald-300">` + template.HTMLEscapeString(l.Text) + `</a>`)
	}

	auth := string(authNav)
	// If no auth nav provided, leave the auth slot empty.
	authDesktop := auth
	authMobile := auth
	if auth == "" {
		authDesktop = ""
		authMobile = ""
	}

	h := `<header class="sticky top-0 z-50 border-b border-emerald-500/25 bg-ink/85 backdrop-blur">
  <div class="mx-auto flex max-w-6xl items-center justify-between px-4 py-3 sm:px-6">
    <a href="` + template.HTMLEscapeString(brandURL) + `" class="whitespace-nowrap font-mono text-base font-bold text-white">
      <span class="glow text-emerald-400">net</span>sekurity<span class="text-emerald-500">.com</span>
      <span class="ml-2 hidden text-[11px] text-gray-500 lg:inline"># ptaas</span>
    </a>
    <nav aria-label="Main navigation" class="hidden items-center gap-3 text-xs text-gray-400 lg:flex">` + desktop.String() + `
    </nav>
    <div class="flex items-center gap-3">
      <!-- The auth CTA stays visible at every width. It used to be md:flex, which
           hid the only call to action on phones behind the burger menu — where the
           majority of ad traffic lands. -->
      <div class="flex items-center gap-3">` + authDesktop + `</div>
      <button type="button" aria-label="Open menu" aria-expanded="false"
        onclick="nskToggleMobile()" id="nsk-burger"
        class="flex h-9 w-9 items-center justify-center rounded border border-white/15 font-mono text-lg text-emerald-300 hover:bg-white/5 lg:hidden">☰</button>
    </div>
  </div>
  <!-- Mobile menu (shared drawer) -->
  <div id="nsk-mobile-menu" class="hidden border-t border-emerald-500/25 bg-ink/95 px-4 pb-4 pt-2 lg:hidden">
    <nav aria-label="Mobile navigation" class="flex flex-col gap-1 font-mono text-sm">` + mobile.String() + `
      <div class="my-1 border-t border-white/10"></div>
      <div class="flex flex-col gap-2 px-2">` + authMobile + `</div>
    </nav>
  </div>
</header>`
	return template.HTML(h)
}

// headerMobileJS is the single shared JS toggle for the mobile drawer. Include it once
// on any page that uses renderHeader.
const headerMobileJS = `<script>
function nskToggleMobile() {
  var b = document.getElementById('nsk-burger');
  var m = document.getElementById('nsk-mobile-menu');
  var open = m.classList.contains('hidden');
  if (open) { m.classList.remove('hidden'); b.setAttribute('aria-expanded','true'); b.textContent='×'; }
  else { m.classList.add('hidden'); b.setAttribute('aria-expanded','false'); b.textContent='☰'; }
  m.querySelectorAll('a').forEach(function(a){ a.addEventListener('click', function(){ m.classList.add('hidden'); b.textContent='☰'; }); });
}
</script>`
