package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/minio/selfupdate"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// Well-known update failure categories. Callers wrap process-level exit codes
// around these so CI / cron / package scripts can tell a retryable network
// failure apart from a non-retryable "no binary for this platform". See
// issue #1449.
var (
	// ErrNetwork indicates a transport-level failure talking to the release
	// registry (DNS, connection refused, TLS, timeout). Typically retryable.
	ErrNetwork = errors.New("update: network failure")
	// ErrAPIStatus indicates the release registry returned a non-2xx status.
	// May be retryable depending on the code (e.g. 5xx yes, 404 no).
	ErrAPIStatus = errors.New("update: release API returned non-OK status")
	// ErrNoAsset indicates no release asset matched the current OS/arch.
	// Not retryable without a new release.
	ErrNoAsset = errors.New("update: no asset for this platform")
	// ErrExtract indicates an archive (tar.gz / zip) failed to decompress or
	// did not contain the expected binary. Not retryable.
	ErrExtract = errors.New("update: failed to extract archive")
	// ErrApply indicates the final binary swap failed (permissions, disk
	// full, integrity check). Not retryable without fixing the environment.
	ErrApply = errors.New("update: failed to apply binary")
	// ErrNonInteractive indicates --yes was not given and stdin was closed
	// (EOF before any answer). Resolve by re-running with --yes.
	ErrNonInteractive = errors.New("update: non-interactive stdin without --yes")
	// ErrUserCancelled indicates the user answered the prompt with a
	// non-affirmative response (empty, "n", "no").
	ErrUserCancelled = errors.New("update: user declined")
	// ErrUpdateAvailable is returned only in check-only mode (SetCheckOnly)
	// when the release API reports a newer version than the current build.
	// Callers map it to a dedicated exit code (e.g. 10) so CI / cron can tell
	// "update exists" apart from "already latest" (nil) and "couldn't check"
	// (ErrNetwork / ErrAPIStatus). See issue #1484.
	ErrUpdateAvailable = errors.New("update: new version available")
)

// Updater checks GitHub Releases for the latest version and self-updates the binary.
type Updater struct {
	currentVersion string
	repoOwner      string
	repoName       string
	reader         io.Reader
	writer         io.Writer
	errWriter      io.Writer
	httpClient     *http.Client
	autoConfirm    bool
	progressIsTTY  bool
	// reportCancelledAsError: when true, Exec returns ErrUserCancelled on
	// a declined prompt. The zero value (false) preserves the legacy
	// behavior of returning nil, so library callers who never call the
	// setter keep their old contract.
	reportCancelledAsError bool
	// checkOnly: when true, Exec fetches the release info, writes a
	// machine-readable one-line summary to writer and a human-friendly
	// message to errWriter, then returns without prompting or downloading.
	// See issue #1484.
	checkOnly bool
}

// SetAutoConfirm enables or disables the automatic confirmation of updates,
// skipping the interactive prompt. Useful for CI/CD and scripted environments.
func (u *Updater) SetAutoConfirm(v bool) {
	u.autoConfirm = v
}

// SetProgressIsTTY tells the Updater whether the progress-output stream is a
// TTY. When false, the download progress is not rendered with carriage returns
// so that logs captured via tee/redirect remain readable.
func (u *Updater) SetProgressIsTTY(v bool) {
	u.progressIsTTY = v
}

// NewUpdater creates a new Updater.
func NewUpdater(currentVersion string, reader io.Reader, writer, errWriter io.Writer) *Updater {
	return &Updater{
		currentVersion: currentVersion,
		repoOwner:      "yuta-yoshinaga",
		repoName:       "go_trumpcards",
		reader:         reader,
		writer:         writer,
		errWriter:      errWriter,
		httpClient:     &http.Client{Timeout: 60 * time.Second},
	}
}

type ghRelease struct {
	TagName string    `json:"tag_name"`
	Assets  []ghAsset `json:"assets"`
}

type ghAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// SetReportCancelledAsError controls how a user-declined prompt is reported.
// When false (the default, for backwards compatibility), Exec returns nil on
// cancellation. When true, Exec returns ErrUserCancelled so callers can map
// it to a dedicated exit code (e.g. 75 / EX_TEMPFAIL).
func (u *Updater) SetReportCancelledAsError(v bool) {
	u.reportCancelledAsError = v
}

// SetCheckOnly toggles check-only (dry-run) mode. When enabled, Exec fetches
// the latest release metadata, writes a tab-separated summary line
// (`<latest>\t<status>\t<current>`) to writer and a human-friendly message to
// errWriter, then returns without prompting or downloading. The return value
// distinguishes "already latest" (nil) from "update available"
// (ErrUpdateAvailable) so CI / cron scripts can branch on exit code. See
// issue #1484.
func (u *Updater) SetCheckOnly(v bool) {
	u.checkOnly = v
}

// Exec runs the self-update flow.
//
// Errors are wrapped with sentinel categories from this package (ErrNetwork,
// ErrAPIStatus, ErrNoAsset, ErrExtract, ErrApply, ErrNonInteractive,
// ErrUserCancelled) so the caller can pick a meaningful exit code via
// errors.Is. See issue #1449.
func (u *Updater) Exec() error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", u.repoOwner, u.repoName)
	resp, err := u.httpClient.Get(url) //nolint:noctx // simple GET to GitHub API
	if err != nil {
		_, _ = fmt.Fprintln(u.errWriter, i18n.Tf("updateFetchError", "error", err.Error()))
		return fmt.Errorf("%w: %w", ErrNetwork, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("HTTP %d", resp.StatusCode)
		_, _ = fmt.Fprintln(u.errWriter, i18n.Tf("updateFetchError", "error", errMsg))
		return fmt.Errorf("%w: %s", ErrAPIStatus, errMsg)
	}

	var release ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		_, _ = fmt.Fprintln(u.errWriter, i18n.Tf("updateFetchError", "error", err.Error()))
		// A malformed body from the API is still an API-level failure as
		// far as callers are concerned — they can't fix the upstream.
		return fmt.Errorf("%w: %w", ErrAPIStatus, err)
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentClean := strings.TrimPrefix(u.currentVersion, "v")

	// Check-only (--check / --dry-run) short-circuit: report status and return.
	// Never prompts, never downloads. See issue #1484.
	if u.checkOnly {
		return u.reportCheckOnly(release.TagName, latestVersion, currentClean)
	}

	// If current version is not "dev" and matches latest, already up to date.
	if currentClean != "dev" && currentClean == latestVersion {
		_, _ = fmt.Fprintln(u.writer, i18n.Tf("updateAlreadyLatest", "version", release.TagName))
		return nil
	}

	// Prompt for confirmation (skipped with --yes).
	if u.autoConfirm {
		_, _ = fmt.Fprintln(u.writer, i18n.Tf("updateAutoConfirm", "current", u.currentVersion, "version", release.TagName))
	} else {
		_, _ = fmt.Fprint(u.writer, i18n.Tf("updateAvailable", "current", u.currentVersion, "version", release.TagName)+" ")
		var answer string
		_, scanErr := fmt.Fscanln(u.reader, &answer)
		if errors.Is(scanErr, io.EOF) {
			_, _ = fmt.Fprintln(u.errWriter, i18n.T("updateNonInteractive"))
			return ErrNonInteractive
		}
		// Fscanln returns a non-EOF error when it reads a blank line (no token)
		// or only whitespace; treat that as the empty-input default path (cancel).
		// Any other unexpected I/O failure should surface to the user.
		if scanErr != nil && !isBlankLineScanErr(scanErr) {
			_, _ = fmt.Fprintln(u.errWriter, i18n.Tf("inputReadError", "error", scanErr.Error()))
			return scanErr
		}
		ans := strings.ToLower(strings.TrimSpace(answer))
		if ans != "y" && ans != "yes" {
			_, _ = fmt.Fprintln(u.writer, i18n.T("updateCancelled"))
			if u.reportCancelledAsError {
				return ErrUserCancelled
			}
			return nil
		}
	}

	// Find matching asset.
	assetName, assetURL := u.findAsset(release.Assets)
	if assetURL == "" {
		_, _ = fmt.Fprintln(u.errWriter, i18n.Tf("updateNoAsset", "os", runtime.GOOS, "arch", runtime.GOARCH))
		return fmt.Errorf("%w: %s/%s", ErrNoAsset, runtime.GOOS, runtime.GOARCH)
	}

	_, _ = fmt.Fprintln(u.writer, i18n.Tf("updateDownloading", "version", release.TagName))

	// Download and apply update.
	if err := u.downloadAndApply(assetName, assetURL); err != nil {
		_, _ = fmt.Fprintln(u.errWriter, i18n.Tf("updateApplyError", "error", err.Error()))
		return err
	}

	_, _ = fmt.Fprintln(u.writer, i18n.Tf("updateSuccess", "version", release.TagName))
	return nil
}

