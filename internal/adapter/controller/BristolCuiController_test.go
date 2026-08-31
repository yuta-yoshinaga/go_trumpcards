package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newMockBristolInteractor() *mockusecase.MockBristolInteractor {
	return new(mockusecase.MockBristolInteractor)
}

func TestBristolCuiControllerQuit(t *testing.T) {
	c := NewBristolCuiController(newMockBristolInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestBristolCuiControllerReset(t *testing.T) {
	bi := newMockBristolInteractor()
	c := NewBristolCuiController(bi)
	bi.On("Reset").Return("reset")
	assert.Equal(t, "reset", c.Exec("r"))
	assert.Equal(t, "reset", c.Exec("reset"))
}

func TestBristolCuiControllerNoArgCommands(t *testing.T) {
	cases := []struct {
		method  string
		aliases []string
	}{
		{"Draw", []string{"d", "draw"}},
		{"GiveUp", []string{"g", "giveup"}},
		{"AutoComplete", []string{"ac", "autocomplete"}},
		{"Undo", []string{"u", "undo"}},
		{"Hint", []string{"h", "hint"}},
		{"ActionLog", []string{"log", "l"}},
	}
	for _, tc := range cases {
		for _, alias := range tc.aliases {
			bi := newMockBristolInteractor()
			c := NewBristolCuiController(bi)
			bi.On(tc.method).Return("out")
			assert.Equal(t, "out", c.Exec(alias), "method=%s alias=%s", tc.method, alias)
		}
	}
}

func TestBristolCuiControllerMoveTableauToTableau(t *testing.T) {
	bi := newMockBristolInteractor()
	c := NewBristolCuiController(bi)
	bi.On("MoveTableauToTableau", 0, 1).Return("mtt")
	assert.Equal(t, "mtt", c.Exec("m t 0 t 1"))
}

func TestBristolCuiControllerMoveTableauToFoundation(t *testing.T) {
	bi := newMockBristolInteractor()
	c := NewBristolCuiController(bi)
	bi.On("MoveTableauToFoundation", 2).Return("mtf")
	assert.Equal(t, "mtf", c.Exec("m t 2 f"))
}

func TestBristolCuiControllerMoveFanToTableau(t *testing.T) {
	bi := newMockBristolInteractor()
	c := NewBristolCuiController(bi)
	bi.On("MoveFanToTableau", 1, 3).Return("mnt")
	assert.Equal(t, "mnt", c.Exec("m n 1 t 3"))
}

func TestBristolCuiControllerMoveFanToFoundation(t *testing.T) {
	bi := newMockBristolInteractor()
	c := NewBristolCuiController(bi)
	bi.On("MoveFanToFoundation", 0).Return("mnf")
	assert.Equal(t, "mnf", c.Exec("m n 0 f"))
}

func TestBristolCuiControllerMoveErrors(t *testing.T) {
	t.Run("empty move prompts", func(t *testing.T) {
		c := NewBristolCuiController(newMockBristolInteractor())
		assert.True(t, cuiutil.IsPromptRequest(c.Exec("m")))
	})
	t.Run("invalid from", func(t *testing.T) {
		c := NewBristolCuiController(newMockBristolInteractor())
		assert.Contains(t, c.Exec("m x 0 f"), "x")
	})
	t.Run("tableau prompt for column", func(t *testing.T) {
		c := NewBristolCuiController(newMockBristolInteractor())
		assert.True(t, cuiutil.IsPromptRequest(c.Exec("m t")))
	})
	t.Run("tableau invalid column", func(t *testing.T) {
		c := NewBristolCuiController(newMockBristolInteractor())
		assert.Contains(t, c.Exec("m t abc f"), "abc")
	})
	t.Run("tableau destination prompt", func(t *testing.T) {
		c := NewBristolCuiController(newMockBristolInteractor())
		assert.True(t, cuiutil.IsPromptRequest(c.Exec("m t 0")))
	})
	t.Run("tableau to-column prompt", func(t *testing.T) {
		c := NewBristolCuiController(newMockBristolInteractor())
		assert.True(t, cuiutil.IsPromptRequest(c.Exec("m t 0 t")))
	})
	t.Run("tableau invalid to-column", func(t *testing.T) {
		c := NewBristolCuiController(newMockBristolInteractor())
		assert.Contains(t, c.Exec("m t 0 t abc"), "abc")
	})
	t.Run("tableau usage when bad destination", func(t *testing.T) {
		c := NewBristolCuiController(newMockBristolInteractor())
		out := c.Exec("m t 0 x")
		assert.NotEmpty(t, out)
		assert.False(t, cuiutil.IsPromptRequest(out))
	})
	t.Run("fan prompt for index", func(t *testing.T) {
		c := NewBristolCuiController(newMockBristolInteractor())
		assert.True(t, cuiutil.IsPromptRequest(c.Exec("m n")))
	})
	t.Run("fan invalid index", func(t *testing.T) {
		c := NewBristolCuiController(newMockBristolInteractor())
		assert.Contains(t, c.Exec("m n abc f"), "abc")
	})
	t.Run("fan destination prompt", func(t *testing.T) {
		c := NewBristolCuiController(newMockBristolInteractor())
		assert.True(t, cuiutil.IsPromptRequest(c.Exec("m n 0")))
	})
	t.Run("fan to-column prompt", func(t *testing.T) {
		c := NewBristolCuiController(newMockBristolInteractor())
		assert.True(t, cuiutil.IsPromptRequest(c.Exec("m n 0 t")))
	})
	t.Run("fan invalid to-column", func(t *testing.T) {
		c := NewBristolCuiController(newMockBristolInteractor())
		assert.Contains(t, c.Exec("m n 0 t abc"), "abc")
	})
	t.Run("fan usage when bad destination", func(t *testing.T) {
		c := NewBristolCuiController(newMockBristolInteractor())
		out := c.Exec("m n 0 x")
		assert.NotEmpty(t, out)
		assert.False(t, cuiutil.IsPromptRequest(out))
	})
}

func TestBristolCuiControllerUnknown(t *testing.T) {
	c := NewBristolCuiController(newMockBristolInteractor())
	assert.Contains(t, c.Exec("xyz"), "コマンドが不明です")
}

func TestBristolCuiControllerEmpty(t *testing.T) {
	c := NewBristolCuiController(newMockBristolInteractor())
	assert.Contains(t, c.Exec(""), "'help' でコマンド一覧を表示します。")
}

// #6427: 8 列 + 3 ファン + 4 組札を「押して弾かれる」で探るしかなかった。
func TestBristolCuiControllerTargets(t *testing.T) {
	t.Run("tableau source with t alias", func(t *testing.T) {
		bi := newMockBristolInteractor()
		c := NewBristolCuiController(bi)
		bi.On("Targets", "tableau", 0).Return("targets_tt0")
		assert.Equal(t, "targets_tt0", c.Exec("t t 0"))
		bi.AssertExpectations(t)
	})

	t.Run("fan source with targets alias", func(t *testing.T) {
		bi := newMockBristolInteractor()
		c := NewBristolCuiController(bi)
		bi.On("Targets", "fan", 1).Return("targets_fn1")
		assert.Equal(t, "targets_fn1", c.Exec("targets n 1"))
		bi.AssertExpectations(t)
	})
}

func TestBristolCuiControllerTargetsErrors(t *testing.T) {
	t.Run("no args prompts for source zone", func(t *testing.T) {
		bi := newMockBristolInteractor()
		c := NewBristolCuiController(bi)
		out := c.Exec("t")
		assert.True(t, cuiutil.IsPromptRequest(out))
		assert.Contains(t, out, i18n.T("bristol.promptSourceZone"))
		bi.AssertNotCalled(t, "Targets", mock.Anything, mock.Anything)
	})

	t.Run("invalid source zone", func(t *testing.T) {
		bi := newMockBristolInteractor()
		c := NewBristolCuiController(bi)
		out := c.Exec("t x 0")
		assert.Contains(t, out, "x")
		bi.AssertNotCalled(t, "Targets", mock.Anything, mock.Anything)
	})

	t.Run("tableau prompts for column", func(t *testing.T) {
		bi := newMockBristolInteractor()
		c := NewBristolCuiController(bi)
		out := c.Exec("t t")
		assert.True(t, cuiutil.IsPromptRequest(out))
		assert.Contains(t, out, i18n.T("bristol.promptFromColumn"))
		bi.AssertNotCalled(t, "Targets", mock.Anything, mock.Anything)
	})

	t.Run("tableau rejects invalid column", func(t *testing.T) {
		bi := newMockBristolInteractor()
		c := NewBristolCuiController(bi)
		out := c.Exec("t t abc")
		assert.Contains(t, out, "abc")
		bi.AssertNotCalled(t, "Targets", mock.Anything, mock.Anything)
	})

	t.Run("fan prompts for index", func(t *testing.T) {
		bi := newMockBristolInteractor()
		c := NewBristolCuiController(bi)
		out := c.Exec("t n")
		assert.True(t, cuiutil.IsPromptRequest(out))
		assert.Contains(t, out, i18n.T("bristol.promptFan"))
		bi.AssertNotCalled(t, "Targets", mock.Anything, mock.Anything)
	})

	t.Run("fan rejects invalid index", func(t *testing.T) {
		bi := newMockBristolInteractor()
		c := NewBristolCuiController(bi)
		out := c.Exec("t n xyz")
		assert.Contains(t, out, "xyz")
		bi.AssertNotCalled(t, "Targets", mock.Anything, mock.Anything)
	})
}

// **追加コマンドが候補一覧に入っていること。**
func TestBristolCuiControllerSuggestsTargets(t *testing.T) {
	bi := newMockBristolInteractor()
	c := NewBristolCuiController(bi)
	assert.Contains(t, c.Exec("target t 0"), "targets")
}
