//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Sly Fox's three divergences from Colorado ---
//
// Colorado is the clone source, and the two are conflated in some references
// (Wikipedia redirects one to the other). PySolFC implements them separately,
// and these are the differences. Each test carries a negative control that the
// Colorado predicate would fail.

// **山札からの札は捨て札を経由しない。**プレイヤーがその場で「どのリザーブ枠に
// 置くか」を決める ── その 20 回の選択がこのゲームそのもの。コロラドは捨て札を
// 挟むので、置き先を後から選び直せる。
func TestSlyFox_ADealtCardGoesStraightToAChosenPile(t *testing.T) {
	c := newTestSlyFox()
	clearSlyFoxBoard(c)
	c.stock = []*Card{NewCard(CardDesignHeart, 7, true)}

	require.NoError(t, c.DealToPile(3))
	assert.Len(t, c.tableau[3], 1, "選んだ枠に乗る")
	assert.Equal(t, 0, c.GetStockCount())
	assert.Equal(t, 1, c.DealtThisCycle())
}

// **20 枚配り切るまでリザーブから組札へ送れない。**これが Sly Fox の核。
// コロラドはいつでも送れる。
func TestSlyFox_ReserveIsLockedUntilTwentyHaveBeenDealt(t *testing.T) {
	setup := func() *SlyFox {
		c := newTestSlyFox()
		clearSlyFoxBoard(c)
		startSlyFoxCycle(c)
		// リザーブ 0 の上に ♠A。組札 0 (昇順) がちょうど受け取れる形。
		c.tableau[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
		for i := range 30 {
			c.stock = append(c.stock, NewCard(CardDesignHeart, i%13+1, true))
		}
		return c
	}

	// 配りかけ (0/20) では送れない。
	c := setup()
	err := c.MoveTableauToFoundation(0)
	require.Error(t, err)
	code, _ := ErrorMessageCode(err)
	assert.Equal(t, "slyfox.errDealInProgress", code)

	// 19 枚では、まだ送れない。境界を 1 手前で見る。
	c = setup()
	for range SlyFoxDealCycle - 1 {
		require.NoError(t, c.DealToPile(1))
	}
	assert.Error(t, c.MoveTableauToFoundation(0), "19 枚では開かない")

	// 20 枚で開く。
	require.NoError(t, c.DealToPile(1))
	assert.Equal(t, SlyFoxDealCycle, c.DealtThisCycle())
	require.NoError(t, c.MoveTableauToFoundation(0))
	assert.Len(t, c.foundation[0], 1)
}

// **配りの途中でも、めくった札そのものは組札へ送れる。**送った札は 20 枚に
// 数えない ── 数えると、運の良い引きが配りの手数を食ってしまう。
func TestSlyFox_ADealtCardMayGoStraightToAFoundationWithoutCounting(t *testing.T) {
	c := newTestSlyFox()
	clearSlyFoxBoard(c)
	startSlyFoxCycle(c)
	c.stock = []*Card{NewCard(CardDesignSpade, 1, true), NewCard(CardDesignHeart, 9, true)}

	require.NoError(t, c.DealToFoundation(0), "♠A は昇順の組札 0 を開ける")
	assert.Len(t, c.foundation[0], 1)
	assert.Equal(t, 0, c.DealtThisCycle(), "組札行きは 20 枚に数えない")

	// 負のコントロール: 置けない札は組札へ送れない。数えないことを利用して
	// 山札を素通しできてしまわないこと。
	assert.Error(t, c.DealToFoundation(0), "♥9 は ♠A の上には積めない")
	assert.Equal(t, 1, c.GetStockCount(), "拒まれた札は山札に残る")
}

// **20 枚を配り切ったらカウンタが戻り、また閉じる。**1 周で開きっぱなしになると、
// 2 周目以降がコロラドと同じになる。
func TestSlyFox_TheCycleClosesAgainOnTheNextDeal(t *testing.T) {
	c := newTestSlyFox()
	clearSlyFoxBoard(c)
	c.tableau[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
	for range 30 {
		c.stock = append(c.stock, NewCard(CardDesignHeart, 9, true))
	}
	for range SlyFoxDealCycle {
		require.NoError(t, c.DealToPile(1))
	}
	require.NoError(t, c.MoveTableauToFoundation(0))

	// 次の 1 枚を配った時点で、また閉じる。
	require.NoError(t, c.DealToPile(2))
	assert.Equal(t, 1, c.DealtThisCycle())
	c.tableau[0] = []*Card{NewCard(CardDesignSpade, 2, true)}
	assert.Error(t, c.MoveTableauToFoundation(0), "次の周が始まったら閉じる")
}

// **山札が尽きたら制限は無くなる。**残ったリザーブだけで詰めていく終盤が
// あるので、ここで閉じたままだと必ず手詰まりになる。
func TestSlyFox_TheLockLiftsOnceTheStockIsEmpty(t *testing.T) {
	c := newTestSlyFox()
	clearSlyFoxBoard(c)
	startSlyFoxCycle(c)
	c.tableau[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
	c.stock = nil

	assert.Equal(t, 0, c.DealtThisCycle())
	require.NoError(t, c.MoveTableauToFoundation(0), "山札が空なら配りかけでも送れる")
}

// **空いたリザーブ枠は補充されない。**コロラドは山札から直接埋められる
// (`MoveStockToTableau`)。Sly Fox にその手は無く、空き枠は次に配るときの
// 置き場所として使うだけ。
func TestSlyFox_AnEmptyPileIsNotRefilledFromTheStock(t *testing.T) {
	c := newTestSlyFox()
	clearSlyFoxBoard(c)
	c.stock = []*Card{NewCard(CardDesignHeart, 7, true)}

	// 空き枠へは「配る」ことでしか札が入らない。配れば入る。
	require.NoError(t, c.DealToPile(5))
	assert.Len(t, c.tableau[5], 1)
}

// **縛りは「送れる手」の定義そのもの。**`MoveTableauToFoundation` だけに書くと、
// 同じ盤を読む他の経路がそれを知らないまま先へ進む（St. Helena で同じ穴を出した）。
func TestSlyFox_TheLockBindsEveryPathToAFoundation(t *testing.T) {
	setup := func() *SlyFox {
		c := newTestSlyFox()
		clearSlyFoxBoard(c)
		startSlyFoxCycle(c)
		// リザーブ 0 の ♠A は組札 0 にちょうど乗る ── 開いてさえいれば。
		c.tableau[0] = []*Card{NewCard(CardDesignSpade, 1, true)}
		c.stock = []*Card{NewCard(CardDesignHeart, 9, true)}
		return c
	}

	// 前提: 手で送るのは拒まれる。ここが通ると以下は何も測らない。
	require.Error(t, setup().MoveTableauToFoundation(0))

	t.Run("GetHint does not point at the reserve", func(t *testing.T) {
		h := setup().GetHint()
		if h != nil && h.FromZone == "tableau" {
			assert.Fail(t, "閉じている間にリザーブの手を指している", "%+v", h)
		}
	})

	t.Run("AutoComplete does not play it", func(t *testing.T) {
		c := setup()
		assert.Error(t, c.AutoComplete(), "送れる手が無いので拒む")
		assert.Empty(t, c.GetFoundation()[0])
		assert.Len(t, c.GetTableau()[0], 1)
	})

	// 負のコントロール: 周を配り切れば、どちらの経路でも同じ手を選ぶ。
	// 「常に送らない」実装でも上の 2 つは通ってしまう。
	t.Run("both play it once the cycle completes", func(t *testing.T) {
		open := func() *SlyFox {
			c := setup()
			c.dealtThisCycle = SlyFoxDealCycle
			return c
		}
		h := open().GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "tableau", h.FromZone)
		assert.Equal(t, "foundation", h.ToZone)

		c := open()
		require.NoError(t, c.AutoComplete())
		assert.Len(t, c.GetFoundation()[0], 1)
	})
}
