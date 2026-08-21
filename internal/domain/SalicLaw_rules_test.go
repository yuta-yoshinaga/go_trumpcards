//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Salic Law's four divergences from Congress ---
//
// Congress is the clone source: 8 piles, 8 foundations, two decks. Everything
// below is a rule Congress does NOT have, so each test carries a negative
// control that the Congress predicate would fail.

// **クイーン 8 枚は使わない。**ゲーム名の由来（女子の継承を認めない）で、
// 96 枚で遊ぶ。Congress は 104 枚すべてを使う。
func TestSalicLaw_QueensAreOutOfPlay(t *testing.T) {
	c := newTestSalicLaw()

	queens := c.GetQueens()
	require.Len(t, queens, SalicLawQueenCnt)
	for _, q := range queens {
		assert.Equal(t, SalicLawQueenValue, q.GetValue())
	}

	// 場に出る枚数の合計が 96。負のコントロール: 104 なら Q を抜いていない。
	inPlay := c.GetStockCount()
	for _, pile := range c.GetTableau() {
		inPlay += len(pile)
	}
	assert.Equal(t, SalicLawTotalCards, inPlay)
	assert.Equal(t, CardCnt*2-SalicLawQueenCnt, inPlay, "104 枚のままなら Q を抜いていない")
}

// **配りは K が出るまで 1 列。**最初に K を 1 枚据えて始め、以降めくった札は
// 今の列に積み、K が出たら次の列の土台になる。Congress は 8 山に 1 枚ずつ配って
// 残りを山札にする。
func TestSalicLaw_DealOpensAPileOnlyOnAKing(t *testing.T) {
	c := newTestSalicLaw()

	// 開始時は 1 列だけ、しかもその中身は K 1 枚。
	assert.Equal(t, 1, c.GetOpenPiles())
	require.Len(t, c.GetTableau()[0], 1)
	assert.Equal(t, CardValueMax, c.GetTableau()[0][0].GetValue(), "土台は K")
	for i := 1; i < SalicLawTableauCnt; i++ {
		assert.Empty(t, c.GetTableau()[i], "pile %d はまだ開いていない", i)
	}

	// 山札を最後まで配ると 8 列すべてが開き、どの列も底が K。
	for c.GetStockCount() > 0 {
		require.NoError(t, c.Draw())
	}
	assert.Equal(t, SalicLawTableauCnt, c.GetOpenPiles())
	for i, pile := range c.GetTableau() {
		require.NotEmpty(t, pile, "pile %d", i)
		assert.Equal(t, CardValueMax, pile[0].GetValue(), "pile %d の底は K", i)
	}
}

