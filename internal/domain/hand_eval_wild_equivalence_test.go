//go:build test

package domain

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

// evalWildHandReference is the exhaustive, unoptimised evaluation.
//
// **This is the oracle, not a second implementation to keep in sync.** It
// walks every ordered assignment of the 52-card alphabet with no early exit,
// which is what `evalWildHand` did before the search was reduced to multisets.
// Keeping it here lets the reduction be proved rather than argued.
func evalWildHandReference(cards []*Card, isWild func(*Card) bool) int {
	if len(cards) != 5 {
		return PokerHandHighCard
	}
	if isWild == nil {
		return evalFiveCardHand(cards)
	}
	var wilds, nonWilds []*Card
	for _, c := range cards {
		if isWild(c) {
			wilds = append(wilds, c)
		} else {
			nonWilds = append(nonWilds, c)
		}
	}
	if len(wilds) == 0 {
		return evalFiveCardHand(cards)
	}
	if len(wilds) >= 4 {
		return PokerHandFiveOfAKind
	}

	best := PokerHandHighCard
	subs := make([]*Card, len(wilds))
	var walk func(depth int)
	walk = func(depth int) {
		if depth == len(wilds) {
			hand := make([]*Card, 0, 5)
			hand = append(hand, nonWilds...)
			hand = append(hand, subs...)
			if r := evalFiveCardHand(hand); r > best {
				best = r
			}
			return
		}
		for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
			for val := 1; val <= 13; val++ {
				subs[depth] = NewCard(suit, val, false)
				walk(depth + 1)
			}
		}
	}
	walk(0)
	return best
}

// **速い探索が遅い探索と同じ答えを返す。** 順列の重複を畳んだのと、上限を
// ロイヤルフラッシュからファイブカードに直したのは、どちらも探索する手の
// 集合を変えないはず ── それを主張ではなく実測で確かめる。
func TestEvalWildHand_MatchesTheExhaustiveSearch(t *testing.T) {
	isWild := func(c *Card) bool { return c != nil && (c.GetValue() == 3 || c.GetValue() == 9) }

	// 決め打ちの盤面 (ワイルド 0〜3 枚、フラッシュ・ストレート・ペアの種)。
	fixed := [][]*Card{
		{NewCard(1, 1, true), NewCard(1, 13, true), NewCard(1, 12, true), NewCard(1, 11, true), NewCard(1, 10, true)},
		{NewCard(1, 3, true), NewCard(2, 1, true), NewCard(3, 1, true), NewCard(4, 1, true), NewCard(1, 1, true)},
		{NewCard(1, 3, true), NewCard(2, 9, true), NewCard(3, 1, true), NewCard(4, 1, true), NewCard(1, 7, true)},
		{NewCard(1, 3, true), NewCard(2, 9, true), NewCard(3, 3, true), NewCard(4, 1, true), NewCard(1, 1, true)},
		{NewCard(1, 3, true), NewCard(2, 9, true), NewCard(3, 3, true), NewCard(4, 5, true), NewCard(1, 7, true)},
		{NewCard(1, 2, true), NewCard(2, 4, true), NewCard(3, 6, true), NewCard(4, 8, true), NewCard(1, 10, true)},
	}
	for i, hand := range fixed {
		got, _ := evalWildHand(hand, isWild)
		require.Equal(t, evalWildHandReference(hand, isWild), got, "決め打ちの盤面 %d", i)
	}

	// **ワイルドの枚数を決め打ちで振る。** 素の無作為だと 3 枚ワイルドは
	// 数百回に 1 度しか出ず、いちばん削った枝が検査されないまま緑になる。
	rng := rand.New(rand.NewSource(20260813))
	wildValues := []int{3, 9}
	for wilds := range 4 {
		for n := range 150 {
			hand := make([]*Card, 0, 5)
			for range wilds {
				hand = append(hand, NewCard(1+rng.Intn(4), wildValues[rng.Intn(2)], true))
			}
			for len(hand) < 5 {
				v := 1 + rng.Intn(13)
				if v == 3 || v == 9 {
					continue // ここは非ワイルドで埋める
				}
				hand = append(hand, NewCard(1+rng.Intn(4), v, true))
			}
			require.Equal(t, wilds, countWilds(hand, isWild), "配ったワイルド枚数")

			got, usedWild := evalWildHand(hand, isWild)
			require.Equal(t, evalWildHandReference(hand, isWild), got,
				"ワイルド %d 枚の盤面 %d: %v", wilds, n, hand)
			require.Equal(t, wilds > 0, usedWild, "ワイルド使用の記録 (%d 枚) %d", wilds, n)
		}
	}

	// **同スートの非ワイルド**は探索アルファベットを 2 スートに広げる枝。
	// 無作為に任せると滅多に出ないので明示的に踏む。
	for wilds := 1; wilds <= 3; wilds++ {
		for n := range 60 {
			hand := make([]*Card, 0, 5)
			for range wilds {
				hand = append(hand, NewCard(1+rng.Intn(4), wildValues[rng.Intn(2)], true))
			}
			suit := 1 + rng.Intn(4)
			for len(hand) < 5 {
				v := 1 + rng.Intn(13)
				if v == 3 || v == 9 {
					continue
				}
				hand = append(hand, NewCard(suit, v, true))
			}
			got, _ := evalWildHand(hand, isWild)
			require.Equal(t, evalWildHandReference(hand, isWild), got,
				"同スート・ワイルド %d 枚の盤面 %d: %v", wilds, n, hand)
		}
	}
}
