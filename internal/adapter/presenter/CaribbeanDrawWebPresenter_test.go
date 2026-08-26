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

func setupCaribbeanDrawWebMockDefaults(m *interfaces.MockCaribbeanDrawGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.CaribbeanDrawPhaseBet).Maybe()
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
	m.On("GetDrawCost").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func parseCaribbeanDrawOutput(t *testing.T, jsonStr string) *controller.CaribbeanDrawWebOutput {
	t.Helper()
	var out controller.CaribbeanDrawWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

func TestCaribbeanDrawWebPresenter_Output_BetPhase(t *testing.T) {
	p := new(CaribbeanDrawWebPresenter)
	m := new(interfaces.MockCaribbeanDrawGame)
	setupCaribbeanDrawWebMockDefaults(m)

	result := parseCaribbeanDrawOutput(t, p.Output(m, nil))
	assert.Equal(t, domain.CaribbeanDrawPhaseBet, result.Phase)
	assert.Equal(t, 1000, result.Chips)
	assert.Empty(t, result.PlayerHand)
	assert.Empty(t, result.DealerHand)
	assert.Empty(t, result.Message)
}

func TestCaribbeanDrawWebPresenter_Output_Error(t *testing.T) {
	p := new(CaribbeanDrawWebPresenter)
	m := new(interfaces.MockCaribbeanDrawGame)
	setupCaribbeanDrawWebMockDefaults(m)

	result := parseCaribbeanDrawOutput(t, p.Output(m, errors.New("test error")))
	assert.Equal(t, "test error", result.Message)
}

func TestCaribbeanDrawWebPresenter_Output_PlayerWins(t *testing.T) {
	p := new(CaribbeanDrawWebPresenter)
	m := new(interfaces.MockCaribbeanDrawGame)
	m.On("GetChips").Return(1400).Maybe()
	m.On("GetPhase").Return(domain.CaribbeanDrawPhaseEnd).Maybe()
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
	m.On("GetDrawCost").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(1).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := parseCaribbeanDrawOutput(t, p.Output(m, nil))
	assert.Equal(t, "Player wins!", result.Message)
	assert.Equal(t, "caribbeandraw.result.playerWins", result.MessageCode)
}

func TestCaribbeanDrawWebPresenter_Output_DealerWins(t *testing.T) {
	p := new(CaribbeanDrawWebPresenter)
	m := new(interfaces.MockCaribbeanDrawGame)
	m.On("GetChips").Return(700).Maybe()
	m.On("GetPhase").Return(domain.CaribbeanDrawPhaseEnd).Maybe()
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
	m.On("GetDrawCost").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(1).Maybe()
	m.On("GetDealerHandRank").Return(2).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := parseCaribbeanDrawOutput(t, p.Output(m, nil))
	assert.Equal(t, "Dealer wins!", result.Message)
	assert.Equal(t, "caribbeandraw.result.dealerWins", result.MessageCode)
}

func TestCaribbeanDrawWebPresenter_Output_Fold(t *testing.T) {
	p := new(CaribbeanDrawWebPresenter)
	m := new(interfaces.MockCaribbeanDrawGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.CaribbeanDrawPhaseEnd).Maybe()
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
	m.On("GetDrawCost").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(1).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := parseCaribbeanDrawOutput(t, p.Output(m, nil))
	assert.Equal(t, "Player folded.", result.Message)
	assert.Equal(t, "caribbeandraw.result.fold", result.MessageCode)
}

func TestCaribbeanDrawWebPresenter_Output_Push(t *testing.T) {
	p := new(CaribbeanDrawWebPresenter)
	m := new(interfaces.MockCaribbeanDrawGame)
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.CaribbeanDrawPhaseEnd).Maybe()
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
	m.On("GetDrawCost").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(1).Maybe()
	m.On("GetDealerHandRank").Return(1).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := parseCaribbeanDrawOutput(t, p.Output(m, nil))
	assert.Equal(t, "Push!", result.Message)
	assert.Equal(t, "caribbeandraw.result.push", result.MessageCode)
}

func TestCaribbeanDrawWebPresenter_Output_DealerNotQualified(t *testing.T) {
	p := new(CaribbeanDrawWebPresenter)
	m := new(interfaces.MockCaribbeanDrawGame)
	m.On("GetChips").Return(1100).Maybe()
	m.On("GetPhase").Return(domain.CaribbeanDrawPhaseEnd).Maybe()
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
	m.On("GetDrawCost").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(1).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := parseCaribbeanDrawOutput(t, p.Output(m, nil))
	assert.Equal(t, "Dealer does not qualify!", result.Message)
	assert.Equal(t, "caribbeandraw.result.dealerNotQualified", result.MessageCode)
}

func TestCaribbeanDrawWebPresenter_Output_ActionPhase_DealerMasking(t *testing.T) {
	p := new(CaribbeanDrawWebPresenter)
	m := new(interfaces.MockCaribbeanDrawGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.CaribbeanDrawPhaseAction).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 10, false),
		domain.NewCard(domain.CardDesignHeart, 11, false),
		domain.NewCard(domain.CardDesignDiamond, 13, false),
		domain.NewCard(domain.CardDesignClover, 5, false),
		domain.NewCard(domain.CardDesignSpade, 7, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 13, false),
		domain.NewCard(domain.CardDesignSpade, 5, false),
		domain.NewCard(domain.CardDesignClover, 3, false),
		domain.NewCard(domain.CardDesignDiamond, 8, false),
		domain.NewCard(domain.CardDesignHeart, 2, false),
	}).Maybe()
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
	m.On("GetDrawCost").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := parseCaribbeanDrawOutput(t, p.Output(m, nil))
	assert.Equal(t, domain.CaribbeanDrawPhaseAction, result.Phase)
	assert.Len(t, result.DealerHand, 5)
	// First card is visible (face-up)
	assert.Equal(t, "HEART", result.DealerHand[0].Design)
	assert.Equal(t, 13, result.DealerHand[0].Value)
	// Remaining cards are masked (face-down)
	for i := 1; i < 5; i++ {
		assert.Equal(t, "", result.DealerHand[i].Design, "card %d should be masked", i)
		assert.Equal(t, 0, result.DealerHand[i].Value, "card %d value should be 0", i)
	}
}

func TestCaribbeanDrawWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(CaribbeanDrawWebPresenter)
	m := new(interfaces.MockCaribbeanDrawGame)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()

	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "entries")
}
