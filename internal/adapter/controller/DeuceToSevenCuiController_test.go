package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mockUsecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newDeuceToSevenMockInteractor() *mockUsecase.MockDeuceToSevenInteractor {
	return new(mockUsecase.MockDeuceToSevenInteractor)
}

func TestDeuceToSevenCuiController_Hint(t *testing.T) {
	mi := newDeuceToSevenMockInteractor()
	mi.On("Hint").Return("hint ok")
	c := NewDeuceToSevenCuiController(mi)
	assert.Equal(t, "hint ok", c.Exec("h"))
	assert.Equal(t, "hint ok", c.Exec("hint"))
	mi.AssertCalled(t, "Hint")
}

func TestDeuceToSevenCuiController_Quit(t *testing.T) {
	c := NewDeuceToSevenCuiController(newDeuceToSevenMockInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestDeuceToSevenCuiController_Reset(t *testing.T) {
	mi := newDeuceToSevenMockInteractor()
	c := NewDeuceToSevenCuiController(mi)
	mi.On("Reset").Return("reset ok")
	assert.Equal(t, "reset ok", c.Exec("r"))
	assert.Equal(t, "reset ok", c.Exec("reset"))
}

func TestDeuceToSevenCuiController_Exchange_ValidIndices(t *testing.T) {
	mi := newDeuceToSevenMockInteractor()
	c := NewDeuceToSevenCuiController(mi)
	mi.On("Exchange", []int{0, 2, 4}).Return("exchange ok")
	assert.Equal(t, "exchange ok", c.Exec("e 0 2 4"))
}

func TestDeuceToSevenCuiController_Exchange_OutOfRangeSkipped(t *testing.T) {
	mi := newDeuceToSevenMockInteractor()
	c := NewDeuceToSevenCuiController(mi)
	// 2-7 uses 5 cards, valid range 0..4. 5 is skipped, 1 is valid.
	result := c.Exec("e 5 1")
	assert.Contains(t, result, "5")
	mi.AssertNotCalled(t, "Exchange", mock.Anything)
}

func TestDeuceToSevenCuiController_Exchange_NoIndices(t *testing.T) {
	mi := newDeuceToSevenMockInteractor()
	c := NewDeuceToSevenCuiController(mi)
	mi.On("Exchange", []int{}).Return("exchange empty")
	assert.Equal(t, "exchange empty", c.Exec("e"))
}

func TestDeuceToSevenCuiController_Stand(t *testing.T) {
	mi := newDeuceToSevenMockInteractor()
	c := NewDeuceToSevenCuiController(mi)
	mi.On("Stand").Return("stand ok")
	assert.Equal(t, "stand ok", c.Exec("s"))
	assert.Equal(t, "stand ok", c.Exec("stand"))
}

func TestDeuceToSevenCuiController_BettingActions(t *testing.T) {
	tests := []struct {
		name    string
		command string
		action  int
		amount  int
	}{
		{"bet", "b 20", domain.DeuceToSevenActionBet, 20},
		{"call", "c", domain.DeuceToSevenActionCall, 0},
		{"raise", "ra 40", domain.DeuceToSevenActionRaise, 40},
		{"fold", "f", domain.DeuceToSevenActionFold, 0},
		{"check", "ck", domain.DeuceToSevenActionCheck, 0},
		{"allin", "a", domain.DeuceToSevenActionAllIn, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mi := newDeuceToSevenMockInteractor()
			c := NewDeuceToSevenCuiController(mi)
			mi.On("Action", tt.action, tt.amount, 0).Return("ok")
			assert.Equal(t, "ok", c.Exec(tt.command))
		})
	}
}

func TestDeuceToSevenCuiController_BetNoAmount(t *testing.T) {
	mi := newDeuceToSevenMockInteractor()
	c := NewDeuceToSevenCuiController(mi)
	mi.On("Action", domain.DeuceToSevenActionBet, 0, 0).Return("bet placeholder")
	assert.Equal(t, "bet placeholder", c.Exec("b"))
}

func TestDeuceToSevenCuiController_BetInvalidAmount(t *testing.T) {
	mi := newDeuceToSevenMockInteractor()
	c := NewDeuceToSevenCuiController(mi)
	mi.On("Action", domain.DeuceToSevenActionBet, 0, 0).Return("bet ok")
	assert.Equal(t, "bet ok", c.Exec("b abc"))
}

func TestDeuceToSevenCuiController_BettingLimit_NoArg(t *testing.T) {
	mi := newDeuceToSevenMockInteractor()
	c := NewDeuceToSevenCuiController(mi)
	result := c.Exec("bl")
	assert.Contains(t, result, "Betting limit type is required")
}

func TestDeuceToSevenCuiController_BettingLimit_Valid(t *testing.T) {
	mi := newDeuceToSevenMockInteractor()
	c := NewDeuceToSevenCuiController(mi)
	cfg := domain.DefaultDeuceToSevenConfig()
	mi.On("GetConfig").Return(cfg)
	wanted := cfg
	wanted.BettingLimit = domain.BettingLimitNoLimit
	mi.On("ResetWithConfig", wanted, mock.Anything).Return("ok")
	assert.Equal(t, "ok", c.Exec("bl 2"))
}

func TestDeuceToSevenCuiController_BettingLimit_Invalid(t *testing.T) {
	mi := newDeuceToSevenMockInteractor()
	c := NewDeuceToSevenCuiController(mi)
	result := c.Exec("bl 7")
	assert.Contains(t, result, "Invalid betting limit")
}

func TestDeuceToSevenCuiController_SetCpuCount(t *testing.T) {
	mi := newDeuceToSevenMockInteractor()
	c := NewDeuceToSevenCuiController(mi)
	cfg := domain.DefaultDeuceToSevenConfig()
	mi.On("GetConfig").Return(cfg)
	wanted := cfg
	wanted.CpuCount = 2
	mi.On("ResetWithConfig", wanted, mock.Anything).Return("ok")
	assert.Equal(t, "ok", c.Exec("scc 2"))
}

func TestDeuceToSevenCuiController_MetaAI_Toggle(t *testing.T) {
	mi := newDeuceToSevenMockInteractor()
	c := NewDeuceToSevenCuiController(mi)
	cfg := domain.DefaultDeuceToSevenConfig()
	mi.On("GetConfig").Return(cfg)
	wanted := cfg
	wanted.CpuMetaAI = true
	mi.On("ResetWithConfig", wanted, mock.Anything).Return("ok")
	assert.Equal(t, "ok", c.Exec("mai 1"))
}

func TestDeuceToSevenCuiController_MetaAI_NoArg(t *testing.T) {
	mi := newDeuceToSevenMockInteractor()
	c := NewDeuceToSevenCuiController(mi)
	assert.Equal(t, i18n.T("metaAIRequired"), c.Exec("mai"))
}

func TestDeuceToSevenCuiController_MetaAI_Invalid(t *testing.T) {
	mi := newDeuceToSevenMockInteractor()
	c := NewDeuceToSevenCuiController(mi)
	result := c.Exec("mai abc")
	assert.Equal(t, invalidArg("invalidMetaAI", "val", "abc"), result)
}

func TestDeuceToSevenCuiController_Log(t *testing.T) {
	mi := newDeuceToSevenMockInteractor()
	c := NewDeuceToSevenCuiController(mi)
	mi.On("ActionLog").Return("log entries")
	assert.Equal(t, "log entries", c.Exec("log"))
	assert.Equal(t, "log entries", c.Exec("l"))
}

func TestDeuceToSevenCuiController_UnknownCommand(t *testing.T) {
	c := NewDeuceToSevenCuiController(newDeuceToSevenMockInteractor())
	result := c.Exec("nonsense")
	assert.NotEmpty(t, result)
}
