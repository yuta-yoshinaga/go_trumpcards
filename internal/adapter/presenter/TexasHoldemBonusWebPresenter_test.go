package presenter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupTexasHoldemBonusWebMockDefaults(m *interfaces.MockTexasHoldemBonusGame) {
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
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func parseTexasHoldemBonusOutput(t *testing.T, jsonStr string) *controller.TexasHoldemBonusWebOutput {
	t.Helper()
	var out controller.TexasHoldemBonusWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

func TestTexasHoldemBonusWebPresenter_Output_BetPhase(t *testing.T) {
	p := new(TexasHoldemBonusWebPresenter)
	m := new(interfaces.MockTexasHoldemBonusGame)
	setupTexasHoldemBonusWebMockDefaults(m)

	result := parseTexasHoldemBonusOutput(t, p.Output(m, nil))
	assert.Equal(t, domain.TexasHoldemBonusPhaseBet, result.Phase)
	assert.Equal(t, 1000, result.Chips)
	assert.Empty(t, result.PlayerHand)
	assert.Empty(t, result.DealerHand)
	assert.Empty(t, result.Community)
	assert.Empty(t, result.Message)
}

func TestTexasHoldemBonusWebPresenter_Output_Error(t *testing.T) {
	p := new(TexasHoldemBonusWebPresenter)
	m := new(interfaces.MockTexasHoldemBonusGame)
	setupTexasHoldemBonusWebMockDefaults(m)

	result := parseTexasHoldemBonusOutput(t, p.Output(m, errors.New("test error")))
	assert.Equal(t, "test error", result.Message)
}

func TestTexasHoldemBonusWebPresenter_Output_PreFlop_DealerMasked(t *testing.T) {
	p := new(TexasHoldemBonusWebPresenter)
	m := new(interfaces.MockTexasHoldemBonusGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.TexasHoldemBonusPhasePreFlop).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignSpade, 13, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignDiamond, 5, false),
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

	result := parseTexasHoldemBonusOutput(t, p.Output(m, nil))
	assert.Len(t, result.DealerHand, 2)
	for i := 0; i < 2; i++ {
		assert.Equal(t, "", result.DealerHand[i].Design)
		assert.Equal(t, 0, result.DealerHand[i].Value)
	}
}

func TestTexasHoldemBonusWebPresenter_Output_PlayerWins(t *testing.T) {
	p := new(TexasHoldemBonusWebPresenter)
	m := new(interfaces.MockTexasHoldemBonusGame)
	m.On("GetChips").Return(1500).Maybe()
	m.On("GetPhase").Return(domain.TexasHoldemBonusPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetCommunity").Return(([]*domain.Card)(nil)).Maybe()
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
	m.On("GetPlayerHandRank").Return(1).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := parseTexasHoldemBonusOutput(t, p.Output(m, nil))
	assert.Equal(t, "Player wins!", result.Message)
	assert.Equal(t, "texasholdembonus.result.playerWins", result.MessageCode)
}

func TestTexasHoldemBonusWebPresenter_Output_DealerWins(t *testing.T) {
	p := new(TexasHoldemBonusWebPresenter)
	m := new(interfaces.MockTexasHoldemBonusGame)
	m.On("GetChips").Return(700).Maybe()
	m.On("GetPhase").Return(domain.TexasHoldemBonusPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetCommunity").Return(([]*domain.Card)(nil)).Maybe()
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
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(1).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := parseTexasHoldemBonusOutput(t, p.Output(m, nil))
	assert.Equal(t, "Dealer wins!", result.Message)
	assert.Equal(t, "texasholdembonus.result.dealerWins", result.MessageCode)
}

func TestTexasHoldemBonusWebPresenter_Output_Fold(t *testing.T) {
	p := new(TexasHoldemBonusWebPresenter)
	m := new(interfaces.MockTexasHoldemBonusGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.TexasHoldemBonusPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
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

	result := parseTexasHoldemBonusOutput(t, p.Output(m, nil))
	assert.Equal(t, "Player folded.", result.Message)
	assert.Equal(t, "texasholdembonus.result.fold", result.MessageCode)
}

func TestTexasHoldemBonusWebPresenter_Output_Push(t *testing.T) {
	p := new(TexasHoldemBonusWebPresenter)
	m := new(interfaces.MockTexasHoldemBonusGame)
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.TexasHoldemBonusPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetCommunity").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetBonusBet").Return(0).Maybe()
	m.On("GetFlopBet").Return(200).Maybe()
	m.On("GetTurnBet").Return(0).Maybe()
	m.On("GetRiverBet").Return(0).Maybe()
	m.On("GetTotalPlayBet").Return(200).Maybe()
	m.On("GetResult").Return(domain.GameResultDraw).Maybe()
	m.On("GetAntePayout").Return(100).Maybe()
	m.On("GetPlayPayout").Return(200).Maybe()
	m.On("GetBonusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(300).Maybe()
	m.On("GetPlayerHandRank").Return(1).Maybe()
	m.On("GetDealerHandRank").Return(1).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := parseTexasHoldemBonusOutput(t, p.Output(m, nil))
	assert.Equal(t, "Push!", result.Message)
	assert.Equal(t, "texasholdembonus.result.push", result.MessageCode)
}

func TestTexasHoldemBonusWebPresenter_Output_EndPhase_DealerVisible(t *testing.T) {
	p := new(TexasHoldemBonusWebPresenter)
	m := new(interfaces.MockTexasHoldemBonusGame)
	m.On("GetChips").Return(1500).Maybe()
	m.On("GetPhase").Return(domain.TexasHoldemBonusPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignClover, 1, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignDiamond, 5, false),
	}).Maybe()
	m.On("GetCommunity").Return(([]*domain.Card)(nil)).Maybe()
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
	m.On("GetPlayerHandRank").Return(1).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := parseTexasHoldemBonusOutput(t, p.Output(m, nil))
	assert.Len(t, result.DealerHand, 2)
	assert.NotEqual(t, "", result.DealerHand[0].Design)
	assert.NotEqual(t, 0, result.DealerHand[0].Value)
}

func TestTexasHoldemBonusWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(TexasHoldemBonusWebPresenter)
	m := new(interfaces.MockTexasHoldemBonusGame)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()

	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "entries")
}
