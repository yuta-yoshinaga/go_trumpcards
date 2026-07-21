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
	m.On("GetHeldIndices").Return([domain.VideoPokerHandSize]bool{true, true, true, true, false}).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "フェーズ: リザルト")
	assert.Contains(t, result, "Four of a Kind! あなたの勝利です！")
	assert.Contains(t, result, "払戻し: 25")
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

	deucesDraw := func(hand []*domain.Card) *interfaces.MockVideoPokerGame {
		m := new(interfaces.MockVideoPokerGame)
		m.On("GetPhase").Return(domain.VideoPokerPhaseDraw)
		m.On("GetVariantName").Return("deuceswild")
		m.On("GetHand").Return(hand)
		return m
	}

	t.Run("holds deuces and a made pair", func(t *testing.T) {
		hand := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 2, false),  // deuce (wild)
			domain.NewCard(domain.CardDesignHeart, 8, false),  // pair of 8s
			domain.NewCard(domain.CardDesignClover, 8, false), // pair of 8s
			domain.NewCard(domain.CardDesignDiamond, 5, false),
			domain.NewCard(domain.CardDesignSpade, 11, false),
		}
		out := p.HintOutput(deucesDraw(hand))
		prefix := strings.SplitN(i18n.T("videopoker.hintHold"), "{{", 2)[0]
		assert.Contains(t, out, prefix)
		assert.Contains(t, out, "[0]") // deuce
		assert.Contains(t, out, "[1]") // pair
		assert.Contains(t, out, "[2]") // pair
		assert.NotContains(t, out, "[3]")
	})

	t.Run("recommends redraw with no deuces or pair", func(t *testing.T) {
		hand := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 3, false),
			domain.NewCard(domain.CardDesignHeart, 7, false),
			domain.NewCard(domain.CardDesignClover, 9, false),
			domain.NewCard(domain.CardDesignDiamond, 11, false),
			domain.NewCard(domain.CardDesignSpade, 13, false),
		}
		assert.Contains(t, p.HintOutput(deucesDraw(hand)), i18n.T("videopoker.hintHoldNone"))
	})

	t.Run("no hint for a non-deuceswild variant", func(t *testing.T) {
		m := new(interfaces.MockVideoPokerGame)
		m.On("GetPhase").Return(domain.VideoPokerPhaseDraw)
		m.On("GetVariantName").Return("jacksorbetter")
		assert.Contains(t, p.HintOutput(m), i18n.T("videopoker.hintNone"))
	})

	t.Run("no hint outside the draw phase", func(t *testing.T) {
		m := new(interfaces.MockVideoPokerGame)
		m.On("GetPhase").Return(domain.VideoPokerPhaseBet)
		assert.Contains(t, p.HintOutput(m), i18n.T("videopoker.hintNone"))
	})
}
