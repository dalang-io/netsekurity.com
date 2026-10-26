package main

// Templates for the public site. Each page is siteLayoutHTML + siteSharedHTML +
// siteSVGHTML (site_svg.go, generated) + its own body (see site.go). The theme is
// the light one shared with the dashboard; component classes live in
// assets/input.css.
//
// Rules the tests enforce (security_test.go, site_test.go):
//   - no inline event handler and no style="…" in markup: the page runs under a
//     nonce-based CSP, so behaviour is attached by the one script in the layout;
//   - every <script> carries the response's nonce;
//   - decorative illustrations are aria-hidden, and the idea each one carries is
//     also in text beside it.

const siteLayoutHTML = `{{define "layout"}}<!DOCTYPE html>
<html lang="{{.Lang.HTMLLang}}">
<head>
<meta charset="UTF-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1"/>
<title>{{.Title}}</title>
<meta name="description" content="{{.Description}}"/>
<meta name="robots" content="{{.Robots}}"/>
<meta name="color-scheme" content="light"/>
<meta name="theme-color" content="#facc15"/>
<meta name="application-name" content="Netsekurity"/>
<meta name="facebook-domain-verification" content="a8ols447zx1ngw2l4xdthddr5t0pm1"/>
<link rel="canonical" href="{{.Canonical}}"/>
{{range .Alternates}}<link rel="alternate" hreflang="{{.Lang}}" href="{{.Href}}"/>
{{end}}<link rel="sitemap" type="application/xml" href="/sitemap.xml"/>
<link rel="icon" type="image/svg+xml" href="{{asset "/favicon.svg"}}"/>
<link rel="icon" type="image/png" sizes="64x64" href="{{asset "/favicon-64.png"}}"/>
<link rel="apple-touch-icon" href="{{asset "/favicon-512.png"}}"/>
<meta property="og:type" content="website"/>
<meta property="og:site_name" content="Netsekurity"/>
<meta property="og:title" content="{{.Title}}"/>
<meta property="og:description" content="{{.Description}}"/>
<meta property="og:url" content="{{.Canonical}}"/>
<meta property="og:locale" content="{{.Lang.OGLocale}}"/>
{{range .Alternates}}{{if ne .Lang "x-default"}}{{if ne .Lang $.Lang.HTMLLang}}<meta property="og:locale:alternate" content="{{.OG}}"/>
{{end}}{{end}}{{end}}<meta property="og:image" content="{{.OGImage}}"/>
<meta property="og:image:width" content="1200"/>
<meta property="og:image:height" content="630"/>
<meta property="og:image:alt" content="{{.Title}}"/>
<meta name="twitter:card" content="summary_large_image"/>
<meta name="twitter:title" content="{{.Title}}"/>
<meta name="twitter:description" content="{{.Description}}"/>
<meta name="twitter:image" content="{{.OGImage}}"/>
<meta name="twitter:image:alt" content="{{.Title}}"/>
<link rel="preload" href="/fonts/PlusJakartaSans-latin-v1.woff2" as="font" type="font/woff2" crossorigin/>
<link rel="stylesheet" href="/css/styles.css?v={{.CSSHash}}"/>
{{.FBPixel}}{{.Analytics}}
<script nonce="{{.Nonce}}" defer src="https://umami.dalang.io/script.js" data-website-id="` + umamiSite + `"></script>
{{if .HTMX}}<meta name="htmx-config" content='{"includeIndicatorStyles":false}'/>
<script nonce="{{.Nonce}}" src="/js/htmx-1.9.12.min.js" defer></script>
{{end}}<script type="application/ld+json" nonce="{{.Nonce}}">{{.JSONLD}}</script>
</head>
<body class="nsk-app min-h-screen overflow-x-hidden">
<a href="#main" class="sr-only focus:not-sr-only focus:fixed focus:left-2 focus:top-2 focus:z-50 focus:bg-white focus:p-3 focus:font-bold">{{.C.Skip}}</a>

<header class="sticky top-0 z-40 border-b-2 border-black bg-white">
  <div class="mx-auto flex h-16 max-w-7xl items-center justify-between gap-3 px-4 sm:px-6">
    <a href="{{.HomeHref}}" class="flex items-center gap-2.5" aria-label="Netsekurity">
      <img src="{{asset "/favicon.svg"}}" width="34" height="34" alt="" class="h-[34px] w-[34px]"/>
      <span class="text-xl font-extrabold tracking-tight">Netsekurity</span>
    </a>
    <nav aria-label="Main" class="hidden items-center gap-0.5 xl:flex">
      {{range .Nav}}<a href="{{.Href}}" data-umami-event="{{.Event}}" class="nsk-nav-link whitespace-nowrap !min-h-[40px] !px-3 !text-sm"{{if eq .Key $.Key}} aria-current="page"{{end}}>{{.Label}}</a>
      {{end}}
    </nav>
    <div class="flex items-center gap-2">
      {{if .LangLinks}}<div class="hidden items-center gap-1 sm:flex" role="group" aria-label="{{.C.LangLabel}}">
        {{range .LangLinks}}<a href="{{.Href}}" lang="{{.Code}}" hreflang="{{.Code}}" title="{{.Name}}" aria-label="{{.Name}}" class="nsk-nav-link !min-h-[36px] !px-2 !text-xs"{{if .Current}} aria-current="true"{{end}}>{{.Label}}</a>{{end}}
      </div>{{end}}
      <a href="{{.AuthHref}}" data-umami-event="netsekurity_header_masuk" data-umami-event-tujuan="{{.AuthHref}}" class="nsk-btn nsk-btn-ghost nsk-btn-sm hidden whitespace-nowrap sm:inline-flex">{{.AuthLabel}}</a>
      <a href="{{.StartHref}}" data-umami-event="netsekurity_header_start" data-umami-event-tujuan="{{.StartHref}}" class="nsk-btn nsk-btn-sm hidden whitespace-nowrap md:inline-flex">{{.C.Start}}</a>
      <button type="button" id="nsk-burger" class="nsk-btn nsk-btn-ghost nsk-btn-sm xl:hidden" aria-label="{{.C.MenuOpen}}" aria-controls="nsk-mobile" aria-expanded="false" data-open="{{.C.MenuOpen}}" data-close="{{.C.MenuClose}}">{{icon "menu" 20}}</button>
    </div>
  </div>
  <div id="nsk-mobile" class="hidden border-t-2 border-black bg-white px-4 pb-4 pt-2 xl:hidden">
    <nav aria-label="Mobile" class="flex flex-col gap-1">
      {{range .Nav}}<a href="{{.Href}}" class="nsk-nav-link"{{if eq .Key $.Key}} aria-current="page"{{end}}>{{.Label}}</a>
      {{end}}
    </nav>
    <div class="mt-3 flex flex-wrap items-center gap-2 border-t-2 border-gray-200 pt-3">
      {{range .LangLinks}}<a href="{{.Href}}" lang="{{.Code}}" hreflang="{{.Code}}" title="{{.Name}}" aria-label="{{.Name}}" class="nsk-nav-link !min-h-[44px]"{{if .Current}} aria-current="true"{{end}}>{{.Label}}</a>{{end}}
      <a href="{{.AuthHref}}" class="nsk-btn nsk-btn-ghost nsk-btn-sm">{{.AuthLabel}}</a>
      <a href="{{.StartHref}}" class="nsk-btn nsk-btn-sm">{{.C.Start}}</a>
    </div>
  </div>
</header>

<main id="main">
{{template "body" .}}
</main>

<footer class="border-t-2 border-black bg-white">
  <div class="mx-auto max-w-7xl px-4 py-12 sm:px-6">
    <div class="grid gap-10 md:grid-cols-[1.3fr_1fr_1fr_1fr]">
      <div>
        <a href="{{.HomeHref}}" class="flex items-center gap-2.5" aria-label="Netsekurity">
          <img src="{{asset "/favicon.svg"}}" width="34" height="34" alt="" class="h-[34px] w-[34px]"/>
          <span class="text-xl font-extrabold tracking-tight">Netsekurity</span>
        </a>
        <p class="mt-3 max-w-xs text-sm text-gray-700">{{.C.Tagline}}</p>
        <p class="mt-3 max-w-xs text-sm text-gray-700">{{.C.Operator}}</p>
      </div>
      <nav aria-label="{{.C.ColProduct}}">
        <p class="nsk-nav-group">{{.C.ColProduct}}</p>
        <ul class="text-sm">
          <li><a class="nsk-footlink" href="{{.HomeHref}}#how">{{.C.NavHow}}</a></li>
          <li><a class="nsk-footlink" href="{{.HomeHref}}#pricing">{{.C.NavPricing}}</a></li>
          <li><a class="nsk-footlink" href="{{.SampleHref}}">{{.C.NavSample}}</a></li>
          <li><a class="nsk-footlink" href="{{.SetupHref}}">{{.C.LinkStartGuide}}</a></li>
          <li><a class="nsk-footlink" href="{{.ScopeHref}}">{{.C.LinkScope}}</a></li>
          <li><a class="nsk-footlink" href="{{.DocsHref}}">{{.C.NavDocs}}</a></li>
        </ul>
      </nav>
      <nav aria-label="{{.C.ColEnterprise}}">
        <p class="nsk-nav-group">{{.C.ColEnterprise}}</p>
        <ul class="text-sm">
          <li><a class="nsk-footlink" href="{{.EntHref}}">{{.C.NavEnterprise}}</a></li>
          <li><a class="nsk-footlink" href="{{.EntHref}}#procurement">{{.C.LinkProcurement}}</a></li>
          <li><a class="nsk-footlink" href="{{.TrustHref}}">{{.C.NavTrust}}</a></li>
          <li><a class="nsk-footlink" href="{{.GuideHref}}">{{.C.LinkGuide}}</a></li>
          <li><a class="nsk-footlink" href="{{.ContactHref}}" data-umami-event="netsekurity_footer_kontak">{{.C.LinkContact}}</a></li>
        </ul>
      </nav>
      <nav aria-label="{{.C.ColResources}}">
        <p class="nsk-nav-group">{{.C.ColResources}}</p>
        <ul class="text-sm">
          <li><a class="nsk-footlink" href="https://dalang.io/" rel="noopener">{{.C.LinkDalang}}</a></li>
          <li><a class="nsk-footlink" href="/llms.txt">{{.C.LinkLLMS}}</a></li>
          <li><a class="nsk-footlink" href="/.well-known/security.txt">{{.C.LinkSecurityTxt}}</a></li>
          <li><a class="nsk-footlink" href="/sitemap.xml">{{.C.LinkSitemap}}</a></li>
          <li><a class="nsk-footlink" href="{{.AuthHref}}" data-umami-event="netsekurity_footer_masuk" data-umami-event-tujuan="{{.AuthHref}}">{{.AuthLabel}}</a></li>
        </ul>
      </nav>
    </div>
    <div class="mt-10 flex flex-wrap items-center gap-x-6 gap-y-1 border-t-2 border-gray-200 pt-6 text-sm text-gray-600">
      <p>{{.C.Rights}}</p>
      <a class="nsk-footlink" rel="noopener" href="https://dalang.io/privacy" data-umami-event="netsekurity_footer_privacy">{{.C.LinkPrivacy}}</a>
      <a class="nsk-footlink" rel="noopener" href="https://dalang.io/terms-of-use" data-umami-event="netsekurity_footer_terms">{{.C.LinkTerms}}</a>
      <a class="nsk-footlink" rel="noopener" href="https://dalang.io/security" data-umami-event="netsekurity_footer_security">{{.C.LinkTrustCenter}}</a>
    </div>
  </div>
</footer>
<script nonce="{{.Nonce}}">
(function () {
  // Mobile menu: a button, Escape closes it and returns focus to the button.
  var b = document.getElementById('nsk-burger'), m = document.getElementById('nsk-mobile');
  function setMenu(open) {
    if (!b || !m) return;
    m.classList.toggle('hidden', !open);
    b.setAttribute('aria-expanded', open ? 'true' : 'false');
    b.setAttribute('aria-label', open ? b.getAttribute('data-close') : b.getAttribute('data-open'));
  }
  if (b && m) {
    b.addEventListener('click', function () { setMenu(m.classList.contains('hidden')); });
    document.addEventListener('keydown', function (e) {
      if (e.key === 'Escape' && !m.classList.contains('hidden')) { setMenu(false); b.focus(); }
    });
  }

  // Copy buttons on code blocks, and print buttons. One delegated listener, so the
  // markup carries no inline handler (the CSP forbids them).
  var timer;
  document.addEventListener('click', function (e) {
    var t = e.target;
    var copy = t.closest ? t.closest('[data-copy]') : null;
    if (copy) {
      var box = copy.closest('.nsk-codebox'), pre = box && box.querySelector('pre');
      if (!pre) return;
      var text = pre.textContent, label = copy.querySelector('span'), status = box.querySelector('[role="status"]');
      var done = function () {
        var was = label.getAttribute('data-was') || label.textContent;
        label.setAttribute('data-was', was);
        label.textContent = box.getAttribute('data-copied');
        if (status) status.textContent = box.getAttribute('data-copied');
        clearTimeout(timer);
        timer = setTimeout(function () { label.textContent = was; if (status) status.textContent = ''; }, 2000);
      };
      if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(text).then(done, function () {});
      } else {
        var r = document.createRange(); r.selectNodeContents(pre);
        var s = window.getSelection(); s.removeAllRanges(); s.addRange(r);
        try { if (document.execCommand('copy')) done(); } catch (x) {}
        s.removeAllRanges();
      }
      return;
    }
    var print = t.closest ? t.closest('[data-print]') : null;
    if (print) {
      var what = print.getAttribute('data-print');
      if (what === 'procurement') {
        document.body.classList.add('nsk-print-procurement');
        window.addEventListener('afterprint', function () { document.body.classList.remove('nsk-print-procurement'); }, { once: true });
      }
      window.print();
    }
  });
})();
</script>
{{.ViewContent}}
</body>
</html>{{end}}
`

