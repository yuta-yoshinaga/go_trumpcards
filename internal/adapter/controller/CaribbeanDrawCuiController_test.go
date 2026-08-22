//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockCaribbeanDrawInteractor() *usecase.MockCaribbeanDrawInteractor {
	m := new(usecase.MockCaribbeanDrawInteractor)
	m.On("Reset").Return("reset result")
	m.On("Bet", 100, 0).Return("bet result")
	m.On("Bet", 100, 10).Return("bet with jackpot result")
	m.On("Play").Return("play result")
	m.On("Fold").Return("fold result")
	m.On("ActionLog").Return("action log result")
	m.On("Hint").Return("hint result")
	return m
}

func TestCaribbeanDrawCuiController_Quit(t *testing.T) {
	m := newMockCaribbeanDrawInteractor()
	c := controller.NewCaribbeanDrawCuiController(m)

	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestCaribbeanDrawCuiController_Reset(t *testing.T) {
	m := newMockCaribbeanDrawInteractor()
	c := controller.NewCaribbeanDrawCuiController(m)

	assert.Equal(t, "reset result", c.Exec("r"))
	assert.Equal(t, "reset result", c.Exec("reset"))
}

func TestCaribbeanDrawCuiController_Bet(t *testing.T) {
	m := newMockCaribbeanDrawInteractor()
	c := controller.NewCaribbeanDrawCuiController(m)

	t.Run("ante only", func(t *testing.T) {
		assert.Equal(t, "bet result", c.Exec("b 100"))
	})

	t.Run("ante with jackpot", func(t *testing.T) {
		assert.Equal(t, "bet with jackpot result", c.Exec("b 100 10"))
	})

	t.Run("bet long form", func(t *testing.T) {
		assert.Equal(t, "bet result", c.Exec("bet 100"))
	})
}

func TestCaribbeanDrawCuiController_Bet_Errors(t *testing.T) {
	m := newMockCaribbeanDrawInteractor()
	c := controller.NewCaribbeanDrawCuiController(m)

	t.Run("missing args", func(t *testing.T) {
		result := c.Exec("b")
		assert.Contains(t, result, msgAnteAmountRequired())
	})

	t.Run("invalid amount", func(t *testing.T) {
		result := c.Exec("b abc")
		assert.Contains(t, result, msgInvalidAnteAmountPrefix())
	})

	t.Run("zero amount", func(t *testing.T) {
		result := c.Exec("b 0")
		assert.Contains(t, result, msgInvalidAnteAmountPrefix())
	})

	t.Run("negative amount", func(t *testing.T) {
		result := c.Exec("b -10")
		assert.Contains(t, result, msgInvalidAnteAmountPrefix())
	})

	t.Run("invalid jackpot", func(t *testing.T) {
		result := c.Exec("b 100 abc")
		assert.Contains(t, result, msgStem("invalidJackpotAmount"))
	})
}

func TestCaribbeanDrawCuiController_Play(t *testing.T) {
	m := newMockCaribbeanDrawInteractor()
	c := controller.NewCaribbeanDrawCuiController(m)

	assert.Equal(t, "play result", c.Exec("p"))
	assert.Equal(t, "play result", c.Exec("play"))
}

func TestCaribbeanDrawCuiController_Fold(t *testing.T) {
	m := newMockCaribbeanDrawInteractor()
	c := controller.NewCaribbeanDrawCuiController(m)

	assert.Equal(t, "fold result", c.Exec("f"))
	assert.Equal(t, "fold result", c.Exec("fold"))
}

func TestCaribbeanDrawCuiController_ActionLog(t *testing.T) {
	m := newMockCaribbeanDrawInteractor()
	c := controller.NewCaribbeanDrawCuiController(m)

	assert.Equal(t, "action log result", c.Exec("log"))
	assert.Equal(t, "action log result", c.Exec("l"))
}

func TestCaribbeanDrawCuiController_Unknown(t *testing.T) {
	m := newMockCaribbeanDrawInteractor()
	c := controller.NewCaribbeanDrawCuiController(m)

	result := c.Exec("xyz")
	assert.Contains(t, result, "コマンドが不明です")
}

func TestCaribbeanDrawCuiController_Empty(t *testing.T) {
	m := newMockCaribbeanDrawInteractor()
	c := controller.NewCaribbeanDrawCuiController(m)

	result := c.Exec("")
	assert.Contains(t, result, "'help' でコマンド一覧を表示します。")
}

// **カリビアンスタッドだけ CUI に戦略アシストが無かった (#4697)。**
func TestCaribbeanDrawCuiController_Hint(t *testing.T) {
	m := newMockCaribbeanDrawInteractor()
	c := controller.NewCaribbeanDrawCuiController(m)

	assert.Equal(t, "hint result", c.Exec("h"))
	assert.Equal(t, "hint result", c.Exec("hint"))
}
