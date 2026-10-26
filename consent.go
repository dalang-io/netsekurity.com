package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Consent gate for the marketing tracker stack.
//
// Owner decision (2026-10-16): nothing on netsekurity.com may talk to a tracker
// before the visitor agrees — the rule dalang.io already enforces with
// src/lib/analytics/consent.js and its `dalang_consent` key. This is the
// netsekurity port of it, same idea, Go + hand-written HTML instead of Svelte.
//
// Gated (all of it hangs off the Meta Pixel):
//   - connect.facebook.net/en_US/fbevents.js — the pixel library itself,
//   - connect.facebook.net/signals/config/<id> — the pixel config,
//   - POST www.facebook.com/tr/ — the beacon fired by the pixel,
//   - POST *.on.aws/events — the Conversions API Gateway the pixel config routes
//     to (Meta-side setting, not ours — it follows the pixel automatically),
//   - the server-side Conversions API calls (Meta Lead / Purchase).
//
// Deliberately NOT gated: Umami (umami.dalang.io). Cookieless, no personal data,
// the same exemption dalang.io documents in src/lib/analytics/consent.js. The
// cookie banner says so instead of pretending it covers everything.
//
// The choice lives in the `nsk_consent` cookie rather than localStorage (which
// is what dalang.io uses) because the server needs it too: the Xendit webhook
// settles a payment with no browser attached, and it must still be able to tell
// whether that buyer agreed to be measured. A cookie is readable by the page,
// by Go, and — if the Owner wants it — by a Cloudflare rule.
const consentCookieName = "nsk_consent"

const (
	consentGrantedValue = "granted"
	consentDeniedValue  = "denied"
)

// consentGranted reports whether this request carries an explicit "granted"
// choice. Absent, "denied" or anything unrecognised all mean no consent, so a
// missing cookie can never be read as permission.
func consentGranted(r *http.Request) bool {
	c, err := r.Cookie(consentCookieName)
	return err == nil && c.Value == consentGrantedValue
}

// consentGate returns the consent banner and the deferred Meta Pixel loader.
//
// It replaces the unconditional metaPixelSnippet: the served HTML never contains
// a <script src="connect.facebook.net/...">, so that request cannot happen
// before consent. The pixel base code travels as a JSON string and is only put
// into the document — as a real <script>, not eval() — once the visitor accepts.
//
// pageView=true fires the PageView event as soon as the pixel loads (i.e. on
// accept, or on the next page load for a visitor who already accepted).
func consentGate(pageView bool) string {
	base, err := json.Marshal(metaPixelBaseJS())
	if err != nil {
		// metaPixelBaseJS is a compile-time constant; the only way this fails is
		// a Go bug, and shipping an unparseable gate would be worse than no pixel.
		base = []byte(`""`)
	}
	pv := "false"
	if pageView {
		pv = "true"
	}
	s := strings.ReplaceAll(consentGateHTML, "__PIXEL_BASE_JSON__", string(base))
	return strings.ReplaceAll(s, "__PAGEVIEW__", pv)
}

// metaTrackJS returns an inline <script> that fires one Meta client event. It
// calls nskTrack (defined by consentGate) instead of fbq directly, so the event
// is dropped when the visitor has not agreed — or when the pixel is not loaded
// yet — rather than throwing on an undefined fbq.
func metaTrackJS(event string, params string) string {
	if params == "" {
		return fmt.Sprintf("<script>nskTrack('%s');</script>", event)
	}
	return fmt.Sprintf("<script>nskTrack('%s', %s);</script>", event, params)
}

