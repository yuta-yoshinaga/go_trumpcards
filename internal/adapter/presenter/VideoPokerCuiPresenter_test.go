package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestFormatSigned(t *testing.T) {
	assert.Equal(t, "+5", formatSigned(5))
	assert.Equal(t, "+0", formatSigned(0))
	assert.Equal(t, "-5", formatSigned(-5))
}

func setupVideoPokerCuiMockDefaults(m *interfaces.MockVideoPokerGame) {
	m.On("RecommendedHold").Return(domain.VideoPokerHoldAdvice{}).Maybe()
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.VideoPokerPhaseBet).Maybe()
	m.On("GetChipsRefilled").Return(false).Maybe()
	m.On("GetCurrentHandKey").Return("").Maybe()
	m.On("GetHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetBetAmount").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetPayout").Return(0).Maybe()
	m.On("GetHands").Return(0).Maybe()
	m.On("GetWins").Return(0).Maybe()
	m.On("GetTotalBet").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetHandRank").Return(0).Maybe()
	m.On("GetHandName").Return("").Maybe()
	m.On("GetHandKey").Return("").Maybe()
	m.On("GetCurrentHandKey").Return("").Maybe()
	m.On("GetHeldIndices").Return([domain.VideoPokerHandSize]bool{}).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetVariantName").Return("videopoker").Maybe()
}

func setupVideoPokerCuiStatsMockDefaults(m *interfaces.MockVideoPokerGame) {
	m.On("GetHands").Return(0).Maybe()
	m.On("GetWins").Return(0).Maybe()
	m.On("GetTotalBet").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
}

func TestVideoPokerCuiPresenter_Output_SessionStats(t *testing.T) {
	p := new(VideoPokerCuiPresenter)
	m := new(interfaces.MockVideoPokerGame)
	m.On("GetChips").Return(1000)
	m.On("GetPhase").Return(domain.VideoPokerPhaseBet)
	m.On("GetHands").Return(4)
	m.On("GetWins").Return(1)
	m.On("GetTotalBet").Return(12)
	m.On("GetTotalPayout").Return(9)
	m.On("GetHand").Return(([]*domain.Card)(nil))
	m.On("GetChipsRefilled").Return(false)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetVariantName").Return("videopoker")

	result := p.Output(m, nil)
	assert.Contains(t, result, "4 ハンド / 勝率 25% / 収支 -3")
	assert.Contains(t, result, "配当表")
}

func TestVideoPokerCuiPresenter_Output_SessionStats_RoundsWinRate(t *testing.T) {
	p := new(VideoPokerCuiPresenter)
	m := new(interfaces.MockVideoPokerGame)
	m.On("GetChips").Return(1000)
	m.On("GetPhase").Return(domain.VideoPokerPhaseBet)
	m.On("GetHands").Return(3)
	m.On("GetWins").Return(2)
	m.On("GetTotalBet").Return(12)
	m.On("GetTotalPayout").Return(9)
	m.On("GetHand").Return(([]*domain.Card)(nil))
	m.On("GetChipsRefilled").Return(false)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetVariantName").Return("videopoker")

	result := p.Output(m, nil)
	assert.Contains(t, result, "3 ハンド / 勝率 67% / 収支 -3")
}

func TestVideoPokerCuiPresenter_Output_SessionStats_ZeroHands(t *testing.T) {
	p := new(VideoPokerCuiPresenter)
	m := new(interfaces.MockVideoPokerGame)
	m.On("GetChips").Return(1000)
	m.On("GetPhase").Return(domain.VideoPokerPhaseBet)
	m.On("GetHands").Return(0)
	m.On("GetWins").Return(0)
	m.On("GetTotalBet").Return(0)
	m.On("GetTotalPayout").Return(0)
	m.On("GetHand").Return(([]*domain.Card)(nil))
	m.On("GetChipsRefilled").Return(false)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetVariantName").Return("videopoker")

	result := p.Output(m, nil)
	assert.Contains(t, result, "0 ハンド / 勝率 0% / 収支 +0")
	assert.NotContains(t, result, "NaN")
}

