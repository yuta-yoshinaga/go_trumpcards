//go:build test

package controller_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// recorded wraps httptest.ResponseRecorder with assertion helpers
// that mirror the go-json-rest/rest/test API for minimal test migration.
type recorded struct {
	*httptest.ResponseRecorder
	t *testing.T
}

// execRequest sends a POST request to the given handler and returns
// a recorded response with assertion helpers.
func execRequest(t *testing.T, handler http.HandlerFunc, body any) *recorded {
	t.Helper()
	var reader io.Reader
	switch v := body.(type) {
	case io.Reader:
		reader = v
	default:
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}
		reader = bytes.NewReader(b)
	}
	req := httptest.NewRequest("POST", "/exec", reader)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler(rec, req)
	return &recorded{ResponseRecorder: rec, t: t}
}

// CodeIs asserts that the response status code matches expected.
func (r *recorded) CodeIs(expected int) {
	r.t.Helper()
	if r.Code != expected {
		r.t.Errorf("expected status %d, got %d; body=%s", expected, r.Code, r.Body.String())
	}
}

// BodyIs asserts that the trimmed response body matches expected.
func (r *recorded) BodyIs(expected string) {
	r.t.Helper()
	actual := strings.TrimSpace(r.Body.String())
	if actual != expected {
		r.t.Errorf("body mismatch:\n  want: %s\n  got:  %s", expected, actual)
	}
}

// ContentTypeIsJson asserts the Content-Type header is application/json.
func (r *recorded) ContentTypeIsJson() {
	r.t.Helper()
	ct := r.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		r.t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}
