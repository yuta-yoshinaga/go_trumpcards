//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockCrazyFourPokerInteractor() *usecase.MockCrazyFourPokerInteractor {
	m := new(usecase.MockCrazyFourPokerInteractor)
	m.On("Reset").Return("reset result")
	m.On("PlaceBet", 50, 20).Return("bet 50/20")
	m.On("PlaceBet", 50, 0).Return("bet 50/0")
	m.On("Play", 1).Return("played x1")
	m.On("Play", 3).Return("played x3")
	m.On("Fold").Return("folded")
	m.On("NextRound").Return("next round")
	m.On("Hint").Return("hint result")
	m.On("ActionLog").Return("action log result")
	return m
}

func TestCrazyFourPokerCuiController_QuitAndReset(t *testing.T) {
	c := controller.NewCrazyFourPokerCuiController(newMockCrazyFourPokerInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "reset result", c.Exec("r"))
}

// **Queens Up は省略できる。** 省略は 0 (置かない)。
func TestCrazyFourPokerCuiController_Bet(t *testing.T) {
	m := newMockCrazyFourPokerInteractor()
	c := controller.NewCrazyFourPokerCuiController(m)

	assert.Equal(t, "bet 50/20", c.Exec("bet 50 20"))
	assert.Equal(t, "bet 50/0", c.Exec("bet 50"))
	m.AssertCalled(t, "PlaceBet", 50, 0)

	assert.Contains(t, c.Exec("bet"), "required")
	assert.Contains(t, c.Exec("bet xyz"), "Invalid")
	assert.Contains(t, c.Exec("bet 50 xyz"), "Invalid")
}

func TestCrazyFourPokerCuiController_Play(t *testing.T) {
	m := newMockCrazyFourPokerInteractor()
	c := controller.NewCrazyFourPokerCuiController(m)

	assert.Equal(t, "played x1", c.Exec("play 1"))
	assert.Equal(t, "played x3", c.Exec("play 3"))
	assert.Contains(t, c.Exec("play"), "required")
	assert.Contains(t, c.Exec("play xyz"), "Invalid")
}

func TestCrazyFourPokerCuiController_RemainingCommands(t *testing.T) {
	c := controller.NewCrazyFourPokerCuiController(newMockCrazyFourPokerInteractor())
	assert.Equal(t, "folded", c.Exec("fold"))
	assert.Equal(t, "folded", c.Exec("f"))
	assert.Equal(t, "next round", c.Exec("next"))
	assert.Equal(t, "hint result", c.Exec("hint"))
	assert.Equal(t, "action log result", c.Exec("log"))
}

func TestCrazyFourPokerCuiController_UnknownCommand(t *testing.T) {
	c := controller.NewCrazyFourPokerCuiController(newMockCrazyFourPokerInteractor())
	assert.NotEmpty(t, c.Exec("zzz"))
	assert.NotEmpty(t, c.Exec(""))
}
