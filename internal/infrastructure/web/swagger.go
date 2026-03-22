package web

import (
	_ "embed"
	"log/slog"
	"net/http"

	trumpapi "github.com/yuta-yoshinaga/go_trumpcards/api"
)

//go:embed swagger.html
var swaggerHTML []byte

// RegisterSwaggerRoutes registers Swagger UI routes on the given mux.
// It serves the embedded OpenAPI spec at /swagger/openapi.yaml and
// the Swagger UI HTML page at /swagger/.
func RegisterSwaggerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/swagger/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		if _, err := w.Write(trumpapi.OpenAPISpec); err != nil {
			slog.Warn("failed to write swagger spec", "error", err, "remote_addr", r.RemoteAddr)
		}
	})
	mux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := w.Write(swaggerHTML); err != nil {
			slog.Warn("failed to write swagger html", "error", err, "remote_addr", r.RemoteAddr)
		}
	})
}
