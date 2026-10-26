package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// A confirmation step between the dashboard and Xendit. The buy button used to
// post straight to /api/topup and redirect the browser off-site, so the last
// thing a customer saw before a payment page was a small "buy" button — no
// summary, no total, and no statement of the currency they were about to be
// charged in. That is where 28 consecutive invoices were abandoned.

type checkoutData struct {
	Name       string
	Email      string
	Balance    float64
	Package    pkg
	Total      string // "$50"
	ChargeNote string // the currency footnote
	Mismatch   bool   // charge currency differs from the quoted USD
}

// handleCheckout renders the order summary for one credit package.
func handleCheckout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store, private")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	claims, err := currentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	id := strings.TrimSpace(r.URL.Query().Get("package"))
	if id == "" {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}

	var p pkg
	if err := db.QueryRow(`SELECT id, name, usd_price, credits FROM credit_packages WHERE id=? AND is_active=1`, id).
		Scan(&p.ID, &p.Name, &p.USD, &p.Credits); err != nil {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}

	data := checkoutData{Package: p, Mismatch: currencyMismatch()}
	db.QueryRow(`SELECT name, email FROM users WHERE id=?`, claims.Sub).Scan(&data.Name, &data.Email)
	db.QueryRow(`SELECT balance FROM credit_balance WHERE user_id=?`, claims.Sub).Scan(&data.Balance)
	data.Total = "$" + trimUSD(p.USD)

	// The footnote states the charge currency plainly. When it matches the
	// quoted USD that is a one-line reassurance; when it does not, the customer
	// is told the other amount here rather than meeting it on the payment page.
	if data.Mismatch {
		cur := chargeCurrency()
		data.ChargeNote = "Prices are shown in USD. Your card will be charged in " + cur +
			" — " + groupDigits(topUpAmountUSD(p.USD, cur, usdRate()), ".") + " " + cur +
			" — because that is the currency our payment provider settles in for this account."
	} else {
		data.ChargeNote = "All prices are in US dollars. You will be charged " + data.Total +
			" USD. No conversion is applied and no other currency is involved."
	}

	if err := checkoutTpl.ExecuteTemplate(w, "checkout", data); err != nil {
		log.Printf("checkout render error: %v", err)
	}
}

// trimUSD renders a package price without trailing zeros ("50", not "50.00").
func trimUSD(v float64) string {
	s := strings.TrimRight(strings.TrimRight(strconv.FormatFloat(v, 'f', 2, 64), "0"), ".")
	if s == "" {
		return "0"
	}
	return s
}

var checkoutTpl = template.Must(template.New("checkout").Funcs(template.FuncMap{
	"cssHash":      func() string { return cssHash },
	"fbpixel":      func() template.HTML { return template.HTML(consentGate(true)) },
	"icon":         icon,
	"credits":      fmtCredits,
	"creditsLabel": creditsLabel,
}).Parse(checkoutHTML))

