package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
)

// captureTransport keeps events in memory so tests can assert on what would
// have been sent, without any network.
type captureTransport struct {
	mu     sync.Mutex
	events []*sentry.Event
}

func (c *captureTransport) Configure(sentry.ClientOptions) {}

func (c *captureTransport) SendEvent(event *sentry.Event) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = append(c.events, event)
}

func (c *captureTransport) Flush(time.Duration) bool              { return true }
func (c *captureTransport) FlushWithContext(context.Context) bool { return true }
func (c *captureTransport) Close()                                {}

func (c *captureTransport) snapshot() []*sentry.Event {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]*sentry.Event(nil), c.events...)
}

// bindTestSentry swaps the current hub's client for one that captures events,
// and restores the previous client when the test ends.
func bindTestSentry(t *testing.T) *captureTransport {
	t.Helper()
	transport := &captureTransport{}
	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:         "https://public@example.com/1",
		Environment: "test",
		BeforeSend:  beforeSend,
		Transport:   transport,
	})
	if err != nil {
		t.Fatalf("new sentry client: %v", err)
	}
	hub := sentry.CurrentHub()
	previous := hub.Client()
	hub.BindClient(client)
	t.Cleanup(func() {
		hub.BindClient(previous)
		client.Close()
	})
	return transport
}

func TestRecoverPanicsReportsAndReturns500(t *testing.T) {
	transport := bindTestSentry(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/boom", func(http.ResponseWriter, *http.Request) {
		panic("kaboom")
	})

	rec := httptest.NewRecorder()
	withSentry(mux).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/boom", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	events := transport.snapshot()
	if len(events) != 1 {
		t.Fatalf("captured %d events, want 1", len(events))
	}
	// The SDK records a recovered panic either as an exception or as the event
	// message ("kaboom"), depending on how the panic was caught.
	described := events[0].Message
	for _, ex := range events[0].Exception {
		described += " " + ex.Type + " " + ex.Value
	}
	if !strings.Contains(described, "kaboom") {
		t.Errorf("event does not describe the panic: %q", described)
	}
	if events[0].Level != sentry.LevelFatal && events[0].Level != sentry.LevelError {
		t.Errorf("level = %q, want error or fatal", events[0].Level)
	}
}

func TestRecoverPanicsKeepsQuietOnErrAbortHandler(t *testing.T) {
	transport := bindTestSentry(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/abort", func(http.ResponseWriter, *http.Request) {
		panic(http.ErrAbortHandler)
	})

	func() {
		defer func() {
			if rec := recover(); rec != http.ErrAbortHandler {
				t.Fatalf("recovered %v, want http.ErrAbortHandler to propagate", rec)
			}
		}()
		withSentry(mux).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/abort", nil))
	}()

	if events := transport.snapshot(); len(events) != 0 {
		t.Fatalf("captured %d events for an aborted response, want 0", len(events))
	}
}

func TestBeforeSendScrubsSecrets(t *testing.T) {
	event := &sentry.Event{
		Request: &sentry.Request{
			Headers: map[string]string{
				"Cookie":        "session=abc",
				"Authorization": "Bearer xyz",
				"User-Agent":    "test-agent",
			},
			Cookies:     "session=abc",
			QueryString: "auth_token=secret&page=2",
			Data:        `{"auth_token":"secret","email":"a@b.c"}`,
		},
		Contexts: map[string]sentry.Context{
			"state": {"cookies": "session=abc", "auth_token": "secret", "path": "/dashboard"},
		},
		Breadcrumbs: []*sentry.Breadcrumb{{
			Data: map[string]interface{}{"auth_token": "secret", "url": "/dashboard"},
		}},
	}

	got := beforeSend(event, nil)
	if got == nil {
		t.Fatal("beforeSend dropped an ordinary event")
	}
	for _, header := range []string{"Cookie", "Authorization"} {
		if _, ok := got.Request.Headers[header]; ok {
			t.Errorf("header %q was not removed", header)
		}
	}
	if got.Request.Headers["User-Agent"] != "test-agent" {
		t.Errorf("User-Agent was scrubbed: %v", got.Request.Headers)
	}
	if got.Request.Cookies != "" {
		t.Errorf("cookies field = %q, want empty", got.Request.Cookies)
	}
	if strings.Contains(got.Request.QueryString, "secret") {
		t.Errorf("query string still has the token: %q", got.Request.QueryString)
	}
	if !strings.Contains(got.Request.QueryString, "page=2") {
		t.Errorf("query string lost benign params: %q", got.Request.QueryString)
	}
	if strings.Contains(got.Request.Data, "secret") || strings.Contains(got.Request.Data, "auth_token") {
		t.Errorf("request body still has the token: %q", got.Request.Data)
	}
	if !strings.Contains(got.Request.Data, "a@b.c") {
		t.Errorf("request body lost benign fields: %q", got.Request.Data)
	}
	if _, ok := got.Contexts["state"]["auth_token"]; ok {
		t.Error("context auth_token was not removed")
	}
	if got.Contexts["state"]["path"] != "/dashboard" {
		t.Error("context path was scrubbed")
	}
	if _, ok := got.Breadcrumbs[0].Data["auth_token"]; ok {
		t.Error("breadcrumb auth_token was not removed")
	}
	if got.Breadcrumbs[0].Data["url"] != "/dashboard" {
		t.Error("breadcrumb url was scrubbed")
	}
}

func TestBeforeSendDropsNoise(t *testing.T) {
	noise := []string{
		"context canceled",
		"write tcp 1.2.3.4:443: broken pipe",
		"read tcp 1.2.3.4:443: connection reset by peer",
		"client disconnected",
		"unexpected EOF",
	}
	for _, msg := range noise {
		event := &sentry.Event{Exception: []sentry.Exception{{Type: "*errors.errorString", Value: msg}}}
		if got := beforeSend(event, nil); got != nil {
			t.Errorf("noise %q was not dropped", msg)
		}
	}

	real := &sentry.Event{Exception: []sentry.Exception{{Type: "*errors.errorString", Value: "database is locked"}}}
	if got := beforeSend(real, nil); got == nil {
		t.Error("a real error was dropped")
	}

	abort := &sentry.Event{}
	if got := beforeSend(abort, &sentry.EventHint{RecoveredException: http.ErrAbortHandler}); got != nil {
		t.Error("http.ErrAbortHandler was reported")
	}
}

func TestSentryEnvDefaults(t *testing.T) {
	want := "https://bd3ff6f325012d29eacd074457aba1ea@mata.dalang.io/8"
	if got := sentryDSN(); got != want {
		t.Errorf("sentryDSN() = %q, want %q", got, want)
	}

	// An explicitly empty value is how a deployment opts out.
	env["SENTRY_DSN"] = ""
	t.Cleanup(func() { delete(env, "SENTRY_DSN") })
	if got := sentryDSN(); got != "" {
		t.Errorf("sentryDSN() with explicit empty = %q, want empty", got)
	}
}
