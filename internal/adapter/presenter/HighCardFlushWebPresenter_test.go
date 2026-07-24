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

func setupHighCardFlushWebMockDefaults(m *interfaces.MockHighCardFlushGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.HighCardFlushPhaseBet).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteBet").Return(0).Maybe()
	m.On("GetFlushBonusBet").Return(0).Maybe()
	m.On("GetStraightFlushBet").Return(0).Maybe()
	m.On("GetRaiseBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetRaisePayout").Return(0).Maybe()
	m.On("GetFlushBonusPayout").Return(0).Maybe()
	m.On("GetStraightFlushPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetPlayerFlushLen").Return(0).Maybe()
	m.On("GetDealerFlushLen").Return(0).Maybe()
	m.On("GetPlayerStraightFlushLen").Return(0).Maybe()
	m.On("MaxRaiseMultiplier").Return(1).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func parseHighCardFlushOutput(t *testing.T, jsonStr string) *controller.HighCardFlushWebOutput {
	t.Helper()
	var out controller.HighCardFlushWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

func TestHighCardFlushWebPresenter_Output_BetPhase(t *testing.T) {
	p := new(HighCardFlushWebPresenter)
	m := new(interfaces.MockHighCardFlushGame)
	setupHighCardFlushWebMockDefaults(m)

	result := parseHighCardFlushOutput(t, p.Output(m, nil))
	assert.Equal(t, domain.HighCardFlushPhaseBet, result.Phase)
	assert.Equal(t, 1000, result.Chips)
	assert.Empty(t, result.PlayerHand)
	assert.Empty(t, result.Message)
}

func TestHighCardFlushWebPresenter_Output_Error(t *testing.T) {
	p := new(HighCardFlushWebPresenter)
	m := new(interfaces.MockHighCardFlushGame)
	setupHighCardFlushWebMockDefaults(m)

	result := parseHighCardFlushOutput(t, p.Output(m, errors.New("test error")))
	assert.Equal(t, "test error", result.Message)
}

func endStateMock(t *testing.T, result domain.GameResult, raise int, qualified bool) *interfaces.MockHighCardFlushGame {
	t.Helper()
	m := new(interfaces.MockHighCardFlushGame)
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.HighCardFlushPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetFlushBonusBet").Return(0).Maybe()
	m.On("GetStraightFlushBet").Return(0).Maybe()
	m.On("GetRaiseBet").Return(raise).Maybe()
	m.On("GetResult").Return(result).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetRaisePayout").Return(0).Maybe()
	m.On("GetFlushBonusPayout").Return(0).Maybe()
	m.On("GetStraightFlushPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetDealerQualified").Return(qualified).Maybe()
	m.On("GetPlayerFlushLen").Return(3).Maybe()
	m.On("GetDealerFlushLen").Return(3).Maybe()
	m.On("GetPlayerStraightFlushLen").Return(0).Maybe()
	m.On("MaxRaiseMultiplier").Return(1).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	return m
}

// The dealer's 7 cards are dealt with the player's but must stay hidden until
// showdown; the web output must not leak them during bet/action (cheat + a11y
// over-share). Only the END phase reveals them.
func TestHighCardFlushWebPresenter_Output_HidesDealerHandUntilShowdown(t *testing.T) {
	dealer := []*domain.Card{domain.NewCard(1, 5, false), domain.NewCard(2, 9, false)}
	build := func(phase int) *interfaces.MockHighCardFlushGame {
		m := new(interfaces.MockHighCardFlushGame)
		// Registered first so these win over the defaults' Bet/nil values.
		m.On("GetPhase").Return(phase).Maybe()
		m.On("GetDealerHand").Return(dealer).Maybe()
		setupHighCardFlushWebMockDefaults(m)
		return m
	}
	p := new(HighCardFlushWebPresenter)

	for _, phase := range []int{domain.HighCardFlushPhaseBet, domain.HighCardFlushPhaseAction} {
		out := parseHighCardFlushOutput(t, p.Output(build(phase), nil))
		assert.Empty(t, out.DealerHand, "dealer hand must not leak before showdown (phase %d)", phase)
	}

	endOut := parseHighCardFlushOutput(t, p.Output(build(domain.HighCardFlushPhaseEnd), nil))
	assert.Len(t, endOut.DealerHand, 2, "dealer hand is revealed at showdown")
}

func TestHighCardFlushWebPresenter_Output_PlayerWins(t *testing.T) {
	p := new(HighCardFlushWebPresenter)
	m := endStateMock(t, domain.GameResultWin, 100, true)
	result := parseHighCardFlushOutput(t, p.Output(m, nil))
	assert.Equal(t, "Player wins!", result.Message)
	assert.Equal(t, "highcardflush.result.playerWins", result.MessageCode)
}

func TestHighCardFlushWebPresenter_Output_DealerWins(t *testing.T) {
	p := new(HighCardFlushWebPresenter)
	m := endStateMock(t, domain.GameResultLose, 100, true)
	result := parseHighCardFlushOutput(t, p.Output(m, nil))
	assert.Equal(t, "Dealer wins!", result.Message)
	assert.Equal(t, "highcardflush.result.dealerWins", result.MessageCode)
}

func TestHighCardFlushWebPresenter_Output_Fold(t *testing.T) {
	p := new(HighCardFlushWebPresenter)
	m := endStateMock(t, domain.GameResultLose, 0, false)
	result := parseHighCardFlushOutput(t, p.Output(m, nil))
	assert.Equal(t, "Player folded.", result.Message)
	assert.Equal(t, "highcardflush.result.fold", result.MessageCode)
}

func TestHighCardFlushWebPresenter_Output_Push(t *testing.T) {
	p := new(HighCardFlushWebPresenter)
	m := endStateMock(t, domain.GameResultDraw, 100, true)
	result := parseHighCardFlushOutput(t, p.Output(m, nil))
	assert.Equal(t, "Push!", result.Message)
	assert.Equal(t, "highcardflush.result.push", result.MessageCode)
}

func TestHighCardFlushWebPresenter_Output_DealerNotQualified(t *testing.T) {
	p := new(HighCardFlushWebPresenter)
	m := endStateMock(t, domain.GameResultWin, 100, false)
	result := parseHighCardFlushOutput(t, p.Output(m, nil))
	assert.Equal(t, "Dealer does not qualify!", result.Message)
	assert.Equal(t, "highcardflush.result.dealerNotQualified", result.MessageCode)
}

// In the rare case where the dealer does not qualify *and* the hand is a draw,
// the "dealer not qualified" message dominates because it is the rule that
// produced the chip outcome (ante pays 1:1, raise pushes — same as the
// dealer-qualified-draw payout, so the displayed message reflecting the
// not-qualified rule is the more useful signal to the player).
func TestHighCardFlushWebPresenter_Output_DealerNotQualifiedOverridesDraw(t *testing.T) {
	p := new(HighCardFlushWebPresenter)
	m := endStateMock(t, domain.GameResultDraw, 100, false)
	result := parseHighCardFlushOutput(t, p.Output(m, nil))
	assert.Equal(t, "highcardflush.result.dealerNotQualified", result.MessageCode)
}

func TestHighCardFlushWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(HighCardFlushWebPresenter)
	m := new(interfaces.MockHighCardFlushGame)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()

	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "entries")
}
