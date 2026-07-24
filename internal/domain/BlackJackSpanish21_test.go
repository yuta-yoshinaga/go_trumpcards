//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewSpanish21BlackJack はファクトリがバリアント設定とSpanishデッキを正しく初期化することを検証する
func TestNewSpanish21BlackJack(t *testing.T) {
	bj := NewSpanish21BlackJack()
	require.NotNil(t, bj)

	v := bj.GetVariant()
	require.NotNil(t, v)
	assert.Equal(t, BJVariantSpanish21, v.Name)
	assert.Equal(t, BJVariantSpanish21, bj.GetConfig().Variant)
	// 初期デッキも48枚 (10を除く)
	assert.Equal(t, SpanishDeckCardCount, bj.trumpCards.GetTotalCount())
	assert.Equal(t, BJDefaultChips, bj.GetPlayer().GetChips())
}

// TestSpanish21ResetUsesVariantDeck はResetでバリアントのDeckBuilderが使われることを検証する
func TestSpanish21ResetUsesVariantDeck(t *testing.T) {
	bj := NewSpanish21BlackJack()
	bj.deckCountChanged = true // 強制リシャッフル
	bj.Reset()
	assert.Equal(t, SpanishDeckCardCount, bj.trumpCards.GetTotalCount())
	// デッキに10が含まれていないことを抜き取りで確認
	for i := 0; i < SpanishDeckCardCount; i++ {
		c := bj.trumpCards.DrawCard()
		require.NotNil(t, c)
		assert.NotEqual(t, 10, c.GetValue())
	}
}

// TestSpanish21Player21AlwaysWins はプレイヤー21がディーラー21にも勝つことを検証する
func TestSpanish21Player21AlwaysWins(t *testing.T) {
	bj := NewSpanish21BlackJack()
	// Player: 7-7-7 (21, 3-card)
	hand := NewBlackJackHand()
	hand.AddCard(NewCard(CardDesignClover, 7, false))
	hand.AddCard(NewCard(CardDesignHeart, 7, false))
	hand.AddCard(NewCard(CardDesignDiamond, 7, false))
	bj.SetPlayerHands([]*BlackJackHand{hand})

	// Dealer: 10は存在しないので J-A=21 (BJではない: J=10扱いだが「ナチュラル」かどうかは2枚で21のとき真)
	bj.GetDealer().AddCard(NewCard(CardDesignSpade, 13, false)) // K=10
	bj.GetDealer().AddCard(NewCard(CardDesignHeart, 1, false))  // A=11

	// 両者21だが、Spanish21ではプレイヤーが勝利
	assert.Equal(t, GameResultWin, bj.judgeHand(hand))
}

// TestSpanish21PlayerBJBeatsDealerBJ は両者ナチュラルBJでプレイヤー勝利を検証する
func TestSpanish21PlayerBJBeatsDealerBJ(t *testing.T) {
	bj := NewSpanish21BlackJack()
	// Player: A-K (BJ)
	hand := NewBlackJackHand()
	hand.AddCard(NewCard(CardDesignClover, 1, false))
	hand.AddCard(NewCard(CardDesignDiamond, 13, false))
	bj.SetPlayerHands([]*BlackJackHand{hand})

	// Dealer: A-K (BJ)
	bj.GetDealer().AddCard(NewCard(CardDesignSpade, 1, false))
	bj.GetDealer().AddCard(NewCard(CardDesignHeart, 13, false))

	// Spanish21: プレイヤーBJはディーラーBJに勝つ
	assert.Equal(t, GameResultWin, bj.judgeHand(hand))
}

// TestStandardBlackJackUnchanged は標準BJで挙動が変わっていないことを検証する
func TestStandardBlackJackUnchanged(t *testing.T) {
	bj := NewDefaultBlackJack()
	// Player: 7-7-7 (21)
	hand := NewBlackJackHand()
	hand.AddCard(NewCard(CardDesignClover, 7, false))
	hand.AddCard(NewCard(CardDesignHeart, 7, false))
	hand.AddCard(NewCard(CardDesignDiamond, 7, false))
	bj.SetPlayerHands([]*BlackJackHand{hand})
	// Dealer: K-A (=21)
	bj.GetDealer().AddCard(NewCard(CardDesignSpade, 13, false))
	bj.GetDealer().AddCard(NewCard(CardDesignHeart, 1, false))

	// 標準BJでは両者21は引き分け (ディーラーは2枚=BJ、プレイヤー3枚)
	assert.Equal(t, GameResultLose, bj.judgeHand(hand))
}

