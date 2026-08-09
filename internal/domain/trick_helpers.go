// Package domain トリックテイキング系ゲーム共通のトリック勝者判定ヘルパー。
//
// No build tag: this file is compiled into every build, including all
// Cloudflare Worker WASM binaries. The helper is game-agnostic (each game
// supplies its own trump suit and card-ranking function), so it must be
// available regardless of which category worker a game lands in. Mirrors the
// rationale in slice_helpers.go / player_helpers.go.
package domain

// TrickCard is a single card played into a trick, tagged with the seat that
// played it. The vast majority of trick-taking games share this exact shape
// (issue #4297), so they reuse this type rather than redeclaring an identical
// per-game struct. Games needing extra per-play state (e.g. Mighty's joker-lead
// flag) keep their own bespoke type.
//
// The JSON tags match every game's historical per-game struct (`pi`/`c`) so the
// serialized KV snapshot shape is unchanged.
type TrickCard struct {
	PlayerIdx int   `json:"pi"`
	Card      *Card `json:"c"`
}

// ResolveTrickWinner returns the PlayerIdx of the card that wins trick.
//
// The first card sets the lead suit. A card of trumpSuit beats any non-trump;
// among trumps (and, when no trump is present, among lead-suit cards) the higher
// rank wins. Ties resolve to the earlier-played card (the comparison is strict).
// Pass trumpSuit < 0 (or any value no card's design can equal) for a no-trump
// deal. rank maps a card to its comparable strength; nil uses (*Card).GetValue
// (natural 2..14 order). An empty trick returns 0.
//
// This consolidates the "trump beats non-trump, else lead-suit highest by rank"
// winner logic that was hand-written identically across the standard
// trick-taking games; games with a bespoke ranking (e.g. suit-independent tarot
// orderings, joker/bower reordering, or delegated `beats` helpers) keep their
// own implementation.
func ResolveTrickWinner(trick []*TrickCard, trumpSuit int, rank func(*Card) int) int {
	if len(trick) == 0 {
		return 0
	}
	if rank == nil {
		rank = (*Card).GetValue
	}
	leadSuit := trick[0].Card.GetDesign()
	winnerIdx := trick[0].PlayerIdx
	winnerRank := rank(trick[0].Card)
	winnerIsTrump := trick[0].Card.GetDesign() == trumpSuit

	for _, tc := range trick[1:] {
		isTrump := tc.Card.GetDesign() == trumpSuit
		r := rank(tc.Card)
		switch {
		case isTrump && !winnerIsTrump:
			// Trump beats a non-trump leader.
			winnerIdx, winnerRank, winnerIsTrump = tc.PlayerIdx, r, true
		case isTrump && winnerIsTrump:
			// Trump vs trump: higher rank wins.
			if r > winnerRank {
				winnerIdx, winnerRank = tc.PlayerIdx, r
			}
		case !isTrump && !winnerIsTrump && tc.Card.GetDesign() == leadSuit && r > winnerRank:
			// Non-trump vs non-trump: higher lead-suit card wins.
			winnerIdx, winnerRank = tc.PlayerIdx, r
		}
	}
	return winnerIdx
}

// indexOfPlayerInTrick returns the position at which playerIdx played into the
// current trick, or -1 when that seat has not played yet. Position is play
// order, not seat order.
//
// 20 games had this loop written out. It needs no type parameter: every one of
// them holds its trick as []*TrickCard, so this is a plain function and adds no
// monomorphised copies. See issue #5185.
func indexOfPlayerInTrick(trick []*TrickCard, playerIdx int) int {
	for i, tc := range trick {
		if tc.PlayerIdx == playerIdx {
			return i
		}
	}
	return -1
}
