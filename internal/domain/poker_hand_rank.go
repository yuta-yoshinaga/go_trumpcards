package domain

// poker_hand_rank.go holds the poker hand-rank constants and names. They live
// in an untagged (core) file — not in a per-category game file — because the
// core hand evaluators (hand_eval.go, kicker.go) and non-casino games such as
// PokerSquares (solo worker) depend on them, so they must compile into every
// Cloudflare Worker (#2126).

// ポーカーハンドランク定数
const (
	PokerHandHighCard      = 0
	PokerHandOnePair       = 1
	PokerHandTwoPair       = 2
	PokerHandThreeOfAKind  = 3
	PokerHandStraight      = 4
	PokerHandFlush         = 5
	PokerHandFullHouse     = 6
	PokerHandFourOfAKind   = 7
	PokerHandStraightFlush = 8
	PokerHandRoyalFlush    = 9
	PokerHandFiveOfAKind   = 10
)

// PokerHandNames ポーカーハンド名
var PokerHandNames = []string{
	"High Card",
	"One Pair",
	"Two Pair",
	"Three of a Kind",
	"Straight",
	"Flush",
	"Full House",
	"Four of a Kind",
	"Straight Flush",
	"Royal Flush",
	"Five of a Kind",
}

// pokerHandName returns the display name for a hand rank, or "Unknown" when the
// rank is outside the table. 5 betting games had this written out.
func pokerHandName(rank int) string {
	if rank >= 0 && rank < len(PokerHandNames) {
		return PokerHandNames[rank]
	}
	return "Unknown"
}

// dealerQualifies reports whether the dealer's hand meets the usual
// ace-king-or-better qualification. 3 casino games had this written out.
func dealerQualifies(handRank int, hand []*Card) bool {
	if handRank >= PokerHandOnePair {
		return true
	}
	hasAce := false
	hasKing := false
	for _, c := range hand {
		switch c.GetValue() {
		case 1:
			hasAce = true
		case 13:
			hasKing = true
		}
	}
	return hasAce && hasKing
}
