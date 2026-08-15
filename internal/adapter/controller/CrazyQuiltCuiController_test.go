package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockCrazyQuiltInteractor() *mockusecase.MockCrazyQuiltInteractor {
	return new(mockusecase.MockCrazyQuiltInteractor)
}

func TestCrazyQuiltCuiControllerSimpleCommands(t *testing.T) {
	for _, tc := range []struct {
		method  string
		aliases []string
	}{
		{"Reset", []string{"r", "reset"}},
		{"Draw", []string{"d", "draw"}},
		{"GiveUp", []string{"g", "giveup"}},
		{"Hint", []string{"h", "hint"}},
		{"AutoComplete", []string{"ac", "autocomplete"}},
		{"ActionLog", []string{"log", "l"}},
		{"Undo", []string{"u", "undo"}},
	} {
		t.Run(tc.method, func(t *testing.T) {
			ci := newMockCrazyQuiltInteractor()
			c := NewCrazyQuiltCuiController(ci)
			ci.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

func TestCrazyQuiltCuiControllerQuit(t *testing.T) {
	c := NewCrazyQuiltCuiController(newMockCrazyQuiltInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestCrazyQuiltCuiControllerMoves(t *testing.T) {
	t.Run("quilt to a foundation", func(t *testing.T) {
		ci := newMockCrazyQuiltInteractor()
		c := NewCrazyQuiltCuiController(ci)
		ci.On("MoveQuiltToFoundation", 2).Return("qf")
		assert.Equal(t, "qf", c.Exec("m q 2 f"))
	})

	// キルトを崩す主要な手。捨て札と数字が 1 つ違いなら置ける。
	t.Run("quilt onto the waste", func(t *testing.T) {
		ci := newMockCrazyQuiltInteractor()
		c := NewCrazyQuiltCuiController(ci)
		ci.On("MoveQuiltToWaste", 5).Return("qw")
		assert.Equal(t, "qw", c.Exec("m q 5 w"))
	})

	t.Run("waste to a foundation", func(t *testing.T) {
		ci := newMockCrazyQuiltInteractor()
		c := NewCrazyQuiltCuiController(ci)
		ci.On("MoveWasteToFoundation").Return("wf")
		assert.Equal(t, "wf", c.Exec("m w f"))
	})

	// 捨て札の行き先は組札しかないので、ゾーンを省いても同じ手になる。
	t.Run("waste without a destination zone", func(t *testing.T) {
		ci := newMockCrazyQuiltInteractor()
		c := NewCrazyQuiltCuiController(ci)
		ci.On("MoveWasteToFoundation").Return("wf")
		assert.Equal(t, "wf", c.Exec("m w"))
	})

	// キルトは移動先にならない（崩す一方）。
	t.Run("rejects the quilt as a destination", func(t *testing.T) {
		ci := newMockCrazyQuiltInteractor()
		c := NewCrazyQuiltCuiController(ci)
		assert.Equal(t, invalidArg("crazyquilt.invalidToZone", "val", "q"), c.Exec("m q 2 q"))
		ci.AssertExpectations(t)
	})

	t.Run("rejects an unknown source zone", func(t *testing.T) {
		ci := newMockCrazyQuiltInteractor()
		c := NewCrazyQuiltCuiController(ci)
		assert.Equal(t, invalidArg("crazyquilt.invalidFromZone", "val", "s"), c.Exec("m s t 3"))
		ci.AssertExpectations(t)
	})
}

func TestCrazyQuiltCuiControllerPrompts(t *testing.T) {
	for _, cmd := range []string{"m", "m q", "m q 2"} {
		t.Run(cmd, func(t *testing.T) {
			c := NewCrazyQuiltCuiController(newMockCrazyQuiltInteractor())
			assert.True(t, cuiutil.IsPromptRequest(c.Exec(cmd)), "%q should prompt for more input", cmd)
		})
	}
}

func TestCrazyQuiltCuiControllerErrors(t *testing.T) {
	for _, tc := range []struct{ cmd, contains string }{
		{"m x f", "x"},
		{"m q abc f", "abc"},
		{"m q 0 z", "z"},
		{"m w z", "z"},
	} {
		t.Run(tc.cmd, func(t *testing.T) {
			c := NewCrazyQuiltCuiController(newMockCrazyQuiltInteractor())
			assert.Contains(t, c.Exec(tc.cmd), tc.contains)
		})
	}
}
