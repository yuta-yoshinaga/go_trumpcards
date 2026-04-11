package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupCaribbeanStudCuiMockDefaults(m *interfaces.MockCaribbeanStudGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.CaribbeanStudPhaseBet).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteBet").Return(0).Maybe()
	m.On("GetJackpotBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetJackpotPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func TestCaribbeanStudCuiPresenter_Output_BetPhase(t *testing.T) {
	p := new(CaribbeanStudCuiPresenter)
	m := new(interfaces.MockCaribbeanStudGame)
	setupCaribbeanStudCuiMockDefaults(m)

	result := p.Output(m, nil)
	assert.Contains(t, result, "chips: 1000")
	assert.Contains(t, result, "phase: BET")
}

func TestCaribbeanStudCuiPresenter_Output_ActionPhase(t *testing.T) {
	p := new(CaribbeanStudCuiPresenter)
	m := new(interfaces.MockCaribbeanStudGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.CaribbeanStudPhaseAction).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignClover, 10, false),
		domain.NewCard(domain.CardDesignHeart, 5, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false),
		domain.NewCard(domain.CardDesignSpade, 2, false),
	}).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetJackpotBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetJackpotPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetPlayerHandRank").Return(domain.PokerHandHighCard).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "phase: ACTION")
	assert.Contains(t, result, "PLAYER")
}

func TestCaribbeanStudCuiPresenter_Output_EndPhase_PlayerWins(t *testing.T) {
	p := new(CaribbeanStudCuiPresenter)
	m := new(interfaces.MockCaribbeanStudGame)
	m.On("GetChips").Return(1400).Maybe()
	m.On("GetPhase").Return(domain.CaribbeanStudPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, false),
		domain.NewCard(domain.CardDesignClover, 7, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignDiamond, 4, false),
		domain.NewCard(domain.CardDesignSpade, 2, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 4, false),
		domain.NewCard(domain.CardDesignHeart, 4, false),
		domain.NewCard(domain.CardDesignClover, 6, false),
		domain.NewCard(domain.CardDesignSpade, 8, false),
		domain.NewCard(domain.CardDesignDiamond, 11, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetJackpotBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(200).Maybe()
	m.On("GetResult").Return(domain.GameResultWin).Maybe()
	m.On("GetAntePayout").Return(200).Maybe()
	m.On("GetPlayPayout").Return(800).Maybe()
	m.On("GetJackpotPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(1000).Maybe()
	m.On("GetDealerQualified").Return(true).Maybe()
	m.On("GetPlayerHandRank").Return(domain.PokerHandThreeOfAKind).Maybe()
	m.On("GetDealerHandRank").Return(domain.PokerHandOnePair).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "phase: END")
	assert.Contains(t, result, "Player wins!")
	assert.Contains(t, result, "DEALER")
	assert.Contains(t, result, "(Qualified)")
	assert.Contains(t, result, "total payout: 1000")
}

func TestCaribbeanStudCuiPresenter_Output_EndPhase_Fold(t *testing.T) {
	p := new(CaribbeanStudCuiPresenter)
	m := new(interfaces.MockCaribbeanStudGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.CaribbeanStudPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 2, false),
		domain.NewCard(domain.CardDesignClover, 4, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignDiamond, 9, false),
		domain.NewCard(domain.CardDesignSpade, 10, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 12, false),
		domain.NewCard(domain.CardDesignHeart, 3, false),
		domain.NewCard(domain.CardDesignClover, 2, false),
		domain.NewCard(domain.CardDesignSpade, 5, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetJackpotBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResultLose).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetJackpotPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetPlayerHandRank").Return(domain.PokerHandHighCard).Maybe()
	m.On("GetDealerHandRank").Return(domain.PokerHandHighCard).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "Player folded.")
}

func TestCaribbeanStudCuiPresenter_Output_Error(t *testing.T) {
	p := new(CaribbeanStudCuiPresenter)
	m := new(interfaces.MockCaribbeanStudGame)
	setupCaribbeanStudCuiMockDefaults(m)

	result := p.Output(m, domain.NewDomainError(domain.ErrWrongPhase, "wrong phase"))
	assert.Contains(t, result, "wrong phase")
}

func TestCaribbeanStudCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(CaribbeanStudCuiPresenter)
	m := new(interfaces.MockCaribbeanStudGame)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()

	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "棋譜はありません")
}
