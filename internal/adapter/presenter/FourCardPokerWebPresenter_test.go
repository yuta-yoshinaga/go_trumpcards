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

func setupFourCardPokerWebMockDefaults(m *interfaces.MockFourCardPokerGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.FourCardPokerPhaseBet).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetPlayerBest").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerBest").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerUpCard").Return((*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteBet").Return(0).Maybe()
	m.On("GetAcesUpBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe()
	m.On("GetPlayMultiplier").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetAnteBonusPayout").Return(0).Maybe()
	m.On("GetAcesUpPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("RecommendPlayMultiplier").Return(0).Maybe()
}

func parseFourCardPokerOutput(t *testing.T, jsonStr string) *controller.FourCardPokerWebOutput {
	t.Helper()
	var out controller.FourCardPokerWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

func TestFourCardPokerWebPresenter_Output_BetPhase(t *testing.T) {
	p := new(FourCardPokerWebPresenter)
	m := new(interfaces.MockFourCardPokerGame)
	setupFourCardPokerWebMockDefaults(m)

	r := parseFourCardPokerOutput(t, p.Output(m, nil))
	assert.Equal(t, domain.FourCardPokerPhaseBet, r.Phase)
	assert.Equal(t, 1000, r.Chips)
	assert.Empty(t, r.PlayerHand)
	assert.Empty(t, r.DealerHand)
	assert.Empty(t, r.Message)
}

func TestFourCardPokerWebPresenter_Output_ActionHidesDealerHoleCards(t *testing.T) {
	p := new(FourCardPokerWebPresenter)
	m := new(interfaces.MockFourCardPokerGame)
	up := domain.NewCard(domain.CardDesignSpade, 13, false)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.FourCardPokerPhaseAction).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignClover, 10, false),
		domain.NewCard(domain.CardDesignHeart, 5, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false),
		domain.NewCard(domain.CardDesignClover, 2, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{up,
		domain.NewCard(domain.CardDesignSpade, 4, false),
		domain.NewCard(domain.CardDesignSpade, 6, false),
		domain.NewCard(domain.CardDesignSpade, 8, false),
		domain.NewCard(domain.CardDesignSpade, 9, false),
		domain.NewCard(domain.CardDesignSpade, 11, false),
	}).Maybe()
	m.On("GetPlayerBest").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerBest").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerUpCard").Return(up).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetAcesUpBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe()
	m.On("GetPlayMultiplier").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetAnteBonusPayout").Return(0).Maybe()
	m.On("GetAcesUpPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(domain.FourCardHandHighCard).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	r := parseFourCardPokerOutput(t, p.Output(m, nil))
	assert.Equal(t, domain.FourCardPokerPhaseAction, r.Phase)
	assert.Len(t, r.PlayerHand, 5)
	assert.Len(t, r.DealerHand, 1)
	assert.Empty(t, r.DealerBest)
}

func setupFourCardPokerEndPhase(m *interfaces.MockFourCardPokerGame, result domain.GameResult, playBet int) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.FourCardPokerPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetPlayerBest").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerBest").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerUpCard").Return((*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetAcesUpBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(playBet).Maybe()
	m.On("GetPlayMultiplier").Return(0).Maybe()
	m.On("GetResult").Return(result).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetAnteBonusPayout").Return(0).Maybe()
	m.On("GetAcesUpPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func TestFourCardPokerWebPresenter_Output_PlayerWins(t *testing.T) {
	p := new(FourCardPokerWebPresenter)
	m := new(interfaces.MockFourCardPokerGame)
	setupFourCardPokerEndPhase(m, domain.GameResultWin, 100)

	r := parseFourCardPokerOutput(t, p.Output(m, nil))
	assert.Equal(t, "Player wins!", r.Message)
	assert.Equal(t, "fourcardpoker.result.playerWins", r.MessageCode)
}

func TestFourCardPokerWebPresenter_Output_DealerWins(t *testing.T) {
	p := new(FourCardPokerWebPresenter)
	m := new(interfaces.MockFourCardPokerGame)
	setupFourCardPokerEndPhase(m, domain.GameResultLose, 100)

	r := parseFourCardPokerOutput(t, p.Output(m, nil))
	assert.Equal(t, "Dealer wins!", r.Message)
	assert.Equal(t, "fourcardpoker.result.dealerWins", r.MessageCode)
}

func TestFourCardPokerWebPresenter_Output_Fold(t *testing.T) {
	p := new(FourCardPokerWebPresenter)
	m := new(interfaces.MockFourCardPokerGame)
	setupFourCardPokerEndPhase(m, domain.GameResultLose, 0)

	r := parseFourCardPokerOutput(t, p.Output(m, nil))
	assert.Equal(t, "Player folded.", r.Message)
	assert.Equal(t, "fourcardpoker.result.fold", r.MessageCode)
}

func TestFourCardPokerWebPresenter_Output_Push(t *testing.T) {
	p := new(FourCardPokerWebPresenter)
	m := new(interfaces.MockFourCardPokerGame)
	setupFourCardPokerEndPhase(m, domain.GameResultDraw, 100)

	r := parseFourCardPokerOutput(t, p.Output(m, nil))
	assert.Equal(t, "Push!", r.Message)
	assert.Equal(t, "fourcardpoker.result.push", r.MessageCode)
}

func TestFourCardPokerWebPresenter_Output_Error(t *testing.T) {
	p := new(FourCardPokerWebPresenter)
	m := new(interfaces.MockFourCardPokerGame)
	setupFourCardPokerWebMockDefaults(m)

	r := parseFourCardPokerOutput(t, p.Output(m, errors.New("boom")))
	assert.Equal(t, "boom", r.Message)
}

func TestFourCardPokerWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(FourCardPokerWebPresenter)
	m := new(interfaces.MockFourCardPokerGame)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()

	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "[")
}

func TestFourCardPokerWebPresenter_HintOutput(t *testing.T) {
	p := new(FourCardPokerWebPresenter)
	m := new(interfaces.MockFourCardPokerGame)
	setupFourCardPokerWebMockDefaults(m)

	r := parseFourCardPokerOutput(t, p.HintOutput(m))
	assert.Equal(t, domain.FourCardPokerPhaseBet, r.Phase)
}
