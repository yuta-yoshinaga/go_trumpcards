package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockAgnesInteractor() *mockusecase.MockAgnesInteractor {
	return new(mockusecase.MockAgnesInteractor)
}

func TestAgnesCuiControllerQuit(t *testing.T) {
	ci := newMockAgnesInteractor()
	c := NewAgnesCuiController(ci)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestAgnesCuiControllerReset(t *testing.T) {
	ci := newMockAgnesInteractor()
	c := NewAgnesCuiController(ci)
	ci.On("Reset").Return("reset")
	assert.Equal(t, "reset", c.Exec("r"))
	assert.Equal(t, "reset", c.Exec("reset"))
}

func TestAgnesCuiControllerDeal(t *testing.T) {
	ci := newMockAgnesInteractor()
	c := NewAgnesCuiController(ci)
	ci.On("DealStock").Return("deal")
	assert.Equal(t, "deal", c.Exec("d"))
	assert.Equal(t, "deal", c.Exec("deal"))
}

func TestAgnesCuiControllerGiveUp(t *testing.T) {
	ci := newMockAgnesInteractor()
	c := NewAgnesCuiController(ci)
	ci.On("GiveUp").Return("giveup")
	assert.Equal(t, "giveup", c.Exec("g"))
	assert.Equal(t, "giveup", c.Exec("giveup"))
}

func TestAgnesCuiControllerUndo(t *testing.T) {
	ci := newMockAgnesInteractor()
	c := NewAgnesCuiController(ci)
	ci.On("Undo").Return("undo")
	assert.Equal(t, "undo", c.Exec("u"))
	assert.Equal(t, "undo", c.Exec("undo"))
}

func TestAgnesCuiControllerHint(t *testing.T) {
	ci := newMockAgnesInteractor()
	c := NewAgnesCuiController(ci)
	ci.On("Hint").Return("hint")
	assert.Equal(t, "hint", c.Exec("h"))
	assert.Equal(t, "hint", c.Exec("hint"))
}

func TestAgnesCuiControllerLog(t *testing.T) {
	ci := newMockAgnesInteractor()
	c := NewAgnesCuiController(ci)
	ci.On("ActionLog").Return("log")
	assert.Equal(t, "log", c.Exec("log"))
	assert.Equal(t, "log", c.Exec("l"))
}

func TestAgnesCuiControllerMoveTableauToTableau(t *testing.T) {
	ci := newMockAgnesInteractor()
	c := NewAgnesCuiController(ci)
	ci.On("MoveTableauToTableau", 0, -1, 3).Return("mtt")
	assert.Equal(t, "mtt", c.Exec("m 0 3"))
}

func TestAgnesCuiControllerMoveTableauToFoundation(t *testing.T) {
	ci := newMockAgnesInteractor()
	c := NewAgnesCuiController(ci)
	ci.On("MoveTableauToFoundation", 2).Return("mtf")
	assert.Equal(t, "mtf", c.Exec("m 2 f"))
}

func TestAgnesCuiControllerMoveErrors(t *testing.T) {
	t.Run("empty move prompts", func(t *testing.T) {
		ci := newMockAgnesInteractor()
		c := NewAgnesCuiController(ci)
		assert.True(t, cuiutil.IsPromptRequest(c.Exec("m")))
	})

	t.Run("from only prompts for destination", func(t *testing.T) {
		ci := newMockAgnesInteractor()
		c := NewAgnesCuiController(ci)
		result := c.Exec("m 0")
		assert.True(t, cuiutil.IsPromptRequest(result))
		_, tmpl := cuiutil.ParsePromptRequest(result)
		assert.Equal(t, "m 0 {0}", tmpl)
	})

	t.Run("invalid from column", func(t *testing.T) {
		ci := newMockAgnesInteractor()
		c := NewAgnesCuiController(ci)
		assert.Contains(t, c.Exec("m abc 1"), "abc")
	})

	t.Run("invalid to column", func(t *testing.T) {
		ci := newMockAgnesInteractor()
		c := NewAgnesCuiController(ci)
		assert.Contains(t, c.Exec("m 0 xyz"), "xyz")
	})
}

func TestAgnesCuiControllerUnknown(t *testing.T) {
	ci := newMockAgnesInteractor()
	c := NewAgnesCuiController(ci)
	assert.Contains(t, c.Exec("zzz"), "コマンドが不明です")
}

func TestAgnesCuiControllerEmpty(t *testing.T) {
	ci := newMockAgnesInteractor()
	c := NewAgnesCuiController(ci)
	assert.Contains(t, c.Exec(""), "'help' でコマンド一覧を表示します。")
}