const siteSharedHTML = `
{{define "crumb"}}<nav aria-label="Breadcrumb" class="mb-5 text-sm text-gray-700">
  <ol class="flex flex-wrap items-center gap-2">
    <li><a class="underline" href="{{.HomeHref}}">{{.C.Breadcrumb}}</a></li>
    <li aria-hidden="true">›</li>
    <li aria-current="page" class="font-bold">{{.Crumb}}</li>
  </ol>
</nav>{{end}}

{{define "faq"}}<div class="mt-8 space-y-3">
{{range .}}<details class="nsk-card-flat group">
  <summary class="flex min-h-[44px] cursor-pointer list-none items-center justify-between gap-4 px-5 py-4 text-lg font-extrabold marker:content-['']">
    <span>{{.Q}}</span><span aria-hidden="true" class="text-2xl leading-none transition-transform group-open:rotate-45">+</span>
  </summary>
  <div class="nsk-prose border-t-2 border-gray-200 px-5 py-4 text-gray-800">{{.A}}</div>
</details>
{{end}}</div>{{end}}

{{define "codebox"}}<div class="nsk-codebox" data-copied="{{.Copied}}">
  <button type="button" class="nsk-copy" data-copy>{{icon "copy" 14}}<span>{{.Copy}}</span></button>
  <span class="sr-only" role="status"></span>
  <pre class="nsk-pre">{{.Code}}</pre>
</div>{{end}}

{{define "coverage"}}{{with .C.Coverage}}<div class="mt-8 grid gap-6 lg:grid-cols-2">
  <div class="nsk-card">
    <h3 class="nsk-card-head flex items-center gap-2">{{icon "check" 18}}{{.InTitle}}</h3>
    <ul class="space-y-3 p-5">
      {{range .In}}<li><p class="font-extrabold">{{.Title}}</p><p class="text-sm text-gray-700">{{.Body}}</p></li>
      {{end}}
    </ul>
  </div>
  <div class="nsk-card">
    <h3 class="nsk-card-head flex items-center gap-2">{{icon "x" 18}}{{.OutTitle}}</h3>
    <ul class="space-y-3 p-5">
      {{range .Out}}<li><p class="font-extrabold">{{.Title}}</p><p class="text-sm text-gray-700">{{.Body}}</p></li>
      {{end}}
    </ul>
  </div>
</div>
<p class="nsk-note mt-6 text-sm">{{.Note}}</p>{{end}}{{end}}

{{define "tiers"}}<div class="mt-8 grid gap-6 lg:grid-cols-3">
  {{range .C.Tiers}}<div class="nsk-card-flat flex flex-col p-5">
    <p class="text-sm font-extrabold uppercase tracking-wide text-gray-700">{{.Name}}</p>
    <p class="mt-2 text-2xl font-extrabold">{{.Price}}</p>
    <p class="mt-1 text-gray-700">{{.Audience}}</p>
    <ul class="mt-4 space-y-2 text-sm">
      {{range .Points}}<li class="flex gap-2"><span class="mt-0.5 text-green-800">{{icon "check" 16}}</span><span>{{.}}</span></li>
      {{end}}
    </ul>
  </div>
  {{end}}
</div>{{end}}

{{define "stepsvg"}}{{if eq . 0}}{{template "svg-step-1"}}{{else if eq . 1}}{{template "svg-step-2"}}{{else if eq . 2}}{{template "svg-step-3"}}{{else}}{{template "svg-step-4"}}{{end}}{{end}}

{{define "contact"}}<section id="contact" class="border-t-2 border-black bg-[#fffbeb] py-14 sm:py-16">
  <div class="mx-auto max-w-3xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.C.Form.Title}}</h2>
    <p class="mt-3 text-lg text-gray-800">{{.C.Form.Lead}}</p>
    <form hx-post="/contact" hx-target="#contact-result" hx-swap="innerHTML" hx-disabled-elt="find button" class="nsk-card mt-6 space-y-4 p-5">
      <input type="hidden" name="source" value="{{.Key}}"/>
      <input type="hidden" name="lang" value="{{.Lang.Code}}"/>
      <input type="text" name="website" tabindex="-1" autocomplete="off" class="hidden" aria-hidden="true"/>
      <div class="grid gap-4 sm:grid-cols-2">
        <div><label for="c-name" class="nsk-label">{{.C.Form.Name}}</label><input id="c-name" name="name" maxlength="200" autocomplete="name" class="nsk-input"/></div>
        <div><label for="c-email" class="nsk-label">{{.C.Form.Email}}</label><input id="c-email" name="email" type="email" required maxlength="200" autocomplete="email" class="nsk-input"/></div>
      </div>
      <div><label for="c-msg" class="nsk-label">{{.C.Form.Message}}</label><textarea id="c-msg" name="message" required rows="4" maxlength="4000" placeholder="{{.C.Form.MessagePlaceholder}}" class="nsk-input"></textarea></div>
      <div class="flex flex-wrap items-center gap-3">
        <button data-umami-event="netsekurity_kontak_form_kirim" class="nsk-btn">{{.C.Form.Send}}</button>
        <span class="text-sm text-gray-700">{{.C.Form.Reply}}</span>
      </div>
      <div id="contact-result" aria-live="polite"></div>
    </form>
    <p class="mt-4 text-sm text-gray-700"><a class="nsk-textlink" data-umami-event="netsekurity_kontak_email" href="mailto:sales@dalang.io?subject=%5Bnetsekurity%5D%20enterprise">sales@dalang.io</a></p>
  </div>
</section>{{end}}

`

