// Package recoverymw provides an HTTP middleware that turns Go panics into
// JSON 500 responses so Cloudflare Workers don't surface their generic 1101
// HTML error page (which loses CORS headers and shows up to users as a
// "通信エラー" with no actionable detail).
//
// The package is named recoverymw rather than "recover" to avoid shadowing
// Go's builtin recover() function.
package recoverymw

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
)

// Middleware wraps next so that any panic in the downstream handler is
// caught, logged, and converted to a JSON 500 response. It must be wrapped
// by the CORS middleware so the recovered response still carries the CORS
// headers needed for the browser to read it.
//
// If the downstream handler already wrote response bytes before panicking,
// WriteHeader below is a no-op on the status line and the JSON body will be
// appended to the partial response, producing a malformed body. We accept
// that limitation because the alternative — letting the panic escape —
// surfaces Cloudflare's 1101 HTML page without CORS headers, which the
// browser cannot read at all. See TestMiddleware_RecoversPanicAfterPartialWrite.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			slog.Error(
				"handler panic",
				"path", r.URL.Path,
				"method", r.Method,
				"panic", fmt.Sprint(rec),
				"stack", string(debug.Stack()),
			)
			// Best-effort: attempt to write a JSON error regardless of whether
			// the handler already committed headers. If it has, WriteHeader is
			// a no-op and the body is appended to whatever was already written.
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"message": fmt.Sprintf("internal error: %v", rec),
			})
		}()
		next.ServeHTTP(w, r)
	})
}
