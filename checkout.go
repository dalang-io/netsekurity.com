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
	"cssHash": func() string { return cssHash },
	"fbpixel": func() template.HTML { return template.HTML(metaPixelSnippet(true)) },
}).Parse(checkoutHTML))

const checkoutHTML = `{{define "checkout"}}<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
<title>checkout — netsekurity</title>
<meta name="robots" content="noindex, nofollow"/>
<link rel="stylesheet" href="/css/styles.css?v={{cssHash}}"/>
<link rel="icon" type="image/svg+xml" href="/favicon.svg" />{{fbpixel}}
<script src="https://unpkg.com/htmx.org@1.9.12/dist/htmx.min.js" defer></script>
</head>
<body class="scanlines bg-ink text-gray-300 min-h-screen">
<header class="border-b border-emerald-500/25 bg-ink/85">
  <div class="mx-auto flex max-w-3xl items-center justify-between px-4 py-3">
    <a href="/" class="font-mono text-base font-bold text-white"><span class="glow text-emerald-400">net</span>sekurity<span class="text-emerald-500">.com</span> <span class="text-xs text-cyan-300">/checkout</span></a>
    <a href="/dashboard" class="prompt font-mono text-xs text-gray-500 hover:text-white">back to dashboard</a>
  </div>
</header>

<main class="mx-auto max-w-3xl px-4 py-8 sm:px-6">
  <h1 class="font-mono text-2xl font-bold text-white"><span class="text-emerald-400">$</span> confirm order</h1>
  <p class="copy mt-2 text-gray-400">Review what you are buying. You will be taken to our payment provider to complete it.</p>

  <section class="mt-6 rounded border border-emerald-500/30 bg-[#04060c]">
    <div class="flex items-center gap-1.5 border-b border-emerald-500/20 px-4 py-2">
      <span class="h-2.5 w-2.5 rounded-full bg-emerald-500/70"></span>
      <span class="ml-1 font-mono text-[11px] text-gray-500">$ order</span>
    </div>

    <dl class="divide-y divide-white/10 px-4">
      <div class="flex items-baseline justify-between py-3">
        <dt class="font-mono text-xs text-gray-500">package</dt>
        <dd class="font-mono text-sm text-white">{{.Package.Name}}</dd>
      </div>
      <div class="flex items-baseline justify-between py-3">
        <dt class="font-mono text-xs text-gray-500">credits</dt>
        <dd class="font-mono text-sm text-emerald-300">{{printf "%.0f" .Package.Credits}}</dd>
      </div>
      <div class="flex items-baseline justify-between py-3">
        <dt class="font-mono text-xs text-gray-500">what 1 credit buys</dt>
        <dd class="font-mono text-xs text-gray-400">1 external assessment · 1 domain</dd>
      </div>
      <div class="flex items-baseline justify-between py-3">
        <dt class="font-mono text-xs text-gray-500">billed to</dt>
        <dd class="font-mono text-xs text-gray-400">{{.Email}}</dd>
      </div>
      <div class="flex items-baseline justify-between py-3">
        <dt class="font-mono text-sm font-bold text-white">total</dt>
        <dd class="font-mono text-2xl font-bold text-emerald-300 glow">{{.Total}} <span class="text-sm font-normal text-gray-500">USD</span></dd>
      </div>
    </dl>

    <div class="px-4 pb-4">
      <p class="copy rounded border {{if .Mismatch}}border-yellow-500/40 bg-yellow-500/10 text-yellow-200{{else}}border-white/10 bg-white/5 text-gray-400{{end}} px-3 py-2 text-[13px]">
        {{if .Mismatch}}<span class="font-bold">Note on currency — </span>{{end}}{{.ChargeNote}}
      </p>

      <form hx-post="/api/topup" hx-target="#pay-result" hx-swap="innerHTML" hx-disabled-elt="find button" class="mt-4">
        <input type="hidden" name="package_id" value="{{.Package.ID}}"/>
        <button
          onclick="fbq('track','InitiateCheckout',{content_name:'{{.Package.Name}}',content_type:'product',value:{{.Package.USD}},currency:'USD'})"
          class="w-full rounded border-2 border-emerald-400 bg-emerald-500/10 px-5 py-3 font-mono text-sm font-bold text-emerald-300 hover:bg-emerald-500/20 glow disabled:cursor-not-allowed disabled:opacity-50">
          proceed to payment — {{.Total}} USD
        </button>
      </form>
      <div id="pay-result" class="mt-3"></div>

      <p class="copy mt-3 text-[13px] text-gray-600">
        Credits are added to your balance as soon as the payment settles, and never expire.
        Your current balance is {{printf "%.1f" .Balance}}.
      </p>
      <p class="copy mt-1 text-[13px] text-gray-600">
        Payment is handled by Xendit. We never see or store your card details.
      </p>
    </div>
  </section>

  <p class="mt-5 text-center font-mono text-xs">
    <a href="/dashboard" class="text-gray-500 hover:text-emerald-300">← pick a different package</a>
  </p>
</main>
</body>
</html>{{end}}`