const siteHomeHTML = `{{define "body"}}{{with .Home}}
<section class="border-b-2 border-black bg-[#fffbeb]">
  <div class="nsk-fit mx-auto grid max-w-7xl items-center gap-10 px-4 py-12 sm:px-6 lg:grid-cols-[1.15fr_0.85fr] lg:py-20">
    <div>
      <span class="nsk-tag">{{.Eyebrow}}</span>
      <h1 class="mt-4 text-4xl font-extrabold leading-[1.1] tracking-tight sm:text-5xl">{{.H1}}</h1>
      <p class="mt-5 max-w-2xl text-lg text-gray-800">{{.Lead}}</p>
      <p class="mt-3 max-w-2xl text-gray-700">{{.SubLead}}</p>
      <div class="mt-7 flex flex-wrap gap-3">
        <a href="{{$.StartHref}}" data-umami-event="netsekurity_hero_cta" class="nsk-btn">{{.CTAPrimary}}</a>
        <a href="{{$.SampleHref}}" data-umami-event="netsekurity_hero_sample" class="nsk-btn nsk-btn-ghost">{{.CTASecondary}}</a>
      </div>
      <p class="text-sm text-gray-700">{{.SignInNote}}</p>
      {{$.ScansProof}}
      <ul class="mt-8 grid gap-2 sm:grid-cols-3">
        {{range .Assurances}}<li class="flex gap-2 border-2 border-black bg-white p-3 text-sm font-bold"><span class="mt-0.5 text-green-800">{{icon "check" 16}}</span><span>{{.}}</span></li>
        {{end}}
      </ul>
    </div>
    <aside aria-label="{{.Card.Label}}" class="nsk-card">
      <div class="flex items-center justify-between gap-2 border-b-2 border-black px-4 py-3">
        <span class="nsk-tag">{{.Card.Label}}</span>
        <span class="nsk-mono text-xs text-gray-700">{{.Card.Domain}}</span>
      </div>
      <div class="p-4">
        <p class="text-lg font-extrabold">{{.Card.Summary}}</p>
        <p class="mt-1 text-sm font-bold text-yellow-900">{{.Card.Posture}}</p>
        <ul class="mt-4 space-y-2">
          {{range .Card.Findings}}<li class="border-2 border-gray-300 bg-white p-2.5 text-sm">{{.}}</li>
          {{end}}
        </ul>
        <p class="mt-3 text-xs text-gray-600">{{.Card.Note}}</p>
        <a href="{{$.SampleHref}}" class="nsk-textlink text-sm">{{.Card.Link}}<span aria-hidden="true"> →</span></a>
      </div>
    </aside>
  </div>
</section>

<section id="audiences" class="nsk-section nsk-section-tint">
  <div class="mx-auto max-w-7xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.AudiencesH2}}</h2>
    <div class="mt-8 grid gap-6 lg:grid-cols-3">
      {{range $i, $a := .Audiences}}<article class="nsk-card-flat flex flex-col p-6">
        <span class="nsk-icon-tile">{{icon (pick "audiences" $i) 22}}</span>
        <p class="mt-4 text-sm font-extrabold uppercase tracking-wide text-gray-700">{{$a.Role}}</p>
        <h3 class="mt-1 text-xl font-extrabold">{{$a.Tagline}}</h3>
        <ul class="mt-4 flex-1 space-y-3 text-gray-800">
          {{range $a.Bullets}}<li class="flex gap-2"><span class="mt-1 text-green-800">{{icon "check" 16}}</span><span>{{.}}</span></li>
          {{end}}
        </ul>
        <a href="{{$.HomeHref}}#{{$a.Anchor}}" class="nsk-textlink mt-3">{{$a.LinkLabel}}<span aria-hidden="true"> →</span></a>
      </article>
      {{end}}
    </div>
  </div>
</section>

<section id="report" class="nsk-section">
  <div class="nsk-fit mx-auto grid max-w-7xl gap-10 px-4 sm:px-6 lg:grid-cols-2">
    <div>
      <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.ReportH2}}</h2>
      <p class="mt-3 text-lg text-gray-700">{{.ReportLead}}</p>
      <ol class="nsk-steps mt-8 space-y-4">
        {{range .Report}}<li><div><p class="font-extrabold">{{.Title}}</p><p class="text-sm text-gray-700">{{.Body}}</p></div></li>
        {{end}}
      </ol>
      <a href="{{$.SampleHref}}" class="nsk-textlink mt-4">{{.ReportLink}}<span aria-hidden="true"> →</span></a>
    </div>
    <div class="space-y-6">
      <div aria-hidden="true" class="nsk-card-flat-tint p-6"><div class="mx-auto w-full max-w-sm">{{template "svg-report"}}</div></div>
      <div class="nsk-card-flat p-5">
        <h3 class="text-lg font-extrabold">{{.PlainH3}}</h3>
        <dl class="mt-3 space-y-3 text-sm">
          {{range .Plain}}<div><dt class="font-extrabold">{{.Title}}</dt><dd class="text-gray-700">{{.Body}}</dd></div>
          {{end}}
        </dl>
        <a href="{{$.SampleHref}}#s4" class="nsk-textlink mt-2 text-sm">{{.PlainLink}}<span aria-hidden="true"> →</span></a>
      </div>
    </div>
  </div>
</section>

<section id="how" class="nsk-section nsk-section-tint">
  <div class="mx-auto max-w-7xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.HowH2}}</h2>
    <ol class="mt-8 grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
      {{range $i, $s := .Steps}}<li class="nsk-card-flat flex flex-col bg-white p-5">
        <div aria-hidden="true" class="flex h-28 items-center justify-center"><div class="w-full max-w-[9rem]">{{template "stepsvg" $i}}</div></div>
        <div class="mt-3 flex items-center gap-3">
          <span class="flex h-8 w-8 shrink-0 items-center justify-center border-2 border-black bg-primary font-extrabold">{{inc $i}}</span>
          <h3 class="text-lg font-extrabold">{{$s.Title}}</h3>
        </div>
        <p class="mt-2 text-sm text-gray-800">{{$s.Body}}</p>
      </li>
      {{end}}
    </ol>
    <p class="mt-5"><a href="{{$.SetupHref}}" class="nsk-textlink">{{$.C.LinkStartGuide}}</a></p>
    <p id="cicd" class="mt-5"><a href="{{$.DocsHref}}" data-umami-event="netsekurity_nav_docs" class="nsk-textlink">{{$.C.NavDocs}}</a></p>
  </div>
</section>

<section id="what" class="nsk-section">
  <div class="nsk-fit mx-auto grid max-w-7xl items-center gap-10 px-4 sm:px-6 lg:grid-cols-2">
    <div>
      <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.DefinitionH2}}</h2>
      <p class="mt-4 text-lg leading-relaxed text-gray-800">{{.Definition}}</p>
      <p class="mt-4"><a href="{{$.ScopeHref}}" class="nsk-textlink">{{$.C.LinkScope}}</a></p>
      <h3 class="mt-8 text-lg font-extrabold">{{.TypicalH3}}</h3>
      <ul class="mt-3 space-y-2 text-gray-800">
        {{range .Typical}}<li class="flex gap-2"><span class="mt-1 text-green-800">{{icon "check" 16}}</span><span>{{.}}</span></li>
        {{end}}
      </ul>
    </div>
    <div class="nsk-card-flat-tint p-5 sm:p-6">
      <div aria-hidden="true" class="mx-auto w-full max-w-sm">{{template "svg-scope"}}</div>
      <ul class="mt-5 space-y-3 text-sm">
        {{range $i, $l := .ScopeLegend}}<li class="flex gap-3"><span class="nsk-swatch {{if eq $i 0}}nsk-swatch-scanned{{else}}nsk-swatch-hatch{{end}}" aria-hidden="true"></span><span><strong class="font-extrabold">{{$l.Title}}.</strong> {{$l.Body}}</span></li>
        {{end}}
      </ul>
    </div>
  </div>
</section>

<section id="trust" class="nsk-section nsk-section-tint">
  <div class="mx-auto max-w-7xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.TrustH2}}</h2>
    <p class="mt-3 max-w-3xl text-lg text-gray-700">{{.TrustLead}}</p>
    <ul class="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
      {{range $i, $t := .Trust}}<li class="nsk-card-flat bg-white p-5"><span class="nsk-icon-tile">{{icon (pick "home-trust" $i) 20}}</span><p class="mt-3 font-extrabold">{{$t.Title}}</p><p class="mt-1 text-sm text-gray-700">{{$t.Body}}</p></li>
      {{end}}
    </ul>
    <a href="{{$.TrustHref}}" class="nsk-textlink mt-4">{{.TrustLink}}<span aria-hidden="true"> →</span></a>
  </div>
</section>

<section id="pricing" class="nsk-section">
  <div class="mx-auto max-w-7xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.PricingH2}}</h2>
    <p class="mt-3 max-w-3xl text-lg text-gray-700">{{.PricingLead}}</p>
    <div class="nsk-card mt-8 bg-white">
      <table class="w-full text-left">
        <thead><tr class="border-b-2 border-black">
          <th scope="col" class="px-3 py-3 text-sm font-extrabold sm:px-4">{{index .PricingCols 0}}</th>
          <th scope="col" class="hidden px-4 py-3 text-sm font-extrabold md:table-cell">{{index .PricingCols 1}}</th>
          <th scope="col" class="px-3 py-3 text-sm font-extrabold sm:px-4">{{index .PricingCols 2}}</th>
          <th scope="col" class="hidden px-4 py-3 text-sm font-extrabold md:table-cell">{{index .PricingCols 3}}</th>
          <th scope="col" class="px-3 py-3 sm:px-4"><span class="sr-only">{{index .PricingCols 4}}</span></th>
        </tr></thead>
        <tbody>
          {{range $.Pricing.Rows}}<tr class="border-b-2 border-gray-200 last:border-b-0{{if .Best}} bg-[#fffbeb]{{end}}">
            <th scope="row" class="px-3 py-3 align-top font-extrabold sm:px-4">{{.Label}}
              {{if .Best}}<span class="mt-1 block"><span class="nsk-badge nsk-badge-ok !whitespace-normal text-left">{{$.Home.PricingBest}}</span></span>{{end}}
              <span class="mt-0.5 block text-sm font-medium text-gray-700 md:hidden">{{index $.Home.PricingCols 3}}: {{.PerScan}}</span>
            </th>
            <td class="hidden px-4 py-3 align-top md:table-cell">{{.Credits}}</td>
            <td class="px-3 py-3 align-top text-lg font-extrabold sm:px-4">{{.Price}}</td>
            <td class="hidden px-4 py-3 align-top text-gray-700 md:table-cell">{{.PerScan}}</td>
            <td class="px-3 py-3 text-right align-top sm:px-4"><a href="/checkout?package={{.ID}}" data-umami-event="netsekurity_pricing_buy" data-umami-event-paket="{{.ID}}" class="nsk-btn nsk-btn-sm" aria-label="{{$.C.Buy}} {{.Label}}">{{$.C.Buy}}</a></td>
          </tr>
          {{end}}
        </tbody>
      </table>
    </div>
    <p class="mt-4 max-w-3xl text-gray-800">{{.PricingNote}}</p>
    <div class="mt-6 flex flex-wrap items-center gap-3">
      <p class="text-gray-800">{{.PricingTalk}}</p>
      <a href="{{$.ContactHref}}" data-umami-event="netsekurity_pricing_talk" class="nsk-btn nsk-btn-ghost nsk-btn-sm">{{.PricingTalkBtn}}</a>
    </div>
  </div>
</section>

<section id="tiers" class="nsk-section nsk-section-tint">
  <div class="nsk-fit mx-auto grid max-w-7xl items-center gap-8 px-4 sm:px-6 lg:grid-cols-[1fr_20rem]">
    <div><h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{$.C.TiersH2}}</h2>
      <p class="mt-3 max-w-3xl text-lg text-gray-700">{{$.C.TiersLead}}</p>
      {{with index $.C.Tiers 2}}<p class="mt-4 font-bold">{{.Price}}</p>{{end}}
      <a href="{{$.EntHref}}" class="nsk-btn nsk-btn-ghost mt-5">{{$.C.NavEnterprise}}</a>
    </div>
    <div aria-hidden="true" class="mx-auto w-full max-w-[16rem]">{{template "svg-tiers"}}</div>
  </div>
</section>

<section id="faq" class="nsk-section">
  <div class="mx-auto max-w-4xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.FAQH2}}</h2>
    {{template "faq" .FAQ}}
  </div>
</section>

<section class="border-y-2 border-black bg-primary">
  <div class="mx-auto flex max-w-7xl flex-col gap-5 px-4 py-10 sm:px-6 md:flex-row md:items-center md:justify-between">
    <div>
      <h2 class="text-2xl font-extrabold tracking-tight">{{.BandH2}}</h2>
      <p class="mt-1 text-gray-900">{{.BandBody}}</p>
    </div>
    <div class="flex flex-wrap gap-3">
      <a href="{{$.StartHref}}" data-umami-event="netsekurity_band_start" data-umami-event-tujuan="{{$.StartHref}}" class="nsk-btn nsk-btn-dark">{{.CTATertiary}}</a>
      <a href="{{$.SampleHref}}" data-umami-event="netsekurity_band_sample" class="nsk-btn nsk-btn-ghost">{{.CTASecondary}}</a>
    </div>
  </div>
</section>
{{end}}
{{template "contact" .}}
{{end}}
`

