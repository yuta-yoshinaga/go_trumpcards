package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockMonteCarloInteractor() *mockusecase.MockMonteCarloInteractor {
	return new(mockusecase.MockMonteCarloInteractor)
}

func TestMonteCarloCuiController_Quit(t *testing.T) {
	mi := newMockMonteCarloInteractor()
	c := NewMonteCarloCuiController(mi)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestMonteCarloCuiController_Reset(t *testing.T) {
	mi := newMockMonteCarloInteractor()
	c := NewMonteCarloCuiController(mi)
	mi.On("Reset").Return("reset_output")
	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestMonteCarloCuiController_Remove(t *testing.T) {
	mi := newMockMonteCarloInteractor()
	c := NewMonteCarloCuiController(mi)
	mi.On("Remove", 0, 1, 1, 2).Return("remove_output")
	assert.Equal(t, "remove_output", c.Exec("m 0 1 1 2"))
	assert.Equal(t, "remove_output", c.Exec("move 0 1 1 2"))
	assert.Equal(t, "remove_output", c.Exec("remove 0 1 1 2"))
}

func TestMonteCarloCuiController_RemoveInvalid(t *testing.T) {
	mi := newMockMonteCarloInteractor()
	c := NewMonteCarloCuiController(mi)

	assert.Contains(t, c.Exec("m"), "Usage:")
	assert.Contains(t, c.Exec("m 0 1 1"), "Usage:")
	assert.True(t, msgRejected(c.Exec("m abc 1 2 3")))
	assert.True(t, msgRejected(c.Exec("m 0 1 2 zzz")))
}

func TestMonteCarloCuiController_Deal(t *testing.T) {
	mi := newMockMonteCarloInteractor()
	c := NewMonteCarloCuiController(mi)
	mi.On("Deal").Return("deal_output")
	assert.Equal(t, "deal_output", c.Exec("d"))
	assert.Equal(t, "deal_output", c.Exec("deal"))
}

func TestMonteCarloCuiController_Undo(t *testing.T) {
	mi := newMockMonteCarloInteractor()
	c := NewMonteCarloCuiController(mi)
	mi.On("Undo").Return("undo_output")
	assert.Equal(t, "undo_output", c.Exec("u"))
	assert.Equal(t, "undo_output", c.Exec("undo"))
}

func TestMonteCarloCuiController_GiveUp(t *testing.T) {
	mi := newMockMonteCarloInteractor()
	c := NewMonteCarloCuiController(mi)
	mi.On("GiveUp").Return("giveup_output")
	assert.Equal(t, "giveup_output", c.Exec("g"))
	assert.Equal(t, "giveup_output", c.Exec("giveup"))
}

func TestMonteCarloCuiController_Hint(t *testing.T) {
	mi := newMockMonteCarloInteractor()
	c := NewMonteCarloCuiController(mi)
	mi.On("Hint").Return("hint_output")
	assert.Equal(t, "hint_output", c.Exec("h"))
	assert.Equal(t, "hint_output", c.Exec("hint"))
}

func TestMonteCarloCuiController_ActionLog(t *testing.T) {
	mi := newMockMonteCarloInteractor()
	c := NewMonteCarloCuiController(mi)
	mi.On("ActionLog").Return("log_output")
	assert.Equal(t, "log_output", c.Exec("l"))
	assert.Equal(t, "log_output", c.Exec("log"))
}
