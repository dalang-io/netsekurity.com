package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// --- the consent cookie is the only thing that grants anything -------------

// Meta's own rule, mirrored for us: absent is not a yes. A missing, empty,
// misspelled or hostile cookie value must all read as "no consent", because the
// tracker is only ever supposed to fire on an explicit choice.
func TestConsentGrantedOnlyForExplicitGrant(t *testing.T) {
	cases := []struct {
		name   string
		cookie *http.Cookie
		want   bool
	}{
		{"no cookie", nil, false},
		{"granted", &http.Cookie{Name: consentCookieName, Value: consentGrantedValue}, true},
		{"denied", &http.Cookie{Name: consentCookieName, Value: consentDeniedValue}, false},
		{"empty", &http.Cookie{Name: consentCookieName, Value: ""}, false},
		{"garbage", &http.Cookie{Name: consentCookieName, Value: "yes"}, false},
		{"substring", &http.Cookie{Name: consentCookieName, Value: "not-granted"}, false},
		{"wrong name", &http.Cookie{Name: "consent", Value: consentGrantedValue}, false},
	}
	for _, c := range cases {
		r := httptest.NewRequest("GET", "/", nil)
		if c.cookie != nil {
			r.AddCookie(c.cookie)
		}
		if got := consentGranted(r); got != c.want {
			t.Errorf("%s: consentGranted = %v, want %v", c.name, got, c.want)
		}
	}
}

// --- the served pages cannot reach Meta before consent ---------------------

// The pixel library and the <noscript> beacon were both in the <head> of every
// page. A page that still ships either of them makes the request on load, and no
// banner can take that back — this is the regression guard for that.
func TestServedPagesDoNotLoadThePixelEagerly(t *testing.T) {
	pages := map[string]string{
		"consent gate": consentGate(true),
		"checkout":     checkoutHTML,
		"dashboard":    dashboardHTML,
	}
	// Every public page, as served: the home pages, the guides and the evidence pages.
	for _, path := range publicPagePaths {
		pages[path] = renderPublic(t, path)
	}
	for name, html := range pages {
		if strings.Contains(html, `src="https://connect.facebook.net`) ||
			strings.Contains(html, "src='https://connect.facebook.net") {
			t.Errorf("%s: loads fbevents.js on page load — that request must wait for consent", name)
		}
		if strings.Contains(html, "facebook.com/tr") {
			t.Errorf("%s: ships the <noscript> tracking pixel, which fires with no JS and no way to consent", name)
		}
		if strings.Contains(html, "fbq('track'") && !strings.Contains(html, "nskTrack") {
			t.Errorf("%s: calls fbq directly instead of the consent-gated nskTrack", name)
		}
	}
}

// The gate itself must be present on every page that used to carry the pixel,
// otherwise a visitor could never agree and the pixel would never run.
func TestServedPagesCarryTheGate(t *testing.T) {
	for _, path := range publicPagePaths {
		page := renderPublic(t, path)
		if !strings.Contains(page, consentCookieName) || !strings.Contains(page, "nskTrack") {
			t.Errorf("%s has no consent gate — a visitor could never enable the pixel", path)
		}
	}

	// The signed-in templates are sources: the gate arrives through their
	// placeholder, which the tests above prove renders gated output.
	templates := map[string]string{
		"checkout":  checkoutHTML,
		"dashboard": dashboardHTML,
	}
	for name, html := range templates {
		if !strings.Contains(html, "{{fbpixel}}") && !strings.Contains(html, "__FBPIXEL__") {
			t.Errorf("%s: carries neither the gate nor its placeholder", name)
		}
	}
}

// The pixel code itself must survive inside the gate: the point is to defer it,
// never to lose the ability to fire it.
func TestGateStillContainsThePixelCode(t *testing.T) {
	g := consentGate(true)
	if !strings.Contains(g, "connect.facebook.net/en_US/fbevents.js") {
		t.Error("the pixel base code is gone from the gate")
	}
	if !strings.Contains(g, fbPixelID) {
		t.Errorf("the gate no longer initialises pixel %s", fbPixelID)
	}
	if !strings.Contains(g, "PAGEVIEW = true") {
		t.Error("pageView=true did not reach the gate")
	}
	if !strings.Contains(consentGate(false), "PAGEVIEW = false") {
		t.Error("pageView=false did not reach the gate")
	}
	// JSON-encoded, so a stray </script> or quote in the pixel code cannot break
	// out of the string it now lives in, and the browser has no code to run until
	// the gate puts it into a <script> itself.
	if !strings.Contains(g, `"!function(f,b,e,v,n,t,s)`) {
		t.Error("the pixel base code is not emitted as a JSON string")
	}
}

// --- the server-side Conversions API obeys the same gate -------------------