// isBlankLineScanErr reports whether err is the "unexpected newline" error
// returned by fmt.Fscanln when the user presses Enter with no input — we want
// to treat that as the default-no path, not a hard I/O failure.
func isBlankLineScanErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "unexpected newline")
}

// reportCheckOnly emits the machine-readable status line to writer plus a
// human-friendly message to errWriter, and returns the sentinel expected by
// the CLI exit-code mapper (nil = latest, ErrUpdateAvailable = newer exists).
//
// Output format on writer (stable; safe for shell parsing with `cut -f1`):
//
//	<latest-tag>\t<status>\t<current>
//
// Status is one of: "latest", "available", "dev".
func (u *Updater) reportCheckOnly(latestTag, latestClean, currentClean string) error {
	switch currentClean {
	case "dev":
		_, _ = fmt.Fprintf(u.writer, "%s\tdev\t%s\n", latestTag, u.currentVersion)
		_, _ = fmt.Fprintln(u.errWriter, i18n.Tf("updateCheckDev", "version", latestTag))
		return nil
	case latestClean:
		_, _ = fmt.Fprintf(u.writer, "%s\tlatest\t%s\n", latestTag, u.currentVersion)
		_, _ = fmt.Fprintln(u.errWriter, i18n.Tf("updateAlreadyLatest", "version", latestTag))
		return nil
	default:
		_, _ = fmt.Fprintf(u.writer, "%s\tavailable\t%s\n", latestTag, u.currentVersion)
		_, _ = fmt.Fprintln(u.errWriter, i18n.Tf("updateCheckAvailable", "current", u.currentVersion, "version", latestTag))
		return ErrUpdateAvailable
	}
}

// progressReader wraps an io.Reader to display download progress.
type progressReader struct {
	reader  io.Reader
	total   int64
	current int64
	writer  io.Writer
	isTTY   bool
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.current += int64(n)
	if !pr.isTTY {
		return n, err
	}
	if pr.total > 0 {
		pct := min(100, pr.current*100/pr.total)
		_, _ = fmt.Fprintf(pr.writer, "\r  %d%% (%s / %s)", pct, formatBytes(pr.current), formatBytes(pr.total))
	} else {
		_, _ = fmt.Fprintf(pr.writer, "\r  %s downloaded", formatBytes(pr.current))
	}
	return n, err
}

// formatBytes formats bytes into a human-readable string (KB/MB/GB).
func formatBytes(b int64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case b >= gb:
		return fmt.Sprintf("%.1f GB", float64(b)/float64(gb))
	case b >= mb:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(mb))
	case b >= kb:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(kb))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

