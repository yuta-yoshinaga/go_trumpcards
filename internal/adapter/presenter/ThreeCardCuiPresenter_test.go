package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupThreeCardCuiMockDefaults(m *interfaces.MockThreeCardGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.ThreeCardPhaseBet).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteBet").Return(0).Maybe()
	m.On("GetPairPlusBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetAnteBonusPayout").Return(0).Maybe()
	m.On("GetPairPlusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func TestThreeCardCuiPresenter_Output_BetPhase(t *testing.T) {
	p := new(ThreeCardCuiPresenter)
	m := new(interfaces.MockThreeCardGame)
	setupThreeCardCuiMockDefaults(m)

	result := p.Output(m, nil)
	assert.Contains(t, result, "chips: 1000")
	assert.Contains(t, result, "phase: BET")
}

func TestThreeCardCuiPresenter_Output_ActionPhase(t *testing.T) {
	p := new(ThreeCardCuiPresenter)
	m := new(interfaces.MockThreeCardGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.ThreeCardPhaseAction).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignClover, 10, false),
		domain.NewCard(domain.CardDesignHeart, 5, false),
	}).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetPairPlusBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetAnteBonusPayout").Return(0).Maybe()
	m.On("GetPairPlusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetPlayerHandRank").Return(domain.ThreeCardHandHighCard).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "phase: ACTION")
	assert.Contains(t, result, "PLAYER")
	assert.Contains(t, result, "High Card")
}

func TestThreeCardCuiPresenter_Output_EndPhase_PlayerWins(t *testing.T) {
	p := new(ThreeCardCuiPresenter)
	m := new(interfaces.MockThreeCardGame)
	m.On("GetChips").Return(1200).Maybe()
	m.On("GetPhase").Return(domain.ThreeCardPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignClover, 10, false),
		domain.NewCard(domain.CardDesignHeart, 5, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 12, false),
		domain.NewCard(domain.CardDesignHeart, 3, false),
		domain.NewCard(domain.CardDesignClover, 2, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetPairPlusBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(100).Maybe()
	m.On("GetResult").Return(domain.GameResultWin).Maybe()
	m.On("GetAntePayout").Return(200).Maybe()
	m.On("GetPlayPayout").Return(200).Maybe()
	m.On("GetAnteBonusPayout").Return(0).Maybe()
	m.On("GetPairPlusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(400).Maybe()
	m.On("GetDealerQualified").Return(true).Maybe()
	m.On("GetPlayerHandRank").Return(domain.ThreeCardHandHighCard).Maybe()
	m.On("GetDealerHandRank").Return(domain.ThreeCardHandHighCard).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "phase: END")
	assert.Contains(t, result, "Player wins!")
	assert.Contains(t, result, "DEALER")
	assert.Contains(t, result, "(Qualified)")
	assert.Contains(t, result, "total payout: 400")
}

func TestThreeCardCuiPresenter_Output_EndPhase_Fold(t *testing.T) {
	p := new(ThreeCardCuiPresenter)
	m := new(interfaces.MockThreeCardGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.ThreeCardPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 2, false),
		domain.NewCard(domain.CardDesignClover, 4, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 12, false),
		domain.NewCard(domain.CardDesignHeart, 3, false),
		domain.NewCard(domain.CardDesignClover, 2, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetPairPlusBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe() // fold
	m.On("GetResult").Return(domain.GameResultLose).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetAnteBonusPayout").Return(0).Maybe()
	m.On("GetPairPlusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetPlayerHandRank").Return(domain.ThreeCardHandHighCard).Maybe()
	m.On("GetDealerHandRank").Return(domain.ThreeCardHandHighCard).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "Player folded.")
}

func TestThreeCardCuiPresenter_Output_Error(t *testing.T) {
	p := new(ThreeCardCuiPresenter)
	m := new(interfaces.MockThreeCardGame)
	setupThreeCardCuiMockDefaults(m)

	result := p.Output(m, domain.NewDomainError(domain.ErrWrongPhase, "wrong phase"))
	assert.Contains(t, result, "wrong phase")
}

func TestThreeCardCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(ThreeCardCuiPresenter)
	m := new(interfaces.MockThreeCardGame)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()

	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "棋譜はありません")
}