// TestSpanish21BonusPayout5Card21 は5枚21ボーナスでチップが3:2配当されることを検証する
func TestSpanish21BonusPayout5Card21(t *testing.T) {
	bj := NewSpanish21BlackJack()
	bj.GetPlayer().SetChips(900)

	// 5-card 21: 2+3+4+5+7
	hand := NewBlackJackHand()
	hand.AddCard(NewCard(CardDesignSpade, 2, false))
	hand.AddCard(NewCard(CardDesignClover, 3, false))
	hand.AddCard(NewCard(CardDesignHeart, 4, false))
	hand.AddCard(NewCard(CardDesignDiamond, 5, false))
	hand.AddCard(NewCard(CardDesignSpade, 7, false))
	hand.SetBet(100)

	bonus := bj.payoutHandWithVariant(bj.GetPlayer(), hand, false, GameResultWin)
	require.NotNil(t, bonus)
	assert.Equal(t, "spanish21.bonus.fivecard21", bonus.NameKey)
	// 100 * 3/2 = 150 利益、+ベット返却100 = 250
	assert.Equal(t, 900+250, bj.GetPlayer().GetChips())
}

// TestSpanish21BonusPayoutSuited777 はスペード7-7-7のスーパーボーナス3:1を検証する
func TestSpanish21BonusPayoutSuited777(t *testing.T) {
	bj := NewSpanish21BlackJack()
	bj.GetPlayer().SetChips(900)

	hand := NewBlackJackHand()
	hand.AddCard(NewCard(CardDesignSpade, 7, false))
	hand.AddCard(NewCard(CardDesignSpade, 7, false))
	hand.AddCard(NewCard(CardDesignSpade, 7, false))
	hand.SetBet(100)

	bonus := bj.payoutHandWithVariant(bj.GetPlayer(), hand, false, GameResultWin)
	require.NotNil(t, bonus)
	assert.Equal(t, "spanish21.bonus.777.spade", bonus.NameKey)
	// 100 * 3/1 = 300 利益、+ベット返却100 = 400
	assert.Equal(t, 900+400, bj.GetPlayer().GetChips())
}

// TestSpanish21BonusNotPaidAfterDouble はダブルダウン後はボーナス対象外を検証する
func TestSpanish21BonusNotPaidAfterDouble(t *testing.T) {
	bj := NewSpanish21BlackJack()
	bj.GetPlayer().SetChips(800)

	// 5-card 21 だがダブルダウン済み
	hand := NewBlackJackHand()
	hand.AddCard(NewCard(CardDesignSpade, 2, false))
	hand.AddCard(NewCard(CardDesignClover, 3, false))
	hand.AddCard(NewCard(CardDesignHeart, 4, false))
	hand.AddCard(NewCard(CardDesignDiamond, 5, false))
	hand.AddCard(NewCard(CardDesignSpade, 7, false))
	hand.SetBet(200) // ダブル後はベット倍
	hand.SetDoubled(true)

	bonus := bj.payoutHandWithVariant(bj.GetPlayer(), hand, false, GameResultWin)
	assert.Nil(t, bonus, "doubled hand must not receive bonus")
	// 通常配当 200*2=400
	assert.Equal(t, 800+400, bj.GetPlayer().GetChips())
}

// TestSpanish21BonusNotPaidOnNaturalBJ はナチュラルBJはBJ配当が優先されることを検証する
func TestSpanish21BonusNotPaidOnNaturalBJ(t *testing.T) {
	bj := NewSpanish21BlackJack()
	bj.GetPlayer().SetChips(900)

	// A-K = ナチュラルBJ
	hand := NewBlackJackHand()
	hand.AddCard(NewCard(CardDesignSpade, 1, false))
	hand.AddCard(NewCard(CardDesignClover, 13, false))
	hand.SetBet(100)

	bonus := bj.payoutHandWithVariant(bj.GetPlayer(), hand, false, GameResultWin)
	assert.Nil(t, bonus, "natural BJ uses 3:2 BJ payout, not bonus")
	// BJ配当 100 + 100*3/2 = 250
	assert.Equal(t, 900+250, bj.GetPlayer().GetChips())
}

// TestSpanish21BonusFallsThroughOnNonQualifyingWin は Spanish 21 で勝利したが
// ボーナス対象外 (3-card 21 で 6-7-8 や 7-7-7 ではない) のケースで、
// 通常の 1:1 配当にフォールバックすることを検証する
func TestSpanish21BonusFallsThroughOnNonQualifyingWin(t *testing.T) {
	bj := NewSpanish21BlackJack()
	bj.GetPlayer().SetChips(900)

	// 3-card 21: 7-5-9 (ボーナス対象外)
	hand := NewBlackJackHand()
	hand.AddCard(NewCard(CardDesignSpade, 7, false))
	hand.AddCard(NewCard(CardDesignClover, 5, false))
	hand.AddCard(NewCard(CardDesignHeart, 9, false))
	hand.SetBet(100)

	bonus := bj.payoutHandWithVariant(bj.GetPlayer(), hand, false, GameResultWin)
	assert.Nil(t, bonus, "non-qualifying 21 must not receive bonus")
	// 通常の 2x 配当 = 200
	assert.Equal(t, 900+200, bj.GetPlayer().GetChips())
}

