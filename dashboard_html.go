package main

// Templates for the customer dashboard (see dashboard.go for the page map).
// dashboardHTML is the concatenation of the shell and every page, parsed once
// into `tmpl`. Written in the light theme shared with dalang.io/dashboard; the
// component classes (nsk-card, nsk-btn, nsk-badge …) live in assets/input.css.
const dashboardHTML = dashShellHTML + dashHomeHTML + dashWebsitesHTML + dashScansHTML + dashCreditsHTML + dashAutomationHTML

// dashShellHTML is the frame every page sits in: top bar, sidebar, flash area.
// A page opens it with {{template "shell_open" shell . "<nav key>" "<title>"}}
// and closes it with {{template "shell_close"}}.
const dashShellHTML = `{{define "shell_open"}}<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
<title>{{.Title}} — Netsekurity</title>
<meta name="robots" content="noindex, nofollow"/>
<link rel="preconnect" href="https://fonts.googleapis.com"/>
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin/>
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;500;600;700;800&display=swap"/>
<link rel="stylesheet" href="/css/styles.css?v={{cssHash}}"/>
<link rel="icon" type="image/svg+xml" href="/favicon.svg" />{{fbpixel}}
<script src="https://unpkg.com/htmx.org@1.9.12/dist/htmx.min.js" defer></script>
</head>
<body class="nsk-app min-h-screen overflow-x-hidden">
<a href="#main" class="sr-only focus:not-sr-only focus:fixed focus:left-2 focus:top-2 focus:z-50 focus:bg-white focus:p-3 focus:font-bold">Skip to content</a>

<header class="fixed inset-x-0 top-0 z-40 h-16 border-b-2 border-black bg-white">
  <div class="flex h-full items-center justify-between gap-3 px-3 sm:px-5">
    <div class="flex items-center gap-2">
      <button type="button" id="nsk-menu-btn" class="nsk-btn nsk-btn-ghost nsk-btn-sm md:hidden" aria-label="Open menu" aria-controls="nsk-sidebar" aria-expanded="false" onclick="nskMenu()">{{icon "menu" 20}}</button>
      <a href="/dashboard" class="flex items-center gap-2.5" aria-label="Netsekurity home">
        <img src="/favicon.svg" width="34" height="34" alt="" class="h-[34px] w-[34px]"/>
        <span class="text-xl font-extrabold tracking-tight">Netsekurity</span>
        <span class="hidden text-xs font-medium text-gray-600 sm:inline">by Dalang.io</span>
      </a>
    </div>
    <div class="flex items-center gap-2 sm:gap-3">
      <a href="/dashboard/credits" class="nsk-btn nsk-btn-ghost nsk-btn-sm" title="Your credit balance">{{icon "card" 16}}<span>Credits: <span id="credit-balance" class="font-extrabold">{{credits .Balance}}</span></span></a>
      <a href="/dashboard/credits" class="nsk-btn nsk-btn-sm hidden sm:inline-flex">Buy credits</a>
    </div>
  </div>
</header>

<div id="nsk-overlay" class="fixed inset-0 top-16 z-30 hidden bg-black/40 md:hidden" onclick="nskMenu()"></div>
<aside id="nsk-sidebar" aria-label="Dashboard menu" class="fixed bottom-0 left-0 top-16 z-30 hidden w-64 flex-col overflow-y-auto border-r-2 border-black bg-white p-4 md:flex">
  <nav class="flex-1" aria-label="Main">
    <p class="nsk-nav-group">Menu</p>
    <ul class="space-y-1.5">
      {{range nav}}
      <li><a href="{{.Href}}" class="nsk-nav-link"{{if eq .Key $.Active}} aria-current="page"{{end}}>{{icon .Icon 18}}<span>{{.Label}}</span></a></li>
      {{end}}
    </ul>

    <p class="nsk-nav-group mt-6">Help</p>
    <ul class="space-y-1.5">
      <li><a href="/sample-report" target="_blank" rel="noopener" class="nsk-nav-link">{{icon "file" 18}}<span>See a sample report</span></a></li>
      <li><a href="/product-scope" target="_blank" rel="noopener" class="nsk-nav-link">{{icon "help" 18}}<span>What we check</span></a></li>
      <li><a href="/#contact" class="nsk-nav-link">{{icon "help" 18}}<span>Contact us</span></a></li>
    </ul>

    {{if .IsAdmin}}<a href="/su" class="nsk-btn nsk-btn-dark nsk-btn-block mt-6">{{icon "lock" 16}}Super Admin</a>{{end}}
  </nav>

  <div class="mt-6 border-t-2 border-black pt-4">
    <div class="flex items-center gap-3">
      <span class="flex h-10 w-10 shrink-0 items-center justify-center border-2 border-black bg-primary text-lg font-extrabold" aria-hidden="true">{{.Initial}}</span>
      <div class="min-w-0">
        <p class="truncate text-sm font-extrabold">{{if .Name}}{{.Name}}{{else}}{{.Email}}{{end}}</p>
        <p class="truncate text-xs text-gray-600">{{.Email}}</p>
      </div>
    </div>
    <a href="/logout" class="nsk-btn nsk-btn-ghost nsk-btn-sm nsk-btn-block mt-3">{{icon "logout" 16}}Log out</a>
  </div>
</aside>

<main id="main" class="px-4 pb-16 pt-24 md:ml-64 md:px-8">
<div class="max-w-5xl">
  <div id="flash" class="nsk-flash fixed left-3 right-3 top-[72px] z-50 max-w-3xl empty:hidden md:left-72 md:right-6" aria-live="polite"></div>
{{end}}

{{define "shell_close"}}
</div>
</main>
<script>
function nskMenu() {
  var sb = document.getElementById('nsk-sidebar'), ov = document.getElementById('nsk-overlay'),
      btn = document.getElementById('nsk-menu-btn'), open = sb.classList.contains('hidden');
  sb.classList.toggle('hidden', !open); sb.classList.toggle('flex', open);
  ov.classList.toggle('hidden', !open);
  btn.setAttribute('aria-expanded', open ? 'true' : 'false');
}
function nskCopy(btn) {
  var text = btn.getAttribute('data-copy'), label = btn.querySelector('.nsk-copy-label');
  function done() {
    if (!label) return;
    var old = label.textContent; label.textContent = 'Copied!';
    setTimeout(function () { label.textContent = old; }, 1500);
  }
  if (navigator.clipboard && window.isSecureContext) {
    navigator.clipboard.writeText(text).then(done, function () {});
    return;
  }
  var ta = document.createElement('textarea');
  ta.value = text; ta.style.position = 'fixed'; ta.style.opacity = '0';
  document.body.appendChild(ta); ta.select();
  try { document.execCommand('copy'); done(); } catch (e) {}
  document.body.removeChild(ta);
}
// Times are stored in UTC; show them in the viewer's own timezone.
function nskLocalize() {
  var els = document.querySelectorAll('time.nsk-time[datetime]:not([data-local])');
  for (var i = 0; i < els.length; i++) {
    var d = new Date(els[i].getAttribute('datetime'));
    if (isNaN(d)) continue;
    els[i].title = els[i].textContent;
    els[i].textContent = d.toLocaleString(undefined, {day: 'numeric', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit'});
    els[i].setAttribute('data-local', '1');
  }
}
document.addEventListener('DOMContentLoaded', nskLocalize);
// HTMX does not swap error responses by default. Our handlers answer a refused
// request (bad address, nothing to remove, payment provider down) with a styled
// alert box and an error status, so swap those — and only those; a proxy's own
// error page never contains the marker.
document.addEventListener('htmx:beforeSwap', function (e) {
  if (e.detail.xhr.status >= 400 && (e.detail.serverResponse || '').indexOf('nsk-alert') !== -1) {
    e.detail.shouldSwap = true; e.detail.isError = false;
  }
});
// A result message floats over the page (so nothing underneath shifts) and can be dismissed.
function nskFlashDecorate() {
  var f = document.getElementById('flash'), a = f && f.firstElementChild;
  if (!a || a.querySelector('.nsk-flash-x')) return;
  var x = document.createElement('button');
  x.type = 'button'; x.className = 'nsk-flash-x'; x.setAttribute('aria-label', 'Dismiss message'); x.textContent = '×';
  x.onclick = function () { f.innerHTML = ''; };
  a.appendChild(x);
}
function nskFlash(kind, html) {
  var f = document.getElementById('flash');
  if (!f) return;
  f.innerHTML = '<div class="nsk-alert nsk-alert-' + kind + '" role="alert">' + html + '</div>';
  nskFlashDecorate();
}
// Anything else that goes wrong must still say so — silence reads as "it worked".
document.addEventListener('htmx:responseError', function (e) {
  var st = e.detail.xhr.status;
  if (st === 401) nskFlash('warn', '<strong>You have been signed out.</strong> <a href="/login" class="font-bold underline">Log in again</a> to continue.');
  else if (st === 429) nskFlash('warn', '<strong>Too many requests.</strong> Please wait a minute and try again.');
  else nskFlash('bad', '<strong>Something went wrong.</strong> Please try again. If it keeps happening, contact us.');
});
document.addEventListener('htmx:sendError', function () {
  nskFlash('bad', '<strong>Could not reach Netsekurity.</strong> Check your internet connection and try again.');
});
var nskFlashTimer;
document.addEventListener('htmx:afterSwap', function (e) {
  nskLocalize();
  // A result message clears itself, so it never sits over the page it describes.
  if (e.detail.target && e.detail.target.id === 'flash') {
    nskFlashDecorate();
    clearTimeout(nskFlashTimer);
    nskFlashTimer = setTimeout(function () { e.detail.target.innerHTML = ''; }, 15000);
  }
});
</script>
</body>
</html>{{end}}
`

