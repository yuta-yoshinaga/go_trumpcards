package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
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
}

func TestVideoPokerCuiPresenter_Output_BetPhase(t *testing.T) {
	p := new(VideoPokerCuiPresenter)
	m := new(interfaces.MockVideoPokerGame)
	setupVideoPokerCuiMockDefaults(m)

	result := p.Output(m, nil)
	assert.Contains(t, result, "チップ: 1000")
	assert.Contains(t, result, "フェーズ: ベット")
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
