package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockSpiteAndMaliceInteractor() *mockusecase.MockSpiteAndMaliceInteractor {
	return new(mockusecase.MockSpiteAndMaliceInteractor)
}

func TestSpiteAndMaliceCuiController_Quit(t *testing.T) {
	si := newMockSpiteAndMaliceInteractor()
	c := NewSpiteAndMaliceCuiController(si)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestSpiteAndMaliceCuiController_Reset(t *testing.T) {
	si := newMockSpiteAndMaliceInteractor()
	c := NewSpiteAndMaliceCuiController(si)
	si.On("Reset").Return("reset_out")
	assert.Equal(t, "reset_out", c.Exec("r"))
}

func TestSpiteAndMaliceCuiController_PlayFromHand(t *testing.T) {
	si := newMockSpiteAndMaliceInteractor()
	c := NewSpiteAndMaliceCuiController(si)
	si.On("PlayFromHand", 2, 1).Return("ok")
	assert.Equal(t, "ok", c.Exec("ph 2 1"))
}

func TestSpiteAndMaliceCuiController_PlayFromGoal(t *testing.T) {
	si := newMockSpiteAndMaliceInteractor()
	c := NewSpiteAndMaliceCuiController(si)
	si.On("PlayFromGoal", 3).Return("ok")
	assert.Equal(t, "ok", c.Exec("pg 3"))
}

func TestSpiteAndMaliceCuiController_PlayFromSide(t *testing.T) {
	si := newMockSpiteAndMaliceInteractor()
	c := NewSpiteAndMaliceCuiController(si)
	si.On("PlayFromSide", 0, 2).Return("ok")
	assert.Equal(t, "ok", c.Exec("ps 0 2"))
}

func TestSpiteAndMaliceCuiController_Discard(t *testing.T) {
	si := newMockSpiteAndMaliceInteractor()
	c := NewSpiteAndMaliceCuiController(si)
	si.On("Discard", 1, 2).Return("ok")
	assert.Equal(t, "ok", c.Exec("d 1 2"))
	assert.Equal(t, "ok", c.Exec("discard 1 2"))
}

func TestSpiteAndMaliceCuiController_CpuStep(t *testing.T) {
	si := newMockSpiteAndMaliceInteractor()
	c := NewSpiteAndMaliceCuiController(si)
	si.On("CpuStep").Return("cpu_out")
	assert.Equal(t, "cpu_out", c.Exec("cpu"))
}

func TestSpiteAndMaliceCuiController_Hint(t *testing.T) {
	si := newMockSpiteAndMaliceInteractor()
	c := NewSpiteAndMaliceCuiController(si)
	si.On("Hint").Return("hint_out")
	assert.Equal(t, "hint_out", c.Exec("h"))
	assert.Equal(t, "hint_out", c.Exec("hint"))
}

func TestSpiteAndMaliceCuiController_ActionLog(t *testing.T) {
	si := newMockSpiteAndMaliceInteractor()
	c := NewSpiteAndMaliceCuiController(si)
	si.On("ActionLog").Return("log_out")
	assert.Equal(t, "log_out", c.Exec("l"))
	assert.Equal(t, "log_out", c.Exec("log"))
}

func TestSpiteAndMaliceCuiController_Prompts(t *testing.T) {
	si := newMockSpiteAndMaliceInteractor()
	c := NewSpiteAndMaliceCuiController(si)
	assert.Contains(t, c.Exec("ph"), cuiutil.PromptPrefix)
	assert.Contains(t, c.Exec("ph 1"), cuiutil.PromptPrefix)
	assert.Contains(t, c.Exec("pg"), cuiutil.PromptPrefix)
	assert.Contains(t, c.Exec("ps"), cuiutil.PromptPrefix)
	assert.Contains(t, c.Exec("ps 0"), cuiutil.PromptPrefix)
	assert.Contains(t, c.Exec("d"), cuiutil.PromptPrefix)
	assert.Contains(t, c.Exec("d 1"), cuiutil.PromptPrefix)
}

func TestSpiteAndMaliceCuiController_InvalidArgs(t *testing.T) {
	si := newMockSpiteAndMaliceInteractor()
	c := NewSpiteAndMaliceCuiController(si)
	assert.NotEmpty(t, c.Exec("ph abc 1"))
	assert.NotEmpty(t, c.Exec("ph 0 abc"))
	assert.NotEmpty(t, c.Exec("pg abc"))
	assert.NotEmpty(t, c.Exec("ps abc 1"))
	assert.NotEmpty(t, c.Exec("ps 0 abc"))
	assert.NotEmpty(t, c.Exec("d abc 1"))
	assert.NotEmpty(t, c.Exec("d 0 abc"))
}

func TestSpiteAndMaliceCuiController_Unknown(t *testing.T) {
	si := newMockSpiteAndMaliceInteractor()
	c := NewSpiteAndMaliceCuiController(si)
	assert.NotEmpty(t, c.Exec("unknowncmd"))
	assert.NotEmpty(t, c.Exec(""))
}
