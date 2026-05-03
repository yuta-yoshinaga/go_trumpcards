package web

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
)

// spaFallbackHandler wraps a static file server so unknown GET/HEAD requests
// fall back to index.html instead of returning 404. The frontend uses
// HashRouter, so direct URLs like /blackjack or /poker have no on-disk file —
// without this fallback, externally shared deep-links return Go's plain
// "404 page not found" body. Any other HTTP method is delegated to the
// underlying file server unchanged.
//
// publicDir must be the same path the underlying file server is rooted at, so
// the fallback can stat-check the requested path relative to that root.
func spaFallbackHandler(fs http.Handler, publicDir string) http.Handler {
	indexPath := filepath.Join(publicDir, "index.html")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			fs.ServeHTTP(w, r)
			return
		}
		clean := path.Clean("/" + r.URL.Path)
		fp := filepath.Join(publicDir, clean)
		// Only fall back when the resolved path doesn't exist on disk. Existing
		// files (and directories with an index.html) keep the original
		// FileServer behaviour, which preserves correct Content-Type, ETag,
		// caching headers, and HEAD support for static assets.
		if info, err := os.Stat(fp); err == nil && !info.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, indexPath)
	})
}
