package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupTexasHoldemBonusCuiMockDefaults(m *interfaces.MockTexasHoldemBonusGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.TexasHoldemBonusPhaseBet).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetCommunity").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteBet").Return(0).Maybe()
	m.On("GetBonusBet").Return(0).Maybe()
	m.On("GetFlopBet").Return(0).Maybe()
	m.On("GetTurnBet").Return(0).Maybe()
	m.On("GetRiverBet").Return(0).Maybe()
	m.On("GetTotalPlayBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetBonusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetPlayerBest").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerBest").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func TestTexasHoldemBonusCuiPresenter_Output_BetPhase(t *testing.T) {
	p := new(TexasHoldemBonusCuiPresenter)
	m := new(interfaces.MockTexasHoldemBonusGame)
	setupTexasHoldemBonusCuiMockDefaults(m)

	result := p.Output(m, nil)
	assert.Contains(t, result, "chips: 1000")
	assert.Contains(t, result, "phase: BET")
}

func TestTexasHoldemBonusCuiPresenter_Output_PreFlopPhase(t *testing.T) {
	p := new(TexasHoldemBonusCuiPresenter)
	m := new(interfaces.MockTexasHoldemBonusGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.TexasHoldemBonusPhasePreFlop).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignClover, 13, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 5, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false),
	}).Maybe()
	m.On("GetCommunity").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetBonusBet").Return(0).Maybe()
	m.On("GetFlopBet").Return(0).Maybe()
	m.On("GetTurnBet").Return(0).Maybe()
	m.On("GetRiverBet").Return(0).Maybe()
	m.On("GetTotalPlayBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetBonusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "phase: PRE-FLOP")
	assert.Contains(t, result, "PLAYER")
	assert.Contains(t, result, "DEALER")
	// Dealer cards hidden in pre-flop
	assert.Contains(t, result, "??")
}

func TestTexasHoldemBonusCuiPresenter_Output_FlopPhase(t *testing.T) {
	p := new(TexasHoldemBonusCuiPresenter)
	m := new(interfaces.MockTexasHoldemBonusGame)
	m.On("GetChips").Return(700).Maybe()
	m.On("GetPhase").Return(domain.TexasHoldemBonusPhaseFlop).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignClover, 13, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 5, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false),
	}).Maybe()
	m.On("GetCommunity").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 11, false),
		domain.NewCard(domain.CardDesignSpade, 12, false),
		domain.NewCard(domain.CardDesignSpade, 10, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetBonusBet").Return(0).Maybe()
	m.On("GetFlopBet").Return(200).Maybe()
	m.On("GetTurnBet").Return(0).Maybe()
	m.On("GetRiverBet").Return(0).Maybe()
	m.On("GetTotalPlayBet").Return(200).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetBonusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "phase: FLOP")
	assert.Contains(t, result, "BOARD")
}

func TestTexasHoldemBonusCuiPresenter_Output_EndPhase_PlayerWins(t *testing.T) {
	p := new(TexasHoldemBonusCuiPresenter)
	m := new(interfaces.MockTexasHoldemBonusGame)
	m.On("GetChips").Return(1500).Maybe()
	m.On("GetPhase").Return(domain.TexasHoldemBonusPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignClover, 1, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 13, false),
		domain.NewCard(domain.CardDesignDiamond, 13, false),
	}).Maybe()
	m.On("GetCommunity").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 2, false),
		domain.NewCard(domain.CardDesignClover, 4, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignSpade, 9, false),
		domain.NewCard(domain.CardDesignDiamond, 12, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetBonusBet").Return(0).Maybe()
	m.On("GetFlopBet").Return(200).Maybe()
	m.On("GetTurnBet").Return(0).Maybe()
	m.On("GetRiverBet").Return(0).Maybe()
	m.On("GetTotalPlayBet").Return(200).Maybe()
	m.On("GetResult").Return(domain.GameResultWin).Maybe()
	m.On("GetAntePayout").Return(200).Maybe()
	m.On("GetPlayPayout").Return(400).Maybe()
	m.On("GetBonusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(600).Maybe()
	m.On("GetPlayerHandRank").Return(domain.PokerHandOnePair).Maybe()
	m.On("GetDealerHandRank").Return(domain.PokerHandOnePair).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "phase: END")
	assert.Contains(t, result, "Player wins!")
	assert.Contains(t, result, "total payout: 600")
}

