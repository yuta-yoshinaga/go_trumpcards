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

func setupCaribbeanStudWebMockDefaults(m *interfaces.MockCaribbeanStudGame) {
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

func parseCaribbeanStudOutput(t *testing.T, jsonStr string) *controller.CaribbeanStudWebOutput {
	t.Helper()
	var out controller.CaribbeanStudWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

func TestCaribbeanStudWebPresenter_Output_BetPhase(t *testing.T) {
	p := new(CaribbeanStudWebPresenter)
	m := new(interfaces.MockCaribbeanStudGame)
	setupCaribbeanStudWebMockDefaults(m)

	result := parseCaribbeanStudOutput(t, p.Output(m, nil))
	assert.Equal(t, domain.CaribbeanStudPhaseBet, result.Phase)
	assert.Equal(t, 1000, result.Chips)
	assert.Empty(t, result.PlayerHand)
	assert.Empty(t, result.DealerHand)
	assert.Empty(t, result.Message)
}

func TestCaribbeanStudWebPresenter_Output_Error(t *testing.T) {
	p := new(CaribbeanStudWebPresenter)
	m := new(interfaces.MockCaribbeanStudGame)
	setupCaribbeanStudWebMockDefaults(m)

	result := parseCaribbeanStudOutput(t, p.Output(m, errors.New("test error")))
	assert.Equal(t, "test error", result.Message)
}

func TestCaribbeanStudWebPresenter_Output_PlayerWins(t *testing.T) {
	p := new(CaribbeanStudWebPresenter)
	m := new(interfaces.MockCaribbeanStudGame)
	m.On("GetChips").Return(1400).Maybe()
	m.On("GetPhase").Return(domain.CaribbeanStudPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetJackpotBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(200).Maybe()
	m.On("GetResult").Return(domain.GameResultWin).Maybe()
	m.On("GetAntePayout").Return(200).Maybe()
	m.On("GetPlayPayout").Return(400).Maybe()
	m.On("GetJackpotPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(600).Maybe()
	m.On("GetDealerQualified").Return(true).Maybe()
	m.On("GetPlayerHandRank").Return(1).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := parseCaribbeanStudOutput(t, p.Output(m, nil))
	assert.Equal(t, "Player wins!", result.Message)
	assert.Equal(t, "caribbeanstud.result.playerWins", result.MessageCode)
}

func TestCaribbeanStudWebPresenter_Output_DealerWins(t *testing.T) {
	p := new(CaribbeanStudWebPresenter)
	m := new(interfaces.MockCaribbeanStudGame)
	m.On("GetChips").Return(700).Maybe()
	m.On("GetPhase").Return(domain.CaribbeanStudPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetJackpotBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(200).Maybe()
	m.On("GetResult").Return(domain.GameResultLose).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetJackpotPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerQualified").Return(true).Maybe()
	m.On("GetPlayerHandRank").Return(1).Maybe()
	m.On("GetDealerHandRank").Return(2).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := parseCaribbeanStudOutput(t, p.Output(m, nil))
	assert.Equal(t, "Dealer wins!", result.Message)
	assert.Equal(t, "caribbeanstud.result.dealerWins", result.MessageCode)
}

func TestCaribbeanStudWebPresenter_Output_Fold(t *testing.T) {
	p := new(CaribbeanStudWebPresenter)
	m := new(interfaces.MockCaribbeanStudGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.CaribbeanStudPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetJackpotBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe() // No play bet = fold
	m.On("GetResult").Return(domain.GameResultLose).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetJackpotPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetPlayerHandRank").Return(1).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := parseCaribbeanStudOutput(t, p.Output(m, nil))
	assert.Equal(t, "Player folded.", result.Message)
	assert.Equal(t, "caribbeanstud.result.fold", result.MessageCode)
}

func TestCaribbeanStudWebPresenter_Output_Push(t *testing.T) {
	p := new(CaribbeanStudWebPresenter)
	m := new(interfaces.MockCaribbeanStudGame)
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.CaribbeanStudPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetJackpotBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(200).Maybe()
	m.On("GetResult").Return(domain.GameResultDraw).Maybe()
	m.On("GetAntePayout").Return(100).Maybe()
	m.On("GetPlayPayout").Return(200).Maybe()
	m.On("GetJackpotPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(300).Maybe()
	m.On("GetDealerQualified").Return(true).Maybe()
	m.On("GetPlayerHandRank").Return(1).Maybe()
	m.On("GetDealerHandRank").Return(1).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := parseCaribbeanStudOutput(t, p.Output(m, nil))
	assert.Equal(t, "Push!", result.Message)
	assert.Equal(t, "caribbeanstud.result.push", result.MessageCode)
}

func TestCaribbeanStudWebPresenter_Output_DealerNotQualified(t *testing.T) {
	p := new(CaribbeanStudWebPresenter)
	m := new(interfaces.MockCaribbeanStudGame)
	m.On("GetChips").Return(1100).Maybe()
	m.On("GetPhase").Return(domain.CaribbeanStudPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetJackpotBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(200).Maybe()
	m.On("GetResult").Return(domain.GameResultWin).Maybe()
	m.On("GetAntePayout").Return(200).Maybe()
	m.On("GetPlayPayout").Return(200).Maybe()
	m.On("GetJackpotPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(400).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetPlayerHandRank").Return(1).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := parseCaribbeanStudOutput(t, p.Output(m, nil))
	assert.Equal(t, "Dealer does not qualify!", result.Message)
	assert.Equal(t, "caribbeanstud.result.dealerNotQualified", result.MessageCode)
}

func TestCaribbeanStudWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(CaribbeanStudWebPresenter)
	m := new(interfaces.MockCaribbeanStudGame)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()

	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "entries")
}