const dashHomeHTML = `{{define "dashboard"}}{{template "shell_open" shell . "home" "Home"}}
  <div class="mb-6">
    <h1 class="text-2xl font-extrabold sm:text-3xl">Hi, {{.FirstName}}</h1>
    <p class="mt-1 max-w-2xl text-gray-700">Netsekurity checks your website from the outside, the way an attacker would, then explains what to fix in plain language.</p>
  </div>

  {{with .PendingPay}}
  <div class="nsk-alert nsk-alert-warn mb-6 flex flex-wrap items-center justify-between gap-3">
    <p><strong>Your payment isn't finished.</strong> {{if .Package}}{{.Package}} pack{{else}}Credit top-up{{end}} · ${{usd .AmountUSD}}. Your credits arrive as soon as it is paid.</p>
    <a href="{{.URL}}" target="_blank" rel="noopener" class="nsk-btn nsk-btn-sm">Finish payment</a>
  </div>
  {{end}}

  <div class="mb-6 grid gap-4 sm:grid-cols-3">
    <a href="/dashboard/credits" class="nsk-card block p-4 no-underline hover:bg-gray-50">
      <p class="text-sm font-bold text-gray-700">Credits left</p>
      <p class="mt-1 text-3xl font-extrabold">{{credits .Balance}}</p>
      <p class="mt-1 text-sm text-gray-700">{{if le .Balance 0.0}}Buy credits to run a scan →{{else}}1 credit = 1 scan of 1 website{{end}}</p>
    </a>
    <a href="/dashboard/websites" class="nsk-card block p-4 no-underline hover:bg-gray-50">
      <p class="text-sm font-bold text-gray-700">Websites</p>
      <p class="mt-1 text-3xl font-extrabold">{{.VerifiedCount}} <span class="text-base font-bold text-gray-700">verified</span></p>
      <p class="mt-1 text-sm text-gray-700">{{len .Domains}} added in total</p>
    </a>
    <a href="/dashboard/scans" class="nsk-card block p-4 no-underline hover:bg-gray-50">
      <p class="text-sm font-bold text-gray-700">Scans run</p>
      <p class="mt-1 text-3xl font-extrabold">{{.ScanCount}}</p>
      <p class="mt-1 text-sm text-gray-700">See your reports →</p>
    </a>
  </div>

  {{if .AllDone}}
  <section class="nsk-card mb-6">
    <h2 class="nsk-card-head">Scan a website</h2>
    <div class="p-4">
      <p class="mb-3 text-gray-700">Pick a website and we'll check it. A scan takes 10 minutes to 24 hours, depending on what it finds. You can leave this page while it runs.</p>
      {{template "scanpicker" .}}
    </div>
  </section>
  {{else}}
  <section class="nsk-card nsk-card-tint mb-6">
    <div class="nsk-card-head flex flex-wrap items-center justify-between gap-2">
      <h2>Get started</h2>
      <span class="text-sm font-bold text-gray-700">{{.StepsDone}} of 4 done</span>
    </div>
    <ol class="space-y-3 p-4">
      {{range .Steps}}
      <li class="flex gap-3 border-2 p-3 {{if .Current}}border-black bg-white shadow-[2px_2px_0_0_#000]{{else}}border-gray-300 bg-white{{end}}">
        <span class="flex h-9 w-9 shrink-0 items-center justify-center border-2 border-black font-extrabold {{if .Done}}bg-green-200{{else if .Current}}bg-primary{{else}}bg-white{{end}}">{{if .Done}}{{icon "check" 18}}<span class="sr-only">Done</span>{{else}}{{.N}}{{end}}</span>
        <div class="min-w-0 flex-1">
          <p class="font-extrabold {{if and (not .Current) (not .Done)}}text-gray-700{{end}}">{{.Title}}</p>
          <p class="text-sm text-gray-700">{{.Body}}</p>
          {{if .Current}}<a href="{{.Href}}" class="nsk-btn nsk-btn-sm mt-3">{{.CTA}}{{icon "arrow" 16}}</a>{{end}}
        </div>
      </li>
      {{end}}
    </ol>
  </section>
  {{end}}

  {{if gt .ScanCount 0}}
  <section class="nsk-card mb-6">
    <div class="nsk-card-head flex items-center justify-between gap-2">
      <h2>Latest scans</h2>
      <a href="/dashboard/scans" class="text-sm font-bold underline">See all</a>
    </div>
    <div class="p-4" id="pentest-list" hx-get="/api/pentests/list?limit=3" hx-trigger="load, every 10s, refresh" hx-swap="innerHTML">
      <p class="text-sm text-gray-700">Loading your scans…</p>
    </div>
  </section>
  {{end}}
{{template "shell_close"}}{{end}}

{{define "scanpicker"}}<div id="scan-picker">
{{with .Verified}}
<ul class="space-y-2">
  {{range .}}
  <li class="flex flex-wrap items-center justify-between gap-3 border-2 border-black bg-white p-3">
    <span class="flex min-w-0 items-center gap-2 font-bold">{{icon "globe" 18}}<span class="break-all">{{.Domain}}</span></span>
    {{if le $.Balance 0.0}}
    <a href="/dashboard/credits" class="nsk-btn nsk-btn-sm" aria-label="Buy a credit to scan {{.Domain}}">Buy a credit to scan</a>
    {{else}}
    <button hx-post="/api/pentests/start" hx-vals='{"domain":"{{.Domain}}","mode":"standard"}'
      hx-target="#flash" hx-swap="innerHTML"
      title="Checks this website from the outside. Read-only, nothing on your site is changed."
      aria-label="Scan {{.Domain}} now, uses 1 credit"
      class="nsk-btn nsk-btn-sm">Scan now · 1 credit</button>
    {{end}}
  </li>
  {{end}}
</ul>
{{else}}
<p class="text-sm text-gray-700">You don't have a verified website yet. <a href="/dashboard/websites" class="font-bold underline">Add and verify one</a> first.</p>
{{end}}
</div>{{end}}
`

