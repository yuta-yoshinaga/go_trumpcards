package recover_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	rec "github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/recover"
)

func TestMiddleware_RecoversPanicAsJSON500(t *testing.T) {
	h := rec.Middleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("boom")
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/sevens/exec", strings.NewReader(`{}`))
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status: got %d, want %d", rr.Code, http.StatusInternalServerError)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("content-type: got %q, want application/json", got)
	}
	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v (raw=%q)", err, rr.Body.String())
	}
	if !strings.Contains(body["message"], "boom") {
		t.Errorf("message should mention panic value, got %q", body["message"])
	}
}

func TestMiddleware_PassesThroughNonPanic(t *testing.T) {
	h := rec.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/sevens/exec", strings.NewReader(`{}`))
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", rr.Code)
	}
	if rr.Body.String() != `{"ok":true}` {
		t.Errorf("body: got %q, want %q", rr.Body.String(), `{"ok":true}`)
	}
}