// TestStandardBlackJackNoBonusPayout は標準BJでボーナスが発動しないことを検証する
func TestStandardBlackJackNoBonusPayout(t *testing.T) {
	bj := NewDefaultBlackJack()
	bj.GetPlayer().SetChips(900)

	// 5-card 21 in standard BJ (no bonus)
	hand := NewBlackJackHand()
	hand.AddCard(NewCard(CardDesignSpade, 2, false))
	hand.AddCard(NewCard(CardDesignClover, 3, false))
	hand.AddCard(NewCard(CardDesignHeart, 4, false))
	hand.AddCard(NewCard(CardDesignDiamond, 5, false))
	hand.AddCard(NewCard(CardDesignSpade, 7, false))
	hand.SetBet(100)

	bonus := bj.payoutHandWithVariant(bj.GetPlayer(), hand, false, GameResultWin)
	assert.Nil(t, bonus)
	// 通常勝利配当 2x = 200
	assert.Equal(t, 900+200, bj.GetPlayer().GetChips())
}

// TestSpanish21JSONRoundTrip はバリアント情報がシリアライズ/デシリアライズで保持されることを検証する
func TestSpanish21JSONRoundTrip(t *testing.T) {
	bj := NewSpanish21BlackJack()
	data, err := json.Marshal(bj)
	require.NoError(t, err)

	var restored BlackJack
	require.NoError(t, json.Unmarshal(data, &restored))
	v := restored.GetVariant()
	require.NotNil(t, v, "variant must be re-resolved from config.Variant on Unmarshal")
	assert.Equal(t, BJVariantSpanish21, v.Name)
	assert.NotNil(t, v.DeckBuilder)
	assert.NotNil(t, v.BonusEval)
}

// TestSpanish21SetConfigSwitchesVariant はSetConfigでバリアント切り替えが反映されることを検証する
func TestSpanish21SetConfigSwitchesVariant(t *testing.T) {
	bj := NewDefaultBlackJack()
	cfg := bj.GetConfig()
	cfg.Variant = BJVariantSpanish21
	require.NoError(t, bj.SetConfig(cfg))
	v := bj.GetVariant()
	require.NotNil(t, v)
	assert.Equal(t, BJVariantSpanish21, v.Name)
	assert.True(t, bj.deckCountChanged, "variant change should trigger deck rebuild")
}

// TestSpanish21BonusKeysCapturedOnResolve は resolvePayouts がボーナス成立時に
// GetBonusKeys へキーを記録すること、およびボーナスなしの再精算でクリアされることを検証する。
func TestSpanish21BonusKeysCapturedOnResolve(t *testing.T) {
	bj := NewSpanish21BlackJack()

	// 5-card 21 (2+3+4+5+7) で勝利 → fivecard21 ボーナス成立
	hand := NewBlackJackHand()
	hand.AddCard(NewCard(CardDesignSpade, 2, false))
	hand.AddCard(NewCard(CardDesignClover, 3, false))
	hand.AddCard(NewCard(CardDesignHeart, 4, false))
	hand.AddCard(NewCard(CardDesignDiamond, 5, false))
	hand.AddCard(NewCard(CardDesignSpade, 7, false))
	hand.SetBet(100)
	bj.playerHands = []*BlackJackHand{hand}
	// ディーラーは 17 (10+7): プレイヤー21 が勝つが BJ ではない
	bj.dealer.AddCard(NewCard(CardDesignClover, 10, false))
	bj.dealer.AddCard(NewCard(CardDesignDiamond, 7, false))

	bj.resolvePayouts()
	assert.Equal(t, []string{"spanish21.bonus.fivecard21"}, bj.GetBonusKeys())

	// ボーナスなしの通常ハンドで再精算 → クリアされる
	plain := NewBlackJackHand()
	plain.AddCard(NewCard(CardDesignSpade, 10, false))
	plain.AddCard(NewCard(CardDesignClover, 9, false)) // 19, 勝ち, ボーナスなし
	plain.SetBet(100)
	bj.playerHands = []*BlackJackHand{plain}
	bj.resolvePayouts()
	assert.Empty(t, bj.GetBonusKeys())
}

// TestBlackJack_SetBonusKeys は SetBonusKeys が成立ボーナスキーを設定し
// GetBonusKeys で取得できることを検証する（テスト用セッター）。
func TestBlackJack_SetBonusKeys(t *testing.T) {
	bj := NewSpanish21BlackJack()
	assert.Empty(t, bj.GetBonusKeys())

	keys := []string{"spanish21.bonus.fivecard21", "spanish21.bonus.678.spade"}
	bj.SetBonusKeys(keys)
	assert.Equal(t, keys, bj.GetBonusKeys())
}
