package main

import (
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestHandleIndex_InjectsClientIDAndFooter(t *testing.T) {
	os.Setenv("GOOGLE_CLIENT_ID", "test-client.apps.googleusercontent.com")
	rr := httptest.NewRecorder()
	handleIndex(rr, httptest.NewRequest("GET", "/", nil))
	if rr.Code != 200 {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, `NSK_CLIENT_ID = "test-client.apps.googleusercontent.com"`) {
		t.Errorf("One Tap client id not injected")
	}
	if !strings.Contains(body, "Dalang Pte Ltd") {
		t.Errorf("footer company name missing")
	}
}

func TestHandleIndex_UnknownPath404(t *testing.T) {
	rr := httptest.NewRecorder()
	handleIndex(rr, httptest.NewRequest("GET", "/nope", nil))
	if rr.Code != 404 {
		t.Fatalf("status = %d, want 404", rr.Code)
	}
}