// **組札はスートを見ない。**A から J まで、スート不問で 1 つずつ上げる。
// Congress は同スートで A→K。
func TestSalicLaw_FoundationIgnoresSuitAndStopsAtJack(t *testing.T) {
	c := newTestSalicLaw()
	clearSalicLawBoard(c)
	c.openPiles = SalicLawTableauCnt
	for i := range SalicLawTableauCnt {
		c.tableau[i] = []*Card{NewCard(CardDesignSpade, CardValueMax, true)}
	}

	// 負のコントロール: Congress ならスート違いで拒む形。
	c.foundation[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
	c.tableau[0] = append(c.tableau[0], NewCard(CardDesignHeart, 2, true))
	require.NoError(t, c.MoveTableauToFoundation(0), "異なるスートでも積める")
	assert.Len(t, c.foundation[0], 2)

	// J で止まる。Q は場に無く、K は土台なので、12 枚目は存在しない。
	c.foundation[1] = make([]*Card, 0, SalicLawFoundationTarget)
	for v := 1; v <= SalicLawFoundationTarget; v++ {
		c.foundation[1] = append(c.foundation[1], NewCard(CardDesignClover, v, true))
	}
	c.tableau[1] = append(c.tableau[1], NewCard(CardDesignClover, SalicLawFoundationTarget+1, true))
	assert.Error(t, c.MoveTableauToFoundation(1), "J の上には積めない")
}

// **タブロー同士は積めない。**唯一の例外が「K だけの列」で、そこには任意の
// 1 枚を置ける。Congress はどの山にも降順で積める。
func TestSalicLaw_OnlyABareKingAcceptsACard(t *testing.T) {
	c := newTestSalicLaw()
	clearSalicLawBoard(c)
	c.openPiles = SalicLawTableauCnt
	for i := range SalicLawTableauCnt {
		c.tableau[i] = []*Card{NewCard(CardDesignSpade, CardValueMax, true)}
	}
	// 列0 は K の上に ♦8 が乗っている（＝空きではない）。
	c.tableau[0] = append(c.tableau[0], NewCard(CardDesignDiamond, 8, true))
	// 列2 は K の上に ♥7 が乗っている。これを動かす側にする。
	c.tableau[2] = append(c.tableau[2], NewCard(CardDesignHeart, 7, true))

	// 負のコントロール: ♦8 の上の ♥7 は Congress では降順で合法。ここでは拒む。
	assert.Error(t, c.MoveTableauToTableau(2, 0), "K だけでない列には置けない")

	// 列1 は K 1 枚 = 空き。ここへは置ける。
	require.NoError(t, c.MoveTableauToTableau(2, 1))
	assert.Len(t, c.tableau[1], 2)
	assert.Len(t, c.tableau[2], 1, "動かした側は K だけに戻る")

	// 土台の K そのものは動かせない。
	assert.Error(t, c.MoveTableauToTableau(2, 3), "K を土台から剥がせない")
}

// **まだ開いていない列の組札は使えない。**PySolFC の "if a king is beneath
// them"。8 つの組札は 8 つの K 列と対で、列が開くまで受け付けない。
func TestSalicLaw_FoundationNeedsItsKingPileOpen(t *testing.T) {
	c := newTestSalicLaw()
	clearSalicLawBoard(c)
	c.openPiles = 1
	c.tableau[0] = []*Card{
		NewCard(CardDesignSpade, CardValueMax, true),
		NewCard(CardDesignHeart, 1, true),
	}

	// A は組札 0 へは行けるが、まだ開いていない 1..7 へは行けない。
	assert.True(t, c.canPlaceOnFoundation(c.tableau[0][1], 0))
	for i := 1; i < SalicLawFoundationCnt; i++ {
		assert.False(t, c.canPlaceOnFoundation(c.tableau[0][1], i), "組札 %d はまだ使えない", i)
	}

	// 負のコントロール: 列が開けば同じ札が同じ組札へ行ける。
	c.openPiles = SalicLawFoundationCnt
	assert.True(t, c.canPlaceOnFoundation(c.tableau[0][1], SalicLawFoundationCnt-1))
}

// **勝利は K 以外の 88 枚。**8 つの K は土台に残る。Congress は 104 枚全部を
// 組札へ積む。
func TestSalicLaw_ClearsWithTheKingsStillOnTheBoard(t *testing.T) {
	c := newTestSalicLaw()
	clearSalicLawBoard(c)
	c.openPiles = SalicLawTableauCnt
	for i := range SalicLawTableauCnt {
		c.tableau[i] = []*Card{NewCard(CardDesignSpade, CardValueMax, true)}
		c.foundation[i] = make([]*Card, 0, SalicLawFoundationTarget)
		for v := 1; v <= SalicLawFoundationTarget; v++ {
			c.foundation[i] = append(c.foundation[i], NewCard(CardDesignSpade, v, true))
		}
	}
	c.checkGameClear()

	assert.Equal(t, SalicLawPhaseGameClear, c.GetPhase())
	assert.Equal(t, SalicLawTableauCnt*SalicLawFoundationTarget, 88, "K を除く 88 枚")
}

// **プレースホルダを持つ文言には必ず値を渡す。**`errPileEmpty` と
// `errInvalidPile` の訳文は `{{pile}}` を含むので、params に nil を渡すと
// 画面に `{{pile}}` がそのまま出る。lint も既存のコードテストも見ていない。
func TestSalicLaw_PileErrorsCarryTheColumnNumber(t *testing.T) {
	c := newTestSalicLaw()
	clearSalicLawBoard(c)
	openAllSalicLawPiles(c)
	c.tableau[4] = nil

	for name, act := range map[string]func() error{
		"empty column to a foundation": func() error { return c.MoveTableauToFoundation(4) },
		"empty column sideways":        func() error { return c.MoveTableauToTableau(4, 1) },
		"out of range to a foundation": func() error { return c.MoveTableauToFoundation(99) },
		"out of range source sideways": func() error { return c.MoveTableauToTableau(99, 1) },
		"out of range target sideways": func() error { return c.MoveTableauToTableau(1, 99) },
	} {
		t.Run(name, func(t *testing.T) {
			err := act()
			require.Error(t, err)
			_, params := ErrorMessageCode(err)
			require.NotNil(t, params, "文言が {{pile}} を持つので params が要る")
			assert.NotEmpty(t, params["pile"])
		})
	}
}
