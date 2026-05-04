package web

import (
	"log/slog"
	"net/http"
)

// spaFallbackHandler wraps a static file server so unknown GET/HEAD requests
// fall back to index.html instead of returning 404. The frontend uses
// HashRouter, so direct URLs like /blackjack or /poker have no on-disk file —
// without this fallback, externally shared deep-links return Go's plain
// "404 page not found" body. Any other HTTP method is delegated to the
// underlying file server unchanged.
//
// publicDir must be the same path the underlying file server is rooted at, so
// the fallback can probe the requested path relative to that root. We use
// http.Dir.Open rather than os.Stat with a joined path so the standard
// library's traversal guards (rejection of "..", NUL bytes, absolute paths
// outside the root) handle the user-controlled URL safely — and so static
// analysis tools recognise the sanitisation.
func spaFallbackHandler(fs http.Handler, publicDir string) http.Handler {
	publicFS := http.Dir(publicDir)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			fs.ServeHTTP(w, r)
			return
		}
		// Defer to the FileServer when the requested path resolves to an
		// existing file. Directories also fall back to index.html so the SPA
		// owns its own routing instead of seeing a directory listing.
		if f, err := publicFS.Open(r.URL.Path); err == nil {
			stat, statErr := f.Stat()
			_ = f.Close()
			if statErr == nil && !stat.IsDir() {
				fs.ServeHTTP(w, r)
				return
			}
		}
		serveSPAIndex(w, r, publicFS)
	})
}

// serveSPAIndex writes the SPA's index.html using a constant path so the
// served content doesn't depend on the user-supplied URL. http.ServeFile would
// reject any request whose URL.Path contains "..", which we want the SPA — not
// the server — to decide on, so we use http.ServeContent with the file we
// opened ourselves.
func serveSPAIndex(w http.ResponseWriter, r *http.Request, publicFS http.FileSystem) {
	f, err := publicFS.Open("/index.html")
	if err != nil {
		slog.Warn("spa: index.html unavailable", "error", err)
		http.Error(w, "index unavailable", http.StatusInternalServerError)
		return
	}
	defer func() { _ = f.Close() }()
	stat, err := f.Stat()
	if err != nil {
		slog.Warn("spa: stat index.html failed", "error", err)
		http.Error(w, "index unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeContent(w, r, "index.html", stat.ModTime(), f)
}
