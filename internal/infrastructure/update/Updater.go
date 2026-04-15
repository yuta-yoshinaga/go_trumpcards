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

// Exec runs the self-update flow.
func (u *Updater) Exec() error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", u.repoOwner, u.repoName)
	resp, err := u.httpClient.Get(url) //nolint:noctx // simple GET to GitHub API
	if err != nil {
		_, _ = fmt.Fprintln(u.errWriter, i18n.Tf("updateFetchError", "error", err.Error()))
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("HTTP %d", resp.StatusCode)
		_, _ = fmt.Fprintln(u.errWriter, i18n.Tf("updateFetchError", "error", errMsg))
		return errors.New(errMsg)
	}

	var release ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		_, _ = fmt.Fprintln(u.errWriter, i18n.Tf("updateFetchError", "error", err.Error()))
		return err
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentClean := strings.TrimPrefix(u.currentVersion, "v")

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
			return errors.New("non-interactive stdin: --yes required")
		}
		ans := strings.ToLower(strings.TrimSpace(answer))
		if ans != "y" && ans != "yes" {
			_, _ = fmt.Fprintln(u.writer, i18n.T("updateCancelled"))
			return nil
		}
	}

	// Find matching asset.
	assetName, assetURL := u.findAsset(release.Assets)
	if assetURL == "" {
		_, _ = fmt.Fprintln(u.errWriter, i18n.Tf("updateNoAsset", "os", runtime.GOOS, "arch", runtime.GOARCH))
		return fmt.Errorf("no matching asset for %s/%s", runtime.GOOS, runtime.GOARCH)
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

// downloadAndApply downloads the asset and applies the update.
func (u *Updater) downloadAndApply(assetName, assetURL string) error {
	assetResp, err := u.httpClient.Get(assetURL) //nolint:noctx // downloading release asset
	if err != nil {
		return err
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
			return err
		}
		endProgress()
		return selfupdate.Apply(binaryReader, selfupdate.Options{})
	}

	// For zip, we need io.ReaderAt so we must read the entire archive into memory.
	assetData, err := io.ReadAll(body)
	if err != nil {
		endProgress()
		return err
	}
	endProgress()

	binaryName := "trumpcards.exe"
	binaryReader, err := u.extractFromZip(assetData, binaryName)
	if err != nil {
		return err
	}
	return selfupdate.Apply(binaryReader, selfupdate.Options{})
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