const siteEnterpriseHTML = `{{define "body"}}{{with .Ent}}
<section class="border-b-2 border-black bg-[#fffbeb]">
  <div class="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:py-16">
    {{template "crumb" $}}
    <span class="nsk-tag">{{.Eyebrow}}</span>
    <h1 class="mt-4 max-w-4xl text-4xl font-extrabold leading-[1.1] tracking-tight sm:text-5xl">{{.H1}}</h1>
    <p class="mt-5 max-w-3xl text-lg text-gray-800">{{.Lead}}</p>
    <div class="mt-6 flex flex-wrap items-center gap-x-5 gap-y-1">
      <a class="nsk-btn" href="#contact" data-umami-event="netsekurity_enterprise_talk">{{.HeroCTA}}</a>
      <a class="nsk-textlink" rel="noopener" href="{{$.ProfileHref}}" data-umami-event="netsekurity_enterprise_profile">{{icon "download" 16}}{{.ProfileLink}}</a>
    </div>
    <ul class="mt-4 flex flex-wrap gap-2">
      {{range .Jump}}<li><a class="nsk-btn nsk-btn-ghost nsk-btn-sm" href="#{{.Anchor}}">{{.Label}}</a></li>
      {{end}}
    </ul>
  </div>
</section>

<section id="board" class="nsk-section">
  <div class="mx-auto max-w-7xl px-4 sm:px-6">
    <h2 class="max-w-4xl text-2xl font-extrabold tracking-tight sm:text-3xl">{{.BoardH2}}</h2>
    <p class="mt-3 max-w-3xl text-lg text-gray-700">{{.BoardIntro}}</p>
    <dl class="mt-8 grid gap-4 md:grid-cols-2">
      {{range .Board}}<div class="nsk-card-flat p-5"><dt class="text-lg font-extrabold">{{.Title}}</dt><dd class="mt-2 text-gray-800">{{.Body}}</dd></div>
      {{end}}
    </dl>
    <p class="nsk-note mt-6">{{.BoardCallout}}</p>
  </div>
</section>

{{end}}
<section id="escalation" class="nsk-section nsk-section-tint">
  <div class="mx-auto max-w-7xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.C.TiersH2}}</h2>
    <p class="mt-3 max-w-3xl text-lg text-gray-700">{{.C.TiersLead}}</p>
    <div class="nsk-card-flat mt-8 divide-y-2 divide-gray-200 bg-white">
      {{range .C.Tiers}}<div class="grid gap-3 p-5 md:grid-cols-[16rem_1fr] md:gap-8">
        <div><p class="font-extrabold">{{.Name}}</p><p class="mt-1 text-lg font-extrabold">{{.Price}}</p></div>
        <div>
          <p class="text-gray-800">{{.Audience}}</p>
          <ul class="mt-2 grid gap-1 text-sm text-gray-700 sm:grid-cols-2">
            {{range .Points}}<li class="flex gap-2"><span class="mt-0.5 text-green-800">{{icon "check" 16}}</span><span>{{.}}</span></li>
            {{end}}
          </ul>
        </div>
      </div>
      {{end}}
    </div>
    <a href="{{.GuideHref}}" class="nsk-textlink mt-4">{{.Ent.EscLink}}<span aria-hidden="true"> →</span></a>
  </div>
</section>
{{with .Ent}}
<section id="engineering" class="nsk-section nsk-section-tint">
  <details class="mx-auto max-w-7xl px-4 sm:px-6">
    <summary class="cursor-pointer text-xl font-extrabold">{{.EngH2}}</summary>
    <p class="mt-3 max-w-3xl text-lg text-gray-700">{{.EngIntro}}</p>
    <div class="nsk-fit mt-8 grid gap-6 lg:grid-cols-2">
      {{range .Eng}}<article class="nsk-card-flat bg-white p-5">
        <h3 class="text-lg font-extrabold">{{.Title}}</h3>
        {{range .Paragraphs}}<p class="mt-2 text-gray-800">{{.}}</p>{{end}}
        {{if .Bullets}}<ul class="mt-3 space-y-1.5 text-sm text-gray-800">{{range .Bullets}}<li class="flex gap-2"><span class="mt-0.5 text-green-800">{{icon "check" 16}}</span><span>{{.}}</span></li>{{end}}</ul>{{end}}
        {{if .Code}}<div class="mt-3">{{template "codebox" (codeArgs $.C .Code)}}</div>{{end}}
      </article>
      {{end}}
    </div>
  </details>
</section>
{{end}}

{{with .Ent}}
<section id="procurement" class="nsk-section">
  <div class="mx-auto max-w-7xl px-4 sm:px-6">
    <p class="hidden print:block text-sm font-bold">Netsekurity · {{$.C.LinkProcurement}} · Dalang Pte. Ltd. · {{$.Modified}}</p>
    <div class="flex flex-wrap items-end justify-between gap-3">
      <h2 class="max-w-4xl text-2xl font-extrabold tracking-tight sm:text-3xl">{{.ProcH2}}</h2>
      <button type="button" class="nsk-btn nsk-btn-ghost nsk-btn-sm nsk-noprint" data-print="procurement">{{icon "printer" 16}}{{$.C.PrintLabel}}</button>
    </div>
    <p class="mt-3 max-w-3xl text-lg text-gray-700">{{.ProcIntro}}</p>
    <dl class="nsk-card-flat mt-8 divide-y-2 divide-gray-200">
      {{range .ProcFacts}}<div class="grid gap-1 px-5 py-4 md:grid-cols-[14rem_1fr] md:gap-6"><dt class="font-extrabold">{{.Title}}</dt><dd class="text-gray-800">{{.Body}}</dd></div>
      {{end}}
    </dl>
    <p class="mt-4 text-sm text-gray-700">{{.ProcNote}}</p>
    <p class="nsk-noprint mt-1 text-sm text-gray-700">{{.ProfileLead}} <a class="font-bold underline" rel="noopener" href="{{$.ProfileHref}}">{{.ProfileLink}}</a></p>
  </div>
</section>
{{end}}



<section id="coverage" class="nsk-section">
  <div class="mx-auto max-w-7xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.C.Coverage.H2}}</h2>
    <p class="mt-3 max-w-3xl text-lg text-gray-700">{{.C.Coverage.Lead}}</p>
    {{template "coverage" .}}
  </div>
</section>

{{with .Ent}}
<section id="faq" class="nsk-section nsk-section-tint">
  <div class="mx-auto max-w-4xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.FAQH2}}</h2>
    {{template "faq" .FAQ}}
  </div>
</section>
{{end}}
{{template "contact" .}}
{{end}}
`