// consentGateHTML is the gate itself. Styles are inline on purpose: the Tailwind
// bundle is built from the .html files, so classes invented here would ship
// unstyled.
const consentGateHTML = `<script>
(function () {
  var KEY = 'nsk_consent';
  // The Meta Pixel base code. A string until consent, never a script tag in the
  // document — see consent.go for why that distinction is the whole point.
  var BASE = __PIXEL_BASE_JSON__;
  var PAGEVIEW = __PAGEVIEW__;
  var loaded = false;

  function read() {
    try {
      var m = document.cookie.match(new RegExp('(?:^|;\\s*)' + KEY + '=([^;]*)'));
      return m ? decodeURIComponent(m[1]) : null;
    } catch (e) { return null; }
  }

  function write(v) {
    var d = new Date();
    d.setTime(d.getTime() + 15552000000); // 180 days
    document.cookie = KEY + '=' + v + '; expires=' + d.toUTCString() + '; path=/; SameSite=Lax' +
      (location.protocol === 'https:' ? '; Secure' : '');
  }

  function load() {
    if (loaded || read() !== 'granted') return;
    loaded = true;
    var s = document.createElement('script');
    s.textContent = BASE;
    (document.head || document.documentElement).appendChild(s);
    if (PAGEVIEW && typeof window.fbq === 'function') window.fbq('track', 'PageView');
  }

  // Every fbq() call in the templates goes through this. Before consent it is a
  // no-op: the old inline onclick="fbq(...)" handlers would have thrown (fbq is
  // undefined) or, on a page that still loaded the pixel, sent the event.
  window.nskTrack = function (name, params) {
    if (!loaded || typeof window.fbq !== 'function') return;
    if (params) { window.fbq('track', name, params); } else { window.fbq('track', name); }
  };

  window.nskConsent = {
    state: function () { return read(); },
    grant: function () { write('granted'); hide(); load(); },
    deny: function () { write('denied'); hide(); }
  };

  function show() {
    var el = document.getElementById('nsk-consent');
    if (el) el.hidden = false;
  }

  function hide() {
    var el = document.getElementById('nsk-consent');
    if (el) el.hidden = true;
  }

  function banner() {
    var el = document.createElement('div');
    el.id = 'nsk-consent';
    el.hidden = true;
    el.setAttribute('role', 'region');
    el.setAttribute('aria-label', 'cookie and tracking consent');
    // The signed-in app (body.nsk-app) is a light theme; everything else keeps
    // the terminal look. Same words, same two choices, in both.
    var light = !!(document.body && document.body.classList.contains('nsk-app'));
    var FONT = light ? '"Plus Jakarta Sans",ui-sans-serif,system-ui,sans-serif' : 'ui-monospace,SFMono-Regular,Menlo,Consolas,monospace';
    el.style.cssText = 'position:fixed;left:0;right:0;bottom:0;z-index:2147483000;padding:12px;font-family:' + FONT + ';font-size:' + (light ? '14px' : '12px');
    var box = light
      ? 'max-width:56rem;margin:0 auto;border:2px solid #000;background:#fff;color:#374151;padding:16px;box-shadow:4px 4px 0 0 #000'
      : 'max-width:56rem;margin:0 auto;border:1px solid rgba(16,185,129,.55);background:#0b0f14;color:#9ca3af;padding:16px;box-shadow:0 12px 40px rgba(0,0,0,.65)';
    var head = light ? 'margin:0 0 8px;color:#111827;font-weight:800;font-size:16px' : 'margin:0 0 8px;color:#fff;font-weight:700;font-size:13px';
    // In the light theme the two buttons carry equal weight: refusing must be as
    // easy to see and press as accepting.
    var btnNo = light
      ? 'cursor:pointer;border:2px solid #1c1a17;background:#fff;color:#111827;padding:10px 20px;font:inherit;font-weight:800;box-shadow:3px 3px 0 0 #1c1a17'
      : 'cursor:pointer;border:1px solid rgba(255,255,255,.25);background:transparent;color:#d1d5db;padding:8px 16px;font:inherit';
    var btnYes = light
      ? 'cursor:pointer;border:2px solid #1c1a17;background:#facc15;color:#000;padding:10px 20px;font:inherit;font-weight:800;box-shadow:3px 3px 0 0 #1c1a17'
      : 'cursor:pointer;border:1px solid #34d399;background:rgba(16,185,129,.15);color:#6ee7b7;padding:8px 16px;font:inherit;font-weight:700';
    el.innerHTML =
      '<div style="' + box + '">' +
        '<p style="' + head + '">' + (light ? 'Cookies &amp; tracking' : 'cookies &amp; tracking') + '</p>' +
        '<p style="margin:0 0 12px;line-height:1.65">' +
          'The Meta (Facebook) pixel measures which ads bring customers. It is not loaded, and nothing is sent to Meta, until you agree here. ' +
          'Our own analytics (Umami) is cookieless and stores nothing about you, so it stays on either way.' +
        '</p>' +
        '<div style="display:flex;gap:10px;flex-wrap:wrap">' +
          '<button type="button" data-nsk-consent="decline" style="' + btnNo + '">' + (light ? 'Decline' : 'decline') + '</button>' +
          '<button type="button" data-nsk-consent="grant" style="' + btnYes + '">' + (light ? 'Accept' : 'accept') + '</button>' +
        '</div>' +
      '</div>';
    (document.body || document.documentElement).appendChild(el);
    el.addEventListener('click', function (ev) {
      var b = ev.target.closest ? ev.target.closest('[data-nsk-consent]') : null;
      if (!b) return;
      if (b.getAttribute('data-nsk-consent') === 'grant') window.nskConsent.grant();
      else window.nskConsent.deny();
    });
    return el;
  }

  function start() {
    if (read() === 'granted') { load(); return; }   // repeat visitor: pixel loads with the page
    if (read() === 'denied') return;                // chose no: never load it, never ask again
    var mount = function () { banner(); show(); };
    if (document.body) { mount(); } else { document.addEventListener('DOMContentLoaded', mount); }
  }

  start();
})();
</script>
`