const dashWebsitesHTML = `{{define "dashboard_websites"}}{{template "shell_open" shell . "sites" "My websites"}}
  <div class="mb-6">
    <h1 class="text-2xl font-extrabold sm:text-3xl">My websites</h1>
    <p class="mt-1 max-w-2xl text-gray-700">Add the websites you own. We confirm ownership first, so nobody can scan a website that isn't theirs.</p>
  </div>

  <section class="nsk-card mb-6">
    <h2 class="nsk-card-head">Add a website</h2>
    <form class="p-4" hx-post="/api/domains" hx-target="#flash" hx-swap="innerHTML">
      <label for="domain" class="nsk-label">Website address</label>
      <div class="flex flex-col gap-3 sm:flex-row">
        <input id="domain" name="domain" required placeholder="www.mybusiness.com" autocomplete="off" autocapitalize="none" spellcheck="false" inputmode="url"
          class="nsk-input flex-1"/>
        <button class="nsk-btn">Add website</button>
      </div>
      <p class="nsk-hint">Just the address, like <strong>www.mybusiness.com</strong>. Pasting a full link works too — we tidy it up.</p>
    </form>
  </section>

  <h2 class="mb-3 text-xl font-extrabold">Your websites</h2>
  {{template "domainlist" .}}
{{template "shell_close"}}{{end}}

{{define "domainlist"}}<div id="domain-list">
{{if .Domains}}
<ul class="space-y-4">
  {{range .Domains}}
  <li class="nsk-card p-4">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div class="flex min-w-0 items-center gap-2">{{icon "globe" 20}}<h3 class="break-all text-lg font-extrabold">{{.Domain}}</h3></div>
      {{if eq .Status "verified"}}<span class="nsk-badge nsk-badge-ok">{{icon "check" 14}}Verified</span>{{else}}<span class="nsk-badge nsk-badge-warn">Not verified yet</span>{{end}}
    </div>

    {{if eq .Status "verified"}}
    <p class="mt-2 text-sm text-gray-700">Ownership confirmed. This website is ready to scan.</p>
    <div class="mt-3 flex flex-wrap items-center gap-3">
      {{if le $.Balance 0.0}}
      <button type="button" disabled title="You have no credits. Buy a credit to run a scan." aria-label="Scan {{.Domain}} now (unavailable: you have no credits)" class="nsk-btn nsk-btn-sm">Scan now · 1 credit</button>
      <a href="/dashboard/credits" class="text-sm font-bold underline">Buy credits first</a>
      {{else}}
      <button hx-post="/api/pentests/start" hx-vals='{"domain":"{{.Domain}}","mode":"standard"}'
        hx-target="#flash" hx-swap="innerHTML"
        title="Checks this website from the outside. Read-only, nothing on your site is changed."
        aria-label="Scan {{.Domain}} now, uses 1 credit"
        class="nsk-btn nsk-btn-sm">Scan now · 1 credit</button>
      {{end}}
    </div>
    {{else}}
    <div class="mt-3 border-2 border-black bg-[#fffbeb] p-3">
      <p class="font-extrabold">Prove this website is yours <span class="font-medium text-gray-700">(about 2 minutes)</span></p>
      <ol class="mt-2 list-decimal space-y-3 pl-5 text-sm text-gray-800">
        <li>Open the <strong>DNS settings</strong> for <strong class="break-all">{{.Domain}}</strong>. They are at the company where you bought the domain.</li>
        <li>Add a new record of type <strong>TXT</strong> with this name and value:
          <div class="mt-2 space-y-2">
            <div>
              <span class="nsk-label">Name (sometimes called "Host")</span>
              <div class="nsk-code"><span>_netsekurity</span>
                <button type="button" data-copy="_netsekurity" onclick="nskCopy(this)" aria-label="Copy the record name" class="nsk-btn nsk-btn-ghost nsk-btn-sm">{{icon "copy" 14}}<span class="nsk-copy-label">Copy</span></button></div>
            </div>
            <div>
              <span class="nsk-label">Value</span>
              <div class="nsk-code"><span>{{.TXT}}</span>
                <button type="button" data-copy="{{.TXT}}" onclick="nskCopy(this)" aria-label="Copy the record value" class="nsk-btn nsk-btn-ghost nsk-btn-sm">{{icon "copy" 14}}<span class="nsk-copy-label">Copy</span></button></div>
            </div>
          </div>
          <p class="nsk-hint">Some providers want the full name instead: <code>_netsekurity.{{.Domain}}</code></p>
        </li>
        <li>Save it, wait a minute or two, then press <strong>Check now</strong>.</li>
      </ol>
      <p class="mt-3 text-sm text-gray-700">Not sure how? Send these steps to whoever looks after your website or domain.</p>
    </div>
    <div class="mt-3 flex flex-wrap gap-3">
      <button hx-post="/api/domains/verify" hx-vals='{"domain":"{{.Domain}}"}'
        hx-target="#flash" hx-swap="innerHTML" aria-label="Check the record for {{.Domain}} now" class="nsk-btn nsk-btn-sm">Check now</button>
      <button hx-post="/api/domains/delete" hx-vals='{"domain":"{{.Domain}}"}'
        hx-target="#flash" hx-swap="innerHTML"
        hx-confirm="Remove {{.Domain}} from your account?"
        aria-label="Remove {{.Domain}}"
        class="nsk-btn nsk-btn-ghost nsk-btn-sm">Remove</button>
    </div>
    {{end}}
  </li>
  {{end}}
</ul>
{{else}}
<div class="nsk-card p-6 text-center">
  <p class="text-lg font-extrabold">No websites yet</p>
  <p class="mt-1 text-gray-700">Add your first website above to get started.</p>
</div>
{{end}}
</div>{{end}}
`

