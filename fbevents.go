package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Meta Pixel / Events Manager integration for netsekurity.com.
// Client events track landing & UI actions; Purchase is fired SERVER-SIDE from the
// Xendit webhook so it is accurate (value + currency) and not lost to ad-blockers.

const fbPixelID = "3401314126707507"
const fbAPIBase = "https://graph.facebook.com/v21.0"

// fbToken reads the Meta Conversions API access token. It must stay a function:
// package-level vars are initialised before main() runs loadEnv(), so a
// `var fbToken = getenv(...)` would never see a value coming from .env and every
// server-side event would be silently skipped. Empty token = skip server calls.
func fbToken() string { return getenv("META_CAPI_TOKEN", "") }

// fbClient bounds every Conversions API call. sendMetaEvent runs on the login
// and payment-webhook paths, so an unbounded client would hang them.
var fbClient = &http.Client{Timeout: 5 * time.Second}

// metaPixelSnippet returns the <head> Meta Pixel base script (init + fbq.queue).
// PageView is fired by default unless pageView=false (callers that fire it manually).
func metaPixelSnippet(pageView bool) string {
	pv := ""
	if pageView {
		pv = "fbq('track', 'PageView');"
	}
	return fmt.Sprintf(`<script>
!function(f,b,e,v,n,t,s)
{if(f.fbq)return;n=f.fbq=function(){n.callMethod?
n.callMethod.apply(n,arguments):n.queue.push(arguments)};
if(!f._fbq)f._fbq=n;n.push=n;n.loaded=!0;n.version='2.0';
n.queue=[];t=b.createElement(e);t.async=!0;
t.src=v;s=b.getElementsByTagName(e)[0];s.parentNode.insertBefore(t,s)}(window,
document,'script','https://connect.facebook.net/en_US/fbevents.js');
fbq('init', '%s');
%s</script>
<noscript><img height="1" width="1" style="display:none"
src="https://www.facebook.com/tr?id=%s&ev=PageView&noscript=1"/></noscript>
`, fbPixelID, pv, fbPixelID)
}

// metaEventJS returns a short inline script that fires an fbq client event.
func metaEventJS(event string, extra string) string {
	if extra != "" {
		return fmt.Sprintf("<script>fbq('track', '%s', %s);</script>", event, extra)
	}
	return fmt.Sprintf("<script>fbq('track', '%s');</script>", event)
}

// metaUser carries the identifiers Meta requires on a Conversions API event.
// At least one must be present or the API rejects the whole call with HTTP 400.
type metaUser struct {
	Email     string
	IP        string
	UserAgent string
	FBP       string // _fbp cookie, set by the pixel
	FBC       string // _fbc cookie, carries the ad click id
}

// metaUserFrom captures the browser context of a live request. The webhook path
// has no request of its own, so it replays what was stored at checkout instead.
func metaUserFrom(r *http.Request, email string) metaUser {
	u := metaUser{Email: email, IP: clientIP(r), UserAgent: r.UserAgent()}
	if c, err := r.Cookie("_fbp"); err == nil {
		u.FBP = c.Value
	}
	if c, err := r.Cookie("_fbc"); err == nil {
		u.FBC = c.Value
	}
	return u
}

// payload renders the user_data block. Email is SHA-256 hashed as Meta requires;
// cookies, IP and user agent are sent raw.
func (u metaUser) payload() map[string]interface{} {
	m := map[string]interface{}{}
	if u.Email != "" {
		m["em"] = []string{fbHash(u.Email)}
	}
	if u.IP != "" {
		m["client_ip_address"] = u.IP
	}
	if u.UserAgent != "" {
		m["client_user_agent"] = u.UserAgent
	}
	if u.FBP != "" {
		m["fbp"] = u.FBP
	}
	if u.FBC != "" {
		m["fbc"] = u.FBC
	}
	return m
}

// fbHash normalises and SHA-256 hashes a PII field, the format Meta expects.
func fbHash(s string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(s))))
	return hex.EncodeToString(sum[:])
}

// clientIP returns the caller's address, honouring the Cloudflare / reverse
// proxy headers this site sits behind.
func clientIP(r *http.Request) string {
	if v := r.Header.Get("CF-Connecting-IP"); v != "" {
		return strings.TrimSpace(v)
	}
	if v := r.Header.Get("X-Forwarded-For"); v != "" {
		if i := strings.Index(v, ","); i > 0 {
			v = v[:i]
		}
		return strings.TrimSpace(v)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// metaEvent is one Conversions API event.
type metaEvent struct {
	Name      string                 // Meta event name, e.g. "Purchase"
	ID        string                 // stable dedup key; must match the client event id when both fire
	SourceURL string                 // page the action happened on
	User      metaUser               // required identifiers
	Custom    map[string]interface{} // value, currency, content_name, ...
}

// sendMetaEvent posts an event to the Meta Conversions API (used for Purchase /
// Lead so they survive ad-blockers and carry accurate value+currency). It blocks
// for up to fbClient.Timeout, so callers on a request path should use `go`.
func sendMetaEvent(ev metaEvent) {
	token := fbToken()
	if token == "" {
		log.Printf("fbevents: META_CAPI_TOKEN not set, skipping server-side %s", ev.Name)
		return
	}
	user := ev.User.payload()
	if len(user) == 0 {
		log.Printf("fbevents: no user identifiers for %s, skipping (Meta would reject it)", ev.Name)
		return
	}
	if ev.SourceURL == "" {
		ev.SourceURL = "https://netsekurity.com/"
	}
	event := map[string]interface{}{
		"event_name":       ev.Name,
		"event_time":       epoch(),
		"action_source":    "website",
		"event_source_url": ev.SourceURL,
		"user_data":        user,
		"custom_data":      ev.Custom,
	}
	if ev.ID != "" {
		event["event_id"] = ev.ID
	}
	body, err := json.Marshal(map[string]interface{}{"data": []map[string]interface{}{event}})
	if err != nil {
		log.Printf("fbevents: marshal %s: %v", ev.Name, err)
		return
	}
	endpoint := fmt.Sprintf("%s/%s/events?access_token=%s", fbAPIBase, fbPixelID, url.QueryEscape(token))
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := fbClient.Do(req)
	if err != nil {
		log.Printf("fbevents: send %s: %v", ev.Name, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		log.Printf("fbevents: %s -> HTTP %d: %s", ev.Name, resp.StatusCode, strings.TrimSpace(string(msg)))
		return
	}
	io.Copy(io.Discard, resp.Body)
}

// epoch returns the current unix timestamp in seconds.
func epoch() int64 {
	return time.Now().Unix()
}
