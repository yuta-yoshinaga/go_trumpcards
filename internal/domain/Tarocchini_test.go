//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func tarocchiniTrick(entries ...[2]int) []*TrickCard {
	trick := make([]*TrickCard, 0, len(entries))
	for seat, e := range entries {
		trick = append(trick, &TrickCard{PlayerIdx: seat, Card: NewCard(e[0], e[1], false)})
	}
	return trick
}

func TestTarocchini_DeckIsSixtyTwoDistinctCards(t *testing.T) {
	deck := buildTarocchiniDeck()
	assert.Len(t, deck, TarocchiniDeckSize)
	assert.Equal(t, 62, TarocchiniDeckSize)

	seen := map[[2]int]bool{}
	suits, trumps, matto := 0, 0, 0
	for _, c := range deck {
		key := [2]int{c.GetDesign(), c.GetValue()}
		assert.False(t, seen[key], "duplicate card %v", key)
		seen[key] = true
		switch {
		case tarocchiniIsMatto(c):
			matto++
		case tarocchiniIsTrump(c):
			trumps++
		default:
			suits++
			// **2..5 は抜かれている。**52 枚デッキの感覚で 1..10 を作ると枚数が合わない。
			assert.NotContains(t, []int{2, 3, 4, 5}, c.GetValue(),
				"low pips must not be in a Bolognese deck")
		}
	}
	assert.Equal(t, 40, suits)
	assert.Equal(t, TarocchiniMaxTrump, trumps)
	assert.Equal(t, 1, matto)
}

// 62 は 4 で割り切れないので、余りが出ることを構造として固定しておく。
func TestTarocchini_DealLeavesASurplus(t *testing.T) {
	assert.Equal(t, 2, TarocchiniSurplus)
	assert.Equal(t, TarocchiniDeckSize, TarocchiniPlayerCnt*TarocchiniHandSize+TarocchiniSurplus)
}

func TestTarocchini_TeamsAreOpposite(t *testing.T) {
	assert.Equal(t, TarocchiniTeamOf(0), TarocchiniTeamOf(2))
	assert.Equal(t, TarocchiniTeamOf(1), TarocchiniTeamOf(3))
	assert.NotEqual(t, TarocchiniTeamOf(0), TarocchiniTeamOf(1))
}

// **これがこのゲームの核心。**同格のパパが複数出たら後から出した方が勝つ。
// 共通の「厳密に強い札だけが勝者を更新する」判定では先に出した方が勝ってしまう。
func TestTarocchini_LaterPapaBeatsEarlierPapa(t *testing.T) {
	led := 1
	cases := []struct {
		name  string
		trick []*TrickCard
		want  int
	}{
		{"two papi: the later one wins", tarocchiniTrick(
			[2]int{TarocchiniTrumpDesign, 2},
			[2]int{led, 14},
			[2]int{TarocchiniTrumpDesign, 5},
			[2]int{led, 13},
		), 2},
		{"the papi order on the table does not matter, only lateness", tarocchiniTrick(
			[2]int{TarocchiniTrumpDesign, 5},
			[2]int{led, 14},
			[2]int{TarocchiniTrumpDesign, 2},
			[2]int{led, 13},
		), 2},
		{"all four papi: the last one wins", tarocchiniTrick(
			[2]int{TarocchiniTrumpDesign, 2},
			[2]int{TarocchiniTrumpDesign, 3},
			[2]int{TarocchiniTrumpDesign, 4},
			[2]int{TarocchiniTrumpDesign, 5},
		), 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tarocchiniTrickWinnerOf(tc.trick, led))
		})
	}
}

// 同格なのはパパ同士だけ。通常の切り札は数字順で、後出しでは勝てない。
func TestTarocchini_OrdinaryTrumpsKeepStrictOrder(t *testing.T) {
	led := 1
	assert.Equal(t, 0, tarocchiniTrickWinnerOf(tarocchiniTrick(
		[2]int{TarocchiniTrumpDesign, 20},
		[2]int{TarocchiniTrumpDesign, 19},
		[2]int{led, 14},
		[2]int{led, 13},
	), led), "a later but lower trump must not win")
}

// パパは中位の同格札であって最強札ではない。上位切り札には負ける。
func TestTarocchini_PapaLosesToAHigherTrump(t *testing.T) {
	led := 1
	assert.Equal(t, 1, tarocchiniTrickWinnerOf(tarocchiniTrick(
		[2]int{TarocchiniTrumpDesign, tarocchiniPapiHigh},
		[2]int{TarocchiniTrumpDesign, TarocchiniMaxTrump},
		[2]int{led, 14},
		[2]int{led, 13},
	), led), "the highest trump must beat a papa played earlier")

	// 席 0 が最強切り札、席 1 が後からパパ。後出しでも上位切り札には届かない。
	assert.Equal(t, 0, tarocchiniTrickWinnerOf(tarocchiniTrick(
		[2]int{TarocchiniTrumpDesign, TarocchiniMaxTrump},
		[2]int{TarocchiniTrumpDesign, tarocchiniPapiLow},
		[2]int{led, 14},
		[2]int{led, 13},
	), led), "a papa played later must not beat a higher trump")
}

func TestTarocchini_TrickBasics(t *testing.T) {
	led := 2
	t.Run("highest of the led suit wins with no trump", func(t *testing.T) {
		assert.Equal(t, 1, tarocchiniTrickWinnerOf(tarocchiniTrick(
			[2]int{led, 9}, [2]int{led, 14}, [2]int{led, 1}, [2]int{3, 14},
		), led))
	})
	t.Run("any trump beats any plain card", func(t *testing.T) {
		assert.Equal(t, 2, tarocchiniTrickWinnerOf(tarocchiniTrick(
			[2]int{led, 14}, [2]int{led, 13}, [2]int{TarocchiniTrumpDesign, 1}, [2]int{led, 12},
		), led))
	})
	t.Run("an off-suit discard never wins", func(t *testing.T) {
		assert.Equal(t, 0, tarocchiniTrickWinnerOf(tarocchiniTrick(
			[2]int{led, 6}, [2]int{4, 14}, [2]int{3, 14}, [2]int{1, 14},
		), led))
	})
	t.Run("the Matto never takes the trick", func(t *testing.T) {
		assert.Equal(t, 1, tarocchiniTrickWinnerOf(tarocchiniTrick(
			[2]int{TarocchiniMattoDesign, TarocchiniMattoValue},
			[2]int{led, 6}, [2]int{4, 14}, [2]int{3, 14},
		), led))
	})
	t.Run("an empty trick does not panic", func(t *testing.T) {
		assert.Equal(t, 0, tarocchiniTrickWinnerOf(nil, led))
	})
}