// recordingRT counts outgoing Meta calls without touching the network.
type recordingRT struct{ hits chan string }

func (rt recordingRT) RoundTrip(r *http.Request) (*http.Response, error) {
	rt.hits <- r.URL.Host
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("{}")),
		Header:     http.Header{},
		Request:    r,
	}, nil
}

func withMetaRecorder(t *testing.T) chan string {
	t.Helper()
	oldToken, hadToken := env["META_CAPI_TOKEN"]
	env["META_CAPI_TOKEN"] = "test-token"
	oldClient := fbClient
	hits := make(chan string, 8)
	fbClient = &http.Client{Transport: recordingRT{hits: hits}, Timeout: 2 * time.Second}
	t.Cleanup(func() {
		fbClient = oldClient
		if hadToken {
			env["META_CAPI_TOKEN"] = oldToken
		} else {
			delete(env, "META_CAPI_TOKEN")
		}
	})
	return hits
}

func waitNoHit(t *testing.T, hits chan string, what string) {
	t.Helper()
	select {
	case host := <-hits:
		t.Fatalf("%s sent a Conversions API event to %s without consent", what, host)
	case <-time.After(200 * time.Millisecond):
	}
}

func waitHit(t *testing.T, hits chan string, what string) {
	t.Helper()
	select {
	case host := <-hits:
		if host != "graph.facebook.com" {
			t.Errorf("%s: event went to %s, want graph.facebook.com", what, host)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("%s: no Conversions API event after consent", what)
	}
}

// A server call is not a way around the browser gate: the signup Lead carries a
// hashed email, IP and user agent, exactly the payload consent is about.
func TestSignupLeadRequiresConsent(t *testing.T) {
	hits := withMetaRecorder(t)

	plain := httptest.NewRequest("GET", "/auth/google/callback", nil)
	plain.Header.Set("CF-Connecting-IP", "203.0.113.9")
	plain.Header.Set("User-Agent", "Mozilla/5.0")
	trackSignupLead(plain, "u_nocookie", "no-consent@example.com")
	waitNoHit(t, hits, "signup Lead with no consent cookie")

	denied := httptest.NewRequest("GET", "/auth/google/callback", nil)
	denied.Header.Set("CF-Connecting-IP", "203.0.113.9")
	denied.AddCookie(&http.Cookie{Name: consentCookieName, Value: consentDeniedValue})
	trackSignupLead(denied, "u_denied", "declined@example.com")
	waitNoHit(t, hits, "signup Lead after declining")

	agreed := httptest.NewRequest("GET", "/auth/google/callback", nil)
	agreed.Header.Set("CF-Connecting-IP", "203.0.113.9")
	agreed.AddCookie(&http.Cookie{Name: consentCookieName, Value: consentGrantedValue})
	trackSignupLead(agreed, "u_granted", "agreed@example.com")
	waitHit(t, hits, "signup Lead after accepting")
}

// The Purchase event is fired by the Xendit webhook, which has no browser of its
// own — it replays the choice stored when the invoice was created. A payment
// made before the gate existed has no stored choice, and must not be reported.
func TestWebhookPurchaseRequiresStoredConsent(t *testing.T) {
	hits := withMetaRecorder(t)

	seed := func(externalID, consent string) {
		t.Helper()
		if _, err := db.Exec(`INSERT INTO payments (user_id, external_id, amount_usd, credits, status, meta_ip, meta_ua, meta_consent)
			VALUES ('u_test','`+externalID+`',50,1,'pending','203.0.113.9','Mozilla/5.0',?)`, consent); err != nil {
			t.Fatalf("seed %s: %v", externalID, err)
		}
	}
	seed("inv_noconsent", "")
	seed("inv_granted", consentGrantedValue)

	credited, err := creditPaymentByExternalID("inv_noconsent")
	if err != nil || !credited {
		t.Fatalf("creditPaymentByExternalID(no consent) = %v, %v; want true, nil", credited, err)
	}
	waitNoHit(t, hits, "webhook Purchase for a payment with no stored consent")

	credited, err = creditPaymentByExternalID("inv_granted")
	if err != nil || !credited {
		t.Fatalf("creditPaymentByExternalID(granted) = %v, %v; want true, nil", credited, err)
	}
	waitHit(t, hits, "webhook Purchase after consent")
}

// The checkout form used to fire InitiateCheckout straight through fbq, which
// would throw once the pixel is no longer loaded on page load.
func TestCheckoutFiresThroughTheGate(t *testing.T) {
	if strings.Contains(checkoutHTML, "fbq(") {
		t.Error("checkout still calls fbq directly")
	}
	if !strings.Contains(checkoutHTML, "nskTrack('InitiateCheckout'") {
		t.Error("checkout no longer tracks InitiateCheckout at all")
	}
}