const siteTrustHTML = `{{define "body"}}{{with .Trust}}
<section class="border-b-2 border-black bg-[#fffbeb]">
  <div class="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:py-16">
    {{template "crumb" $}}
    <span class="nsk-tag">{{.Eyebrow}}</span>
    <h1 class="mt-4 max-w-4xl text-4xl font-extrabold leading-[1.1] tracking-tight sm:text-5xl">{{.H1}}</h1>
    <p class="mt-5 max-w-3xl text-lg text-gray-800">{{.Lead}}</p>
    <ul class="mt-6 flex flex-wrap gap-2">
      {{range .Jump}}<li><a class="nsk-btn nsk-btn-ghost nsk-btn-sm" href="#{{.Anchor}}">{{.Label}}</a></li>
      {{end}}
    </ul>
  </div>
</section>

<section id="responsible" class="nsk-section">
  <div class="mx-auto max-w-7xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.RespH2}}</h2>
    <p class="mt-3 max-w-3xl text-lg text-gray-700">{{.RespIntro}}</p>
    <dl class="mt-8 grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      {{range $i, $r := .Resp}}<div class="nsk-card-flat p-5"><dt class="flex flex-col items-start gap-3 text-lg font-extrabold"><span class="nsk-icon-tile">{{icon (pick "trust-resp" $i) 20}}</span>{{$r.Title}}</dt><dd class="mt-2 text-gray-800">{{$r.Body}}</dd></div>
      {{end}}
    </dl>
  </div>
</section>

<section id="flow" class="nsk-section nsk-section-tint">
  <div class="nsk-fit mx-auto grid max-w-7xl gap-10 px-4 sm:px-6 lg:grid-cols-[1fr_16rem]">
    <div class="flex flex-col">
      <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.FlowH2}}</h2>
      <p class="mt-3 max-w-3xl text-lg text-gray-700">{{.FlowLead}}</p>
      <ol class="nsk-steps mt-8 flex-1 space-y-5 lg:flex lg:flex-col lg:justify-between lg:space-y-0">
        {{range .Flow}}<li><div><p class="font-extrabold">{{.Title}}</p><p class="text-gray-800">{{.Body}}</p></div></li>
        {{end}}
      </ol>
    </div>
    <div aria-hidden="true" class="nsk-card-flat mx-auto w-full max-w-[14rem] bg-white p-4">{{template "svg-flow"}}</div>
  </div>
</section>

<section id="data" class="nsk-section">
  <div class="mx-auto max-w-7xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.DataH2}}</h2>
    <p class="mt-3 max-w-3xl text-lg text-gray-700">{{.DataIntro}}</p>
    <div class="nsk-card-flat mt-8 bg-white">
      <table class="w-full text-left">
        <thead><tr class="border-b-2 border-black">{{range .DataCols}}<th scope="col" class="px-4 py-3 text-sm font-extrabold">{{.}}</th>{{end}}</tr></thead>
        <tbody>
          {{range .Data}}<tr class="border-b-2 border-gray-200 last:border-b-0"><th scope="row" class="w-32 px-4 py-3 align-top font-extrabold sm:w-44">{{.Item}}</th><td class="px-4 py-3 text-gray-800">{{.Detail}}</td></tr>
          {{end}}
        </tbody>
      </table>
    </div>
  </div>
</section>

<section id="third-parties" class="nsk-section nsk-section-tint">
  <div class="mx-auto max-w-7xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.SubH2}}</h2>
    <p class="mt-3 max-w-3xl text-lg text-gray-700">{{.SubIntro}}</p>
    <div class="nsk-card-flat mt-8 bg-white">
      <table class="nsk-stack w-full text-left" role="table">
        <thead role="rowgroup"><tr class="border-b-2 border-black" role="row">{{range .SubCols}}<th scope="col" role="columnheader" class="px-4 py-3 text-sm font-extrabold">{{.}}</th>{{end}}</tr></thead>
        <tbody role="rowgroup">
          {{range .Subs}}<tr class="border-b-2 border-gray-200 last:border-b-0" role="row">
            <th scope="row" role="rowheader" class="w-56 px-4 py-3 align-top font-extrabold">{{.Name}}</th>
            <td role="cell" data-label="{{index $.Trust.SubCols 1}}" class="px-4 py-3 align-top text-gray-800">{{.Purpose}}</td>
            <td role="cell" data-label="{{index $.Trust.SubCols 2}}" class="px-4 py-3 align-top text-gray-800">{{.Receives}}</td>
          </tr>
          {{end}}
        </tbody>
      </table>
    </div>
  </div>
</section>

<section id="access" class="nsk-section">
  <div class="mx-auto max-w-7xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.AccessH2}}</h2>
    <dl class="mt-8 grid gap-4 md:grid-cols-3">
      {{range .Access}}<div class="nsk-card-flat p-5"><dt class="text-lg font-extrabold">{{.Title}}</dt><dd class="mt-2 text-gray-800">{{.Body}}</dd></div>
      {{end}}
    </dl>
  </div>
</section>

<section id="disclosure" class="nsk-section nsk-section-tint">
  <div class="mx-auto max-w-4xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.DiscH2}}</h2>
    <p class="mt-3 text-lg text-gray-800">{{.DiscBody}}</p>
    <ul class="mt-4 space-y-2 text-gray-800">{{range .Disc}}<li class="flex gap-2"><span class="mt-1 text-green-800">{{icon "check" 16}}</span><span>{{.}}</span></li>{{end}}</ul>
  </div>
</section>

<section id="operator" class="nsk-section">
  <div class="mx-auto max-w-7xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.OpH2}}</h2>
    {{range .OpBody}}<p class="mt-3 max-w-3xl text-lg text-gray-800">{{.}}</p>{{end}}
    <dl class="nsk-card-flat mt-8 divide-y-2 divide-gray-200 bg-white">
      {{range .OpFacts}}<div class="grid gap-1 px-5 py-3 md:grid-cols-[12rem_1fr] md:gap-6"><dt class="font-extrabold">{{.Title}}</dt><dd class="text-gray-800">{{.Body}}</dd></div>
      {{end}}
    </dl>
    <p class="mt-4 text-sm text-gray-700">{{.OpSource}}</p>
    <p class="mt-2 text-sm text-gray-700">{{.ProfileLead}} <a class="font-bold underline" rel="noopener" href="{{$.ProfileHref}}">{{.ProfileLink}}</a></p>
    <p class="mt-6 max-w-3xl text-gray-800">{{.PolicyLead}}</p>
    <ul class="mt-3 flex flex-wrap gap-3">
      <li><a class="nsk-btn nsk-btn-ghost nsk-btn-sm" rel="noopener" href="https://dalang.io/privacy">{{$.C.LinkPrivacy}}</a></li>
      <li><a class="nsk-btn nsk-btn-ghost nsk-btn-sm" rel="noopener" href="https://dalang.io/terms-of-use">{{$.C.LinkTerms}}</a></li>
      <li><a class="nsk-btn nsk-btn-ghost nsk-btn-sm" rel="noopener" href="https://dalang.io/security">{{$.C.LinkTrustCenter}}</a></li>
    </ul>
  </div>
</section>

<section class="nsk-section nsk-section-tint">
  <div class="mx-auto max-w-4xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.LimitsH2}}</h2>
    <p class="mt-3 text-lg text-gray-800">{{.LimitsBody}}</p>
  </div>
</section>
{{end}}
{{end}}
`

