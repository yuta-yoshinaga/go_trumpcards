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

// discardTop returns the card most recently added to a discard pile, or nil
// when the pile is empty.
//
// 22 games had this written out. Like indexOfPlayerInTrick it needs no type
// parameter -- every discard pile is a []*Card -- so it adds no monomorphised
// copies. See issue #5185.
func discardTop(pile []*Card) *Card {
	if len(pile) == 0 {
		return nil
	}
	return pile[len(pile)-1]
}

// copyOf returns a copy of s, so callers can hand out a slice without exposing
// the backing array. Four accessors had make+copy written out.
func copyOf[T any](s []T) []T {
	out := make([]T, len(s))
	copy(out, s)
	return out
}

// elemAt returns s[i], or the zero value when i is outside the slice. Callers
// treat the zero value as "absent" rather than checking bounds themselves.
func elemAt[T any](s []T, i int) T {
	var zero T
	if i < 0 || i >= len(s) {
		return zero
	}
	return s[i]
}

// dropLast returns s without its final element, or s unchanged when empty.
func dropLast[T any](s []T) []T {
	if len(s) == 0 {
		return s
	}
	return s[:len(s)-1]
}

// maxIndexBy returns the index of the highest-scoring element, keeping the
// earliest on a tie and returning 0 for an empty slice -- the behaviour of the
// bodies it replaces, which seeded best := 0 and compared with a strict >.
func maxIndexBy[T any](s []T, score func(T) int) int {
	best := 0
	for i := range s {
		if i > 0 && score(s[i]) > score(s[best]) {
			best = i
		}
	}
	return best
}

// percentOf returns count as an integer percentage of total, or 0 when total is
// zero. Three betting statistics had this written out.
func percentOf(count, total int) int {
	if total == 0 {
		return 0
	}
	return count * 100 / total
}