func TestVideoPokerCuiPresenter_Output_BetPhase(t *testing.T) {
	p := new(VideoPokerCuiPresenter)
	m := new(interfaces.MockVideoPokerGame)
	setupVideoPokerCuiMockDefaults(m)

	result := p.Output(m, nil)
	assert.Contains(t, result, "チップ: 1000")
	assert.Contains(t, result, "フェーズ: ベット")
	// ベットフェーズでは配当表を表示する（デフォルト videopoker バリアント）。
	assert.Contains(t, result, i18n.T("videopoker.payoutTitle"))
	assert.Contains(t, result, "ロイヤルフラッシュ x250")
	assert.Contains(t, result, i18n.T("videopoker.payoutMaxBetNote"))
	assert.Contains(t, result, "ジャックス・オア・ベター x1")
}

func TestVideoPokerCuiPresenter_Output_BetPhase_JokerPokerPaytable(t *testing.T) {
	p := new(VideoPokerCuiPresenter)
	m := new(interfaces.MockVideoPokerGame)
	setupVideoPokerCuiMockDefaults(m)
	// バリアント固有の役（Kings or Better / Five of a Kind / Wild Royal Flush）が出ること。
	m.ExpectedCalls = nil
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetChipsRefilled").Return(false).Maybe()
	m.On("GetPhase").Return(domain.VideoPokerPhaseBet).Maybe()
	m.On("GetCurrentHandKey").Return("").Maybe()
	m.On("GetHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetVariantName").Return("jokerpoker").Maybe()
	setupVideoPokerCuiStatsMockDefaults(m)

	result := p.Output(m, nil)
	assert.Contains(t, result, "キングス・オア・ベター x1")
	assert.Contains(t, result, "ファイブカード x200")
	assert.Contains(t, result, "ワイルドロイヤルフラッシュ x100")
	// jacksorbetter 固有行は出ないこと。
	assert.NotContains(t, result, "ジャックス・オア・ベター")
}

func TestVideoPokerCuiPresenter_Output_BetPhase_Paytable_EnLocale(t *testing.T) {
	i18n.SetLang("en")
	t.Cleanup(func() { i18n.SetLang("ja") })
	p := new(VideoPokerCuiPresenter)
	m := new(interfaces.MockVideoPokerGame)
	setupVideoPokerCuiMockDefaults(m)
	m.ExpectedCalls = nil
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetChipsRefilled").Return(false).Maybe()
	m.On("GetPhase").Return(domain.VideoPokerPhaseBet).Maybe()
	m.On("GetCurrentHandKey").Return("").Maybe()
	m.On("GetHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetVariantName").Return("jokerpoker").Maybe()
	setupVideoPokerCuiStatsMockDefaults(m)

	result := p.Output(m, nil)
	assert.Contains(t, result, "Kings or Better x1")
	assert.Contains(t, result, "Natural Royal Flush x250")
}

