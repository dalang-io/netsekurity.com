package main

import (
	"strings"
	"testing"
)

func TestFmtCreditsAndLabel(t *testing.T) {
	for in, want := range map[float64]string{0: "0", 1: "1", 3: "3", 20: "20", 2.5: "2.5", 0.5: "0.5"} {
		if got := fmtCredits(in); got != want {
			t.Errorf("fmtCredits(%v) = %q, want %q", in, got, want)
		}
	}
	for in, want := range map[float64]string{0: "0 credits", 1: "1 credit", 2: "2 credits", 2.5: "2.5 credits"} {
		if got := creditsLabel(in); got != want {
			t.Errorf("creditsLabel(%v) = %q, want %q", in, got, want)
		}
	}
}

// Go writes RFC3339 and SQLite writes datetime('now'); both must read as UTC.
func TestTimeTagParsesBothStoredFormats(t *testing.T) {
	for _, in := range []string{"2026-08-18T15:59:38Z", "2026-08-18 15:59:38"} {
		out := string(timeTag(in))
		if !strings.Contains(out, `datetime="2026-08-18T15:59:38Z"`) || !strings.Contains(out, "18 Aug 2026, 15:59 UTC") {
			t.Errorf("timeTag(%q) = %s", in, out)
		}
	}
	// Garbage is shown, escaped, never dropped or trusted as markup.
	if got := string(timeTag("<b>x</b>")); strings.Contains(got, "<b>") {
		t.Errorf("unparseable time was not escaped: %s", got)
	}
	if got := string(timeTag("")); got != "" {
		t.Errorf("empty time = %q", got)
	}
}

func TestFirstNameAndInitial(t *testing.T) {
	if got := firstName("Hanjara Adhyatma", "h@x.io"); got != "Hanjara" {
		t.Errorf("firstName = %q", got)
	}
	if got := firstName("", "siti@example.com"); got != "siti" {
		t.Errorf("firstName from email = %q", got)
	}
	if got := firstName("", ""); got != "there" {
		t.Errorf("firstName fallback = %q", got)
	}
	if got := initial("budi", "x@y.z"); got != "B" {
		t.Errorf("initial = %q", got)
	}
	if got := initial("", ""); got != "?" {
		t.Errorf("initial fallback = %q", got)
	}
}

func TestPerScan(t *testing.T) {
	for _, tc := range []struct {
		usd, credits float64
		want         string
	}{{50, 1, "$50"}, {100, 3, "$33.33"}, {500, 20, "$25"}, {1000, 50, "$20"}, {10, 0, ""}} {
		if got := perScan(tc.usd, tc.credits); got != tc.want {
			t.Errorf("perScan(%v,%v) = %q, want %q", tc.usd, tc.credits, got, tc.want)
		}
	}
}

func TestTxnLabelAndAmount(t *testing.T) {
	if txnLabel("topup") != "Credits added" || txnLabel("spend") != "Scan" || txnLabel("admin_credit") != "Credit adjustment" {
		t.Error("transaction labels changed")
	}
	if got := txnLabel("weird"); got != "weird" {
		t.Errorf("unknown type = %q", got)
	}
	for _, tc := range []struct {
		kind   string
		amount float64
		want   string
	}{
		{"topup", 3, "+3"}, {"spend", 1, "−1"}, {"spend", 2, "−2"},
		{"admin_credit", 5, "+5"}, {"admin_credit", -2, "−2"}, {"topup", 2.5, "+2.5"},
	} {
		if got := txnAmount(tc.kind, tc.amount); got != tc.want {
			t.Errorf("txnAmount(%s,%v) = %q, want %q", tc.kind, tc.amount, got, tc.want)
		}
	}
}

func TestIconAndAlert(t *testing.T) {
	if got := string(icon("nope", 16)); got != "" {
		t.Errorf("unknown icon = %q", got)
	}
	if got := string(icon("check", 20)); !strings.Contains(got, `width="20"`) || !strings.Contains(got, "<svg") {
		t.Errorf("icon = %s", got)
	}
	if got := alertHTML("bad", "x"); !strings.Contains(got, `role="alert"`) || !strings.Contains(got, "nsk-alert-bad") {
		t.Errorf("bad alert = %s", got)
	}
	if got := alertHTML("ok", "x"); !strings.Contains(got, `role="status"`) {
		t.Errorf("ok alert = %s", got)
	}
}
