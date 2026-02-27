package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestHoldemCuiController_Quit(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestHoldemCuiController_Reset(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("Reset").Return("reset ok")
	assert.Equal(t, "reset ok", c.Exec("r"))
	assert.Equal(t, "reset ok", c.Exec("reset"))
}

func TestHoldemCuiController_Fold(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("Action", domain.HoldemActionFold, 0).Return("fold ok")
	assert.Equal(t, "fold ok", c.Exec("f"))
	assert.Equal(t, "fold ok", c.Exec("fold"))
}

func TestHoldemCuiController_Check(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("Action", domain.HoldemActionCheck, 0).Return("check ok")
	assert.Equal(t, "check ok", c.Exec("ck"))
	assert.Equal(t, "check ok", c.Exec("check"))
}

func TestHoldemCuiController_Call(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("Action", domain.HoldemActionCall, 0).Return("call ok")
	assert.Equal(t, "call ok", c.Exec("c"))
	assert.Equal(t, "call ok", c.Exec("call"))
}

func TestHoldemCuiController_Bet(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("Action", domain.HoldemActionBet, 50).Return("bet ok")
	assert.Equal(t, "bet ok", c.Exec("b 50"))
	assert.Equal(t, "bet ok", c.Exec("bet 50"))
}

func TestHoldemCuiController_Bet_NoAmount(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("Action", domain.HoldemActionBet, 0).Return("bet ok")
	assert.Equal(t, "bet ok", c.Exec("b"))
}

func TestHoldemCuiController_Raise(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("Action", domain.HoldemActionRaise, 30).Return("raise ok")
	assert.Equal(t, "raise ok", c.Exec("ra 30"))
	assert.Equal(t, "raise ok", c.Exec("raise 30"))
}

func TestHoldemCuiController_Raise_NoAmount(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("Action", domain.HoldemActionRaise, 0).Return("raise ok")
	assert.Equal(t, "raise ok", c.Exec("ra"))
}

func TestHoldemCuiController_AllIn(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	mi.On("Action", domain.HoldemActionAllIn, 0).Return("allin ok")
	assert.Equal(t, "allin ok", c.Exec("a"))
	assert.Equal(t, "allin ok", c.Exec("allin"))
}

func TestHoldemCuiController_Empty(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	assert.Contains(t, c.Exec(""), "コマンドが不明です")
}

func TestHoldemCuiController_Unknown(t *testing.T) {
	mi := new(usecase.MockHoldemInteractor)
	c := NewHoldemCuiController(mi)
	assert.Contains(t, c.Exec("xyz"), "コマンドが不明です")
}
