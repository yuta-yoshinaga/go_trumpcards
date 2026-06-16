//go:build !js || !wasm

package i18n

import "embed"

// localesFS embeds every translation JSON. This embed (and its ~430 KB of
// locale data) is compiled into the server and CLI binaries only — the
// Cloudflare Worker WASM builds (js && wasm) use the i18n_stub.go variant so
// the locales do not bloat each worker toward the 1 MB gzip free-tier limit.
//
//go:embed locales/ja
//go:embed locales/en
var localesFS embed.FS

// loadLocale returns the merged translation map for lang from the embedded
// locale files.
func loadLocale(lang string) map[string]string {
	return loadTranslations(localesFS, lang)
}
