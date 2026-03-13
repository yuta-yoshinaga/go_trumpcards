package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"

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
	resp, err := http.Get(url) //nolint:gosec,noctx // simple GET to GitHub API
	if err != nil {
		fmt.Fprintln(u.errWriter, i18n.Tf("updateFetchError", "error", err.Error()))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("HTTP %d", resp.StatusCode)
		fmt.Fprintln(u.errWriter, i18n.Tf("updateFetchError", "error", errMsg))
		return fmt.Errorf("%s", errMsg)
	}

	var release ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		fmt.Fprintln(u.errWriter, i18n.Tf("updateFetchError", "error", err.Error()))
		return err
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentClean := strings.TrimPrefix(u.currentVersion, "v")

	// If current version is not "dev" and matches latest, already up to date.
	if currentClean != "dev" && currentClean == latestVersion {
		fmt.Fprintln(u.writer, i18n.Tf("updateAlreadyLatest", "version", release.TagName))
		return nil
	}

	// Prompt for confirmation.
	fmt.Fprint(u.writer, i18n.Tf("updateAvailable", "version", release.TagName)+" ")
	var answer string
	fmt.Fscanln(u.reader, &answer)
	if answer != "y" && answer != "Y" {
		fmt.Fprintln(u.writer, i18n.T("updateCancelled"))
		return nil
	}

	// Find matching asset.
	assetName, assetURL := u.findAsset(release.Assets)
	if assetURL == "" {
		fmt.Fprintln(u.errWriter, i18n.Tf("updateNoAsset", "os", runtime.GOOS, "arch", runtime.GOARCH))
		return fmt.Errorf("no matching asset for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	fmt.Fprintln(u.writer, i18n.Tf("updateDownloading", "version", release.TagName))

	// Download asset.
	assetResp, err := http.Get(assetURL) //nolint:gosec,noctx // downloading release asset
	if err != nil {
		fmt.Fprintln(u.errWriter, i18n.Tf("updateApplyError", "error", err.Error()))
		return err
	}
	defer assetResp.Body.Close()

	assetData, err := io.ReadAll(assetResp.Body)
	if err != nil {
		fmt.Fprintln(u.errWriter, i18n.Tf("updateApplyError", "error", err.Error()))
		return err
	}

	// Extract binary from archive.
	binaryReader, err := u.extractBinary(assetName, assetData)
	if err != nil {
		fmt.Fprintln(u.errWriter, i18n.Tf("updateApplyError", "error", err.Error()))
		return err
	}

	// Apply update.
	if err := selfupdate.Apply(binaryReader, selfupdate.Options{}); err != nil {
		fmt.Fprintln(u.errWriter, i18n.Tf("updateApplyError", "error", err.Error()))
		return err
	}

	fmt.Fprintln(u.writer, i18n.Tf("updateSuccess", "version", release.TagName))
	return nil
}

// findAsset finds a matching release asset for the current OS/arch.
func (u *Updater) findAsset(assets []ghAsset) (name, url string) {
	osName := runtime.GOOS
	archName := runtime.GOARCH

	var ext string
	if osName == "windows" {
		ext = ".zip"
	} else {
		ext = ".tar.gz"
	}

	for _, a := range assets {
		lower := strings.ToLower(a.Name)
		if strings.Contains(lower, osName) && strings.Contains(lower, archName) && strings.HasSuffix(lower, ext) {
			return a.Name, a.BrowserDownloadURL
		}
	}
	return "", ""
}

// extractBinary extracts the binary from a tar.gz or zip archive.
func (u *Updater) extractBinary(assetName string, data []byte) (io.Reader, error) {
	binaryName := "trumpcards"
	if runtime.GOOS == "windows" {
		binaryName = "trumpcards.exe"
	}

	if strings.HasSuffix(strings.ToLower(assetName), ".zip") {
		return u.extractFromZip(data, binaryName)
	}
	return u.extractFromTarGz(data, binaryName)
}

// extractFromTarGz extracts the named file from a tar.gz archive.
func (u *Updater) extractFromTarGz(data []byte, name string) (io.Reader, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr.Name == name || strings.HasSuffix(hdr.Name, "/"+name) {
			buf := &bytes.Buffer{}
			if _, err := io.Copy(buf, tr); err != nil { //nolint:gosec // trusted archive from GitHub
				return nil, err
			}
			return buf, nil
		}
	}
	return nil, fmt.Errorf("binary %q not found in archive", name)
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
			defer rc.Close()
			buf := &bytes.Buffer{}
			if _, err := io.Copy(buf, rc); err != nil { //nolint:gosec // trusted archive from GitHub
				return nil, err
			}
			return buf, nil
		}
	}
	return nil, fmt.Errorf("binary %q not found in archive", name)
}
