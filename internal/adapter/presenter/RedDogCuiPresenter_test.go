package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupRedDogCuiMockDefaults(m *interfaces.MockRedDogGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.RedDogPhaseBet).Maybe()
	m.On("GetInitialCards").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetThirdCard").Return((*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnte").Return(0).Maybe()
	m.On("GetRaise").Return(0).Maybe()
	m.On("GetSpread").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func TestRedDogCuiPresenter_Output_BetPhase(t *testing.T) {
	p := new(RedDogCuiPresenter)
	m := new(interfaces.MockRedDogGame)
	setupRedDogCuiMockDefaults(m)

	result := p.Output(m, nil)
	assert.Contains(t, result, "chips: 1000")
	assert.Contains(t, result, "BET")
}

func TestRedDogCuiPresenter_Output_Error(t *testing.T) {
	p := new(RedDogCuiPresenter)
	m := new(interfaces.MockRedDogGame)
	setupRedDogCuiMockDefaults(m)
	result := p.Output(m, errors.New("oops"))
	assert.Contains(t, result, "oops")
}

func TestRedDogCuiPresenter_Output_SpreadDecision(t *testing.T) {
	p := new(RedDogCuiPresenter)
	m := new(interfaces.MockRedDogGame)
	cards := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 5, false),
		domain.NewCard(domain.CardDesignHeart, 10, false),
	}
	m.On("GetChips").Return(900)
	m.On("GetPhase").Return(domain.RedDogPhaseSpreadDecision)
	m.On("GetInitialCards").Return(cards)
	m.On("GetThirdCard").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetAnte").Return(100)
	m.On("GetRaise").Return(0)
	m.On("GetSpread").Return(4)
	m.On("GetResult").Return(domain.GameResult(0))
	m.On("GetTotalPayout").Return(0)

	result := p.Output(m, nil)
	assert.Contains(t, result, "SPREAD DECISION")
	assert.Contains(t, result, "INITIAL")
	assert.Contains(t, result, "spread: 4")
	assert.Contains(t, result, "ante: 100")
}

func TestRedDogCuiPresenter_Output_EndWin(t *testing.T) {
	p := new(RedDogCuiPresenter)
	m := new(interfaces.MockRedDogGame)
	cards := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 5, false),
		domain.NewCard(domain.CardDesignHeart, 10, false),
	}
	third := domain.NewCard(domain.CardDesignClover, 7, false)
	m.On("GetChips").Return(1100)
	m.On("GetPhase").Return(domain.RedDogPhaseEnd)
	m.On("GetInitialCards").Return(cards)
	m.On("GetThirdCard").Return(third)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetAnte").Return(100)
	m.On("GetRaise").Return(100)
	m.On("GetSpread").Return(4)
	m.On("GetResult").Return(domain.GameResultWin)
	m.On("GetTotalPayout").Return(400)

	result := p.Output(m, nil)
	assert.Contains(t, result, "Player wins")
	assert.Contains(t, result, "THIRD")
	assert.Contains(t, result, "raise: 100")
	assert.Contains(t, result, "total payout: 400")
}

func TestRedDogCuiPresenter_Output_EndLose(t *testing.T) {
	p := new(RedDogCuiPresenter)
	m := new(interfaces.MockRedDogGame)
	setupRedDogCuiMockDefaults(m)
	m2 := new(interfaces.MockRedDogGame)
	m2.On("GetChips").Return(900)
	m2.On("GetPhase").Return(domain.RedDogPhaseEnd)
	m2.On("GetInitialCards").Return(([]*domain.Card)(nil))
	m2.On("GetThirdCard").Return((*domain.Card)(nil))
	m2.On("GetGameEndFlag").Return(true)
	m2.On("GetAnte").Return(100)
	m2.On("GetRaise").Return(0)
	m2.On("GetSpread").Return(3)
	m2.On("GetResult").Return(domain.GameResultLose)
	m2.On("GetTotalPayout").Return(0)
	result := p.Output(m2, nil)
	assert.Contains(t, result, "Player loses")
}

func TestRedDogCuiPresenter_Output_EndPush(t *testing.T) {
	p := new(RedDogCuiPresenter)
	m := new(interfaces.MockRedDogGame)
	m.On("GetChips").Return(1000)
	m.On("GetPhase").Return(domain.RedDogPhaseEnd)
	m.On("GetInitialCards").Return(([]*domain.Card)(nil))
	m.On("GetThirdCard").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(true)
	m.On("GetAnte").Return(100)
	m.On("GetRaise").Return(0)
	m.On("GetSpread").Return(0)
	m.On("GetResult").Return(domain.GameResultDraw)
	m.On("GetTotalPayout").Return(100)
	result := p.Output(m, nil)
	assert.Contains(t, result, "Push")
}

func TestRedDogCuiPresenter_PhaseStr_AllBranches(t *testing.T) {
	p := new(RedDogCuiPresenter)
	for phase, expect := range map[int]string{
		domain.RedDogPhaseBet:            "BET",
		domain.RedDogPhaseInitialDealt:   "INITIAL DEALT",
		domain.RedDogPhaseSpreadDecision: "SPREAD DECISION",
		domain.RedDogPhasePairThird:      "PAIR THIRD",
		domain.RedDogPhaseEnd:            "END",
		999:                              "UNKNOWN",
	} {
		assert.Equal(t, expect, p.phaseStr(phase))
	}
}

func TestRedDogCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(RedDogCuiPresenter)
	m := new(interfaces.MockRedDogGame)
	m.On("GetGameEndFlag").Return(false)
	result := p.ActionLogOutput(m)
	assert.NotEmpty(t, result)
}
