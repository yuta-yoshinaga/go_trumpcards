// Package domain 汎用スライスヘルパー。
//
// No build tag: this file is compiled into every build, including all
// Cloudflare Worker WASM binaries. The helper is generic and game-agnostic,
// so it must be available regardless of which category worker a game lands
// in (previously it lived in a casino-only game file, which prevented
// rebucketing any game that used it to another worker).
package domain

// sliceOrEmpty replaces nil with an empty slice for JSON round-trip stability.
func sliceOrEmpty[T any](s []T) []T {
	if s == nil {
		return make([]T, 0)
	}
	return s
}
