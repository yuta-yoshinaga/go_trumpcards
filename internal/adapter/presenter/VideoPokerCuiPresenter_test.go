package presenter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupVideoPokerCuiMockDefaults(m *interfaces.MockVideoPokerGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.VideoPokerPhaseBet).Maybe()
	m.On("GetHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetBetAmount").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetPayout").Return(0).Maybe()
	m.On("GetHandRank").Return(0).Maybe()
	m.On("GetHandName").Return("").Maybe()
	m.On("GetHandKey").Return("").Maybe()
	m.On("GetHeldIndices").Return([domain.VideoPokerHandSize]bool{}).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetVariantName").Return("videopoker").Maybe()
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
	m.On("GetPhase").Return(domain.VideoPokerPhaseBet).Maybe()
	m.On("GetHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetVariantName").Return("jokerpoker").Maybe()

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
	m.On("GetPhase").Return(domain.VideoPokerPhaseBet).Maybe()
	m.On("GetHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetVariantName").Return("jokerpoker").Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "Kings or Better x1")
	assert.Contains(t, result, "Natural Royal Flush x250")
}

func TestVideoPokerCuiPresenter_Output_DrawPhase_WithHand(t *testing.T) {
	p := new(VideoPokerCuiPresenter)
	m := new(interfaces.MockVideoPokerGame)
	m.On("GetChips").Return(997).Maybe()
	m.On("GetPhase").Return(domain.VideoPokerPhaseDraw).Maybe()
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
	m.On("GetHeldIndices").Return([domain.VideoPokerHandSize]bool{true, true, true, true, false}).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetVariantName").Return("videopoker").Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "フェーズ: リザルト")
	// **役名は全バリアントで訳す。**以前はここが英語のままで、その挙動を
	// このテストが固定していた (#4693 / #4694)。
	assert.Contains(t, result, "フォーカード! あなたの勝利です！")
	assert.NotContains(t, result, "Four of a Kind!")
	assert.Contains(t, result, "払戻し: 25")
}

func TestVideoPokerCuiPresenter_Output_ResultPhase_Win_DeucesWildTranslated(t *testing.T) {
	winMock := func(variant, handName, handKey string) *interfaces.MockVideoPokerGame {
		m := new(interfaces.MockVideoPokerGame)
		m.On("GetChips").Return(1025).Maybe()
		m.On("GetPhase").Return(domain.VideoPokerPhaseResult).Maybe()
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
		m.On("GetHeldIndices").Return([domain.VideoPokerHandSize]bool{}).Maybe()
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
		m.On("GetVariantName").Return(variant).Maybe()
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

	result := p.Output(m, nil)
	assert.Contains(t, result, "役なし。")
	assert.Contains(t, result, "払戻し: 0")
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
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetHeldIndices").Return([domain.VideoPokerHandSize]bool{}).Maybe()
	m.On("GetVariantName").Return("jokerpoker").Maybe()
	m.On("GetHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignJoker, 0, false),
		domain.NewCard(domain.CardDesignSpade, 5, false),
	}).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, color.BoldYellow("JOKER"))
	assert.Contains(t, result, "SPADE 5")
}

func TestVideoPokerCuiPresenter_Output_DeucesWildTwosHighlighted(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(false)
	defer color.SetNoColor(origNoColor)

	p := new(VideoPokerCuiPresenter)
	m := new(interfaces.MockVideoPokerGame)
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.VideoPokerPhaseDraw).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetHeldIndices").Return([domain.VideoPokerHandSize]bool{}).Maybe()
	m.On("GetVariantName").Return("deuceswild").Maybe()
	m.On("GetHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 2, false),
		domain.NewCard(domain.CardDesignSpade, 2, false),
		domain.NewCard(domain.CardDesignSpade, 5, false),
	}).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, color.Yellow("HEART 2"))
	assert.Contains(t, result, color.Yellow("SPADE 2"))
	assert.NotContains(t, result, color.Yellow("SPADE 5"))
}

func TestVideoPokerCuiPresenter_Output_PlainVariantTwoNotHighlighted(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(false)
	defer color.SetNoColor(origNoColor)

	p := new(VideoPokerCuiPresenter)
	m := new(interfaces.MockVideoPokerGame)
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.VideoPokerPhaseDraw).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetHeldIndices").Return([domain.VideoPokerHandSize]bool{}).Maybe()
	m.On("GetVariantName").Return("videopoker").Maybe()
	m.On("GetHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 2, false),
	}).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "SPADE 2")
	assert.NotContains(t, result, color.Yellow("SPADE 2"))
}

func TestVideoPokerCuiPresenter_cardStr_NilCard(t *testing.T) {
	p := new(VideoPokerCuiPresenter)
	m := new(interfaces.MockVideoPokerGame)
	assert.Equal(t, "??", p.cardStr(m, nil))
}

func TestVideoPokerCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(VideoPokerCuiPresenter)

	drawGame := func(variant string, hand []*domain.Card) *interfaces.MockVideoPokerGame {
		m := new(interfaces.MockVideoPokerGame)
		m.On("GetPhase").Return(domain.VideoPokerPhaseDraw)
		m.On("GetVariantName").Return(variant)
		m.On("GetHand").Return(hand)
		return m
	}
	holdPrefix := strings.SplitN(i18n.T("videopoker.hintHold"), "{{", 2)[0]

	t.Run("deuces wild holds deuces and a made pair", func(t *testing.T) {
		hand := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 2, false),  // deuce (wild)
			domain.NewCard(domain.CardDesignHeart, 8, false),  // pair of 8s
			domain.NewCard(domain.CardDesignClover, 8, false), // pair of 8s
			domain.NewCard(domain.CardDesignDiamond, 5, false),
			domain.NewCard(domain.CardDesignHeart, 9, false),
		}
		out := p.HintOutput(drawGame("deuceswild", hand))
		assert.Contains(t, out, holdPrefix)
		assert.Contains(t, out, i18n.T("videopoker.holdWildAndPair"))
		assert.Contains(t, out, "[0]") // deuce
		assert.Contains(t, out, "[1]") // pair
		assert.Contains(t, out, "[2]") // pair
		assert.NotContains(t, out, "[3]")
	})

	t.Run("deuces wild holds the lone deuce", func(t *testing.T) {
		hand := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 2, false), // deuce (wild)
			domain.NewCard(domain.CardDesignHeart, 4, false),
			domain.NewCard(domain.CardDesignClover, 7, false),
			domain.NewCard(domain.CardDesignDiamond, 9, false),
			domain.NewCard(domain.CardDesignHeart, 5, false),
		}
		out := p.HintOutput(drawGame("deuceswild", hand))
		assert.Contains(t, out, i18n.T("videopoker.holdWild"))
		assert.Contains(t, out, "[0]")
		assert.NotContains(t, out, "[1]")
	})

	t.Run("joker poker holds the joker", func(t *testing.T) {
		hand := []*domain.Card{
			domain.NewCard(domain.CardDesignJoker, 0, false), // joker (wild)
			domain.NewCard(domain.CardDesignHeart, 4, false),
			domain.NewCard(domain.CardDesignClover, 7, false),
			domain.NewCard(domain.CardDesignDiamond, 9, false),
			domain.NewCard(domain.CardDesignHeart, 5, false),
		}
		out := p.HintOutput(drawGame("jokerpoker", hand))
		assert.Contains(t, out, i18n.T("videopoker.holdWild"))
		assert.Contains(t, out, "[0]")
	})

	t.Run("jacks or better holds a pair", func(t *testing.T) {
		hand := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 8, false),
			domain.NewCard(domain.CardDesignHeart, 8, false),
			domain.NewCard(domain.CardDesignClover, 3, false),
			domain.NewCard(domain.CardDesignDiamond, 5, false),
			domain.NewCard(domain.CardDesignHeart, 9, false),
		}
		out := p.HintOutput(drawGame("jacksorbetter", hand))
		assert.Contains(t, out, i18n.T("videopoker.holdPair"))
		assert.Contains(t, out, "[0]")
		assert.Contains(t, out, "[1]")
		assert.NotContains(t, out, "[2]")
	})

	// **配当のつかない低いペアが、強いドローを潰していた (#4691)。**ペア判定が
	// 最初に無条件でヒットするため、4枚ロイヤル・4枚フラッシュが同居しても
	// 常に弱いペアを勧めていた。順序は Jacks or Better の標準戦略に合わせる:
	//   4枚ロイヤル > 4枚フラッシュ > 低いペア > 4枚ストレート
	t.Run("jacks or better prefers a royal draw over a non-paying pair", func(t *testing.T) {
		hand := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 10, false),
			domain.NewCard(domain.CardDesignSpade, 11, false),
			domain.NewCard(domain.CardDesignSpade, 12, false),
			domain.NewCard(domain.CardDesignSpade, 13, false),
			domain.NewCard(domain.CardDesignHeart, 10, false),
		}
		out := p.HintOutput(drawGame("jacksorbetter", hand))
		assert.Contains(t, out, i18n.T("videopoker.holdRoyalDraw"))
	})

	t.Run("jacks or better prefers a flush draw over a non-paying pair", func(t *testing.T) {
		hand := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 3, false),
			domain.NewCard(domain.CardDesignSpade, 5, false),
			domain.NewCard(domain.CardDesignSpade, 8, false),
			domain.NewCard(domain.CardDesignSpade, 13, false),
			domain.NewCard(domain.CardDesignHeart, 3, false),
		}
		out := p.HintOutput(drawGame("jacksorbetter", hand))
		assert.Contains(t, out, i18n.T("videopoker.holdFlushDraw"))
	})

	// **逆側。**低ペアは4枚ストレートより上。ここを一緒くたに「ドロー優先」と
	// すると、標準戦略から外れる方向に壊れる。
	t.Run("jacks or better keeps a low pair over a straight draw", func(t *testing.T) {
		hand := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 4, false),
			domain.NewCard(domain.CardDesignSpade, 5, false),
			domain.NewCard(domain.CardDesignHeart, 6, false),
			domain.NewCard(domain.CardDesignClover, 7, false),
			domain.NewCard(domain.CardDesignHeart, 4, false),
		}
		out := p.HintOutput(drawGame("jacksorbetter", hand))
		assert.Contains(t, out, i18n.T("videopoker.holdPair"))
	})

	// 配当のつくペア (J 以上) は据え置き。4枚フラッシュより上。
	t.Run("jacks or better keeps a paying high pair over a flush draw", func(t *testing.T) {
		hand := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 12, false),
			domain.NewCard(domain.CardDesignSpade, 5, false),
			domain.NewCard(domain.CardDesignSpade, 8, false),
			domain.NewCard(domain.CardDesignSpade, 3, false),
			domain.NewCard(domain.CardDesignHeart, 12, false),
		}
		out := p.HintOutput(drawGame("jacksorbetter", hand))
		assert.Contains(t, out, i18n.T("videopoker.holdPair"))
	})

	t.Run("jacks or better holds a flush draw", func(t *testing.T) {
		hand := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 3, false),
			domain.NewCard(domain.CardDesignSpade, 6, false),
			domain.NewCard(domain.CardDesignSpade, 8, false),
			domain.NewCard(domain.CardDesignSpade, 10, false),
			domain.NewCard(domain.CardDesignHeart, 4, false),
		}
		out := p.HintOutput(drawGame("jacksorbetter", hand))
		assert.Contains(t, out, i18n.T("videopoker.holdFlushDraw"))
		assert.Contains(t, out, "[0]")
		assert.Contains(t, out, "[3]")
		assert.NotContains(t, out, "[4]")
	})

	t.Run("jacks or better holds a straight draw", func(t *testing.T) {
		hand := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 5, false),
			domain.NewCard(domain.CardDesignHeart, 6, false),
			domain.NewCard(domain.CardDesignClover, 7, false),
			domain.NewCard(domain.CardDesignDiamond, 8, false),
			domain.NewCard(domain.CardDesignHeart, 10, false),
		}
		out := p.HintOutput(drawGame("jacksorbetter", hand))
		assert.Contains(t, out, i18n.T("videopoker.holdStraightDraw"))
		assert.Contains(t, out, "[0]")
		assert.Contains(t, out, "[3]")
		assert.NotContains(t, out, "[4]")
	})

	t.Run("jacks or better holds the high cards", func(t *testing.T) {
		hand := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 3, false),
			domain.NewCard(domain.CardDesignHeart, 7, false),
			domain.NewCard(domain.CardDesignClover, 9, false),
			domain.NewCard(domain.CardDesignDiamond, 11, false), // jack
			domain.NewCard(domain.CardDesignSpade, 13, false),   // king
		}
		out := p.HintOutput(drawGame("jacksorbetter", hand))
		assert.Contains(t, out, i18n.T("videopoker.holdHighCards"))
		assert.Contains(t, out, "[3]")
		assert.Contains(t, out, "[4]")
		assert.NotContains(t, out, "[0]")
	})

	t.Run("recommends redraw when nothing is worth keeping", func(t *testing.T) {
		hand := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 3, false),
			domain.NewCard(domain.CardDesignHeart, 4, false),
			domain.NewCard(domain.CardDesignClover, 6, false),
			domain.NewCard(domain.CardDesignDiamond, 8, false),
			domain.NewCard(domain.CardDesignHeart, 10, false),
		}
		assert.Contains(t, p.HintOutput(drawGame("jacksorbetter", hand)), i18n.T("videopoker.hintHoldNone"))
	})

	t.Run("no hint with an empty hand", func(t *testing.T) {
		assert.Contains(t, p.HintOutput(drawGame("jacksorbetter", nil)), i18n.T("videopoker.hintNone"))
	})

	t.Run("no hint outside the draw phase", func(t *testing.T) {
		m := new(interfaces.MockVideoPokerGame)
		m.On("GetPhase").Return(domain.VideoPokerPhaseBet)
		assert.Contains(t, p.HintOutput(m), i18n.T("videopoker.hintNone"))
	})
}
