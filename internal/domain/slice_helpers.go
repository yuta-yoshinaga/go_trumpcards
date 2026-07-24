// Package domain 汎用スライスヘルパー。
//
// No build tag: this file is compiled into every build, including all
// Cloudflare Worker WASM binaries. The helper is generic and game-agnostic,
// so it must be available regardless of which category worker a game lands
// in (previously it lived in a casino-only game file, which prevented
// rebucketing any game that used it to another worker).
package domain

// collectValidIndices returns the indices in [0, size) for which ok reports
// true. It factors out the "scan a hand and keep the playable card indices"
// loop duplicated across dozens of trick-taking / shedding games (issue #4299):
// each game passes a closure wrapping its own validation (e.g. validatePlay ==
// nil, or isValidPlay). The result is pre-sized to size and is non-nil even
// when empty.
func collectValidIndices(size int, ok func(i int) bool) []int {
	valid := make([]int, 0, size)
	for i := range size {
		if ok(i) {
			valid = append(valid, i)
		}
	}
	return valid
}

// sliceOrEmpty replaces nil with an empty slice for JSON round-trip stability.
func sliceOrEmpty[T any](s []T) []T {
	if s == nil {
		return make([]T, 0)
	}
	return s
}
