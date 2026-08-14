package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockCribbageSquaresInteractor() *mockusecase.MockCribbageSquaresInteractor {
	return new(mockusecase.MockCribbageSquaresInteractor)
}

func TestCribbageSquaresCuiController_Quit(t *testing.T) {
	pi := newMockCribbageSquaresInteractor()
	c := NewCribbageSquaresCuiController(pi)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestCribbageSquaresCuiController_Reset(t *testing.T) {
	pi := newMockCribbageSquaresInteractor()
	c := NewCribbageSquaresCuiController(pi)
	pi.On("Reset").Return("reset_output")
	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestCribbageSquaresCuiController_Place(t *testing.T) {
	pi := newMockCribbageSquaresInteractor()
	c := NewCribbageSquaresCuiController(pi)
	pi.On("Place", 1, 2).Return("place_output")
	assert.Equal(t, "place_output", c.Exec("p 1 2"))
	assert.Equal(t, "place_output", c.Exec("place 1 2"))
}

func TestCribbageSquaresCuiController_PlaceInvalid(t *testing.T) {
	pi := newMockCribbageSquaresInteractor()
	c := NewCribbageSquaresCuiController(pi)

	assert.Contains(t, c.Exec("p"), "Usage:")
	assert.Contains(t, c.Exec("p 1"), "Usage:")
	assert.Contains(t, c.Exec("p abc 0"), "Invalid")
	assert.Contains(t, c.Exec("p 0 abc"), "Invalid")
}

func TestCribbageSquaresCuiController_Undo(t *testing.T) {
	pi := newMockCribbageSquaresInteractor()
	c := NewCribbageSquaresCuiController(pi)
	pi.On("Undo").Return("undo_output")
	assert.Equal(t, "undo_output", c.Exec("u"))
	assert.Equal(t, "undo_output", c.Exec("undo"))
}

func TestCribbageSquaresCuiController_GiveUp(t *testing.T) {
	pi := newMockCribbageSquaresInteractor()
	c := NewCribbageSquaresCuiController(pi)
	pi.On("GiveUp").Return("giveup_output")
	assert.Equal(t, "giveup_output", c.Exec("g"))
	assert.Equal(t, "giveup_output", c.Exec("giveup"))
}

func TestCribbageSquaresCuiController_Hint(t *testing.T) {
	pi := newMockCribbageSquaresInteractor()
	c := NewCribbageSquaresCuiController(pi)
	pi.On("Hint").Return("hint_output")
	assert.Equal(t, "hint_output", c.Exec("h"))
	assert.Equal(t, "hint_output", c.Exec("hint"))
}

func TestCribbageSquaresCuiController_ActionLog(t *testing.T) {
	pi := newMockCribbageSquaresInteractor()
	c := NewCribbageSquaresCuiController(pi)
	pi.On("ActionLog").Return("log_output")
	assert.Equal(t, "log_output", c.Exec("l"))
	assert.Equal(t, "log_output", c.Exec("log"))
}
