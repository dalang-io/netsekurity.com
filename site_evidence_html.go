package main

// Bodies of the three evidence pages: the sample report, the docs and the product
// scope. They are registry pages like the rest (see sitePages), so each exists in
// both languages and shares the layout, the nonce policy and the sitemap.

// docsStartExample is the one request the quickstart shows. The CI examples below
// it send the same request.
const docsStartExample = "curl -s -X POST https://netsekurity.com/api/v1/assessments \\\n" +
	"  -H \"X-API-Token: $NETSEKURITY_API_TOKEN\" \\\n" +
	"  -d \"domain=app.example.com\""

const siteSampleHTML = `{{define "body"}}{{with .Sample}}
<section class="border-b-2 border-black bg-[#fffbeb]">
  <div class="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:py-16">
    <p class="hidden print:block text-sm font-bold">Netsekurity · {{.Eyebrow}} · netsekurity.com · Dalang Pte. Ltd.</p>
    <div class="nsk-noprint">{{template "crumb" $}}</div>
    <span class="nsk-tag">{{.Eyebrow}}</span>
    <h1 class="mt-4 max-w-4xl text-4xl font-extrabold leading-[1.1] tracking-tight sm:text-5xl">{{.H1}}</h1>
    <p class="mt-5 max-w-3xl text-lg text-gray-800">{{.Lead}}</p>
    <div class="nsk-noprint mt-6 flex flex-wrap items-center gap-x-5 gap-y-2">
      <a href="{{$.StartHref}}" data-umami-event="netsekurity_sample_start" data-umami-event-tujuan="{{$.StartHref}}" class="nsk-btn">{{.CTA}}</a>
      <a href="{{$.HomeHref}}#pricing" class="nsk-textlink">{{.CTAAll}}<span aria-hidden="true"> →</span></a>
      <a href="{{$.SamplePDF}}" class="nsk-textlink" data-umami-event="netsekurity_sample_pdf">{{icon "download" 16}}{{.PDFLabel}}</a>
      <button type="button" class="nsk-btn nsk-btn-ghost nsk-btn-sm" data-print="page">{{icon "printer" 16}}{{.PrintLabel}}</button>
    </div>
  </div>
</section>

<div class="mx-auto max-w-7xl space-y-8 px-4 py-12 sm:px-6">

  <section class="nsk-card" aria-labelledby="s1">
    <h2 id="s1" class="nsk-card-head">{{.DocH2}}</h2>
    <dl class="grid divide-y-2 divide-gray-200 text-sm sm:grid-cols-2 sm:divide-y-0">
      {{range .Doc}}<div class="flex flex-wrap justify-between gap-x-4 gap-y-1 px-5 py-3"><dt class="font-extrabold">{{.Title}}</dt><dd{{if eq .Title "Target"}} class="nsk-mono"{{end}}>{{.Body}}</dd></div>
      {{end}}
    </dl>
  </section>

  <section class="nsk-card" aria-labelledby="s2">
    <h2 id="s2" class="nsk-card-head">{{.SumH2}}</h2>
    <div class="p-5">
      <p class="flex flex-wrap gap-2">{{range .Badges}}<span class="nsk-badge nsk-badge-{{.Class}}">{{.Label}}</span>{{end}}</p>
      <p class="mt-4 text-gray-800"><strong class="font-extrabold">{{.Posture}}</strong> {{.PostureBody}}</p>
    </div>
  </section>

  <section class="nsk-card" aria-labelledby="s3">
    <h2 id="s3" class="nsk-card-head">{{.RiskH2}}</h2>
    <table class="w-full text-left text-sm">
      <thead><tr class="border-b-2 border-black bg-[#fffbeb]">
        <th scope="col" class="px-3 py-3 sm:px-4">{{index .RiskCols 0}}</th>
        <th scope="col" class="px-3 py-3 sm:px-4">{{index .RiskCols 1}}</th>
        <th scope="col" class="px-3 py-3 sm:px-4">{{index .RiskCols 2}}</th>
        <th scope="col" class="hidden px-4 py-3 md:table-cell">{{index .RiskCols 3}}</th>
        <th scope="col" class="hidden px-4 py-3 md:table-cell">{{index .RiskCols 4}}</th>
        <th scope="col" class="hidden px-4 py-3 md:table-cell">{{index .RiskCols 5}}</th>
      </tr></thead>
      <tbody>
        {{range .Risks}}<tr class="border-b-2 border-gray-200 align-top last:border-b-0">
          <td class="whitespace-nowrap px-3 py-3 font-extrabold sm:px-4">{{.ID}}</td>
          <td class="px-3 py-3 sm:px-4">{{.Finding}}{{if ne .CVSS "—"}}<span class="mt-1 block text-xs text-gray-700 md:hidden">{{index $.Sample.RiskCols 3}} {{.CVSS}} · {{.CWE}} · {{index $.Sample.RiskCols 5}} {{.OWASP}}</span>{{end}}</td>
          <td class="px-3 py-3 sm:px-4"><span class="nsk-badge nsk-badge-{{.Class}}">{{.Severity}}</span></td>
          <td class="hidden px-4 py-3 md:table-cell">{{.CVSS}}</td>
          <td class="hidden px-4 py-3 md:table-cell">{{.CWE}}</td>
          <td class="hidden px-4 py-3 md:table-cell">{{.OWASP}}</td>
        </tr>
        {{end}}
      </tbody>
    </table>
  </section>

  <section class="nsk-card" aria-labelledby="s4">
    <h2 id="s4" class="nsk-card-head">{{.FindH2}}</h2>
    <div class="divide-y-2 divide-gray-200">
      {{range .Findings}}<article class="space-y-2 p-5">
        <h3 class="text-lg font-extrabold">{{.Title}} <span class="nsk-badge nsk-badge-{{.Class}} ml-2">{{.Badge}}</span></h3>
        <dl class="space-y-2 text-sm text-gray-800">
          {{range .Rows}}<div><dt class="inline font-extrabold">{{.Label}}: </dt><dd class="inline{{if .Mono}} nsk-mono break-all{{end}}">{{.Body}}</dd></div>
          {{end}}
        </dl>
      </article>
      {{end}}
      <article class="p-5">
        <h3 class="text-lg font-extrabold">{{.MoreSummary}}</h3>
        <div class="mt-3 space-y-2 text-sm text-gray-800">
          {{range .More}}<p>{{.}}</p>
          {{end}}
        </div>
      </article>
    </div>
  </section>

  <section class="nsk-card" aria-labelledby="s5">
    <h2 id="s5" class="nsk-card-head">{{.PlanH2}}</h2>
    <div class="grid gap-0 divide-y-2 divide-gray-200 md:grid-cols-3 md:divide-x-2 md:divide-y-0">
      {{range .Plan}}<div class="p-5"><p class="text-sm font-extrabold uppercase tracking-wide text-gray-700">{{.Title}}</p><p class="mt-2 text-gray-800">{{.Body}}</p></div>
      {{end}}
    </div>
  </section>

  <section class="nsk-note" aria-labelledby="s6">
    <h2 id="s6" class="text-lg font-extrabold">{{.LimitsH2}}</h2>
    <p class="mt-2">{{.LimitsBody}}</p>
  </section>

  <p class="nsk-noprint flex flex-wrap items-center gap-x-5 gap-y-2"><a href="{{$.StartHref}}" class="nsk-btn">{{.CTA}}</a><a href="{{$.HomeHref}}#pricing" class="nsk-textlink">{{.CTAAll}}<span aria-hidden="true"> →</span></a></p>
</div>
{{end}}
{{end}}
`

