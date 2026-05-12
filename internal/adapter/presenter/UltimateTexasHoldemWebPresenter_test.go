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

func setupUltimateTexasHoldemWebMockDefaults(m *interfaces.MockUltimateTexasHoldemGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.UltimateTexasHoldemPhaseBet).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetCommunity").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteBet").Return(0).Maybe()
	m.On("GetBlindBet").Return(0).Maybe()
	m.On("GetTripsBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe()
	m.On("GetFolded").Return(false).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetBlindPayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetTripsPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func parseUltimateTexasHoldemOutput(t *testing.T, jsonStr string) *controller.UltimateTexasHoldemWebOutput {
	t.Helper()
	var out controller.UltimateTexasHoldemWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

func TestUltimateTexasHoldemWebPresenter_Output_BetPhase(t *testing.T) {
	p := new(UltimateTexasHoldemWebPresenter)
	m := new(interfaces.MockUltimateTexasHoldemGame)
	setupUltimateTexasHoldemWebMockDefaults(m)

	result := parseUltimateTexasHoldemOutput(t, p.Output(m, nil))
	assert.Equal(t, domain.UltimateTexasHoldemPhaseBet, result.Phase)
	assert.Equal(t, 1000, result.Chips)
	assert.Empty(t, result.PlayerHand)
	assert.Empty(t, result.DealerHand)
	assert.Empty(t, result.Community)
	assert.Empty(t, result.Message)
}

func TestUltimateTexasHoldemWebPresenter_Output_Error(t *testing.T) {
	p := new(UltimateTexasHoldemWebPresenter)
	m := new(interfaces.MockUltimateTexasHoldemGame)
	setupUltimateTexasHoldemWebMockDefaults(m)

	result := parseUltimateTexasHoldemOutput(t, p.Output(m, errors.New("test error")))
	assert.Equal(t, "test error", result.Message)
}

func TestUltimateTexasHoldemWebPresenter_Output_PreFlop_DealerMasked(t *testing.T) {
	p := new(UltimateTexasHoldemWebPresenter)
	m := new(interfaces.MockUltimateTexasHoldemGame)
	m.On("GetChips").Return(800).Maybe()
	m.On("GetPhase").Return(domain.UltimateTexasHoldemPhasePreFlop).Maybe()
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
	m.On("GetBlindBet").Return(100).Maybe()
	m.On("GetTripsBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe()
	m.On("GetFolded").Return(false).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetBlindPayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetTripsPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := parseUltimateTexasHoldemOutput(t, p.Output(m, nil))
	assert.Len(t, result.DealerHand, 2)
	for _, c := range result.DealerHand {
		assert.Equal(t, 0, c.Value, "dealer card value must be masked in pre-flop")
		assert.Empty(t, c.Design, "dealer card design must be masked in pre-flop")
	}
}

func TestUltimateTexasHoldemWebPresenter_Output_End_PlayerWins(t *testing.T) {
	p := new(UltimateTexasHoldemWebPresenter)
	m := new(interfaces.MockUltimateTexasHoldemGame)
	m.On("GetChips").Return(1500).Maybe()
	m.On("GetPhase").Return(domain.UltimateTexasHoldemPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 13, false),
		domain.NewCard(domain.CardDesignClover, 13, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignDiamond, 5, false),
	}).Maybe()
	m.On("GetCommunity").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 11, false),
		domain.NewCard(domain.CardDesignSpade, 12, false),
		domain.NewCard(domain.CardDesignClover, 3, false),
		domain.NewCard(domain.CardDesignHeart, 8, false),
		domain.NewCard(domain.CardDesignDiamond, 2, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetBlindBet").Return(100).Maybe()
	m.On("GetTripsBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(400).Maybe()
	m.On("GetFolded").Return(false).Maybe()
	m.On("GetResult").Return(domain.GameResultWin).Maybe()
	m.On("GetDealerQualified").Return(true).Maybe()
	m.On("GetAntePayout").Return(200).Maybe()
	m.On("GetBlindPayout").Return(100).Maybe()
	m.On("GetPlayPayout").Return(800).Maybe()
	m.On("GetTripsPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(1100).Maybe()
	m.On("GetPlayerHandRank").Return(domain.PokerHandOnePair).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := parseUltimateTexasHoldemOutput(t, p.Output(m, nil))
	assert.Equal(t, "Player wins!", result.Message)
	assert.Equal(t, "ultimatetexasholdem.result.playerWins", result.MessageCode)
	assert.Len(t, result.DealerHand, 2)
	assert.Equal(t, 7, result.DealerHand[0].Value, "dealer cards revealed at end")
	assert.Equal(t, 1100, result.TotalPayout)
}

func setupUltimateTexasHoldemWebEndPhaseMock(m *interfaces.MockUltimateTexasHoldemGame, folded bool, result domain.GameResult) {
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.UltimateTexasHoldemPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetCommunity").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetBlindBet").Return(100).Maybe()
	m.On("GetTripsBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe()
	m.On("GetFolded").Return(folded).Maybe()
	m.On("GetResult").Return(result).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetBlindPayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetTripsPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func TestUltimateTexasHoldemWebPresenter_Output_End_Folded(t *testing.T) {
	p := new(UltimateTexasHoldemWebPresenter)
	m := new(interfaces.MockUltimateTexasHoldemGame)
	setupUltimateTexasHoldemWebEndPhaseMock(m, true, domain.GameResultLose)

	result := parseUltimateTexasHoldemOutput(t, p.Output(m, nil))
	assert.Equal(t, "Player folded.", result.Message)
	assert.Equal(t, "ultimatetexasholdem.result.fold", result.MessageCode)
}

func TestUltimateTexasHoldemWebPresenter_Output_End_DealerWins(t *testing.T) {
	p := new(UltimateTexasHoldemWebPresenter)
	m := new(interfaces.MockUltimateTexasHoldemGame)
	setupUltimateTexasHoldemWebEndPhaseMock(m, false, domain.GameResultLose)

	result := parseUltimateTexasHoldemOutput(t, p.Output(m, nil))
	assert.Equal(t, "Dealer wins!", result.Message)
}

func TestUltimateTexasHoldemWebPresenter_Output_End_Push(t *testing.T) {
	p := new(UltimateTexasHoldemWebPresenter)
	m := new(interfaces.MockUltimateTexasHoldemGame)
	setupUltimateTexasHoldemWebEndPhaseMock(m, false, domain.GameResultDraw)

	result := parseUltimateTexasHoldemOutput(t, p.Output(m, nil))
	assert.Equal(t, "Push!", result.Message)
}

func TestUltimateTexasHoldemWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(UltimateTexasHoldemWebPresenter)
	m := new(interfaces.MockUltimateTexasHoldemGame)
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "bet", Detail: "ante=100"},
	}).Maybe()

	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "bet")
}
