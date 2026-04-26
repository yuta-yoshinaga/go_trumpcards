package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockNertzInteractor() *mockusecase.MockNertzInteractor {
	return new(mockusecase.MockNertzInteractor)
}

func TestNertzCuiController_Quit(t *testing.T) {
	ni := newMockNertzInteractor()
	c := NewNertzCuiController(ni)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestNertzCuiController_Reset(t *testing.T) {
	ni := newMockNertzInteractor()
	c := NewNertzCuiController(ni)
	ni.On("Reset").Return("reset_out")
	assert.Equal(t, "reset_out", c.Exec("r"))
}

func TestNertzCuiController_Draw(t *testing.T) {
	ni := newMockNertzInteractor()
	c := NewNertzCuiController(ni)
	ni.On("Draw", 0).Return("ok")
	assert.Equal(t, "ok", c.Exec("d 0"))
	assert.Equal(t, "ok", c.Exec("draw 0"))
}

func TestNertzCuiController_MoveNF(t *testing.T) {
	ni := newMockNertzInteractor()
	c := NewNertzCuiController(ni)
	ni.On("MoveNertzToFoundation", 0, 1).Return("ok")
	assert.Equal(t, "ok", c.Exec("mnf 0 1"))
}

func TestNertzCuiController_MoveNT(t *testing.T) {
	ni := newMockNertzInteractor()
	c := NewNertzCuiController(ni)
	ni.On("MoveNertzToTableau", 0, 2).Return("ok")
	assert.Equal(t, "ok", c.Exec("mnt 0 2"))
}

func TestNertzCuiController_MoveWF(t *testing.T) {
	ni := newMockNertzInteractor()
	c := NewNertzCuiController(ni)
	ni.On("MoveWasteToFoundation", 0, 0).Return("ok")
	assert.Equal(t, "ok", c.Exec("mwf 0 0"))
}

func TestNertzCuiController_MoveWT(t *testing.T) {
	ni := newMockNertzInteractor()
	c := NewNertzCuiController(ni)
	ni.On("MoveWasteToTableau", 0, 3).Return("ok")
	assert.Equal(t, "ok", c.Exec("mwt 0 3"))
}

func TestNertzCuiController_MoveTF(t *testing.T) {
	ni := newMockNertzInteractor()
	c := NewNertzCuiController(ni)
	ni.On("MoveTableauToFoundation", 0, 1, 2).Return("ok")
	assert.Equal(t, "ok", c.Exec("mtf 0 1 2"))
}

func TestNertzCuiController_MoveTT(t *testing.T) {
	ni := newMockNertzInteractor()
	c := NewNertzCuiController(ni)
	ni.On("MoveTableauToTableau", 0, 1, 0, 2).Return("ok")
	assert.Equal(t, "ok", c.Exec("mtt 0 1 0 2"))
}

func TestNertzCuiController_Tick(t *testing.T) {
	ni := newMockNertzInteractor()
	c := NewNertzCuiController(ni)
	ni.On("Tick").Return("tick_out")
	assert.Equal(t, "tick_out", c.Exec("tick"))
}

func TestNertzCuiController_NextRound(t *testing.T) {
	ni := newMockNertzInteractor()
	c := NewNertzCuiController(ni)
	ni.On("NextRound").Return("nr_out")
	assert.Equal(t, "nr_out", c.Exec("nr"))
}

func TestNertzCuiController_Undo(t *testing.T) {
	ni := newMockNertzInteractor()
	c := NewNertzCuiController(ni)
	ni.On("Undo").Return("undo_out")
	assert.Equal(t, "undo_out", c.Exec("u"))
	assert.Equal(t, "undo_out", c.Exec("undo"))
}

func TestNertzCuiController_Hint(t *testing.T) {
	ni := newMockNertzInteractor()
	c := NewNertzCuiController(ni)
	ni.On("Hint").Return("hint_out")
	assert.Equal(t, "hint_out", c.Exec("h"))
	assert.Equal(t, "hint_out", c.Exec("hint"))
}

func TestNertzCuiController_ActionLog(t *testing.T) {
	ni := newMockNertzInteractor()
	c := NewNertzCuiController(ni)
	ni.On("ActionLog").Return("log_out")
	assert.Equal(t, "log_out", c.Exec("l"))
	assert.Equal(t, "log_out", c.Exec("log"))
}

func TestNertzCuiController_Prompts(t *testing.T) {
	ni := newMockNertzInteractor()
	c := NewNertzCuiController(ni)
	cases := []string{
		"d", "mnf", "mnf 0", "mnt", "mnt 0", "mwf", "mwf 0", "mwt", "mwt 0",
		"mtf", "mtf 0", "mtf 0 1", "mtt", "mtt 0", "mtt 0 1", "mtt 0 1 0",
	}
	for _, cmd := range cases {
		t.Run(cmd, func(t *testing.T) {
			assert.Contains(t, c.Exec(cmd), cuiutil.PromptPrefix)
		})
	}
}

func TestNertzCuiController_InvalidArgs(t *testing.T) {
	ni := newMockNertzInteractor()
	c := NewNertzCuiController(ni)
	cases := []string{
		"d abc",
		"mnf abc 0",
		"mnf 0 abc",
		"mnt abc 0",
		"mnt 0 abc",
		"mwf abc 0",
		"mwf 0 abc",
		"mwt abc 0",
		"mwt 0 abc",
		"mtf abc 0 0",
		"mtf 0 abc 0",
		"mtf 0 0 abc",
		"mtt abc 0 0 0",
		"mtt 0 abc 0 0",
		"mtt 0 0 abc 0",
		"mtt 0 0 0 abc",
	}
	for _, cmd := range cases {
		t.Run(cmd, func(t *testing.T) {
			assert.NotEmpty(t, c.Exec(cmd))
		})
	}
}

func TestNertzCuiController_UnknownCommand(t *testing.T) {
	ni := newMockNertzInteractor()
	c := NewNertzCuiController(ni)
	assert.NotEmpty(t, c.Exec("foobar"))
}