const siteDocsHTML = `{{define "body"}}{{with .Docs}}
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

<div class="nsk-fit mx-auto max-w-7xl space-y-8 px-4 py-12 sm:px-6">

  <section id="quickstart" class="nsk-card" aria-labelledby="h-quick">
    <h2 id="h-quick" class="nsk-card-head">{{.QuickH2}}</h2>
    <div class="p-5">
      <ol class="nsk-prose list-decimal space-y-2 pl-5 text-gray-800">
        {{range .Quick}}<li>{{.}}</li>
        {{end}}
      </ol>
      <div class="mt-4">{{template "codebox" (codeArgs $.C $.DocsStart)}}</div>
      <p class="mt-2 text-sm text-gray-700">{{.QuickNote}}</p>
    </div>
  </section>

  <section id="api" class="nsk-card" aria-labelledby="h-api">
    <h2 id="h-api" class="nsk-card-head">{{.APIH2}}</h2>
    <div class="space-y-8 p-5">
      <div>
        <h3 class="nsk-mono break-all text-base font-extrabold">{{.StartH3}}</h3>
        <p class="mt-1 text-gray-800">{{.StartLead}}</p>
        <div class="mt-3 border-2 border-black">
          <table class="w-full text-left text-sm">
            <thead><tr class="border-b-2 border-black bg-[#fffbeb]">
              <th scope="col" class="px-3 py-2 sm:px-4">{{index .ParamCols 0}}</th>
              <th scope="col" class="hidden px-4 py-2 md:table-cell">{{index .ParamCols 1}}</th>
              <th scope="col" class="px-3 py-2 sm:px-4">{{index .ParamCols 2}}</th>
            </tr></thead>
            <tbody>
              {{range .Params}}<tr class="border-b-2 border-gray-200 align-top last:border-b-0">
                <td class="nsk-mono px-3 py-2 sm:px-4">{{.Name}}<span class="block text-xs text-gray-700 md:hidden">{{.Type}}</span></td>
                <td class="hidden px-4 py-2 md:table-cell">{{.Type}}</td>
                <td class="px-3 py-2 sm:px-4">{{.Desc}}</td>
              </tr>
              {{end}}
            </tbody>
          </table>
        </div>
        <div class="mt-3">{{template "codebox" (codeArgs $.C .StartResp)}}</div>
      </div>
      <div id="result">
        <h3 class="nsk-mono break-all text-base font-extrabold">{{.ResultH3}}</h3>
        <p class="mt-1 text-gray-800">{{.ResultLead}}</p>
        <p class="mt-1 text-sm text-gray-700">{{.ResultParam}}</p>
        <div class="mt-3">{{template "codebox" (codeArgs $.C .ResultResp)}}</div>
        <dl class="mt-4 grid gap-2 sm:grid-cols-2">
          {{range .Statuses}}<div class="nsk-card-flat p-3"><dt class="nsk-mono font-extrabold">{{.Title}}</dt><dd class="text-sm text-gray-800">{{.Body}}</dd></div>
          {{end}}
        </dl>
        <p class="nsk-prose mt-3 text-sm text-gray-800">{{.ResultNote}}</p>
      </div>
    </div>
  </section>

  <section id="gate" class="nsk-card" aria-labelledby="h-gate">
    <h2 id="h-gate" class="nsk-card-head">{{.GateH3}}</h2>
    <div class="p-5">
      <p class="text-gray-800">{{.GateLead}}</p>
      <div class="mt-3">{{template "codebox" (codeArgs $.C .GateCode)}}</div>
      <p class="nsk-note mt-4 text-sm">{{.GateNote}}</p>
    </div>
  </section>

  <section id="ci" aria-labelledby="h-ci">
    <h2 id="h-ci" class="text-2xl font-extrabold tracking-tight sm:text-3xl">{{.CIH2}}</h2>
    <p class="mt-2 max-w-3xl text-gray-700">{{.CILead}}</p>
    <div class="mt-6 space-y-6">
      {{range $i, $e := $.Examples}}<article class="nsk-card-flat" aria-labelledby="ci-{{$i}}">
        <h3 id="ci-{{$i}}" class="nsk-card-head">{{$e.Title}}</h3>
        <div class="p-5">
          {{template "codebox" (codeArgs $.C $e.Code)}}
          <p class="nsk-prose mt-3 text-sm text-gray-700">{{$e.Note}}</p>
        </div>
      </article>
      {{end}}
    </div>
  </section>

  <section id="notes" class="nsk-card" aria-labelledby="h-notes">
    <h2 id="h-notes" class="nsk-card-head">{{.NotesH2}}</h2>
    <ul class="nsk-prose list-disc space-y-2 p-5 pl-10 text-gray-800">
      {{range .Notes}}<li>{{.}}</li>
      {{end}}
    </ul>
  </section>
</div>
{{end}}
{{end}}
`

const siteScopeHTML = `{{define "body"}}
<div class="mx-auto max-w-4xl px-4 pt-8 sm:px-6">{{template "crumb" .}}</div>
<article class="nsk-doc mx-auto max-w-4xl px-4 pb-12 sm:px-6">{{.Doc}}</article>
{{end}}
`
