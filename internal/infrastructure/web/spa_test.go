package web

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// makeSpaFixture creates a temp public dir with index.html and one asset, and
// returns the dir path. Tests use this to drive spaFallbackHandler against a
// realistic on-disk layout.
func makeSpaFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html>SPA</html>"), 0o644); err != nil {
		t.Fatalf("write index.html: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0o755); err != nil {
		t.Fatalf("mkdir assets: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "assets", "app.js"), []byte("console.log('app');"), 0o644); err != nil {
		t.Fatalf("write app.js: %v", err)
	}
	return dir
}

func TestSpaFallbackHandler(t *testing.T) {
	dir := makeSpaFixture(t)
	h := spaFallbackHandler(http.FileServer(http.Dir(dir)), dir)

	tests := []struct {
		name       string
		method     string
		target     string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "existing file is served as-is",
			method:     http.MethodGet,
			target:     "/assets/app.js",
			wantStatus: http.StatusOK,
			wantBody:   "console.log('app');",
		},
		{
			name:       "root serves index.html via underlying FileServer",
			method:     http.MethodGet,
			target:     "/",
			wantStatus: http.StatusOK,
			wantBody:   "<html>SPA</html>",
		},
		{
			name:       "unknown deep-link falls back to index.html",
			method:     http.MethodGet,
			target:     "/blackjack",
			wantStatus: http.StatusOK,
			wantBody:   "<html>SPA</html>",
		},
		{
			name:       "nested unknown path also falls back",
			method:     http.MethodGet,
			target:     "/some/random/path",
			wantStatus: http.StatusOK,
			wantBody:   "<html>SPA</html>",
		},
		{
			name:       "HEAD on unknown path falls back without body",
			method:     http.MethodHead,
			target:     "/holdem",
			wantStatus: http.StatusOK,
			wantBody:   "",
		},
		{
			name:       "POST is delegated to FileServer instead of falling back",
			method:     http.MethodPost,
			target:     "/something",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "path traversal is neutralised: ../../etc/passwd cannot escape publicDir",
			method:     http.MethodGet,
			target:     "/../../etc/passwd",
			wantStatus: http.StatusOK,
			wantBody:   "<html>SPA</html>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.target, nil)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			body, _ := io.ReadAll(rec.Body)
			if tt.wantBody != "" && !strings.Contains(string(body), tt.wantBody) {
				t.Fatalf("body %q does not contain %q", string(body), tt.wantBody)
			}
		})
	}
}
