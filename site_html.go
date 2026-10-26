package main

// Templates for the public site. Each page is siteLayoutHTML + siteSharedHTML +
// its own body (see site.go). The theme is the light one shared with the
// dashboard; component classes live in assets/input.css.

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
<link rel="icon" type="image/svg+xml" href="/favicon.svg"/>
<link rel="icon" type="image/png" sizes="64x64" href="/favicon-64.png"/>
<link rel="apple-touch-icon" href="/favicon-512.png"/>
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
<link rel="preconnect" href="https://fonts.googleapis.com"/>
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin/>
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;500;600;700;800&display=swap"/>
<link rel="stylesheet" href="/css/styles.css?v={{.CSSHash}}"/>
{{.FBPixel}}{{.Analytics}}
<script defer src="https://umami.dalang.io/script.js" data-website-id="` + umamiSite + `"></script>
{{if .HTMX}}<script src="https://unpkg.com/htmx.org@1.9.12/dist/htmx.min.js" defer></script>
{{end}}<script type="application/ld+json">{{.JSONLD}}</script>
</head>
<body class="nsk-app min-h-screen overflow-x-hidden">
<a href="#main" class="sr-only focus:not-sr-only focus:fixed focus:left-2 focus:top-2 focus:z-50 focus:bg-white focus:p-3 focus:font-bold">{{.C.Skip}}</a>

<header class="sticky top-0 z-40 border-b-2 border-black bg-white">
  <div class="mx-auto flex h-16 max-w-7xl items-center justify-between gap-3 px-4 sm:px-6">
    <a href="{{.HomeHref}}" class="flex items-center gap-2.5" aria-label="Netsekurity">
      <img src="/favicon.svg" width="34" height="34" alt="" class="h-[34px] w-[34px]"/>
      <span class="text-xl font-extrabold tracking-tight">Netsekurity</span>
    </a>
    <nav aria-label="Main" class="hidden items-center gap-0.5 xl:flex">
      {{range .Nav}}<a href="{{.Href}}" data-umami-event="{{.Event}}" class="nsk-nav-link whitespace-nowrap !min-h-[40px] !px-3 !text-sm"{{if eq .Key $.Key}} aria-current="page"{{end}}>{{.Label}}</a>
      {{end}}
    </nav>
    <div class="flex items-center gap-2">
      {{if .LangLinks}}<div class="hidden items-center gap-1 sm:flex" role="group" aria-label="{{.C.LangLabel}}">
        {{range .LangLinks}}<a href="{{.Href}}" lang="{{.Code}}" hreflang="{{.Code}}" class="nsk-nav-link !min-h-[36px] !px-2 !text-xs"{{if .Current}} aria-current="true"{{end}}>{{.Label}}</a>{{end}}
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
      {{range .LangLinks}}<a href="{{.Href}}" lang="{{.Code}}" hreflang="{{.Code}}" class="nsk-nav-link !min-h-[40px]"{{if .Current}} aria-current="true"{{end}}>{{.Label}}</a>{{end}}
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
          <img src="/favicon.svg" width="34" height="34" alt="" class="h-[34px] w-[34px]"/>
          <span class="text-xl font-extrabold tracking-tight">Netsekurity</span>
        </a>
        <p class="mt-3 max-w-xs text-sm text-gray-700">{{.C.Tagline}}</p>
        <p class="mt-3 max-w-xs text-sm text-gray-700">{{.C.Operator}}</p>
      </div>
      <nav aria-label="{{.C.ColProduct}}">
        <p class="nsk-nav-group">{{.C.ColProduct}}</p>
        <ul class="space-y-2 text-sm">
          <li><a class="underline" href="{{.HomeHref}}#how">{{.C.NavHow}}</a></li>
          <li><a class="underline" href="{{.HomeHref}}#pricing">{{.C.NavPricing}}</a></li>
          <li><a class="underline" href="/sample-report">{{.C.NavSample}}</a></li>
          <li><a class="underline" href="/product-scope">{{.C.LinkScope}}</a></li>
          <li><a class="underline" href="/docs">{{.C.NavDocs}}</a></li>
        </ul>
      </nav>
      <nav aria-label="{{.C.ColEnterprise}}">
        <p class="nsk-nav-group">{{.C.ColEnterprise}}</p>
        <ul class="space-y-2 text-sm">
          <li><a class="underline" href="{{.EntHref}}">{{.C.NavEnterprise}}</a></li>
          <li><a class="underline" href="{{.EntHref}}#procurement">{{.C.LinkProcurement}}</a></li>
          <li><a class="underline" href="{{.TrustHref}}">{{.C.NavTrust}}</a></li>
          <li><a class="underline" href="{{.GuideHref}}">{{.C.LinkGuide}}</a></li>
          <li><a class="underline" href="{{.ContactHref}}" data-umami-event="netsekurity_footer_kontak">{{.C.LinkContact}}</a></li>
        </ul>
      </nav>
      <nav aria-label="{{.C.ColResources}}">
        <p class="nsk-nav-group">{{.C.ColResources}}</p>
        <ul class="space-y-2 text-sm">
          <li><a class="underline" href="https://dalang.io/" rel="noopener">{{.C.LinkDalang}}</a></li>
          <li><a class="underline" href="/llms.txt">{{.C.LinkLLMS}}</a></li>
          <li><a class="underline" href="/.well-known/security.txt">{{.C.LinkSecurityTxt}}</a></li>
          <li><a class="underline" href="/sitemap.xml">{{.C.LinkSitemap}}</a></li>
          <li><a class="underline" href="{{.AuthHref}}" data-umami-event="netsekurity_footer_masuk" data-umami-event-tujuan="{{.AuthHref}}">{{.AuthLabel}}</a></li>
        </ul>
      </nav>
    </div>
    <div class="mt-10 flex flex-wrap items-center gap-x-6 gap-y-2 border-t-2 border-gray-200 pt-6 text-sm text-gray-600">
      <p>{{.C.Rights}}</p>
      <a class="underline" rel="noopener" href="https://dalang.io/privacy" data-umami-event="netsekurity_footer_privacy">{{.C.LinkPrivacy}}</a>
      <a class="underline" rel="noopener" href="https://dalang.io/terms-of-use" data-umami-event="netsekurity_footer_terms">{{.C.LinkTerms}}</a>
      <a class="underline" rel="noopener" href="https://dalang.io/security" data-umami-event="netsekurity_footer_security">{{.C.LinkTrustCenter}}</a>
    </div>
  </div>