const dashScansHTML = `{{define "dashboard_scans"}}{{template "shell_open" shell . "scans" "Security scans"}}
  <div class="mb-6">
    <h1 class="text-2xl font-extrabold sm:text-3xl">Security scans</h1>
    <p class="mt-1 max-w-2xl text-gray-700">Each scan checks your website from the outside, the way an attacker would. It is read-only and safe: nothing on your website is changed.</p>
  </div>

  <section class="nsk-card mb-6">
    <h2 class="nsk-card-head">Start a scan</h2>
    <div class="p-4">
      {{template "scanpicker" .}}
      <p class="mt-3 text-sm text-gray-700">Each scan uses 1 credit and takes 10 minutes to 24 hours, depending on what it finds. One scan runs at a time.</p>
    </div>
  </section>

  <section class="nsk-card mb-6">
    <h2 class="nsk-card-head">Your scans</h2>
    <div class="p-4" id="pentest-list" hx-get="/api/pentests/list" hx-trigger="load, every 10s, refresh" hx-swap="innerHTML">
      <p class="text-sm text-gray-700">Loading your scans…</p>
    </div>
  </section>

  <details class="nsk-card mb-6">
    <summary class="nsk-card-head cursor-pointer">How to read your results</summary>
    <ul class="space-y-2 p-4 text-sm text-gray-800">
      <li><span class="nsk-badge nsk-badge-bad">Critical</span> and <span class="nsk-badge nsk-badge-bad">High</span> — fix these first.</li>
      <li><span class="nsk-badge nsk-badge-warn">Medium</span> and <span class="nsk-badge nsk-badge-blue">Low</span> — worth fixing when you can.</li>
      <li><span class="nsk-badge nsk-badge-muted">Info</span> — good to know. No action needed.</li>
      <li>The PDF report explains every finding and how to fix it. You can forward it to your developer.</li>
    </ul>
  </details>
{{template "shell_close"}}{{end}}
`

