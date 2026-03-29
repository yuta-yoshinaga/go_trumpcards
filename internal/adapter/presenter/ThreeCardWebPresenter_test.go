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

func setupThreeCardWebMockDefaults(m *interfaces.MockThreeCardGame) {
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

func parseThreeCardOutput(t *testing.T, jsonStr string) *controller.ThreeCardWebOutput {
	t.Helper()
	var out controller.ThreeCardWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

func TestThreeCardWebPresenter_Output_BetPhase(t *testing.T) {
	p := new(ThreeCardWebPresenter)
	m := new(interfaces.MockThreeCardGame)
	setupThreeCardWebMockDefaults(m)

	result := parseThreeCardOutput(t, p.Output(m, nil))
	assert.Equal(t, domain.ThreeCardPhaseBet, result.Phase)
	assert.Equal(t, 1000, result.Chips)
	assert.Empty(t, result.PlayerHand)
	assert.Empty(t, result.DealerHand)
	assert.Empty(t, result.Message)
}

func TestThreeCardWebPresenter_Output_Error(t *testing.T) {
	p := new(ThreeCardWebPresenter)
	m := new(interfaces.MockThreeCardGame)
	setupThreeCardWebMockDefaults(m)

	result := parseThreeCardOutput(t, p.Output(m, errors.New("test error")))
	assert.Equal(t, "test error", result.Message)
}

func TestThreeCardWebPresenter_Output_PlayerWins(t *testing.T) {
	p := new(ThreeCardWebPresenter)
	m := new(interfaces.MockThreeCardGame)
	m.On("GetChips").Return(1200).Maybe()
	m.On("GetPhase").Return(domain.ThreeCardPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
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
	m.On("GetPlayerHandRank").Return(1).Maybe()
	m.On("GetDealerHandRank").Return(1).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := parseThreeCardOutput(t, p.Output(m, nil))
	assert.Equal(t, "Player wins!", result.Message)
	assert.Equal(t, "threecard.result.playerWins", result.MessageCode)
}

func TestThreeCardWebPresenter_Output_DealerWins(t *testing.T) {
	p := new(ThreeCardWebPresenter)
	m := new(interfaces.MockThreeCardGame)
	m.On("GetChips").Return(800).Maybe()
	m.On("GetPhase").Return(domain.ThreeCardPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetPairPlusBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(100).Maybe()
	m.On("GetResult").Return(domain.GameResultLose).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetAnteBonusPayout").Return(0).Maybe()
	m.On("GetPairPlusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerQualified").Return(true).Maybe()
	m.On("GetPlayerHandRank").Return(1).Maybe()
	m.On("GetDealerHandRank").Return(1).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := parseThreeCardOutput(t, p.Output(m, nil))
	assert.Equal(t, "Dealer wins!", result.Message)
	assert.Equal(t, "threecard.result.dealerWins", result.MessageCode)
}

func TestThreeCardWebPresenter_Output_Fold(t *testing.T) {
	p := new(ThreeCardWebPresenter)
	m := new(interfaces.MockThreeCardGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.ThreeCardPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetPairPlusBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe() // No play bet = fold
	m.On("GetResult").Return(domain.GameResultLose).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetAnteBonusPayout").Return(0).Maybe()
	m.On("GetPairPlusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetPlayerHandRank").Return(1).Maybe()
	m.On("GetDealerHandRank").Return(1).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := parseThreeCardOutput(t, p.Output(m, nil))
	assert.Equal(t, "Player folded.", result.Message)
	assert.Equal(t, "threecard.result.fold", result.MessageCode)
}

func TestThreeCardWebPresenter_Output_Push(t *testing.T) {
	p := new(ThreeCardWebPresenter)
	m := new(interfaces.MockThreeCardGame)
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.ThreeCardPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetPairPlusBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(100).Maybe()
	m.On("GetResult").Return(domain.GameResultDraw).Maybe()
	m.On("GetAntePayout").Return(100).Maybe()
	m.On("GetPlayPayout").Return(100).Maybe()
	m.On("GetAnteBonusPayout").Return(0).Maybe()
	m.On("GetPairPlusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(200).Maybe()
	m.On("GetDealerQualified").Return(true).Maybe()
	m.On("GetPlayerHandRank").Return(1).Maybe()
	m.On("GetDealerHandRank").Return(1).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := parseThreeCardOutput(t, p.Output(m, nil))
	assert.Equal(t, "Push!", result.Message)
	assert.Equal(t, "threecard.result.push", result.MessageCode)
}

func TestThreeCardWebPresenter_Output_DealerNotQualified(t *testing.T) {
	p := new(ThreeCardWebPresenter)
	m := new(interfaces.MockThreeCardGame)
	m.On("GetChips").Return(1100).Maybe()
	m.On("GetPhase").Return(domain.ThreeCardPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetPairPlusBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(100).Maybe()
	m.On("GetResult").Return(domain.GameResultWin).Maybe()
	m.On("GetAntePayout").Return(200).Maybe()
	m.On("GetPlayPayout").Return(100).Maybe()
	m.On("GetAnteBonusPayout").Return(0).Maybe()
	m.On("GetPairPlusPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(300).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetPlayerHandRank").Return(1).Maybe()
	m.On("GetDealerHandRank").Return(1).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := parseThreeCardOutput(t, p.Output(m, nil))
	assert.Equal(t, "Dealer does not qualify!", result.Message)
	assert.Equal(t, "threecard.result.dealerNotQualified", result.MessageCode)
}

func TestThreeCardWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(ThreeCardWebPresenter)
	m := new(interfaces.MockThreeCardGame)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()

	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "entries")
}