const checkoutHTML = `{{define "checkout"}}<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
<title>Confirm your order — Netsekurity</title>
<meta name="robots" content="noindex, nofollow"/>
<link rel="preconnect" href="https://fonts.googleapis.com"/>
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin/>
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;500;600;700;800&display=swap"/>
<link rel="stylesheet" href="/css/styles.css?v={{cssHash}}"/>
<link rel="icon" type="image/svg+xml" href="/favicon.svg" />{{fbpixel}}
<script src="https://unpkg.com/htmx.org@1.9.12/dist/htmx.min.js" defer></script>
</head>
<body class="nsk-app min-h-screen overflow-x-hidden">
<header class="fixed inset-x-0 top-0 z-40 h-16 border-b-2 border-black bg-white">
  <div class="mx-auto flex h-full max-w-5xl items-center justify-between gap-3 px-4">
    <a href="/dashboard" class="flex items-center gap-2.5" aria-label="Netsekurity home">
      <img src="/favicon.svg" width="34" height="34" alt="" class="h-[34px] w-[34px]"/>
      <span class="text-xl font-extrabold tracking-tight">Netsekurity</span>
    </a>
    <a href="/dashboard/credits" class="nsk-btn nsk-btn-ghost nsk-btn-sm">Back to credits</a>
  </div>
</header>

<main class="mx-auto max-w-xl px-4 pb-16 pt-28">
  <h1 class="text-2xl font-extrabold sm:text-3xl">Confirm your order</h1>
  <p class="mt-1 text-gray-700">Check the details below. Next you go to our secure payment page to finish.</p>

  <section class="nsk-card mt-6">
    <h2 class="nsk-card-head">Order summary</h2>

    <dl class="divide-y-2 divide-gray-200 px-4">
      <div class="flex items-baseline justify-between gap-4 py-3">
        <dt class="text-gray-700">Pack</dt>
        <dd class="font-extrabold">{{.Package.Name}}</dd>
      </div>
      <div class="flex items-baseline justify-between gap-4 py-3">
        <dt class="text-gray-700">Credits</dt>
        <dd class="font-extrabold">{{creditsLabel .Package.Credits}}</dd>
      </div>
      <div class="flex items-baseline justify-between gap-4 py-3">
        <dt class="text-gray-700">What you get</dt>
        <dd class="text-right text-sm">Each credit runs 1 security scan of 1 website</dd>
      </div>
      <div class="flex items-baseline justify-between gap-4 py-3">
        <dt class="text-gray-700">Billed to</dt>
        <dd class="break-all text-right text-sm">{{.Email}}</dd>
      </div>
      <div class="flex items-baseline justify-between gap-4 py-4">
        <dt class="text-lg font-extrabold">Total</dt>
        <dd class="text-3xl font-extrabold">{{.Total}} <span class="text-sm font-bold text-gray-700">USD</span></dd>
      </div>
    </dl>

    <div class="px-4 pb-4">
      <p class="nsk-alert {{if .Mismatch}}nsk-alert-warn{{else}}nsk-alert-info{{end}} text-sm">
        {{if .Mismatch}}<strong>Note on currency — </strong>{{end}}{{.ChargeNote}}
      </p>

      <form hx-post="/api/topup" hx-target="#pay-result" hx-swap="innerHTML" hx-disabled-elt="find button" class="mt-4">
        <input type="hidden" name="package_id" value="{{.Package.ID}}"/>
        <button
          onclick="nskTrack('InitiateCheckout',{content_name:'{{.Package.Name}}',content_type:'product',value:{{.Package.USD}},currency:'USD'})"
          class="nsk-btn nsk-btn-block min-h-[52px] text-base">
          {{icon "lock" 18}}Proceed to payment — {{.Total}} USD
        </button>
      </form>
      <div id="pay-result" class="mt-3" aria-live="polite"></div>

      <ul class="mt-4 space-y-1.5 text-sm text-gray-700">
        <li>Credits are added as soon as the payment goes through, and they never expire. Your balance now: <strong>{{credits .Balance}}</strong>.</li>
        <li>Payment is handled by Xendit. We never see or store your card details.</li>
      </ul>
    </div>
  </section>

  <p class="mt-5 text-center">
    <a href="/dashboard/credits" class="font-bold underline">← Choose a different pack</a>
  </p>
</main>
<script>
// HTMX does not swap error responses by default; show our own alert boxes (and only those).
document.addEventListener('htmx:beforeSwap', function (e) {
  if (e.detail.xhr.status >= 400 && (e.detail.serverResponse || '').indexOf('nsk-alert') !== -1) {
    e.detail.shouldSwap = true; e.detail.isError = false;
  }
});
document.addEventListener('htmx:responseError', function (e) {
  var r = document.getElementById('pay-result');
  if (r && !r.innerHTML) r.innerHTML = '<div class="nsk-alert nsk-alert-bad" role="alert"><strong>Something went wrong.</strong> Please try again.</div>';
});
</script>
</body>
</html>{{end}}`