const dashCreditsHTML = `{{define "dashboard_credits"}}{{template "shell_open" shell . "credits" "Credits & payments"}}
  <div class="mb-6">
    <h1 class="text-2xl font-extrabold sm:text-3xl">Credits &amp; payments</h1>
    <p class="mt-1 max-w-2xl text-gray-700">1 credit = 1 scan of 1 website. Credits never expire.</p>
  </div>

  <section class="nsk-card nsk-card-tint mb-6 p-4">
    <p class="text-sm font-bold text-gray-700">Your balance</p>
    <p class="mt-1 text-4xl font-extrabold">{{creditsLabel .Balance}}</p>
    {{if le .Balance 0.0}}<p class="mt-2 text-gray-800">You have no credits yet. Choose a pack below to run your first scan.</p>{{end}}
  </section>

  <h2 class="mb-3 text-xl font-extrabold">Buy credits</h2>
  <div class="mb-2 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
    {{range .Packages}}
    <div class="nsk-card flex flex-col p-4">
      <p class="text-sm font-extrabold uppercase tracking-wide text-gray-700">{{.Name}}</p>
      <p class="mt-2 text-3xl font-extrabold">${{usd .USD}}</p>
      <p class="mt-1 font-bold">{{creditsLabel .Credits}}</p>
      <p class="text-sm text-gray-700">{{perScan .USD .Credits}} per scan</p>
      <a href="/checkout?package={{.ID}}" class="nsk-btn nsk-btn-block mt-4">Buy {{.Name}}</a>
    </div>
    {{end}}
  </div>
  <p class="mb-8 text-sm text-gray-700">All prices are in US dollars. You will see the total before paying.</p>

  <section class="nsk-card mb-6">
    <h2 class="nsk-card-head">Payments</h2>
    <div class="p-4">
      {{if .Payments}}
      <ul class="divide-y-2 divide-gray-200">
        {{range .Payments}}
        <li class="flex flex-wrap items-center justify-between gap-3 py-3 first:pt-0 last:pb-0">
          <div class="min-w-0">
            <p class="font-bold">{{if .Package}}{{.Package}} pack{{else}}Credit top-up{{end}} · ${{usd .AmountUSD}}</p>
            <p class="text-sm text-gray-700">{{timeTag .CreatedAt}} · {{creditsLabel .Credits}}</p>
          </div>
          <div class="flex items-center gap-2">
            {{if eq .Status "paid"}}<span class="nsk-badge nsk-badge-ok">Paid</span>
            {{else if eq .Status "pending"}}<span class="nsk-badge nsk-badge-warn">Waiting for payment</span>{{if .URL}}<a href="{{.URL}}" target="_blank" rel="noopener" class="nsk-btn nsk-btn-sm">Pay now</a>{{end}}
            {{else}}<span class="nsk-badge nsk-badge-muted">{{.Status}}</span>{{end}}
          </div>
        </li>
        {{end}}
      </ul>
      {{else}}
      <p class="text-gray-700">No payments yet.</p>
      {{end}}

      {{if .Expired}}
      <details class="mt-4 border-t-2 border-gray-200 pt-3">
        <summary class="cursor-pointer text-sm font-bold text-gray-700">{{len .Expired}} unpaid invoice(s) expired after 24 hours</summary>
        <ul class="mt-2 divide-y divide-gray-200 text-sm text-gray-700">
          {{range .Expired}}
          <li class="flex flex-wrap justify-between gap-2 py-2"><span>{{if .Package}}{{.Package}} pack{{else}}Credit top-up{{end}} · ${{usd .AmountUSD}}</span><span>{{timeTag .CreatedAt}}</span></li>
          {{end}}
        </ul>
      </details>
      {{end}}
    </div>
  </section>

  <section class="nsk-card mb-6">
    <h2 class="nsk-card-head">Credit activity</h2>
    <div class="p-4">
      {{if .Transactions}}
      <ul class="divide-y-2 divide-gray-200">
        {{range .Transactions}}
        <li class="flex items-center justify-between gap-3 py-3 first:pt-0 last:pb-0">
          <div class="min-w-0">
            <p class="font-bold">{{txnLabel .Type}}</p>
            <p class="text-sm text-gray-700">{{if .Description}}{{.Description}} · {{end}}{{timeTag .CreatedAt}}</p>
          </div>
          <span class="shrink-0 text-lg font-extrabold {{if or (eq .Type "spend") (lt .Amount 0.0)}}text-gray-900{{else}}text-green-800{{end}}">{{txnAmount .Type .Amount}}</span>
        </li>
        {{end}}
      </ul>
      {{else}}
      <p class="text-gray-700">Nothing here yet. Your credits and scans will show up in this list.</p>
      {{end}}
    </div>
  </section>
{{template "shell_close"}}{{end}}
`