// downloadAndApply downloads the asset and applies the update. Errors are
// wrapped with sentinel categories so callers can distinguish network,
// extract, and apply failures.
func (u *Updater) downloadAndApply(assetName, assetURL string) error {
	assetResp, err := u.httpClient.Get(assetURL) //nolint:noctx // downloading release asset
	if err != nil {
		return fmt.Errorf("%w: %w", ErrNetwork, err)
	}
	defer func() { _ = assetResp.Body.Close() }()

	body := &progressReader{
		reader: assetResp.Body,
		total:  assetResp.ContentLength,
		writer: u.writer,
		isTTY:  u.progressIsTTY,
	}
	endProgress := func() {
		if u.progressIsTTY {
			_, _ = fmt.Fprintln(u.writer)
		}
	}

	// For tar.gz, stream directly without buffering the entire archive in memory.
	if strings.HasSuffix(strings.ToLower(assetName), ".tar.gz") {
		binaryReader, err := u.extractFromTarGzStream(body)
		if err != nil {
			endProgress()
			return fmt.Errorf("%w: %w", ErrExtract, err)
		}
		endProgress()
		if err := selfupdate.Apply(binaryReader, selfupdate.Options{}); err != nil {
			return fmt.Errorf("%w: %w", ErrApply, err)
		}
		return nil
	}

	// For zip, we need io.ReaderAt so we must read the entire archive into memory.
	assetData, err := io.ReadAll(body)
	if err != nil {
		endProgress()
		return fmt.Errorf("%w: %w", ErrNetwork, err)
	}
	endProgress()

	binaryName := "trumpcards.exe"
	binaryReader, err := u.extractFromZip(assetData, binaryName)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrExtract, err)
	}
	if err := selfupdate.Apply(binaryReader, selfupdate.Options{}); err != nil {
		return fmt.Errorf("%w: %w", ErrApply, err)
	}
	return nil
}

// findAsset finds a matching release asset for the current OS/arch.
// Uses GoReleaser naming convention: trumpcards_<version>_<os>_<arch>.<ext>
func (u *Updater) findAsset(assets []ghAsset) (name, url string) {
	osName := runtime.GOOS
	archName := runtime.GOARCH

	var ext string
	if osName == "windows" {
		ext = ".zip"
	} else {
		ext = ".tar.gz"
	}

	// Match GoReleaser suffix pattern: _<os>_<arch>.<ext>
	suffix := fmt.Sprintf("_%s_%s%s", osName, archName, ext)

	for _, a := range assets {
		lower := strings.ToLower(a.Name)
		if strings.HasSuffix(lower, suffix) {
			return a.Name, a.BrowserDownloadURL
		}
	}
	return "", ""
}

// extractFromTarGzStream extracts the trumpcards binary from a tar.gz stream.
func (u *Updater) extractFromTarGzStream(r io.Reader) (io.Reader, error) {
	binaryName := "trumpcards"

	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer func() { _ = gz.Close() }()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr.Name == binaryName || strings.HasSuffix(hdr.Name, "/"+binaryName) {
			buf := &bytes.Buffer{}
			if _, err := io.Copy(buf, tr); err != nil { //nolint:gosec // trusted archive from GitHub
				return nil, err
			}
			return buf, nil
		}
	}
	return nil, fmt.Errorf("binary %q not found in archive", binaryName)
}

// extractFromZip extracts the named file from a zip archive.
func (u *Updater) extractFromZip(data []byte, name string) (io.Reader, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}

	for _, f := range zr.File {
		if f.Name == name || strings.HasSuffix(f.Name, "/"+name) {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			buf := &bytes.Buffer{}
			if _, cpErr := io.Copy(buf, rc); cpErr != nil { //nolint:gosec // trusted archive from GitHub
				_ = rc.Close()
				return nil, cpErr
			}
			_ = rc.Close()
			return buf, nil
		}
	}
	return nil, fmt.Errorf("binary %q not found in archive", name)
}