func TestVideoPokerCuiPresenter_Output_DrawPhase_WithHand(t *testing.T) {
	p := new(VideoPokerCuiPresenter)
	m := new(interfaces.MockVideoPokerGame)
	m.On("GetChips").Return(997).Maybe()
	m.On("GetPhase").Return(domain.VideoPokerPhaseDraw).Maybe()
	m.On("GetChipsRefilled").Return(false).Maybe()
	m.On("GetCurrentHandKey").Return("").Maybe()
	m.On("GetHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignHeart, 11, false),
		domain.NewCard(domain.CardDesignClover, 5, false),
		domain.NewCard(domain.CardDesignDiamond, 8, false),
		domain.NewCard(domain.CardDesignSpade, 13, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetBetAmount").Return(3).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetPayout").Return(0).Maybe()
	m.On("GetHandRank").Return(0).Maybe()
	m.On("GetHandName").Return("").Maybe()
	m.On("GetHeldIndices").Return([domain.VideoPokerHandSize]bool{true, true, false, false, true}).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	setupVideoPokerCuiStatsMockDefaults(m)

	result := p.Output(m, nil)
	assert.Contains(t, result, "フェーズ: ドロー")
	assert.Contains(t, result, "[ホールド]")
	assert.Contains(t, result, "手札")
}

func TestVideoPokerCuiPresenter_Output_ResultPhase_Win(t *testing.T) {
	p := new(VideoPokerCuiPresenter)
	m := new(interfaces.MockVideoPokerGame)
	m.On("GetChips").Return(1025).Maybe()
	m.On("GetPhase").Return(domain.VideoPokerPhaseResult).Maybe()
	m.On("GetChipsRefilled").Return(false).Maybe()
	m.On("GetCurrentHandKey").Return("").Maybe()
	m.On("GetHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, false),
		domain.NewCard(domain.CardDesignClover, 7, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false),
		domain.NewCard(domain.CardDesignSpade, 3, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetBetAmount").Return(1).Maybe()
	m.On("GetResult").Return(domain.GameResultWin).Maybe()
	m.On("GetPayout").Return(25).Maybe()
	m.On("GetHandRank").Return(domain.PokerHandFourOfAKind).Maybe()
	m.On("GetHandName").Return("Four of a Kind").Maybe()
	m.On("GetHandKey").Return("fourOfAKind").Maybe()
	m.On("GetCurrentHandKey").Return("").Maybe()
	m.On("GetHeldIndices").Return([domain.VideoPokerHandSize]bool{true, true, true, true, false}).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetVariantName").Return("videopoker").Maybe()
	setupVideoPokerCuiStatsMockDefaults(m)

	result := p.Output(m, nil)
	assert.Contains(t, result, "フェーズ: リザルト")
	// **役名は全バリアントで訳す。**以前はここが英語のままで、その挙動を
	// このテストが固定していた (#4693 / #4694)。
	assert.Contains(t, result, "フォーカード! あなたの勝利です！")
	assert.NotContains(t, result, "Four of a Kind!")
	assert.Contains(t, result, "払戻し: 25")
	assert.Contains(t, result, "純増減: +24")
}

func TestVideoPokerCuiPresenter_Output_ResultPhase_Win_DeucesWildTranslated(t *testing.T) {
	winMock := func(variant, handName, handKey string) *interfaces.MockVideoPokerGame {
		m := new(interfaces.MockVideoPokerGame)
		m.On("GetChips").Return(1025).Maybe()
		m.On("GetPhase").Return(domain.VideoPokerPhaseResult).Maybe()
		m.On("GetChipsRefilled").Return(false).Maybe()
		m.On("GetCurrentHandKey").Return("").Maybe()
		m.On("GetHand").Return([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 2, false),
			domain.NewCard(domain.CardDesignSpade, 10, false),
			domain.NewCard(domain.CardDesignSpade, 11, false),
			domain.NewCard(domain.CardDesignSpade, 12, false),
			domain.NewCard(domain.CardDesignSpade, 13, false),
		}).Maybe()
		m.On("GetGameEndFlag").Return(true).Maybe()
		m.On("GetBetAmount").Return(1).Maybe()
		m.On("GetResult").Return(domain.GameResultWin).Maybe()
		m.On("GetPayout").Return(25).Maybe()
		m.On("GetHandRank").Return(domain.PokerHandRoyalFlush).Maybe()
		m.On("GetHandName").Return(handName).Maybe()
		m.On("GetHandKey").Return(handKey).Maybe()
		m.On("GetCurrentHandKey").Return("").Maybe()
		m.On("GetHeldIndices").Return([domain.VideoPokerHandSize]bool{}).Maybe()
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
		m.On("GetVariantName").Return(variant).Maybe()
		setupVideoPokerCuiStatsMockDefaults(m)
		return m
	}
	p := new(VideoPokerCuiPresenter)

	t.Run("ja shows the translated hand name", func(t *testing.T) {
		result := p.Output(winMock("deuceswild", "Wild Royal Flush", "wildRoyalFlush"), nil)
		assert.Contains(t, result, "ワイルドロイヤルフラッシュ! あなたの勝利です！")
		assert.NotContains(t, result, "Wild Royal Flush!")
	})

	t.Run("en shows the translated hand name", func(t *testing.T) {
		i18n.SetLang("en")
		t.Cleanup(func() { i18n.SetLang("ja") })
		result := p.Output(winMock("deuceswild", "Wild Royal Flush", "wildRoyalFlush"), nil)
		assert.Contains(t, result, "Wild Royal Flush! You win!")
	})

	t.Run("empty key falls back to the raw hand name", func(t *testing.T) {
		result := p.Output(winMock("deuceswild", "Wild Royal Flush", ""), nil)
		assert.Contains(t, result, "Wild Royal Flush! あなたの勝利です！")
	})

	// **Joker Poker も訳す。**deuceswild だけ分岐していたので、日本語ロケール
	// でも勝敗行だけ英語で出ていた (#4693)。
	t.Run("joker poker is translated too", func(t *testing.T) {
		result := p.Output(winMock("jokerpoker", "Five of a Kind", "fiveOfAKind"), nil)
		assert.Contains(t, result, "ファイブカード! あなたの勝利です！")
		assert.NotContains(t, result, "Five of a Kind!")
	})

	// 訳の無いキーは英語名に落とす。キー文字列を画面に出さない。
	t.Run("an untranslated key falls back rather than printing the key", func(t *testing.T) {
		result := p.Output(winMock("jokerpoker", "Mystery Hand", "mysteryHand"), nil)
		assert.Contains(t, result, "Mystery Hand! あなたの勝利です！")
		assert.NotContains(t, result, "pokerhand.")
	})
}

func TestVideoPokerCuiPresenter_Output_ResultPhase_Lose(t *testing.T) {
	p := new(VideoPokerCuiPresenter)
	m := new(interfaces.MockVideoPokerGame)
	m.On("GetChips").Return(999).Maybe()
	m.On("GetPhase").Return(domain.VideoPokerPhaseResult).Maybe()
	m.On("GetChipsRefilled").Return(false).Maybe()
	m.On("GetCurrentHandKey").Return("").Maybe()
	m.On("GetHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 2, false),
		domain.NewCard(domain.CardDesignClover, 5, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignDiamond, 9, false),
		domain.NewCard(domain.CardDesignSpade, 11, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetBetAmount").Return(1).Maybe()
	m.On("GetResult").Return(domain.GameResultLose).Maybe()
	m.On("GetPayout").Return(0).Maybe()
	m.On("GetHandRank").Return(domain.PokerHandHighCard).Maybe()
	m.On("GetHandName").Return("").Maybe()
	m.On("GetHeldIndices").Return([domain.VideoPokerHandSize]bool{}).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetVariantName").Return("videopoker").Maybe()
	setupVideoPokerCuiStatsMockDefaults(m)

	result := p.Output(m, nil)
	assert.Contains(t, result, "役なし。")
	assert.Contains(t, result, "払戻し: 0")
	assert.Contains(t, result, "純増減: -1")
}

func TestVideoPokerCuiPresenter_Output_Error(t *testing.T) {
	p := new(VideoPokerCuiPresenter)
	m := new(interfaces.MockVideoPokerGame)
	setupVideoPokerCuiMockDefaults(m)

	result := p.Output(m, domain.NewDomainError(domain.ErrInvalidAmount, "Invalid bet amount."))
	assert.Contains(t, result, "Invalid bet amount.")
}

func TestVideoPokerCuiPresenter_phaseStr(t *testing.T) {
	p := new(VideoPokerCuiPresenter)
	assert.Equal(t, "ベット", p.phaseStr(domain.VideoPokerPhaseBet))
	assert.Equal(t, "ドロー", p.phaseStr(domain.VideoPokerPhaseDraw))
	assert.Equal(t, "リザルト", p.phaseStr(domain.VideoPokerPhaseResult))
	assert.Equal(t, "不明", p.phaseStr(99))
}

func TestVideoPokerCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(VideoPokerCuiPresenter)
	m := new(interfaces.MockVideoPokerGame)
	m.On("GetGameEndFlag").Return(false)
	result := p.ActionLogOutput(m)
	assert.NotEmpty(t, result)
}

func TestVideoPokerCuiPresenter_Output_JokerHighlighted(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(false)
	defer color.SetNoColor(origNoColor)

	p := new(VideoPokerCuiPresenter)
	m := new(interfaces.MockVideoPokerGame)
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.VideoPokerPhaseDraw).Maybe()
	m.On("GetChipsRefilled").Return(false).Maybe()
	m.On("GetCurrentHandKey").Return("").Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetHeldIndices").Return([domain.VideoPokerHandSize]bool{}).Maybe()
	m.On("GetVariantName").Return("jokerpoker").Maybe()
	m.On("GetHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignJoker, 0, false),
		domain.NewCard(domain.CardDesignSpade, 5, false),
	}).Maybe()
	setupVideoPokerCuiStatsMockDefaults(m)

	result := p.Output(m, nil)
	assert.Contains(t, result, color.BoldYellow("JOKER"))
	assert.Contains(t, result, "♠5")
}

func TestVideoPokerCuiPresenter_Output_DeucesWildTwosHighlighted(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(false)
	defer color.SetNoColor(origNoColor)

	p := new(VideoPokerCuiPresenter)
	m := new(interfaces.MockVideoPokerGame)
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.VideoPokerPhaseDraw).Maybe()
	m.On("GetChipsRefilled").Return(false).Maybe()
	m.On("GetCurrentHandKey").Return("").Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetHeldIndices").Return([domain.VideoPokerHandSize]bool{}).Maybe()
	m.On("GetVariantName").Return("deuceswild").Maybe()
	m.On("GetHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 2, false),
		domain.NewCard(domain.CardDesignSpade, 2, false),
		domain.NewCard(domain.CardDesignSpade, 5, false),
	}).Maybe()
	setupVideoPokerCuiStatsMockDefaults(m)

	result := p.Output(m, nil)
	assert.Contains(t, result, color.Yellow("HEART 2"))
	assert.Contains(t, result, color.Yellow("SPADE 2"))
	assert.NotContains(t, result, color.Yellow("♠5"))
}