const dashAutomationHTML = `{{define "dashboard_automation"}}{{template "shell_open" shell . "auto" "Automation"}}
  <div class="mb-6">
    <span class="nsk-tag">For developers</span>
    <h1 class="mt-2 text-2xl font-extrabold sm:text-3xl">Automatic scans</h1>
    <p class="mt-1 max-w-2xl text-gray-700">Start a scan automatically every time your website is published.</p>
  </div>

  <div class="nsk-alert nsk-alert-info mb-6">
    <strong>Most people can skip this page.</strong> Scans you start from the dashboard work exactly the same. This page is for a developer who wants to connect Netsekurity to a deployment pipeline.
  </div>

  <section class="nsk-card mb-6">
    <h2 class="nsk-card-head">Create an access token</h2>
    <form class="p-4" hx-post="/api/tokens/create" hx-target="#token-result" hx-swap="innerHTML">
      <div class="grid gap-3 sm:grid-cols-2">
        <div>
          <label for="token-name" class="nsk-label">Name</label>
          <input id="token-name" name="name" type="text" value="CI/CD" class="nsk-input"/>
          <p class="nsk-hint">So you can recognise it later.</p>
        </div>
        <div>
          <label for="token-expiry" class="nsk-label">Valid for</label>
          <select id="token-expiry" name="expiry_days" class="nsk-input">
            <option value="7">7 days</option>
            <option value="14">14 days</option>
            <option value="30" selected>30 days</option>
            <option value="60">60 days</option>
            <option value="90">90 days</option>
          </select>
        </div>
      </div>
      <button class="nsk-btn mt-4">Create token</button>
      <div id="token-result" class="mt-3" aria-live="polite"></div>
    </form>
  </section>

  <section class="nsk-card mb-6">
    <h2 class="nsk-card-head">Your tokens</h2>
    <div class="p-4" id="token-list">
      <div hx-get="/api/tokens" hx-trigger="load, every 15s, tokens-changed from:body" hx-swap="innerHTML"><p class="text-sm text-gray-700">Loading…</p></div>
    </div>
  </section>

  <section class="nsk-card mb-6">
    <h2 class="nsk-card-head">How to use it</h2>
    <div class="space-y-3 p-4 text-sm text-gray-800">
      <p>On every deploy, have your pipeline send this request. Keep the token secret.</p>
      <pre class="nsk-code block overflow-x-auto whitespace-pre-wrap">curl -s -X POST https://netsekurity.com/api/v1/pentests \
  -H "X-API-Token: YOUR_TOKEN" \
  -d "domain=www.mybusiness.com&amp;mode=standard"</pre>
      <p>Each scan uses 1 credit. <a href="/docs" target="_blank" rel="noopener" class="font-bold underline">Read the full setup guide</a></p>
    </div>
  </section>
{{template "shell_close"}}{{end}}
`
