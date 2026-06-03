//go:build !js || !wasm || casino

package domain

import "sort"

// BadugiHandNames maps Size (1..4) to the canonical English display name.
// Index 0 is unused ("" placeholder).
var BadugiHandNames = []string{
	"", // 0 unused
	"1-card",
	"2-card",
	"3-card",
	"Badugi",
}

// BadugiHand is the result of evaluating a 4-card Badugi hand:
// the best subset whose ranks and suits are all distinct. In Badugi,
// smaller Size hands lose to larger ones; among equal-sized hands the
// hand with the lowest high card wins (Ace = 1, compared top-down).
type BadugiHand struct {
	// Cards is the chosen subset, sorted by GetValue() in descending order
	// so Cards[0] is the "top" (worst) card used for lowball comparison.
	Cards []*Card
	// Size is len(Cards); 1..4 for a valid hand, 0 for invalid input.
	Size int
}

// evalBadugiHand returns the best BadugiHand extractable from a 4-card hand.
// Input must have exactly 4 cards; any other length returns a zero BadugiHand
// (Size == 0). Selection rule:
//  1. Prefer the largest subset with all-distinct ranks and all-distinct suits.
//  2. Break ties by picking the subset with the lowest high card, then second
//     card, and so on (standard Badugi lowball ordering with Ace = 1).
func evalBadugiHand(hand []*Card) BadugiHand {
	if len(hand) != 4 {
		return BadugiHand{}
	}

	var best BadugiHand
	// Enumerate all 15 non-empty subsets via a 4-bit mask.
	for mask := range 1 << 4 {
		if mask == 0 {
			continue
		}
		subset := badugiMaskSubset(hand, mask)
		if !badugiSubsetValid(subset) {
			continue
		}
		sortBadugiDesc(subset)
		cand := BadugiHand{Cards: subset, Size: len(subset)}
		if badugiCandidateBeats(cand, best) {
			best = cand
		}
	}
	return best
}

// compareBadugiHands returns -1 if a is the stronger (lower) hand, 1 if b is
// stronger, 0 on tie. Larger Size wins; on equal Size, compare sorted values
// descending — the lower hand wins. Ace is always 1.
func compareBadugiHands(a, b BadugiHand) int {
	if a.Size != b.Size {
		if a.Size > b.Size {
			return -1
		}
		return 1
	}
	for i := 0; i < a.Size; i++ {
		av := a.Cards[i].GetValue()
		bv := b.Cards[i].GetValue()
		if av < bv {
			return -1
		}
		if av > bv {
			return 1
		}
	}
	return 0
}

// badugiMaskSubset selects cards from hand according to the bits set in mask.
func badugiMaskSubset(hand []*Card, mask int) []*Card {
	out := make([]*Card, 0, 4)
	for i, card := range hand {
		if mask&(1<<i) != 0 {
			out = append(out, card)
		}
	}
	return out
}

// badugiSubsetValid reports whether the subset has all-distinct ranks and
// all-distinct suits. Badugi uses a 52-card deck with no wild cards, so this
// path is never exercised with Jokers.
func badugiSubsetValid(cards []*Card) bool {
	var suits, ranks uint32
	for _, c := range cards {
		sb := uint32(1) << c.GetDesign()
		rb := uint32(1) << c.GetValue()
		if suits&sb != 0 || ranks&rb != 0 {
			return false
		}
		suits |= sb
		ranks |= rb
	}
	return true
}

// sortBadugiDesc sorts cards by value descending so Cards[0] is the high card.
func sortBadugiDesc(cards []*Card) {
	sort.Slice(cards, func(i, j int) bool {
		return cards[i].GetValue() > cards[j].GetValue()
	})
}

// badugiCandidateBeats reports whether cand is a strictly better Badugi hand
// than best. Empty best is always beaten.
func badugiCandidateBeats(cand, best BadugiHand) bool {
	if best.Size == 0 {
		return true
	}
	return compareBadugiHands(cand, best) < 0
}

// FindPotWinnersBadugi selects the best Badugi hand among the eligible
// players for pot distribution; equal hands split. Designed to be passed to
// DistributePotsWithWinnerFunc.
//
// Each player's GetComparisonCards() must return the pre-sorted best subset
// cached by BadugiPlayer.EvalHand — we rebuild BadugiHand{Cards,Size} from
// that slice rather than re-running the 4-card evaluator here.
func FindPotWinnersBadugi(players []BettingPlayer, eligible []int) []int {
	var best BadugiHand
	bestInit := false
	var winners []int

	for _, idx := range eligible {
		pl := players[idx]
		if pl.GetFolded() {
			continue
		}
		cards := pl.GetComparisonCards()
		cand := BadugiHand{Cards: cards, Size: len(cards)}

		if !bestInit {
			best = cand
			winners = []int{idx}
			bestInit = true
			continue
		}
		switch compareBadugiHands(cand, best) {
		case -1:
			best = cand
			winners = []int{idx}
		case 0:
			winners = append(winners, idx)
		}
	}
	return winners
}
