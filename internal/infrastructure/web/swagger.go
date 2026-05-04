package web

import (
	_ "embed"
	"log/slog"
	"net/http"
	"strconv"

	trumpapi "github.com/yuta-yoshinaga/go_trumpcards/api"
)

//go:embed swagger.html
var swaggerHTML []byte

// RegisterSwaggerRoutes registers Swagger UI routes on the given mux.
// It serves the embedded OpenAPI spec at /swagger/openapi.yaml and
// the Swagger UI HTML page at /swagger/.
func RegisterSwaggerRoutes(mux *http.ServeMux) {
	// HEAD is treated like GET so generic clients (curl -I, link checkers,
	// reverse proxies probing health) don't get a spurious 405.
	mux.HandleFunc("/swagger/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/yaml")
		w.Header().Set("Content-Length", strconv.Itoa(len(trumpapi.OpenAPISpec)))
		if r.Method == http.MethodHead {
			return
		}
		if _, err := w.Write(trumpapi.OpenAPISpec); err != nil {
			slog.Warn("failed to write swagger spec", "error", err, "remote_addr", r.RemoteAddr)
		}
	})
	mux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Length", strconv.Itoa(len(swaggerHTML)))
		if r.Method == http.MethodHead {
			return
		}
		if _, err := w.Write(swaggerHTML); err != nil {
			slog.Warn("failed to write swagger html", "error", err, "remote_addr", r.RemoteAddr)
		}
	})
}
