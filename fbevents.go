package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

// Meta Pixel / Events Manager integration for netsekurity.com.
// Client events track landing & UI actions; Purchase is fired SERVER-SIDE from the
// Xendit webhook so it is accurate (value + currency) and not lost to ad-blockers.

const fbPixelID = "3401314126707507"
const fbAPIBase = "https://graph.facebook.com/v21.0"

// fbToken is the Meta Conversions API access token. Leave empty to skip server calls.
var fbToken = getenv("META_CAPI_TOKEN", "")

// fbConsent holds a consent flag for data-processing (optional; GDPR-basic default on).
// You can set META_CONSENT=auto for default-consent or wire an explicit banner later.
var fbConsent = getenv("META_CONSENT", "auto")

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

// metaLeadResp is the JSON Meta returns for a Conversions API call.
type metaLeadResp struct{}

// sendMetaEvent server-side sends an event to the Meta Conversions API (used for
// Purchase / Lead so they survive ad-blockers and carry accurate value+currency).
func sendMetaEvent(eventName string, params map[string]interface{}) {
	if fbToken == "" {
		log.Printf("fbevents: META_CAPI_TOKEN not set, skipping server-side %s", eventName)
		return
	}
	// User data: minimal browser/IP is optional for server-side; we send event only.
	payload := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"event_name":       eventName,
				"event_time":       epoch(),
				"action_source":    "website",
				"event_source_url": "https://netsekurity.com/",
				"custom_data":      params,
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("fbevents: marshal %s: %v", eventName, err)
		return
	}
	endpoint := fmt.Sprintf("%s/%s/events?access_token=%s", fbAPIBase, fbPixelID, url.QueryEscape(fbToken))
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("fbevents: send %s: %v", eventName, err)
		return
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		log.Printf("fbevents: %s -> HTTP %d", eventName, resp.StatusCode)
	}
}

// epoch returns the current unix timestamp in seconds.
func epoch() int64 {
	return time.Now().Unix()
}
