//go:build !js || !wasm || casino

package domain

// soko_hand_eval.go holds Soko's (Canadian Stud) own hand-rank scale.
//
// **Why a separate scale instead of inserting into poker_hand_rank.go.**
// Soko adds two ranks between One Pair and Two Pair, and the obvious
// implementation -- inserting two constants into the shared `PokerHand*`
// enum -- renumbers every rank above the insertion. That enum is not a private
// detail:
//
//   - `poker_hand_rank.go` is an untagged core file compiled into all six
//     Cloudflare Workers, and non-casino games (PokerSquares, solo worker)
//     depend on it;
//   - `PokerHandNames` is indexed positionally (`PokerHandNames[rank]`) from
//     five call sites, so a shift renames every hand in every poker game;
//   - the ranks leave the process as integers -- `handRank int` appears in
//     several Web outputs and `getHandRankHint(handRank: number)` branches on
//     the number in the frontend. A renumber would break the API contract and
//     the frontend hint mapping *without* a compile error on either side.
//
// So Soko keeps its own scale here and converts from the standard evaluator.
// Nothing outside Soko sees these constants.

// Soko ハンドランク定数。
// 標準ポーカーの序列に「4枚ストレート」と「4枚フラッシュ」をワンペアとツーペアの
// 間に挿入したもの。この2つの前後関係は Sökö の公表ルールに従い、
// 4枚フラッシュが4枚ストレートより上。
const (
	SokoHandHighCard      = 0
	SokoHandOnePair       = 1
	SokoHandFourStraight  = 2
	SokoHandFourFlush     = 3
	SokoHandTwoPair       = 4
	SokoHandThreeOfAKind  = 5
	SokoHandStraight      = 6
	SokoHandFlush         = 7
	SokoHandFullHouse     = 8
	SokoHandFourOfAKind   = 9
	SokoHandStraightFlush = 10
	SokoHandRoyalFlush    = 11
)

// sokoHandNames は Soko ハンド名。SokoHand* の値でインデックスする。
var sokoHandNames = []string{
	"High Card",
	"One Pair",
	"Four-Card Straight",
	"Four-Card Flush",
	"Two Pair",
	"Three of a Kind",
	"Straight",
	"Flush",
	"Full House",
	"Four of a Kind",
	"Straight Flush",
	"Royal Flush",
}

// sokoHandName returns the display name for a Soko rank, or "Unknown" when the
// rank is outside the table.
func sokoHandName(rank int) string {
	if rank < 0 || rank >= len(sokoHandNames) {
		return "Unknown"
	}
	return sokoHandNames[rank]
}

// sokoRankShift は標準ランクを Soko ランクへ写すときのオフセット。
// ツーペア以上は、下に2つ挿入された分だけ上へずれる。
const sokoRankShift = 2

// evalSokoHand は5枚の手札を Soko の序列で評価する。
//
// 標準評価をベースに、ツーペア未満のときだけ4枚役を見る。ツーペア以上の手は
// 定義上4枚役より強いので調べる必要がない。
//
// 4枚役はワンペアと共存しうる（ペアがあると異なるランクが4つ残るので、その4枚が
// 同スートだったり連続したりする）。標準評価はペアしか見ないため、そこで打ち切ると
// 4枚フラッシュを過小評価する。
func evalSokoHand(cards []*Card) int {
	if len(cards) != 5 {
		return SokoHandHighCard
	}

	base := evalFiveCardHand(cards)
	if base >= PokerHandTwoPair {
		return base + sokoRankShift
	}

	// base は HighCard(0) か OnePair(1) のみ。Soko の同名ランクと値が一致する。
	best := base
	if hasFourCardStraight(cards) && SokoHandFourStraight > best {
		best = SokoHandFourStraight
	}
	if hasFourCardFlush(cards) && SokoHandFourFlush > best {
		best = SokoHandFourFlush
	}
	return best
}

// hasFourCardFlush はちょうど4枚が同スートかを返す。
// 5枚同スートは本物のフラッシュで、そちらは標準評価が拾うのでここには来ない。
func hasFourCardFlush(cards []*Card) bool {
	counts := make(map[int]int, 4)
	for _, c := range cards {
		counts[c.GetDesign()]++
	}
	for _, n := range counts {
		if n == 4 {
			return true
		}
	}
	return false
}

// hasFourCardStraight は4枚が連続ランクかを返す。
// エースは A-2-3-4 の下端と J-Q-K-A の上端の両方で使えるので、値 1 は 14 としても
// 数える。
func hasFourCardStraight(cards []*Card) bool {
	present := make(map[int]bool, 6)
	for _, c := range cards {
		v := c.GetValue()
		present[v] = true
		if v == 1 {
			present[CardValueMax+1] = true // ace high
		}
	}
	// 開始値は A から 11 まで（11,12,13,14 が最上の並び）。
	for start := 1; start <= CardValueMax+1-3; start++ {
		run := true
		for i := range 4 {
			if !present[start+i] {
				run = false
				break
			}
		}
		if run {
			return true
		}
	}
	return false
}
