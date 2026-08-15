package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockPokerSquaresInteractor() *mockusecase.MockPokerSquaresInteractor {
	return new(mockusecase.MockPokerSquaresInteractor)
}

func TestPokerSquaresCuiController_Quit(t *testing.T) {
	pi := newMockPokerSquaresInteractor()
	c := NewPokerSquaresCuiController(pi)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestPokerSquaresCuiController_Reset(t *testing.T) {
	pi := newMockPokerSquaresInteractor()
	c := NewPokerSquaresCuiController(pi)
	pi.On("Reset").Return("reset_output")
	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestPokerSquaresCuiController_Place(t *testing.T) {
	pi := newMockPokerSquaresInteractor()
	c := NewPokerSquaresCuiController(pi)
	pi.On("Place", 1, 2).Return("place_output")
	assert.Equal(t, "place_output", c.Exec("p 1 2"))
	assert.Equal(t, "place_output", c.Exec("place 1 2"))
}

func TestPokerSquaresCuiController_PlaceInvalid(t *testing.T) {
	pi := newMockPokerSquaresInteractor()
	c := NewPokerSquaresCuiController(pi)

	assert.Contains(t, c.Exec("p"), msgUsage("usagePRowCol"))
	assert.Contains(t, c.Exec("p 1"), msgUsage("usagePRowCol"))
	assert.True(t, msgRejected(c.Exec("p abc 0")))
	assert.True(t, msgRejected(c.Exec("p 0 abc")))
}

func TestPokerSquaresCuiController_Undo(t *testing.T) {
	pi := newMockPokerSquaresInteractor()
	c := NewPokerSquaresCuiController(pi)
	pi.On("Undo").Return("undo_output")
	assert.Equal(t, "undo_output", c.Exec("u"))
	assert.Equal(t, "undo_output", c.Exec("undo"))
}

func TestPokerSquaresCuiController_GiveUp(t *testing.T) {
	pi := newMockPokerSquaresInteractor()
	c := NewPokerSquaresCuiController(pi)
	pi.On("GiveUp").Return("giveup_output")
	assert.Equal(t, "giveup_output", c.Exec("g"))
	assert.Equal(t, "giveup_output", c.Exec("giveup"))
}

func TestPokerSquaresCuiController_Hint(t *testing.T) {
	pi := newMockPokerSquaresInteractor()
	c := NewPokerSquaresCuiController(pi)
	pi.On("Hint").Return("hint_output")
	assert.Equal(t, "hint_output", c.Exec("h"))
	assert.Equal(t, "hint_output", c.Exec("hint"))
}

func TestPokerSquaresCuiController_ActionLog(t *testing.T) {
	pi := newMockPokerSquaresInteractor()
	c := NewPokerSquaresCuiController(pi)
	pi.On("ActionLog").Return("log_output")
	assert.Equal(t, "log_output", c.Exec("l"))
	assert.Equal(t, "log_output", c.Exec("log"))
}
