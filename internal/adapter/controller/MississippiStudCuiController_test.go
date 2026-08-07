//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockMississippiStudInteractor() *usecase.MockMississippiStudInteractor {
	m := new(usecase.MockMississippiStudInteractor)
	m.On("Reset").Return("reset result")
	m.On("Bet", 100).Return("bet result")
	m.On("Play", 1).Return("play 1 result")
	m.On("Play", 2).Return("play 2 result")
	m.On("Play", 3).Return("play 3 result")
	m.On("Fold").Return("fold result")
	m.On("ActionLog").Return("action log result")
	m.On("Hint").Return("hint result")
	return m
}

func TestMississippiStudCuiController_Quit(t *testing.T) {
	m := newMockMississippiStudInteractor()
	c := controller.NewMississippiStudCuiController(m)

	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestMississippiStudCuiController_Reset(t *testing.T) {
	m := newMockMississippiStudInteractor()
	c := controller.NewMississippiStudCuiController(m)

	assert.Equal(t, "reset result", c.Exec("r"))
	assert.Equal(t, "reset result", c.Exec("reset"))
}

func TestMississippiStudCuiController_Bet(t *testing.T) {
	m := newMockMississippiStudInteractor()
	c := controller.NewMississippiStudCuiController(m)

	t.Run("short", func(t *testing.T) {
		assert.Equal(t, "bet result", c.Exec("b 100"))
	})
	t.Run("long", func(t *testing.T) {
		assert.Equal(t, "bet result", c.Exec("bet 100"))
	})
}

func TestMississippiStudCuiController_Bet_Errors(t *testing.T) {
	m := newMockMississippiStudInteractor()
	c := controller.NewMississippiStudCuiController(m)

	t.Run("missing args", func(t *testing.T) {
		assert.Contains(t, c.Exec("b"), "Bet amount is required")
	})
	t.Run("invalid amount", func(t *testing.T) {
		assert.Contains(t, c.Exec("b abc"), "Invalid bet amount")
	})
	t.Run("zero amount", func(t *testing.T) {
		assert.Contains(t, c.Exec("b 0"), "Invalid bet amount")
	})
}

func TestMississippiStudCuiController_Play(t *testing.T) {
	m := newMockMississippiStudInteractor()
	c := controller.NewMississippiStudCuiController(m)

	assert.Equal(t, "play 1 result", c.Exec("p 1"))
	assert.Equal(t, "play 2 result", c.Exec("p 2"))
	assert.Equal(t, "play 3 result", c.Exec("play 3"))
}

func TestMississippiStudCuiController_Play_Errors(t *testing.T) {
	m := newMockMississippiStudInteractor()
	c := controller.NewMississippiStudCuiController(m)

	t.Run("missing multiplier", func(t *testing.T) {
		assert.Contains(t, c.Exec("p"), "Multiplier")
	})
	t.Run("invalid multiplier", func(t *testing.T) {
		assert.Contains(t, c.Exec("p 4"), "Invalid multiplier")
	})
	t.Run("non-numeric multiplier", func(t *testing.T) {
		assert.Contains(t, c.Exec("p abc"), "Invalid multiplier")
	})
}

func TestMississippiStudCuiController_Fold(t *testing.T) {
	m := newMockMississippiStudInteractor()
	c := controller.NewMississippiStudCuiController(m)

	assert.Equal(t, "fold result", c.Exec("f"))
	assert.Equal(t, "fold result", c.Exec("fold"))
}

func TestMississippiStudCuiController_ActionLog(t *testing.T) {
	m := newMockMississippiStudInteractor()
	c := controller.NewMississippiStudCuiController(m)

	assert.Equal(t, "action log result", c.Exec("log"))
}

func TestMississippiStudCuiController_Unknown(t *testing.T) {
	m := newMockMississippiStudInteractor()
	c := controller.NewMississippiStudCuiController(m)

	assert.Contains(t, c.Exec("xyz"), "コマンドが不明です")
}

func TestMississippiStudCuiController_Empty(t *testing.T) {
	m := newMockMississippiStudInteractor()
	c := controller.NewMississippiStudCuiController(m)

	assert.Contains(t, c.Exec(""), "'help' でコマンド一覧を表示します。")
}

// **CUI に 1x/3x/fold を選ぶ材料が無かった (#4710)。**
func TestMississippiStudCuiController_Hint(t *testing.T) {
	m := newMockMississippiStudInteractor()
	c := controller.NewMississippiStudCuiController(m)

	assert.Equal(t, "hint result", c.Exec("h"))
	assert.Equal(t, "hint result", c.Exec("hint"))
}
