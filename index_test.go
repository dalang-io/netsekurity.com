package main

import (
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// The marketing pages no longer load Google One Tap: it is a third-party script
// and an unrequested pop-up on a page read by a board or a procurement team. The
// login button remains.
func TestHandleIndex_NoOneTapAndNamesTheOperator(t *testing.T) {
	os.Setenv("GOOGLE_CLIENT_ID", "test-client.apps.googleusercontent.com")
	rr := httptest.NewRecorder()
	handleIndex(rr, httptest.NewRequest("GET", "/", nil))
	if rr.Code != 200 {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	body := rr.Body.String()
	if strings.Contains(body, "accounts.google.com/gsi/client") || strings.Contains(body, "test-client.apps.googleusercontent.com") {
		t.Errorf("One Tap is back on the landing page")
	}
	if !strings.Contains(body, "Dalang Pte. Ltd.") {
		t.Errorf("operator name missing")
	}
}

func TestHandleIndex_UnknownPath404(t *testing.T) {
	rr := httptest.NewRecorder()
	handleIndex(rr, httptest.NewRequest("GET", "/nope", nil))
	if rr.Code != 404 {
		t.Fatalf("status = %d, want 404", rr.Code)
	}
}
