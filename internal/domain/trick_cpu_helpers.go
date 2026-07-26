// Package domain トリックテイキング系 CPU AI 共通の戦術サブステップ。
//
// No build tag: compiled into every build including all Cloudflare Worker WASM
// binaries, matching slice_helpers.go / player_helpers.go / trick_helpers.go.
//
// These are the small, game-agnostic building blocks the trick-taking CPU
// strategies compose (issue #4300): "filter the candidate indices" and "pick the
// weakest / strongest of them". The per-game `cpuPlayHard` keeps its own
// game-specific assembly of these steps — the AI itself stays game-specific; only
// the mechanical sub-steps are shared.
//
// Card strength is supplied by the caller as `rank func(*Card) int`; nil means
// the natural (*Card).GetValue order. Games with a bespoke ordering (e.g. Court
// Piece / Tarneeb rank tables) pass their own function.
package domain

// cardIndexer is the minimal view of a player needed by the tactical helpers:
// index-addressable access to its hand.
type cardIndexer interface {
	GetCard(int) *Card
}

// cardRankOr returns rank, or (*Card).GetValue when rank is nil.
func cardRankOr(rank func(*Card) int) func(*Card) int {
	if rank == nil {
		return (*Card).GetValue
	}
	return rank
}

// filterByDesign returns the subset of indices whose card has the given suit.
func filterByDesign[P cardIndexer](player P, indices []int, design int) []int {
	out := make([]int, 0, len(indices))
	for _, idx := range indices {
		if player.GetCard(idx).GetDesign() == design {
			out = append(out, idx)
		}
	}
	return out
}

// filterAbove returns the subset of indices whose card ranks strictly above
// threshold.
func filterAbove[P cardIndexer](player P, indices []int, threshold int, rank func(*Card) int) []int {
	r := cardRankOr(rank)
	out := make([]int, 0, len(indices))
	for _, idx := range indices {
		if r(player.GetCard(idx)) > threshold {
			out = append(out, idx)
		}
	}
	return out
}

// filterBelow returns the subset of indices whose card ranks strictly below
// threshold.
func filterBelow[P cardIndexer](player P, indices []int, threshold int, rank func(*Card) int) []int {
	r := cardRankOr(rank)
	out := make([]int, 0, len(indices))
	for _, idx := range indices {
		if r(player.GetCard(idx)) < threshold {
			out = append(out, idx)
		}
	}
	return out
}

// pickLowest returns the index of the lowest-ranked card among indices, or -1
// when indices is empty. Ties keep the earlier index.
func pickLowest[P cardIndexer](player P, indices []int, rank func(*Card) int) int {
	if len(indices) == 0 {
		return -1
	}
	r := cardRankOr(rank)
	best := indices[0]
	bestR := r(player.GetCard(best))
	for _, idx := range indices[1:] {
		if v := r(player.GetCard(idx)); v < bestR {
			bestR = v
			best = idx
		}
	}
	return best
}

// pickHighest returns the index of the highest-ranked card among indices, or -1
// when indices is empty. Ties keep the earlier index.
func pickHighest[P cardIndexer](player P, indices []int, rank func(*Card) int) int {
	if len(indices) == 0 {
		return -1
	}
	r := cardRankOr(rank)
	best := indices[0]
	bestR := r(player.GetCard(best))
	for _, idx := range indices[1:] {
		if v := r(player.GetCard(idx)); v > bestR {
			bestR = v
			best = idx
		}
	}
	return best
}
