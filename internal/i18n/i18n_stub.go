//go:build js && wasm

package i18n

// loadLocale is the Cloudflare Worker (WASM) variant: it returns an empty map
// so the ~430 KB embedded locale set is NOT linked into each worker binary
// (which must stay under the 1 MB gzip free-tier limit). In the worker the Web
// API returns messageCode/messageParams for the frontend to translate, so
// server-side T()/Tf() simply fall back to the key — no user-facing locale data
// is needed inside the worker.
func loadLocale(_ string) map[string]string {
	return map[string]string{}
}
