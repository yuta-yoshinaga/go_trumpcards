package update

import (
	"bytes"
	"encoding/json"
	"errors"
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
		httpClient:     &http.Client{Transport: &rewriteTransport{base: srv.Client().Transport, url: srv.URL}},
	}

	err := u.Exec()
	require.NoError(t, err)
	assert.Contains(t, out.String(), "v1.2.3")
}

// TestExec_UserCancels_Default verifies the default behavior (kept for
// backwards compatibility): cancellation returns nil. Callers that want a
// distinct exit code should call SetReportCancelledAsError(true) — see
// TestExec_UserCancels_AsSentinelError below.
func TestExec_UserCancels_Default(t *testing.T) {
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
		// reportCancelledAsError is false by default — old callers via NewUpdater
		// get nil (no error) on cancellation, preserving backwards compatibility.
	}

	err := u.Exec()
	require.NoError(t, err)
	assert.Contains(t, out.String(), "v2.0.0")
}

// TestExec_UserCancels_AsSentinelError verifies that with
// SetReportCancelledAsError(true), an "n" response returns ErrUserCancelled so
// the CLI can map it to exit code 75 (EX_TEMPFAIL). See issue #1449.
func TestExec_UserCancels_AsSentinelError(t *testing.T) {
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
	u.SetReportCancelledAsError(true)

	err := u.Exec()
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrUserCancelled), "want ErrUserCancelled, got %v", err)
}

func TestExec_UserAccepts(t *testing.T) {
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
		reader:         strings.NewReader("y\n"),
		writer:         out,
		errWriter:      errOut,
		httpClient:     &http.Client{Transport: &rewriteTransport{base: srv.Client().Transport, url: srv.URL}},
	}

	// Exec proceeds past the prompt but fails at findAsset (no matching asset).
	err := u.Exec()
	require.Error(t, err)
	assert.Contains(t, out.String(), "[y/N]")
	assert.True(t, errors.Is(err, ErrNoAsset), "missing asset must wrap ErrNoAsset; got %v", err)
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
	assert.True(t, errors.Is(err, ErrAPIStatus), "HTTP 500 must wrap ErrAPIStatus; got %v", err)
}

// TestExec_NetworkError_WrapsErrNetwork verifies that a transport-level
// failure (unreachable host) surfaces as ErrNetwork so the CLI can map it
// to an exit code that signals "retry later". See issue #1449.
func TestExec_NetworkError_WrapsErrNetwork(t *testing.T) {
	// Set up a server, immediately close it to force a dial failure.
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	srv.Close()

	u := &Updater{
		currentVersion: "1.0.0",
		repoOwner:      "test",
		repoName:       "test",
		reader:         strings.NewReader(""),
		writer:         &bytes.Buffer{},
		errWriter:      &bytes.Buffer{},
		httpClient:     &http.Client{Transport: &rewriteTransport{base: srv.Client().Transport, url: srv.URL}},
	}

	err := u.Exec()
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNetwork), "dial failure must wrap ErrNetwork; got %v", err)
}

// TestExec_NonInteractiveEOF_WrapsErrNonInteractive verifies that an EOF on
// stdin without --yes surfaces as ErrNonInteractive so the CLI can map it
// to exit 2 (usage error). See issue #1449.
func TestExec_NonInteractiveEOF_WrapsErrNonInteractive(t *testing.T) {
	release := ghRelease{TagName: "v2.0.0"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(release)
	}))
	defer srv.Close()

	u := &Updater{
		currentVersion: "1.0.0",
		repoOwner:      "test",
		repoName:       "test",
		reader:         strings.NewReader(""), // immediate EOF
		writer:         &bytes.Buffer{},
		errWriter:      &bytes.Buffer{},
		httpClient:     &http.Client{Transport: &rewriteTransport{base: srv.Client().Transport, url: srv.URL}},
	}
	err := u.Exec()
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNonInteractive), "EOF must wrap ErrNonInteractive; got %v", err)
}

func TestExec_NonInteractiveEOF_ReturnsErrorWithHint(t *testing.T) {
	release := ghRelease{TagName: "v2.0.0"}
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
		reader:         strings.NewReader(""), // immediate EOF, no newline
		writer:         out,
		errWriter:      errOut,
		httpClient:     &http.Client{Transport: &rewriteTransport{base: srv.Client().Transport, url: srv.URL}},
	}

	err := u.Exec()
	require.Error(t, err, "EOF on stdin must be a hard error, not a silent cancel")
	assert.Contains(t, errOut.String(), "--yes", "should hint at --yes")
	assert.NotContains(t, out.String(), "cancelled", "must not print the 'cancelled' path on EOF")
	assert.NotContains(t, out.String(), "キャンセル", "must not print the 'cancelled' path on EOF")
}

type errReader struct{ err error }

func (r *errReader) Read(p []byte) (int, error) { return 0, r.err }

func TestExec_ScanIOError_Surfaces(t *testing.T) {
	release := ghRelease{TagName: "v2.0.0"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(release)
	}))
	defer srv.Close()

	errOut := &bytes.Buffer{}
	ioErr := errors.New("stdin broken pipe")
	u := &Updater{
		currentVersion: "1.0.0",
		repoOwner:      "test",
		repoName:       "test",
		reader:         &errReader{err: ioErr},
		writer:         &bytes.Buffer{},
		errWriter:      errOut,
		httpClient:     &http.Client{Transport: &rewriteTransport{base: srv.Client().Transport, url: srv.URL}},
	}

	err := u.Exec()
	require.Error(t, err)
	assert.Contains(t, errOut.String(), "stdin broken pipe", "unexpected scan I/O error must be surfaced")
}

func TestProgressReader_NonTTY_DoesNotWriteCarriageReturns(t *testing.T) {
	src := strings.NewReader("hello world")
	out := &bytes.Buffer{}
	pr := &progressReader{reader: src, total: 11, writer: out, isTTY: false}
	buf := make([]byte, 4)
	for {
		if _, err := pr.Read(buf); err != nil {
			break
		}
	}
	assert.Empty(t, out.String(), "non-TTY progress reader must not write progress output")
}

func TestProgressReader_TTY_WritesCarriageReturns(t *testing.T) {
	src := strings.NewReader("hello world")
	out := &bytes.Buffer{}
	pr := &progressReader{reader: src, total: 11, writer: out, isTTY: true}
	buf := make([]byte, 4)
	for {
		if _, err := pr.Read(buf); err != nil {
			break
		}
	}
	assert.Contains(t, out.String(), "\r")
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