</footer>
<script>
(function () {
  var b = document.getElementById('nsk-burger'), m = document.getElementById('nsk-mobile');
  if (!b || !m) return;
  b.addEventListener('click', function () {
    var open = m.classList.contains('hidden');
    m.classList.toggle('hidden', !open);
    b.setAttribute('aria-expanded', open ? 'true' : 'false');
    b.setAttribute('aria-label', open ? b.getAttribute('data-close') : b.getAttribute('data-open'));
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
{{range .}}<details class="nsk-card group">
  <summary class="flex cursor-pointer list-none items-center justify-between gap-4 px-5 py-4 text-lg font-extrabold marker:content-['']">
    <span>{{.Q}}</span><span aria-hidden="true" class="text-2xl leading-none transition-transform group-open:rotate-45">+</span>
  </summary>
  <div class="nsk-prose border-t-2 border-black px-5 py-4 text-gray-800">{{.A}}</div>
</details>
{{end}}</div>{{end}}

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
<p class="nsk-alert nsk-alert-warn mt-6 text-sm">{{.Note}}</p>{{end}}{{end}}

{{define "tiers"}}<div class="mt-8 grid gap-6 lg:grid-cols-3">
  {{range .C.Tiers}}<div class="nsk-card flex flex-col p-5">
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
    <p class="mt-4 text-sm text-gray-700"><a class="font-bold underline" data-umami-event="netsekurity_kontak_email" href="mailto:sales@dalang.io?subject=%5Bnetsekurity%5D%20enterprise">sales@dalang.io</a></p>
  </div>
</section>{{end}}

`

const siteHomeHTML = `{{define "body"}}{{with .Home}}
<section class="border-b-2 border-black bg-[#fffbeb]">
  <div class="mx-auto grid max-w-7xl items-center gap-10 px-4 py-12 sm:px-6 lg:grid-cols-[1.15fr_0.85fr] lg:py-20">
    <div>
      <span class="nsk-tag">{{.Eyebrow}}</span>
      <h1 class="mt-4 text-4xl font-extrabold leading-[1.1] tracking-tight sm:text-5xl">{{.H1}}</h1>
      <p class="mt-5 max-w-2xl text-lg text-gray-800">{{.Lead}}</p>
      <p class="mt-3 text-gray-700">{{.SubLead}}</p>
      <div class="mt-7 flex flex-wrap gap-3">
        <a href="/sample-report" data-umami-event="netsekurity_hero_sample" class="nsk-btn">{{.CTAPrimary}}</a>
        <a href="{{$.StartHref}}" data-umami-event="netsekurity_hero_cta" data-umami-event-tujuan="{{$.StartHref}}" class="nsk-btn nsk-btn-ghost">{{.CTASecondary}}</a>
      </div>
      <p class="mt-4"><a href="{{$.ContactHref}}" data-umami-event="netsekurity_hero_enterprise" class="font-bold underline">{{.CTATertiary}} →</a></p>
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
        <a href="/sample-report" class="mt-3 inline-block text-sm font-bold underline">{{.Card.Link}} →</a>
      </div>
    </aside>
  </div>
</section>

<section class="py-14 sm:py-16">
  <div class="mx-auto max-w-4xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.DefinitionH2}}</h2>
    <p class="mt-4 text-lg leading-relaxed text-gray-800">{{.Definition}}</p>
  </div>
</section>

<section id="audiences" class="border-t-2 border-black bg-[#fffbeb] py-14 sm:py-16">
  <div class="mx-auto max-w-7xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.AudiencesH2}}</h2>
    <div class="mt-8 grid gap-6 lg:grid-cols-3">
      {{range .Audiences}}<article class="nsk-card flex flex-col bg-white p-5">
        <p class="text-sm font-extrabold uppercase tracking-wide text-gray-700">{{.Role}}</p>
        <h3 class="mt-1 text-xl font-extrabold">{{.Tagline}}</h3>
        <ul class="mt-4 flex-1 space-y-3 text-gray-800">
          {{range .Bullets}}<li class="flex gap-2"><span class="mt-1 text-green-800">{{icon "check" 16}}</span><span>{{.}}</span></li>
          {{end}}
        </ul>
        <a href="{{$.EntHref}}#{{.Anchor}}" class="mt-5 font-bold underline">{{.LinkLabel}} →</a>
      </article>
      {{end}}
    </div>
  </div>
</section>

<section id="report" class="border-t-2 border-black py-14 sm:py-16">
  <div class="mx-auto max-w-7xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.ReportH2}}</h2>
    <p class="mt-3 max-w-3xl text-lg text-gray-700">{{.ReportLead}}</p>
    <ul class="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-5">
      {{range .Report}}<li class="nsk-card p-4"><p class="font-extrabold">{{.Title}}</p><p class="mt-1 text-sm text-gray-700">{{.Body}}</p></li>
      {{end}}
    </ul>
    <a href="/sample-report" class="mt-6 inline-block font-bold underline">{{.ReportLink}} →</a>
  </div>
</section>

<section id="how" class="border-t-2 border-black bg-[#fffbeb] py-14 sm:py-16">
  <div class="mx-auto max-w-7xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.HowH2}}</h2>
    <ol class="mt-8 grid gap-6 lg:grid-cols-4">
      {{range $i, $s := .Steps}}<li class="nsk-card bg-white p-5">
        <span class="flex h-10 w-10 items-center justify-center border-2 border-black bg-primary text-lg font-extrabold">{{inc $i}}</span>
        <h3 class="mt-3 text-lg font-extrabold">{{$s.Title}}</h3>
        <p class="mt-2 text-sm text-gray-800">{{$s.Body}}</p>
      </li>
      {{end}}
    </ol>
  </div>
</section>
{{end}}

<section id="coverage" class="border-t-2 border-black py-14 sm:py-16">
  <div class="mx-auto max-w-7xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.C.Coverage.H2}}</h2>
    <p class="mt-3 max-w-3xl text-lg text-gray-700">{{.C.Coverage.Lead}}</p>
    {{template "coverage" .}}
  </div>
</section>

<section id="tiers" class="border-t-2 border-black bg-[#fffbeb] py-14 sm:py-16">
  <div class="mx-auto max-w-7xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.C.TiersH2}}</h2>
    <p class="mt-3 max-w-3xl text-lg text-gray-700">{{.C.TiersLead}}</p>
    {{template "tiers" .}}
    <p class="mt-6"><a href="{{.GuideHref}}" class="font-bold underline">{{.C.LinkGuide}} →</a></p>
  </div>
</section>
{{with .Home}}
<section id="cicd" class="border-t-2 border-black py-14 sm:py-16">
  <div class="mx-auto grid max-w-7xl gap-8 px-4 sm:px-6 lg:grid-cols-2">
    <div>
      <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.CICDH2}}</h2>
      <p class="mt-3 text-lg text-gray-700">{{.CICDLead}}</p>
      <ul class="mt-5 space-y-3 text-gray-800">
        {{range .CICD}}<li class="flex gap-2"><span class="mt-1 text-green-800">{{icon "check" 16}}</span><span>{{.}}</span></li>
        {{end}}
      </ul>
      <p class="nsk-alert nsk-alert-warn mt-5 text-sm">{{.CICDNote}}</p>
      <a href="/docs" class="mt-4 inline-block font-bold underline">{{.CICDLink}} →</a>
    </div>
    <div class="nsk-card self-start">
      <p class="nsk-card-head nsk-mono text-sm">POST /api/v1/pentests</p>
      <pre class="nsk-mono overflow-x-auto whitespace-pre-wrap p-5 text-sm leading-relaxed">curl -s -X POST https://netsekurity.com/api/v1/pentests \
  -H "X-API-Token: $NETSEKURITY_API_TOKEN" \
  -d "domain=app.example.com&amp;mode=standard"</pre>
    </div>
  </div>
</section>

<section id="pricing" class="border-t-2 border-black bg-[#fffbeb] py-14 sm:py-16">
  <div class="mx-auto max-w-7xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.PricingH2}}</h2>
    <p class="mt-3 max-w-3xl text-lg text-gray-700">{{.PricingLead}}</p>
    <div class="nsk-card mt-8 overflow-x-auto bg-white">
      <table class="w-full min-w-[520px] text-left">
        <thead><tr class="border-b-2 border-black">{{range .PricingCols}}<th scope="col" class="px-4 py-3 text-sm font-extrabold">{{.}}</th>{{end}}</tr></thead>
        <tbody>
          {{range $.Pricing.Rows}}<tr class="border-b-2 border-gray-200 last:border-b-0">
            <th scope="row" class="px-4 py-3 font-extrabold">{{.Name}}</th>
            <td class="px-4 py-3">{{.Credits}}</td>
            <td class="px-4 py-3 text-lg font-extrabold">{{.Price}}</td>
            <td class="px-4 py-3 text-gray-700">{{.PerScan}}</td>
            <td class="px-4 py-3 text-right"><a href="/checkout?package={{.ID}}" data-umami-event="netsekurity_pricing_buy" data-umami-event-paket="{{.ID}}" class="nsk-btn nsk-btn-sm" aria-label="{{$.C.Buy}} {{.Name}}">{{$.C.Buy}}</a></td>
          </tr>
          {{end}}
        </tbody>
      </table>
    </div>
    <p class="mt-4 max-w-3xl text-gray-800">{{.PricingNote}}</p>
  </div>
</section>

<section id="trust" class="border-t-2 border-black py-14 sm:py-16">
  <div class="mx-auto max-w-7xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.TrustH2}}</h2>
    <p class="mt-3 max-w-3xl text-lg text-gray-700">{{.TrustLead}}</p>
    <ul class="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
      {{range .Trust}}<li class="nsk-card p-5"><span class="text-gray-900">{{icon "lock" 20}}</span><p class="mt-2 font-extrabold">{{.Title}}</p><p class="mt-1 text-sm text-gray-700">{{.Body}}</p></li>
      {{end}}
    </ul>
    <a href="{{$.TrustHref}}" class="mt-6 inline-block font-bold underline">{{.TrustLink}} →</a>
  </div>
</section>

<section id="faq" class="border-t-2 border-black bg-[#fffbeb] py-14 sm:py-16">
  <div class="mx-auto max-w-4xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.FAQH2}}</h2>
    {{template "faq" .FAQ}}
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
    <ul class="mt-6 flex flex-wrap gap-2">
      <li><a class="nsk-btn nsk-btn-ghost nsk-btn-sm" href="#board">{{index .Jump 0}}</a></li>
      <li><a class="nsk-btn nsk-btn-ghost nsk-btn-sm" href="#engineering">{{index .Jump 1}}</a></li>
      <li><a class="nsk-btn nsk-btn-ghost nsk-btn-sm" href="#procurement">{{index .Jump 2}}</a></li>
      <li><a class="nsk-btn nsk-btn-ghost nsk-btn-sm" href="#escalation">{{index .Jump 3}}</a></li>
      <li><a class="nsk-btn nsk-btn-ghost nsk-btn-sm" href="#contact">{{index .Jump 4}}</a></li>
    </ul>
  </div>
</section>

<section id="board" class="py-14 sm:py-16">
  <div class="mx-auto max-w-7xl px-4 sm:px-6">
    <h2 class="max-w-4xl text-2xl font-extrabold tracking-tight sm:text-3xl">{{.BoardH2}}</h2>
    <p class="mt-3 max-w-3xl text-lg text-gray-700">{{.BoardIntro}}</p>
    <dl class="mt-8 grid gap-4 md:grid-cols-2">
      {{range .Board}}<div class="nsk-card p-5"><dt class="text-lg font-extrabold">{{.Title}}</dt><dd class="mt-2 text-gray-800">{{.Body}}</dd></div>
      {{end}}
    </dl>
    <p class="nsk-alert nsk-alert-info mt-6">{{.BoardCallout}}</p>
  </div>
</section>

<section id="engineering" class="border-t-2 border-black bg-[#fffbeb] py-14 sm:py-16">
  <div class="mx-auto max-w-7xl px-4 sm:px-6">
    <h2 class="max-w-4xl text-2xl font-extrabold tracking-tight sm:text-3xl">{{.EngH2}}</h2>
    <p class="mt-3 max-w-3xl text-lg text-gray-700">{{.EngIntro}}</p>
    <div class="mt-8 grid gap-6 lg:grid-cols-2">
      {{range .Eng}}<article class="nsk-card bg-white p-5">
        <h3 class="text-lg font-extrabold">{{.Title}}</h3>
        {{range .Paragraphs}}<p class="mt-2 text-gray-800">{{.}}</p>{{end}}
        {{if .Bullets}}<ul class="mt-3 space-y-1.5 text-sm text-gray-800">{{range .Bullets}}<li class="flex gap-2"><span class="mt-0.5 text-green-800">{{icon "check" 16}}</span><span>{{.}}</span></li>{{end}}</ul>{{end}}
        {{if .Code}}<pre class="nsk-code mt-3 block overflow-x-auto whitespace-pre-wrap text-xs">{{.Code}}</pre>{{end}}
      </article>
      {{end}}
    </div>
  </div>
</section>
{{end}}

{{with .Ent}}
<section id="procurement" class="border-t-2 border-black py-14 sm:py-16">
  <div class="mx-auto max-w-7xl px-4 sm:px-6">
    <h2 class="max-w-4xl text-2xl font-extrabold tracking-tight sm:text-3xl">{{.ProcH2}}</h2>
    <p class="mt-3 max-w-3xl text-lg text-gray-700">{{.ProcIntro}}</p>
    <dl class="nsk-card mt-8 divide-y-2 divide-gray-200">
      {{range .ProcFacts}}<div class="grid gap-1 px-5 py-4 md:grid-cols-[14rem_1fr] md:gap-6"><dt class="font-extrabold">{{.Title}}</dt><dd class="text-gray-800">{{.Body}}</dd></div>
      {{end}}
    </dl>
    <p class="mt-4 text-sm text-gray-700">{{.ProcNote}}</p>
  </div>
</section>
{{end}}

<section id="escalation" class="border-t-2 border-black bg-[#fffbeb] py-14 sm:py-16">
  <div class="mx-auto max-w-7xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.C.TiersH2}}</h2>
    <p class="mt-3 max-w-3xl text-lg text-gray-700">{{.C.TiersLead}}</p>
    {{template "tiers" .}}
  </div>
</section>

<section class="border-t-2 border-black py-14 sm:py-16">
  <div class="mx-auto max-w-7xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.C.Coverage.H2}}</h2>
    <p class="mt-3 max-w-3xl text-lg text-gray-700">{{.C.Coverage.Lead}}</p>
    {{template "coverage" .}}
  </div>
</section>

{{with .Ent}}
<section class="border-t-2 border-black bg-[#fffbeb] py-14 sm:py-16">
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
      <li><a class="nsk-btn nsk-btn-ghost nsk-btn-sm" href="#responsible">{{index .Jump 0}}</a></li>
      <li><a class="nsk-btn nsk-btn-ghost nsk-btn-sm" href="#data">{{index .Jump 1}}</a></li>
      <li><a class="nsk-btn nsk-btn-ghost nsk-btn-sm" href="#access">{{index .Jump 2}}</a></li>
      <li><a class="nsk-btn nsk-btn-ghost nsk-btn-sm" href="#disclosure">{{index .Jump 3}}</a></li>
      <li><a class="nsk-btn nsk-btn-ghost nsk-btn-sm" href="#operator">{{index .Jump 4}}</a></li>
    </ul>
  </div>
</section>

<section id="responsible" class="py-14 sm:py-16">
  <div class="mx-auto max-w-7xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.RespH2}}</h2>
    <p class="mt-3 max-w-3xl text-lg text-gray-700">{{.RespIntro}}</p>
    <dl class="mt-8 grid gap-4 md:grid-cols-2">
      {{range .Resp}}<div class="nsk-card p-5"><dt class="flex items-center gap-2 text-lg font-extrabold">{{icon "lock" 18}}{{.Title}}</dt><dd class="mt-2 text-gray-800">{{.Body}}</dd></div>
      {{end}}
    </dl>
  </div>
</section>

<section id="data" class="border-t-2 border-black bg-[#fffbeb] py-14 sm:py-16">
  <div class="mx-auto max-w-7xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.DataH2}}</h2>
    <p class="mt-3 max-w-3xl text-lg text-gray-700">{{.DataIntro}}</p>
    <div class="nsk-card mt-8 overflow-x-auto bg-white">
      <table class="w-full min-w-[480px] text-left">
        <thead><tr class="border-b-2 border-black">{{range .DataCols}}<th scope="col" class="px-4 py-3 text-sm font-extrabold">{{.}}</th>{{end}}</tr></thead>
        <tbody>
          {{range .Data}}<tr class="border-b-2 border-gray-200 last:border-b-0"><th scope="row" class="w-44 px-4 py-3 align-top font-extrabold">{{.Item}}</th><td class="px-4 py-3 text-gray-800">{{.Detail}}</td></tr>
          {{end}}
        </tbody>
      </table>
    </div>
  </div>
</section>

<section id="access" class="border-t-2 border-black py-14 sm:py-16">
  <div class="mx-auto max-w-7xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.AccessH2}}</h2>
    <dl class="mt-8 grid gap-4 md:grid-cols-3">
      {{range .Access}}<div class="nsk-card p-5"><dt class="text-lg font-extrabold">{{.Title}}</dt><dd class="mt-2 text-gray-800">{{.Body}}</dd></div>
      {{end}}
    </dl>
  </div>
</section>
{{end}}

<section class="border-t-2 border-black bg-[#fffbeb] py-14 sm:py-16">
  <div class="mx-auto max-w-7xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.C.Coverage.H2}}</h2>
    <p class="mt-3 max-w-3xl text-lg text-gray-700">{{.C.Coverage.Lead}}</p>
    {{template "coverage" .}}
  </div>
</section>

{{with .Trust}}
<section id="disclosure" class="border-t-2 border-black py-14 sm:py-16">
  <div class="mx-auto max-w-4xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.DiscH2}}</h2>
    <p class="mt-3 text-lg text-gray-800">{{.DiscBody}}</p>
    <ul class="mt-4 space-y-2 text-gray-800">{{range .Disc}}<li class="flex gap-2"><span class="mt-1 text-green-800">{{icon "check" 16}}</span><span>{{.}}</span></li>{{end}}</ul>
  </div>
</section>

<section id="operator" class="border-t-2 border-black bg-[#fffbeb] py-14 sm:py-16">
  <div class="mx-auto max-w-7xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.OpH2}}</h2>
    {{range .OpBody}}<p class="mt-3 max-w-3xl text-lg text-gray-800">{{.}}</p>{{end}}
    <dl class="nsk-card mt-8 divide-y-2 divide-gray-200 bg-white">
      {{range .OpFacts}}<div class="grid gap-1 px-5 py-3 md:grid-cols-[12rem_1fr] md:gap-6"><dt class="font-extrabold">{{.Title}}</dt><dd class="text-gray-800">{{.Body}}</dd></div>
      {{end}}
    </dl>
    <p class="mt-4 text-sm text-gray-700">{{.OpSource}}</p>
    <p class="mt-6 max-w-3xl text-gray-800">{{.PolicyLead}}</p>
    <ul class="mt-3 flex flex-wrap gap-3">
      <li><a class="nsk-btn nsk-btn-ghost nsk-btn-sm" rel="noopener" href="https://dalang.io/privacy">{{$.C.LinkPrivacy}}</a></li>
      <li><a class="nsk-btn nsk-btn-ghost nsk-btn-sm" rel="noopener" href="https://dalang.io/terms-of-use">{{$.C.LinkTerms}}</a></li>
      <li><a class="nsk-btn nsk-btn-ghost nsk-btn-sm" rel="noopener" href="https://dalang.io/security">{{$.C.LinkTrustCenter}}</a></li>
    </ul>
  </div>
</section>

<section class="border-t-2 border-black py-14 sm:py-16">
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

<section class="py-14 sm:py-16">
  <div class="mx-auto max-w-7xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.TableH2}}</h2>
    <div class="nsk-card mt-8 overflow-x-auto">
      <table class="w-full min-w-[640px] text-left">
        <thead><tr class="border-b-2 border-black bg-[#fffbeb]">{{range .Cols}}<th scope="col" class="px-4 py-3 text-sm font-extrabold">{{.}}</th>{{end}}</tr></thead>
        <tbody>
          {{range .Rows}}<tr class="border-b-2 border-gray-200 last:border-b-0"><th scope="row" class="w-44 px-4 py-3 align-top font-extrabold">{{.Label}}</th>{{range .Cells}}<td class="px-4 py-3 align-top text-gray-800">{{.}}</td>{{end}}</tr>
          {{end}}
        </tbody>
      </table>
    </div>
  </div>
</section>

<section class="border-t-2 border-black bg-[#fffbeb] py-14 sm:py-16">
  <div class="mx-auto grid max-w-7xl gap-6 px-4 sm:px-6 lg:grid-cols-2">
    <div class="nsk-card bg-white p-5">
      <h2 class="flex items-center gap-2 text-xl font-extrabold">{{icon "check" 20}}{{.GoodH2}}</h2>
      <ul class="mt-4 space-y-2 text-gray-800">{{range .Good}}<li class="flex gap-2"><span class="mt-0.5 text-green-800">{{icon "check" 16}}</span><span>{{.}}</span></li>{{end}}</ul>
    </div>
    <div class="nsk-card bg-white p-5">
      <h2 class="flex items-center gap-2 text-xl font-extrabold">{{icon "x" 20}}{{.CannotH2}}</h2>
      <ul class="mt-4 space-y-2 text-gray-800">{{range .Cannot}}<li class="flex gap-2"><span class="mt-0.5 text-red-800">{{icon "x" 16}}</span><span>{{.}}</span></li>{{end}}</ul>
    </div>
  </div>
</section>

<section class="border-t-2 border-black py-14 sm:py-16">
  <div class="mx-auto max-w-4xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.ChooseH2}}</h2>
    <ol class="mt-8 space-y-4">
      {{range $i, $c := .Choose}}<li class="nsk-card flex gap-4 p-5">
        <span class="flex h-10 w-10 shrink-0 items-center justify-center border-2 border-black bg-primary text-lg font-extrabold">{{inc $i}}</span>
        <div><p class="text-lg font-extrabold">{{$c.Title}}</p><p class="mt-1 text-gray-800">{{$c.Body}}</p></div>
      </li>
      {{end}}
    </ol>
  </div>
</section>

<section class="border-t-2 border-black bg-[#fffbeb] py-14 sm:py-16">
  <div class="mx-auto max-w-4xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.GlossH2}}</h2>
    <dl class="mt-8 space-y-3">
      {{range .Gloss}}<div class="nsk-card bg-white p-4"><dt class="font-extrabold">{{.Title}}</dt><dd class="mt-1 text-gray-800">{{.Body}}</dd></div>
      {{end}}
    </dl>
  </div>
</section>

<section class="border-t-2 border-black py-14 sm:py-16">
  <div class="mx-auto max-w-4xl px-4 sm:px-6">
    <h2 class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.FAQH2}}</h2>
    {{template "faq" .FAQ}}
    <p class="mt-8"><a href="{{.CTALink}}" class="nsk-btn">{{.CTA}}</a></p>
  </div>
</section>
{{end}}
{{end}}
`
