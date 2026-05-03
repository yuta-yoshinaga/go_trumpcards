package recoverymw_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/recoverymw"
)

func TestMiddleware_RecoversPanicAsJSON500(t *testing.T) {
	h := recoverymw.Middleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
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
	h := recoverymw.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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

// TestMiddleware_RecoversPanicAfterPartialWrite documents the known limitation
// that when a handler commits headers (or writes body bytes) before panicking,
// the recovery middleware cannot rewrite the status line. The deferred
// WriteHeader is a no-op, so the original 200 status is preserved, and the
// JSON error body is appended to whatever the handler already wrote — yielding
// a syntactically invalid response. We still prefer this over letting the
// panic escape, because the alternative is Cloudflare's 1101 HTML page with no
// CORS headers, which the browser cannot read at all.
func TestMiddleware_RecoversPanicAfterPartialWrite(t *testing.T) {
	h := recoverymw.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"partial":true`))
		panic("boom-after-write")
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/sevens/exec", strings.NewReader(`{}`))

	// The middleware must not let the panic propagate.
	defer func() {
		if rec := recover(); rec != nil {
			t.Fatalf("panic escaped middleware: %v", rec)
		}
	}()
	h.ServeHTTP(rr, req)

	// Original 200 status is preserved (httptest.ResponseRecorder honours the
	// first WriteHeader call, matching the production net/http behaviour).
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200 (already committed by handler)", rr.Code)
	}
	// The recovery body is appended to the partial write.
	body := rr.Body.String()
	if !strings.HasPrefix(body, `{"partial":true`) {
		t.Errorf("partial write should be preserved, got %q", body)
	}
	if !strings.Contains(body, "boom-after-write") {
		t.Errorf("recovery JSON should still be appended with panic value, got %q", body)
	}
}
