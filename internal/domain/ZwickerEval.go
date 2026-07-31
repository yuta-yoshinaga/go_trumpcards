//go:build !js || !wasm || extra2

package domain

import "sort"

// ZwickerJokerSmall / Middle / Large はジョーカー 3 枚に固定で割り当てられる
// マッチ値。**プレイヤーは選べない。**
//
// issue #4387 は「ジョーカーを出すときに任意の値を宣言してワイルドにする」と
// するが、原典 (pagat) ではジョーカーの値は固定である。値を選べるのはむしろ
// A と絵札のほう ([[ZwickerCardValues]])。
const (
	// ZwickerJokerSmall 小ジョーカーのマッチ値
	ZwickerJokerSmall = 15
	// ZwickerJokerMiddle 中ジョーカーのマッチ値
	ZwickerJokerMiddle = 20
	// ZwickerJokerLarge 大ジョーカーのマッチ値
	ZwickerJokerLarge = 25
)

// ZwickerCardValues は c が取りうるマッチ値をすべて返す。
//
// **A と絵札は 2 つの値を持ち、どちらで扱うかはプレイヤーが決める。**ここが
// cassino と決定的に違う点で、cassino では絵札はランク一致でしか取れず合計に
// 参加しない。Zwicker では絵札も合計に参加する。
//
//	A → 1 か 11 / J → 2 か 12 / Q → 3 か 13 / K → 4 か 14
//	2〜10 → 額面
//	ジョーカー → 15 / 20 / 25 のいずれか (固定)
func ZwickerCardValues(c *Card) []int {
	if c == nil {
		return nil
	}
	if c.GetDesign() == CardDesignJoker {
		switch c.GetValue() {
		case 1:
			return []int{ZwickerJokerSmall}
		case 2:
			return []int{ZwickerJokerMiddle}
		default:
			return []int{ZwickerJokerLarge}
		}
	}
	switch v := c.GetValue(); v {
	case 1:
		return []int{1, 11}
	case 11:
		return []int{2, 12}
	case 12:
		return []int{3, 13}
	case 13:
		return []int{4, 14}
	default:
		return []int{v}
	}
}

// ZwickerHasValue は c を value として扱えるかを返す。
func ZwickerHasValue(c *Card, value int) bool {
	for _, v := range ZwickerCardValues(c) {
		if v == value {
			return true
		}
	}
	return false
}

// ZwickerScoreOfCard は 1 枚が最終集計にもたらす点を返す。
//
// 大 7 / 中 6 / 小 5 / ♦10 3 / ♠10 1 / ♠2 1 / 各 A 1。
// これに「枚数最多 3」を足して**ちょうど 30 点**になる。
func ZwickerScoreOfCard(c *Card) int {
	if c == nil {
		return 0
	}
	if c.GetDesign() == CardDesignJoker {
		switch c.GetValue() {
		case 1:
			return ZwickerScoreSmallJoker
		case 2:
			return ZwickerScoreMiddleJoker
		default:
			return ZwickerScoreLargeJoker
		}
	}
	switch {
	case c.GetDesign() == CardDesignDiamond && c.GetValue() == 10:
		return ZwickerScoreDiamondTen
	case c.GetDesign() == CardDesignSpade && c.GetValue() == 10:
		return ZwickerScoreSpadeTen
	case c.GetDesign() == CardDesignSpade && c.GetValue() == 2:
		return ZwickerScoreSpadeTwo
	case c.GetValue() == 1:
		return ZwickerScoreAce
	default:
		return 0
	}
}

// zwickerCanPartition は cards を「それぞれ合計が target になる 1 つ以上の
// グループ」に余りなく分けられるかを返す。
//
// 各札は 2 つの値を持ちうるので、値の選び方まで含めて探索する。**1 グループに
// 限らないのが要点**で、10 を出して 7+3 と 6+4 を同時に取るのが Zwicker の
// 気持ちよさそのものである。
func zwickerCanPartition(cards []*Card, target int) bool {
	if len(cards) == 0 || target <= 0 {
		return false
	}
	used := make([]bool, len(cards))
	return zwickerFillGroup(cards, used, target, target, len(cards))
}

// zwickerFillGroup は「今のグループの残り remaining」を埋めながら再帰する。
// left は未使用の枚数。
func zwickerFillGroup(cards []*Card, used []bool, remaining, target, left int) bool {
	if remaining == 0 {
		if left == 0 {
			return true
		}
		// グループが 1 つ埋まった。次のグループを始める。
		return zwickerFillGroup(cards, used, target, target, left)
	}
	if left == 0 {
		return false
	}
	// 未使用の最小添字から順に試す。グループ内の順序は結果に影響しないので、
	// 「最小の未使用札は必ず今のグループに入る」としてよい。
	first := -1
	for i := range cards {
		if !used[i] {
			first = i
			break
		}
	}
	if first < 0 {
		return false
	}
	start := first
	if remaining != target {
		// グループ途中なら、どの未使用札も候補になる。
		start = 0
	}
	for i := start; i < len(cards); i++ {
		if used[i] {
			continue
		}
		if remaining == target && i != first {
			// 新しいグループの先頭は最小添字に固定 (重複探索の枝刈り)。
			break
		}
		for _, v := range ZwickerCardValues(cards[i]) {
			if v > remaining {
				continue
			}
			used[i] = true
			if zwickerFillGroup(cards, used, remaining-v, target, left-1) {
				used[i] = false
				return true
			}
			used[i] = false
		}
	}
	return false
}

// zwickerSortedUnique は添字集合を昇順・重複なしにする。範囲外があれば false。
func zwickerSortedUnique(idxs []int, size int) ([]int, bool) {
	seen := make(map[int]struct{}, len(idxs))
	out := make([]int, 0, len(idxs))
	for _, i := range idxs {
		if i < 0 || i >= size {
			return nil, false
		}
		if _, dup := seen[i]; dup {
			return nil, false
		}
		seen[i] = struct{}{}
		out = append(out, i)
	}
	sort.Ints(out)
	return out, true
}
