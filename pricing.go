package main

import (
	"strconv"
	"strings"
)

// The product is sold internationally and every price is quoted in USD. Nothing
// customer-facing renders a converted amount.
//
// Xendit still bills in the account's configured currency, though, and that is a
// payment-page surprise the customer only meets after clicking buy: every
// invoice created before 2026-08-28 was raised in IDR and none was ever paid.
// So the conversion is not displayed — it is surfaced to operators instead, via
// a startup warning and a banner on /su, until the charge currency matches the
// quoted one.

// chargeCurrency is the currency Xendit will actually bill in. USD is the
// default because the product is sold internationally and every price is quoted
// in USD; billing in anything else means the customer meets a number they have
// never seen, on the payment page, after they have already decided to buy.
func chargeCurrency() string {
	return strings.ToUpper(getenv("XENDIT_CURRENCY", "USD"))
}

// usdRate is the USD -> chargeCurrency rate used to build the invoice.
func usdRate() float64 {
	return parseFloat(getenv("XENDIT_USD_RATE", "16500"))
}

// currencyMismatch reports whether checkout bills in something other than the
// USD every price is quoted in.
func currencyMismatch() bool {
	return !strings.EqualFold(chargeCurrency(), "USD")
}

// operatorCurrencyWarning describes the mismatch for operators. It is never shown
// to customers — they see USD, and would be charged the converted amount.
func operatorCurrencyWarning() string {
	if !currencyMismatch() {
		return ""
	}
	cur := chargeCurrency()
	rate := usdRate()
	if rate <= 0 {
		rate = 16500
	}
	return "Prices display in USD but Xendit bills in " + cur +
		" at 1 USD = " + groupDigits(int64(rate), ".") + " " + cur +
		". A customer clicking $50 reaches a payment page for " +
		groupDigits(topUpAmountUSD(50, cur, rate), ".") + " " + cur +
		". Set XENDIT_CURRENCY=USD once the Xendit account supports it."
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
