package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
)

// sentryFlushTimeout bounds how long shutdown waits for queued events.
const sentryFlushTimeout = 2 * time.Second

// initSentry configures the SDK as early as possible in main. Reporting is
// error-only: performance tracing and profiling stay off (traces_sample_rate 0),
// so a busy site costs the Sentry instance nothing. An empty DSN disables the
// SDK (it then becomes a no-op client).
func initSentry() {
	dsn := sentryDSN()
	if dsn == "" {
		log.Printf("sentry: disabled (SENTRY_DSN empty)")
		return
	}
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Environment:      sentryEnvironment(),
		Release:          sentryRelease(),
		AttachStacktrace: true,
		TracesSampleRate: 0,
		SendDefaultPII:   false,
		BeforeSend:       beforeSend,
	})
	if err != nil {
		log.Printf("WARNING: sentry init failed: %v", err)
		return
	}
	log.Printf("sentry: reporting errors (env: %s)", sentryEnvironment())
}

// flushSentry drains the transport queue; call it on every exit path.
func flushSentry() { sentry.Flush(sentryFlushTimeout) }

// reportFatal captures a startup error and flushes before log.Fatalf exits the
// process (log.Fatalf calls os.Exit, which would skip the deferred flush).
func reportFatal(err error) {
	if err == nil {
		return
	}
	sentry.CurrentHub().CaptureException(err)
	flushSentry()
}

// noisyErrorPatterns are failures that say nothing about our code: a client
// hanging up or a request context being cancelled mid-flight.
var noisyErrorPatterns = []string{
	"context canceled",
	"context deadline exceeded",
	"broken pipe",
	"connection reset by peer",
	"client disconnected",
}

// eofPattern catches io.EOF ("EOF", "unexpected EOF") without matching words
// that merely contain the letters.
var eofPattern = regexp.MustCompile(`(?i)\beof\b`)

// sensitiveKeys are never sent to the Sentry instance.
func sensitiveKey(key string) bool {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "cookie", "cookies", "set-cookie", "authorization", "auth_token":
		return true
	}
	return false
}

// beforeSend drops noise and strips secrets before an event leaves the process.
func beforeSend(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
	if event == nil {
		return nil
	}
	if hint != nil && hint.RecoveredException == http.ErrAbortHandler {
		return nil // client went away mid-response; net/http suppresses these too
	}
	if isNoisyEvent(event) {
		return nil
	}
	scrubEvent(event)
	return event
}

// isNoisyEvent reports whether every message in the event matches a known-noise
// pattern. A real error sharing the same event must still be reported.
func isNoisyEvent(event *sentry.Event) bool {
	texts := make([]string, 0, len(event.Exception)+1)
	if event.Message != "" {
		texts = append(texts, event.Message)
	}
	for _, ex := range event.Exception {
		texts = append(texts, ex.Type+" "+ex.Value)
	}
	if len(texts) == 0 {
		return false
	}
	for _, t := range texts {
		if !noisyText(t) {
			return false
		}
	}
	return true
}

func noisyText(text string) bool {
	lower := strings.ToLower(text)
	for _, p := range noisyErrorPatterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return eofPattern.MatchString(text)
}

// scrubEvent removes cookies, authorization headers and auth-token fields from
// the request, its query string and any structured context or breadcrumb.
func scrubEvent(event *sentry.Event) {
	if req := event.Request; req != nil {
		for k := range req.Headers {
			if sensitiveKey(k) {
				delete(req.Headers, k)
			}
		}
		req.Cookies = ""
		req.QueryString = scrubQuery(req.QueryString)
		req.Data = scrubJSON(req.Data)
	}
	for _, ctx := range event.Contexts {
		scrubMap(ctx)
	}
	for _, bc := range event.Breadcrumbs {
		scrubMap(bc.Data)
	}
}

// scrubMap deletes sensitive keys recursively from a decoded JSON-like map.
func scrubMap(m map[string]interface{}) {
	for k, v := range m {
		if sensitiveKey(k) {
			delete(m, k)
			continue
		}
		switch val := v.(type) {
		case map[string]interface{}:
			scrubMap(val)
		case []interface{}:
			for _, item := range val {
				if sub, ok := item.(map[string]interface{}); ok {
					scrubMap(sub)
				}
			}
		}
	}
}

// scrubQuery drops sensitive query parameters, leaving the rest readable.
func scrubQuery(raw string) string {
	if raw == "" {
		return raw
	}
	values, err := url.ParseQuery(raw)
	if err != nil {
		return raw
	}
	dropped := false
	for k := range values {
		if sensitiveKey(k) {
			values.Del(k)
			dropped = true
		}
	}
	if !dropped {
		return raw
	}
	return values.Encode()
}

// scrubJSON assumes a JSON object body and removes sensitive keys from it. When
// the body is not a JSON object the string is returned unchanged.
func scrubJSON(data string) string {
	if data == "" {
		return data
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(data), &m); err != nil {
		return data
	}
	scrubMap(m)
	b, err := json.Marshal(m)
	if err != nil {
		return data
	}
	return string(b)
}

// statusWriter remembers whether the handler already started the response, so
// the recover middleware never writes a second header after a partial write.
// Unwrap keeps http.ResponseController working through the wrapper.
type statusWriter struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (w *statusWriter) WriteHeader(code int) {
	if !w.wrote {
		w.status = code
		w.wrote = true
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if !w.wrote {
		w.status = http.StatusOK
		w.wrote = true
	}
	return w.ResponseWriter.Write(b)
}

func (w *statusWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

// recoverPanics is the outermost handler: it turns a panic that escaped the
// router into a logged 500 instead of a killed connection. Reporting is already
// done by sentryhttp (wrapped underneath with Repanic), so this must not capture
// again; it only keeps the process serving.
func recoverPanics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			if rec == http.ErrAbortHandler {
				panic(rec) // net/http swallows this one on purpose
			}
			log.Printf("panic serving %s %s: %v", r.Method, r.URL.Path, rec)
			if !sw.wrote {
				http.Error(sw, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(sw, r)
	})
}

// withSentry wraps mux for panic reporting, then for crash safety.
func withSentry(mux http.Handler) http.Handler {
	return recoverPanics(sentryhttp.New(sentryhttp.Options{Repanic: true}).Handle(mux))
}

// captureMainPanic reports a panic that reached the main goroutine and flushes
// before the process dies.
func captureMainPanic(rec interface{}) {
	sentry.CurrentHub().RecoverWithContext(context.Background(), rec)
	flushSentry()
}
