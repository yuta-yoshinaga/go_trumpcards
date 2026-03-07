package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mockUsecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// --- quit ---

func TestPokerCuiController_Quit(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

// --- reset ---

func TestPokerCuiController_Reset(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Reset").Return("reset ok")
	assert.Equal(t, "reset ok", c.Exec("r"))
	assert.Equal(t, "reset ok", c.Exec("reset"))
}

// --- exchange ---

func TestPokerCuiController_Exchange_WithValidIndices(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Exchange", []int{0, 2, 4}).Return("exchange ok")
	assert.Equal(t, "exchange ok", c.Exec("e 0 2 4"))
}

func TestPokerCuiController_Exchange_LongCommand(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Exchange", []int{1, 3}).Return("exchange ok")
	assert.Equal(t, "exchange ok", c.Exec("exchange 1 3"))
}

func TestPokerCuiController_Exchange_NoIndices(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Exchange", []int{}).Return("exchange empty")
	assert.Equal(t, "exchange empty", c.Exec("e"))
}

func TestPokerCuiController_Exchange_InvalidIndex_NonNumeric(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	// "abc" fails strconv.Atoi => err != nil => skipped; "2" is valid
	mi.On("Exchange", []int{2}).Return("exchange ok")
	assert.Equal(t, "exchange ok", c.Exec("e abc 2"))
}

func TestPokerCuiController_Exchange_InvalidIndex_OutOfRange_Negative(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	// -1 fails 0 <= idx check => skipped; "3" is valid
	mi.On("Exchange", []int{3}).Return("exchange ok")
	assert.Equal(t, "exchange ok", c.Exec("e -1 3"))
}

func TestPokerCuiController_Exchange_InvalidIndex_OutOfRange_Above(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	// 5 fails idx <= 4 check => skipped; "0" is valid
	mi.On("Exchange", []int{0}).Return("exchange ok")
	assert.Equal(t, "exchange ok", c.Exec("e 5 0"))
}

func TestPokerCuiController_Exchange_AllInvalid(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	// All indices invalid => empty slice
	mi.On("Exchange", []int{}).Return("exchange empty")
	assert.Equal(t, "exchange empty", c.Exec("e abc -1 5 99"))
}

// --- stand ---

func TestPokerCuiController_Stand(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Stand").Return("stand ok")
	assert.Equal(t, "stand ok", c.Exec("s"))
	assert.Equal(t, "stand ok", c.Exec("stand"))
}

// --- bet ---

func TestPokerCuiController_Bet_WithValidAmount(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Action", domain.PokerActionBet, 50).Return("bet ok")
	assert.Equal(t, "bet ok", c.Exec("b 50"))
	assert.Equal(t, "bet ok", c.Exec("bet 50"))
}

func TestPokerCuiController_Bet_NoAmount(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	// No amount => amount stays 0
	mi.On("Action", domain.PokerActionBet, 0).Return("bet zero")
	assert.Equal(t, "bet zero", c.Exec("b"))
}

func TestPokerCuiController_Bet_InvalidAmount_NonNumeric(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	// "abc" fails strconv.Atoi => amount stays 0
	mi.On("Action", domain.PokerActionBet, 0).Return("bet zero")
	assert.Equal(t, "bet zero", c.Exec("b abc"))
}

func TestPokerCuiController_Bet_InvalidAmount_Negative(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	// -20 succeeds Atoi but fails a > 0 check => amount stays 0
	mi.On("Action", domain.PokerActionBet, 0).Return("bet zero")
	assert.Equal(t, "bet zero", c.Exec("b -20"))
}

func TestPokerCuiController_Bet_InvalidAmount_Zero(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	// 0 succeeds Atoi but fails a > 0 check => amount stays 0
	mi.On("Action", domain.PokerActionBet, 0).Return("bet zero")
	assert.Equal(t, "bet zero", c.Exec("b 0"))
}

// --- call ---

func TestPokerCuiController_Call(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Action", domain.PokerActionCall, 0).Return("call ok")
	assert.Equal(t, "call ok", c.Exec("c"))
	assert.Equal(t, "call ok", c.Exec("call"))
}

// --- raise ---

func TestPokerCuiController_Raise_WithValidAmount(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Action", domain.PokerActionRaise, 30).Return("raise ok")
	assert.Equal(t, "raise ok", c.Exec("ra 30"))
	assert.Equal(t, "raise ok", c.Exec("raise 30"))
}

func TestPokerCuiController_Raise_NoAmount(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Action", domain.PokerActionRaise, 0).Return("raise zero")
	assert.Equal(t, "raise zero", c.Exec("ra"))
}

func TestPokerCuiController_Raise_InvalidAmount_NonNumeric(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Action", domain.PokerActionRaise, 0).Return("raise zero")
	assert.Equal(t, "raise zero", c.Exec("ra abc"))
}

func TestPokerCuiController_Raise_InvalidAmount_Negative(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Action", domain.PokerActionRaise, 0).Return("raise zero")
	assert.Equal(t, "raise zero", c.Exec("ra -30"))
}

func TestPokerCuiController_Raise_InvalidAmount_Zero(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Action", domain.PokerActionRaise, 0).Return("raise zero")
	assert.Equal(t, "raise zero", c.Exec("ra 0"))
}

// --- fold ---

func TestPokerCuiController_Fold(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Action", domain.PokerActionFold, 0).Return("fold ok")
	assert.Equal(t, "fold ok", c.Exec("f"))
	assert.Equal(t, "fold ok", c.Exec("fold"))
}

// --- check ---

func TestPokerCuiController_Check(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Action", domain.PokerActionCheck, 0).Return("check ok")
	assert.Equal(t, "check ok", c.Exec("ck"))
	assert.Equal(t, "check ok", c.Exec("check"))
}

// --- allin ---

func TestPokerCuiController_AllIn(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	mi.On("Action", domain.PokerActionAllIn, 0).Return("allin ok")
	assert.Equal(t, "allin ok", c.Exec("a"))
	assert.Equal(t, "allin ok", c.Exec("allin"))
}

// --- empty / unknown ---

func TestPokerCuiController_Empty(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	assert.Equal(t, "コマンドが不明です: ", c.Exec(""))
}

func TestPokerCuiController_Unknown(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	assert.Equal(t, "コマンドが不明です: xyz", c.Exec("xyz"))
}

// --- betting limit ---

func TestPokerCuiController_BettingLimit_Valid(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	cfg := domain.DefaultPokerConfig()
	cfg.BettingLimit = domain.BettingLimitPotLimit
	mi.On("ResetWithConfig", cfg).Return("bl ok")
	assert.Equal(t, "bl ok", c.Exec("bl 1"))
}

func TestPokerCuiController_BettingLimit_LongCommand(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	cfg := domain.DefaultPokerConfig()
	cfg.BettingLimit = domain.BettingLimitNoLimit
	mi.On("ResetWithConfig", cfg).Return("bl ok")
	assert.Equal(t, "bl ok", c.Exec("bettinglimit 2"))
}

func TestPokerCuiController_BettingLimit_NoArgs(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	assert.Contains(t, c.Exec("bl"), "Betting limit type is required")
}

func TestPokerCuiController_BettingLimit_InvalidValue(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerCuiController(mi)
	assert.Contains(t, c.Exec("bl 5"), "Invalid betting limit: 5")
	assert.Contains(t, c.Exec("bl abc"), "Invalid betting limit: abc")
	assert.Contains(t, c.Exec("bl -1"), "Invalid betting limit: -1")
}
