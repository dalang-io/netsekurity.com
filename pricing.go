package main

import (
	"strconv"
	"strings"
)

// Prices are quoted in USD everywhere on the site, but Xendit charges in the
// account's configured currency (IDR by default). Showing only the USD figure
// sends the customer to a payment page demanding an amount they have never
// seen, which is why every invoice created before 2026-08-28 expired unpaid.
// Everything user-facing goes through these helpers so the two numbers are
// always presented together.

// chargeCurrency is the currency Xendit will actually bill in.
func chargeCurrency() string {
	return strings.ToUpper(getenv("XENDIT_CURRENCY", "IDR"))
}

// usdRate is the USD -> chargeCurrency rate used to build the invoice.
func usdRate() float64 {
	return parseFloat(getenv("XENDIT_USD_RATE", "16500"))
}

// localPrice renders the amount the customer will be charged, e.g. "Rp 825.000".
// It returns "" when the charge currency is USD and no conversion happens.
func localPrice(usd float64) string {
	cur := chargeCurrency()
	if strings.EqualFold(cur, "USD") {
		return ""
	}
	amount := topUpAmountUSD(usd, cur, usdRate())
	if cur == "IDR" {
		return "Rp " + groupDigits(amount, ".")
	}
	return cur + " " + groupDigits(amount, ",")
}

// localRateNote explains the conversion once, for use under a price list.
func localRateNote() string {
	cur := chargeCurrency()
	if strings.EqualFold(cur, "USD") {
		return ""
	}
	rate := usdRate()
	if rate <= 0 {
		rate = 16500
	}
	return "Prices are quoted in USD. Payment is charged in " + cur +
		" at 1 USD = " + groupDigits(int64(rate), ".") + " " + cur + "."
}

// groupDigits inserts a thousands separator, e.g. 825000 -> "825.000".
func groupDigits(n int64, sep string) string {
	neg := n < 0
	if neg {
		n = -n
	}
	s := strconv.FormatInt(n, 10)
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteString(sep)
		}
		b.WriteRune(c)
	}
	if neg {
		return "-" + b.String()
	}
	return b.String()
}