func TestTexasHoldemBonusCuiPresenter_Output_EndPhase_Fold(t *testing.T) {
	p := new(TexasHoldemBonusCuiPresenter)
	m := new(interfaces.MockTexasHoldemBonusGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.TexasHoldemBonusPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, false),
		domain.NewCard(domain.CardDesignClover, 2, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 13, false),
		domain.NewCard(domain.CardDesignDiamond, 13, false),
	}).Maybe()
	m.On("GetCommunity").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetBonusBet").Return(0).Maybe()
	m.On("GetFlopBet").Return(0).Maybe()
	m.On("GetTurnBet").Return(0).Maybe()
	m.On("GetRiverBet").Return(0).Maybe()
	m.On("GetTotalPlayBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResultLose).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetBonusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "Player folded.")
}

func TestTexasHoldemBonusCuiPresenter_Output_EndPhase_Push(t *testing.T) {
	p := new(TexasHoldemBonusCuiPresenter)
	m := new(interfaces.MockTexasHoldemBonusGame)
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.TexasHoldemBonusPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 2, false),
		domain.NewCard(domain.CardDesignSpade, 3, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignClover, 2, false),
		domain.NewCard(domain.CardDesignClover, 3, false),
	}).Maybe()
	m.On("GetCommunity").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 5, false),
		domain.NewCard(domain.CardDesignDiamond, 6, false),
		domain.NewCard(domain.CardDesignSpade, 7, false),
		domain.NewCard(domain.CardDesignDiamond, 8, false),
		domain.NewCard(domain.CardDesignClover, 9, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetBonusBet").Return(0).Maybe()
	m.On("GetFlopBet").Return(200).Maybe()
	m.On("GetTurnBet").Return(0).Maybe()
	m.On("GetRiverBet").Return(0).Maybe()
	m.On("GetTotalPlayBet").Return(200).Maybe()
	m.On("GetResult").Return(domain.GameResultDraw).Maybe()
	m.On("GetAntePayout").Return(200).Maybe()
	m.On("GetPlayPayout").Return(200).Maybe()
	m.On("GetBonusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(400).Maybe()
	m.On("GetPlayerHandRank").Return(domain.PokerHandStraight).Maybe()
	m.On("GetDealerHandRank").Return(domain.PokerHandStraight).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "Push!")
}

func TestTexasHoldemBonusCuiPresenter_Output_EndPhase_DealerWins(t *testing.T) {
	p := new(TexasHoldemBonusCuiPresenter)
	m := new(interfaces.MockTexasHoldemBonusGame)
	m.On("GetChips").Return(700).Maybe()
	m.On("GetPhase").Return(domain.TexasHoldemBonusPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 2, false),
		domain.NewCard(domain.CardDesignClover, 3, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 1, false),
		domain.NewCard(domain.CardDesignDiamond, 1, false),
	}).Maybe()
	m.On("GetCommunity").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 5, false),
		domain.NewCard(domain.CardDesignClover, 7, false),
		domain.NewCard(domain.CardDesignHeart, 9, false),
		domain.NewCard(domain.CardDesignSpade, 11, false),
		domain.NewCard(domain.CardDesignClover, 13, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetBonusBet").Return(0).Maybe()
	m.On("GetFlopBet").Return(200).Maybe()
	m.On("GetTurnBet").Return(0).Maybe()
	m.On("GetRiverBet").Return(0).Maybe()
	m.On("GetTotalPlayBet").Return(200).Maybe()
	m.On("GetResult").Return(domain.GameResultLose).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetBonusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(domain.PokerHandHighCard).Maybe()
	m.On("GetDealerHandRank").Return(domain.PokerHandOnePair).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "Dealer wins!")
}

func TestTexasHoldemBonusCuiPresenter_Output_Error(t *testing.T) {
	p := new(TexasHoldemBonusCuiPresenter)
	m := new(interfaces.MockTexasHoldemBonusGame)
	setupTexasHoldemBonusCuiMockDefaults(m)

	result := p.Output(m, domain.NewDomainError(domain.ErrWrongPhase, "wrong phase"))
	assert.Contains(t, result, "wrong phase")
}

func TestTexasHoldemBonusCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(TexasHoldemBonusCuiPresenter)
	m := new(interfaces.MockTexasHoldemBonusGame)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()

	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "棋譜はありません")
}

func TestTexasHoldemBonusCuiPresenter_PhaseStr_Unknown(t *testing.T) {
	p := new(TexasHoldemBonusCuiPresenter)
	m := new(interfaces.MockTexasHoldemBonusGame)
	m.On("GetChips").Return(0).Maybe()
	m.On("GetPhase").Return(999).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetCommunity").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "phase: UNKNOWN")
}
