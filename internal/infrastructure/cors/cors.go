package cors

import (
	"net/http"
	"strconv"
	"strings"
)

const maxAge = 3600

// Middleware returns an http.Handler that applies CORS headers for the
// given set of allowed origins. Only POST requests with Content-Type are
// permitted. Preflight OPTIONS requests receive the appropriate headers and
// a 204 response.
func Middleware(allowedOrigins map[string]bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		origin := r.Header.Get("Origin")
		if origin != "" && allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "POST")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Max-Age", strconv.Itoa(maxAge))
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ParseOrigins parses a comma-separated list of origins into a map.
func ParseOrigins(csv string) map[string]bool {
	if csv == "" {
		return nil
	}
	origins := make(map[string]bool, strings.Count(csv, ",")+1)
	for _, origin := range strings.Split(csv, ",") {
		if o := strings.TrimSpace(origin); o != "" {
			origins[o] = true
		}
	}
	return origins
}
