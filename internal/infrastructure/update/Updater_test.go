package update

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetAutoConfirm(t *testing.T) {
	u := NewUpdater("1.0.0", nil, &bytes.Buffer{}, &bytes.Buffer{})
	assert.False(t, u.autoConfirm)
	u.SetAutoConfirm(true)
	assert.True(t, u.autoConfirm)
	u.SetAutoConfirm(false)
	assert.False(t, u.autoConfirm)
}

func TestExec_AlreadyLatest(t *testing.T) {
	release := ghRelease{TagName: "v1.2.3"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(release)
	}))
	defer srv.Close()

	out := &bytes.Buffer{}
	u := &Updater{
		currentVersion: "1.2.3",
		repoOwner:      "test",
		repoName:       "test",
		reader:         strings.NewReader(""),
		writer:         out,
		errWriter:      &bytes.Buffer{},
		httpClient:     srv.Client(),
	}
	// Override the URL by replacing the httpClient transport.
	u.httpClient = &http.Client{Transport: &rewriteTransport{base: srv.Client().Transport, url: srv.URL}}

	err := u.Exec()
	require.NoError(t, err)
	assert.Contains(t, out.String(), "v1.2.3")
}

func TestExec_UserCancels(t *testing.T) {
	release := ghRelease{TagName: "v2.0.0"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(release)
	}))
	defer srv.Close()

	out := &bytes.Buffer{}
	u := &Updater{
		currentVersion: "1.0.0",
		repoOwner:      "test",
		repoName:       "test",
		reader:         strings.NewReader("n\n"),
		writer:         out,
		errWriter:      &bytes.Buffer{},
		httpClient:     &http.Client{Transport: &rewriteTransport{base: srv.Client().Transport, url: srv.URL}},
	}

	err := u.Exec()
	require.NoError(t, err)
	assert.Contains(t, out.String(), "v2.0.0")
}

func TestExec_AutoConfirmSkipsPrompt(t *testing.T) {
	release := ghRelease{TagName: "v2.0.0", Assets: []ghAsset{}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(release)
	}))
	defer srv.Close()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	u := &Updater{
		currentVersion: "1.0.0",
		repoOwner:      "test",
		repoName:       "test",
		reader:         strings.NewReader(""), // no input — would hang without autoConfirm
		writer:         out,
		errWriter:      errOut,
		httpClient:     &http.Client{Transport: &rewriteTransport{base: srv.Client().Transport, url: srv.URL}},
		autoConfirm:    true,
	}

	// Exec will proceed past the prompt but fail at findAsset (no matching asset).
	// That's expected — we're testing that it skips the prompt.
	err := u.Exec()
	require.Error(t, err) // no asset for current platform
	// The output should NOT contain the "[y/N]" prompt.
	assert.NotContains(t, out.String(), "[y/N]")
	// The output should contain the auto-confirm message.
	assert.Contains(t, out.String(), "v2.0.0")
}

func TestExec_FetchError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	errOut := &bytes.Buffer{}
	u := &Updater{
		currentVersion: "1.0.0",
		repoOwner:      "test",
		repoName:       "test",
		reader:         strings.NewReader(""),
		writer:         &bytes.Buffer{},
		errWriter:      errOut,
		httpClient:     &http.Client{Transport: &rewriteTransport{base: srv.Client().Transport, url: srv.URL}},
	}

	err := u.Exec()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestFindAsset(t *testing.T) {
	u := NewUpdater("1.0.0", nil, &bytes.Buffer{}, &bytes.Buffer{})

	tests := []struct {
		name      string
		assets    []ghAsset
		wantEmpty bool
	}{
		{
			name:      "empty assets",
			assets:    nil,
			wantEmpty: true,
		},
		{
			name: "no matching asset",
			assets: []ghAsset{
				{Name: "trumpcards_1.0.0_freebsd_riscv.tar.gz", BrowserDownloadURL: "https://example.com/a"},
			},
			wantEmpty: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, url := u.findAsset(tt.assets)
			if tt.wantEmpty {
				assert.Empty(t, name)
				assert.Empty(t, url)
			} else {
				assert.NotEmpty(t, name)
				assert.NotEmpty(t, url)
			}
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, formatBytes(tt.input))
		})
	}
}

// rewriteTransport rewrites all request URLs to the test server.
type rewriteTransport struct {
	base http.RoundTripper
	url  string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	req.URL.Host = strings.TrimPrefix(t.url, "http://")
	if t.base != nil {
		return t.base.RoundTrip(req)
	}
	return http.DefaultTransport.RoundTrip(req)
}
