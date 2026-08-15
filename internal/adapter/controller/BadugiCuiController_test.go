package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mockUsecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newBadugiMockInteractor() *mockUsecase.MockBadugiInteractor {
	return new(mockUsecase.MockBadugiInteractor)
}

// --- quit ---

func TestBadugiCuiController_Quit(t *testing.T) {
	c := NewBadugiCuiController(newBadugiMockInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

// --- reset ---

func TestBadugiCuiController_Reset(t *testing.T) {
	mi := newBadugiMockInteractor()
	c := NewBadugiCuiController(mi)
	mi.On("Reset").Return("reset ok")
	assert.Equal(t, "reset ok", c.Exec("r"))
	assert.Equal(t, "reset ok", c.Exec("reset"))
}

// --- exchange / stand ---

func TestBadugiCuiController_Exchange_ValidIndices(t *testing.T) {
	mi := newBadugiMockInteractor()
	c := NewBadugiCuiController(mi)
	mi.On("Exchange", []int{0, 2, 3}).Return("exchange ok")
	assert.Equal(t, "exchange ok", c.Exec("e 0 2 3"))
}

func TestBadugiCuiController_Exchange_OutOfRangeSkipped(t *testing.T) {
	mi := newBadugiMockInteractor()
	c := NewBadugiCuiController(mi)
	// Badugi uses 4 cards, valid range 0..3. 4 is skipped, 1 is valid.
	result := c.Exec("e 4 1")
	assert.Contains(t, result, "4")
	mi.AssertNotCalled(t, "Exchange", mock.Anything)
}

func TestBadugiCuiController_Exchange_NoIndices(t *testing.T) {
	mi := newBadugiMockInteractor()
	c := NewBadugiCuiController(mi)
	mi.On("Exchange", []int{}).Return("exchange empty")
	assert.Equal(t, "exchange empty", c.Exec("e"))
}

func TestBadugiCuiController_Stand(t *testing.T) {
	mi := newBadugiMockInteractor()
	c := NewBadugiCuiController(mi)
	mi.On("Stand").Return("stand ok")
	assert.Equal(t, "stand ok", c.Exec("s"))
	assert.Equal(t, "stand ok", c.Exec("stand"))
}

// --- betting actions ---

func TestBadugiCuiController_BettingActions(t *testing.T) {
	tests := []struct {
		name    string
		command string
		action  int
		amount  int
	}{
		{"bet", "b 20", domain.BadugiActionBet, 20},
		{"call", "c", domain.BadugiActionCall, 0},
		{"raise", "ra 40", domain.BadugiActionRaise, 40},
		{"fold", "f", domain.BadugiActionFold, 0},
		{"check", "ck", domain.BadugiActionCheck, 0},
		{"allin", "a", domain.BadugiActionAllIn, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mi := newBadugiMockInteractor()
			c := NewBadugiCuiController(mi)
			mi.On("Action", tt.action, tt.amount, 0).Return("ok")
			assert.Equal(t, "ok", c.Exec(tt.command))
		})
	}
}

func TestBadugiCuiController_BetNoAmount(t *testing.T) {
	mi := newBadugiMockInteractor()
	c := NewBadugiCuiController(mi)
	// Missing amount → defaults to 0, interactor decides fallback.
	mi.On("Action", domain.BadugiActionBet, 0, 0).Return("bet placeholder")
	assert.Equal(t, "bet placeholder", c.Exec("b"))
}

func TestBadugiCuiController_BetInvalidAmount(t *testing.T) {
	mi := newBadugiMockInteractor()
	c := NewBadugiCuiController(mi)
	mi.On("Action", domain.BadugiActionBet, 0, 0).Return("bet ok")
	assert.Equal(t, "bet ok", c.Exec("b abc"))
}

// --- config commands ---

func TestBadugiCuiController_BettingLimit_NoArg(t *testing.T) {
	mi := newBadugiMockInteractor()
	c := NewBadugiCuiController(mi)
	result := c.Exec("bl")
	assert.Contains(t, result, "Betting limit type is required")
}

func TestBadugiCuiController_BettingLimit_Valid(t *testing.T) {
	mi := newBadugiMockInteractor()
	c := NewBadugiCuiController(mi)
	cfg := domain.DefaultBadugiConfig()
	mi.On("GetConfig").Return(cfg)
	wanted := cfg
	wanted.BettingLimit = domain.BettingLimitNoLimit
	mi.On("ResetWithConfig", wanted, mock.Anything).Return("ok")
	assert.Equal(t, "ok", c.Exec("bl 2"))
}

func TestBadugiCuiController_BettingLimit_Invalid(t *testing.T) {
	mi := newBadugiMockInteractor()
	c := NewBadugiCuiController(mi)
	result := c.Exec("bl 7")
	assert.Contains(t, result, "Invalid betting limit")
}

func TestBadugiCuiController_SetCpuCount(t *testing.T) {
	mi := newBadugiMockInteractor()
	c := NewBadugiCuiController(mi)
	cfg := domain.DefaultBadugiConfig()
	mi.On("GetConfig").Return(cfg)
	wanted := cfg
	wanted.CpuCount = 2
	mi.On("ResetWithConfig", wanted, mock.Anything).Return("ok")
	assert.Equal(t, "ok", c.Exec("scc 2"))
}

func TestBadugiCuiController_MetaAI_Toggle(t *testing.T) {
	mi := newBadugiMockInteractor()
	c := NewBadugiCuiController(mi)
	cfg := domain.DefaultBadugiConfig()
	mi.On("GetConfig").Return(cfg)
	wanted := cfg
	wanted.CpuMetaAI = true
	mi.On("ResetWithConfig", wanted, mock.Anything).Return("ok")
	assert.Equal(t, "ok", c.Exec("mai 1"))
}

func TestBadugiCuiController_MetaAI_NoArg(t *testing.T) {
	mi := newBadugiMockInteractor()
	c := NewBadugiCuiController(mi)
	assert.Equal(t, i18n.T("metaAIRequired"), c.Exec("mai"))
}

func TestBadugiCuiController_MetaAI_Invalid(t *testing.T) {
	mi := newBadugiMockInteractor()
	c := NewBadugiCuiController(mi)
	result := c.Exec("mai abc")
	assert.Equal(t, invalidArg("invalidMetaAI", "val", "abc"), result)
}

// --- log ---

func TestBadugiCuiController_Log(t *testing.T) {
	mi := newBadugiMockInteractor()
	c := NewBadugiCuiController(mi)
	mi.On("ActionLog").Return("log entries")
	assert.Equal(t, "log entries", c.Exec("log"))
	assert.Equal(t, "log entries", c.Exec("l"))
}

func TestBadugiCuiController_UnknownCommand(t *testing.T) {
	c := NewBadugiCuiController(newBadugiMockInteractor())
	result := c.Exec("nonsense")
	assert.NotEmpty(t, result)
}