const siteCompareHTML = `{{define "body"}}{{with .Cmp}}
<section class="border-b-2 border-black bg-[#fffbeb]">
  <div class="mx-auto max-w-4xl px-4 py-12 sm:px-6 lg:py-16">
    {{template "crumb" $}}
    <span class="nsk-tag">{{.Eyebrow}}</span>
    <h1 class="mt-4 text-4xl font-extrabold leading-[1.1] tracking-tight sm:text-5xl">{{.H1}}</h1>
    <p class="mt-5 text-lg leading-relaxed text-gray-800">{{.Lead}}</p>
    <p class="mt-3 text-sm text-gray-600"><time datetime="` + siteModified + `">{{.Updated}}</time></p>
  </div>
</section>

<section class="nsk-section">
  <div class="mx-auto max-w-7xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.TableH2}}</h2>
    <div class="nsk-card mt-8 bg-white">
      <table class="nsk-stack w-full text-left" role="table">
        <thead role="rowgroup"><tr class="border-b-2 border-black bg-[#fffbeb]" role="row">{{range .Cols}}<th scope="col" role="columnheader" class="px-4 py-3 text-sm font-extrabold">{{.}}</th>{{end}}</tr></thead>
        <tbody role="rowgroup">
          {{range .Rows}}<tr class="border-b-2 border-gray-200 last:border-b-0" role="row"><th scope="row" role="rowheader" class="w-48 px-4 py-3 align-top font-extrabold">{{.Label}}</th>{{range $i, $c := .Cells}}<td role="cell" data-label="{{index $.Cmp.Cols (inc $i)}}" class="px-4 py-3 align-top text-gray-800">{{$c}}</td>{{end}}</tr>
          {{end}}
        </tbody>
      </table>
    </div>
  </div>
</section>

<section class="nsk-section nsk-section-tint">
  <div class="mx-auto grid max-w-7xl gap-6 px-4 sm:px-6 lg:grid-cols-2">
    <div class="nsk-card-flat bg-white p-5">
      <h2 class="flex items-center gap-2 text-xl font-extrabold">{{icon "check" 20}}{{.GoodH2}}</h2>
      <ul class="mt-4 space-y-2 text-gray-800">{{range .Good}}<li class="flex gap-2"><span class="mt-0.5 text-green-800">{{icon "check" 16}}</span><span>{{.}}</span></li>{{end}}</ul>
    </div>
    <div class="nsk-card-flat bg-white p-5">
      <h2 class="flex items-center gap-2 text-xl font-extrabold">{{icon "x" 20}}{{.CannotH2}}</h2>
      <ul class="mt-4 space-y-2 text-gray-800">{{range .Cannot}}<li class="flex gap-2"><span class="mt-0.5 text-red-800">{{icon "x" 16}}</span><span>{{.}}</span></li>{{end}}</ul>
    </div>
  </div>
</section>

<section class="nsk-section">
  <div class="mx-auto max-w-4xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.ChooseH2}}</h2>
    <ol class="mt-8 space-y-4">
      {{range $i, $c := .Choose}}<li class="nsk-card-flat flex gap-4 p-5">
        <span class="flex h-10 w-10 shrink-0 items-center justify-center border-2 border-black bg-primary text-lg font-extrabold">{{inc $i}}</span>
        <div><p class="text-lg font-extrabold">{{$c.Title}}</p><p class="mt-1 text-gray-800">{{$c.Body}}</p></div>
      </li>
      {{end}}
    </ol>
  </div>
</section>

<section id="glossary" class="nsk-section nsk-section-tint">
  <div class="mx-auto max-w-4xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.GlossH2}}</h2>
    <dl class="mt-8 space-y-3">
      {{range .Gloss}}<div class="nsk-card-flat bg-white p-4"><dt class="font-extrabold">{{.Title}}</dt><dd class="mt-1 text-gray-800">{{.Body}}</dd></div>
      {{end}}
    </dl>
  </div>
</section>

<section class="nsk-section">
  <div class="mx-auto max-w-4xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.FAQH2}}</h2>
    {{template "faq" .FAQ}}
    <p class="mt-8"><a href="{{.CTALink}}" class="nsk-btn">{{.CTA}}</a></p>
  </div>
</section>
{{end}}
{{end}}
`