func TestVideoPokerCuiPresenter_Output_PlainVariantTwoNotHighlighted(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(false)
	defer color.SetNoColor(origNoColor)

	p := new(VideoPokerCuiPresenter)
	m := new(interfaces.MockVideoPokerGame)
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.VideoPokerPhaseDraw).Maybe()
	m.On("GetChipsRefilled").Return(false).Maybe()
	m.On("GetCurrentHandKey").Return("").Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetHeldIndices").Return([domain.VideoPokerHandSize]bool{}).Maybe()
	m.On("GetVariantName").Return("videopoker").Maybe()
	m.On("GetHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 2, false),
	}).Maybe()
	setupVideoPokerCuiStatsMockDefaults(m)

	result := p.Output(m, nil)
	assert.Contains(t, result, "♠2")
	assert.NotContains(t, result, color.Yellow("♠2"))
}

func TestVideoPokerCuiPresenter_cardStr_NilCard(t *testing.T) {
	p := new(VideoPokerCuiPresenter)
	m := new(interfaces.MockVideoPokerGame)
	assert.Equal(t, "??", p.cardStr(m, nil))
}

func TestVideoPokerCuiPresenter_HintOutput_UsesDomainRecommendation(t *testing.T) {
	i18n.SetLang("ja")
	p := new(VideoPokerCuiPresenter)
	m := new(interfaces.MockVideoPokerGame)
	hand := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 10, false),
		domain.NewCard(domain.CardDesignSpade, 11, false),
		domain.NewCard(domain.CardDesignSpade, 12, false),
		domain.NewCard(domain.CardDesignSpade, 13, false),
		domain.NewCard(domain.CardDesignHeart, 2, false),
	}
	m.On("GetPhase").Return(domain.VideoPokerPhaseDraw)
	m.On("GetHand").Return(hand)
	m.On("RecommendedHold").Return(domain.VideoPokerHoldAdvice{Hold: [domain.VideoPokerHandSize]bool{true, true, true, true}, RuleKey: "royalDraw4"})
	result := p.HintOutput(m)
	assert.Contains(t, result, "[0]")
	assert.Contains(t, result, "[3]")
	assert.Contains(t, result, "ロイヤルフラッシュへの4枚をホールド")
	assert.NotContains(t, result, "videopoker.strategy.royalDraw4")
}

func TestVideoPokerCuiPresenter_HintOutput_DrawAll(t *testing.T) {
	p := new(VideoPokerCuiPresenter)
	m := new(interfaces.MockVideoPokerGame)
	m.On("GetPhase").Return(domain.VideoPokerPhaseDraw)
	m.On("GetHand").Return([]*domain.Card{})
	m.On("RecommendedHold").Return(domain.VideoPokerHoldAdvice{RuleKey: "drawAll"})
	assert.Contains(t, p.HintOutput(m), i18n.T("videopoker.hintHoldNone"))
}

func TestVideoPokerStrategyRuleKeysHaveJaAndEnTranslations(t *testing.T) {
	for _, key := range domain.VideoPokerStrategyRuleKeys() {
		fullKey := "videopoker.strategy." + key
		for _, lang := range []string{"ja", "en"} {
			assert.NotEqual(t, fullKey, i18n.TForLang(lang, fullKey), "%s missing %s translation", lang, fullKey)
		}
	}
}
