package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterSwaggerRoutes_MethodHandling(t *testing.T) {
	mux := http.NewServeMux()
	RegisterSwaggerRoutes(mux)

	tests := []struct {
		name        string
		method      string
		target      string
		wantStatus  int
		wantContent string // when set, asserts the Content-Type header prefix
	}{
		{
			name:        "GET /swagger/openapi.yaml returns spec",
			method:      http.MethodGet,
			target:      "/swagger/openapi.yaml",
			wantStatus:  http.StatusOK,
			wantContent: "application/yaml",
		},
		{
			name:        "HEAD /swagger/openapi.yaml returns headers (no body) without 405",
			method:      http.MethodHead,
			target:      "/swagger/openapi.yaml",
			wantStatus:  http.StatusOK,
			wantContent: "application/yaml",
		},
		{
			name:       "POST /swagger/openapi.yaml is rejected as 405",
			method:     http.MethodPost,
			target:     "/swagger/openapi.yaml",
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:        "GET /swagger/ returns Swagger UI HTML",
			method:      http.MethodGet,
			target:      "/swagger/",
			wantStatus:  http.StatusOK,
			wantContent: "text/html",
		},
		{
			name:        "HEAD /swagger/ returns headers (no body) without 405",
			method:      http.MethodHead,
			target:      "/swagger/",
			wantStatus:  http.StatusOK,
			wantContent: "text/html",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.target, nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantContent != "" {
				ct := rec.Header().Get("Content-Type")
				if got := ct; len(got) < len(tt.wantContent) || got[:len(tt.wantContent)] != tt.wantContent {
					t.Fatalf("Content-Type = %q, want prefix %q", ct, tt.wantContent)
				}
			}
			if tt.method == http.MethodHead && rec.Body.Len() != 0 {
				t.Fatalf("HEAD body should be empty, got %d bytes", rec.Body.Len())
			}
			// Successful GET/HEAD must advertise Content-Length so reverse
			// proxies and download estimators don't have to fetch the body.
			if rec.Code == http.StatusOK {
				if cl := rec.Header().Get("Content-Length"); cl == "" || cl == "0" {
					t.Fatalf("Content-Length missing or zero on %s %s, got %q", tt.method, tt.target, cl)
				}
			}
		})
	}
}
