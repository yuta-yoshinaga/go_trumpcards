package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// paiGowSetHandsFixture は7枚を固定したセットハンドフェーズを作る。
// Reset() を通さないのは、配り次第で決まる手札を固定するため。
func paiGowSetHandsFixture(cards []*Card) *PaiGow {
	pg := NewDefaultPaiGow()
	pg.phase = PaiGowPhaseSetHands
	pg.playerCards = cards
	pg.dealerCards = []*Card{
		NewCard(CardDesignClover, 2, false), NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignClover, 4, false), NewCard(CardDesignClover, 6, false),
		NewCard(CardDesignHeart, 8, false), NewCard(CardDesignHeart, 10, false),
		NewCard(CardDesignDiamond, 12, false),
	}
	return pg
}

func TestPaiGow_GetHint(t *testing.T) {
	// **手札は配られているがフェーズが違う、を踏む。**素の NewDefaultPaiGow() は
	// 手札が空なのでハウスウェイ側で弾かれてしまい、フェーズ判定を踏まない。
	t.Run("returns nothing outside the set hands phase", func(t *testing.T) {
		pg := paiGowSetHandsFixture([]*Card{
			NewCard(CardDesignSpade, 14, false), NewCard(CardDesignHeart, 14, false),
			NewCard(CardDesignSpade, 13, false), NewCard(CardDesignHeart, 13, false),
			NewCard(CardDesignClover, 5, false), NewCard(CardDesignDiamond, 7, false),
			NewCard(CardDesignSpade, 9, false),
		})
		require.NotNil(t, pg.GetHint(), "セットハンド中はヒントが出ること")
		pg.phase = PaiGowPhaseBet
		assert.Nil(t, pg.GetHint())
	})

	// **推奨はディーラーのハウスウェイと同じ分割でなければならない。**別実装だと
	// ディーラー自身がやらない置き方を人間に勧めることになる。
	t.Run("recommends the same split the dealer's house way would pick", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 14, false), NewCard(CardDesignHeart, 14, false),
			NewCard(CardDesignSpade, 5, false), NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignClover, 9, false), NewCard(CardDesignDiamond, 11, false),
			NewCard(CardDesignSpade, 13, false),
		}
		pg := paiGowSetHandsFixture(cards)
		hint := pg.GetHint()
		require.NotNil(t, hint)

		_, houseLow := paiGowHouseWay(cards)
		require.Len(t, houseLow, 2)
		assert.ElementsMatch(t,
			[]*Card{cards[hint.LowIdx0], cards[hint.LowIdx1]}, houseLow,
			"推奨はハウスウェイと同じ2枚であるべき")
	})

	// **推奨どおりに置いたら必ず通らなければならない。**通らない推奨は害しかない。
	t.Run("the recommended split is always legal", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 14, false), NewCard(CardDesignHeart, 14, false),
			NewCard(CardDesignSpade, 5, false), NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignClover, 9, false), NewCard(CardDesignDiamond, 11, false),
			NewCard(CardDesignSpade, 13, false),
		}
		pg := paiGowSetHandsFixture(cards)
		hint := pg.GetHint()
		require.NotNil(t, hint)
		assert.NoError(t, pg.SetHands(hint.LowIdx0, hint.LowIdx1))
	})

	// **ローにペアを置けるとは限らない。**♠A ♥A + ばらばらの5枚だと、エースの
	// ペアをローに回した瞬間ハイがハイカードになって反則。ハウスウェイは
	// 合法な分割しか候補にしないので、ここはハイカードのローになる。
	t.Run("does not offer a pair that would foul the hand", func(t *testing.T) {
		pg := paiGowSetHandsFixture([]*Card{
			NewCard(CardDesignSpade, 14, false), NewCard(CardDesignHeart, 14, false),
			NewCard(CardDesignSpade, 5, false), NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignClover, 9, false), NewCard(CardDesignDiamond, 11, false),
			NewCard(CardDesignSpade, 13, false),
		})
		hint := pg.GetHint()
		require.NotNil(t, hint)
		assert.False(t, hint.LowIsPair)
		assert.Equal(t, "house_way_high", hint.Reason)
	})

	// ツーペアなら弱い方のペアをローに置いても合法なので、ペアが推奨される。
	t.Run("offers the lower pair when two pairs make it legal", func(t *testing.T) {
		pg := paiGowSetHandsFixture([]*Card{
			NewCard(CardDesignSpade, 14, false), NewCard(CardDesignHeart, 14, false),
			NewCard(CardDesignSpade, 13, false), NewCard(CardDesignHeart, 13, false),
			NewCard(CardDesignClover, 5, false), NewCard(CardDesignDiamond, 7, false),
			NewCard(CardDesignSpade, 9, false),
		})
		hint := pg.GetHint()
		require.NotNil(t, hint)
		assert.True(t, hint.LowIsPair)
		assert.Equal(t, "house_way_pair", hint.Reason)
		assert.NoError(t, pg.SetHands(hint.LowIdx0, hint.LowIdx1))
	})

}

func TestPaiGow_AutoSetHands(t *testing.T) {
	sample := func() []*Card {
		return []*Card{
			NewCard(CardDesignSpade, 14, false), NewCard(CardDesignHeart, 14, false),
			NewCard(CardDesignSpade, 5, false), NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignClover, 9, false), NewCard(CardDesignDiamond, 11, false),
			NewCard(CardDesignSpade, 13, false),
		}
	}

	t.Run("splits the hand and resolves the round", func(t *testing.T) {
		pg := paiGowSetHandsFixture(sample())
		require.NoError(t, pg.AutoSetHands())
		assert.Len(t, pg.GetPlayerLowHand(), 2)
		assert.Len(t, pg.GetPlayerHighHand(), 5)
		assert.Equal(t, PaiGowPhaseEnd, pg.GetPhase())
	})

	// **自動でも反則の抜け道にはならない。**手作業と同じ SetHands を通す。
	t.Run("never produces a fouled split", func(t *testing.T) {
		pg := paiGowSetHandsFixture(sample())
		require.NoError(t, pg.AutoSetHands())
		high := pg.GetPlayerHighHand()
		low := pg.GetPlayerLowHand()
		assert.True(t, paiGowHighBeatsLow(
			evalPaiGowHighHand(high), high, evalPaiGowLowHand(low), low))
	})

	t.Run("applies exactly what GetHint recommended", func(t *testing.T) {
		pg := paiGowSetHandsFixture(sample())
		hint := pg.GetHint()
		require.NotNil(t, hint)
		require.NoError(t, pg.AutoSetHands())
		assert.ElementsMatch(t,
			[]*Card{sample()[hint.LowIdx0], sample()[hint.LowIdx1]},
			pg.GetPlayerLowHand())
	})

	t.Run("rejected outside the set hands phase", func(t *testing.T) {
		pg := NewDefaultPaiGow()
		assert.Error(t, pg.AutoSetHands())
	})
}
