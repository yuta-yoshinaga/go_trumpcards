//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockOasisPokerInteractor() *usecase.MockOasisPokerInteractor {
	m := new(usecase.MockOasisPokerInteractor)
	m.On("Reset").Return("reset result")
	m.On("Bet", 100, 0).Return("bet result")
	m.On("Bet", 100, 10).Return("bet with jackpot result")
	m.On("Exchange", []int{}).Return("exchange empty result")
	m.On("Exchange", []int{0}).Return("exchange 0 result")
	m.On("Exchange", []int{0, 2, 4}).Return("exchange multi result")
	m.On("Stand").Return("stand result")
	m.On("Play").Return("play result")
	m.On("Fold").Return("fold result")
	m.On("ActionLog").Return("action log result")
	m.On("Hint").Return("hint result")
	return m
}

func TestOasisPokerCuiController_Quit(t *testing.T) {
	m := newMockOasisPokerInteractor()
	c := controller.NewOasisPokerCuiController(m)

	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestOasisPokerCuiController_Reset(t *testing.T) {
	m := newMockOasisPokerInteractor()
	c := controller.NewOasisPokerCuiController(m)

	assert.Equal(t, "reset result", c.Exec("r"))
	assert.Equal(t, "reset result", c.Exec("reset"))
}

func TestOasisPokerCuiController_Bet(t *testing.T) {
	m := newMockOasisPokerInteractor()
	c := controller.NewOasisPokerCuiController(m)

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

func TestOasisPokerCuiController_Bet_Errors(t *testing.T) {
	m := newMockOasisPokerInteractor()
	c := controller.NewOasisPokerCuiController(m)

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
		assert.Contains(t, result, "Invalid jackpot amount")
	})
}

func TestOasisPokerCuiController_Exchange(t *testing.T) {
	m := newMockOasisPokerInteractor()
	c := controller.NewOasisPokerCuiController(m)

	t.Run("single index", func(t *testing.T) {
		assert.Equal(t, "exchange 0 result", c.Exec("e 0"))
	})

	t.Run("multiple indices", func(t *testing.T) {
		assert.Equal(t, "exchange multi result", c.Exec("exchange 0 2 4"))
	})

	t.Run("no indices", func(t *testing.T) {
		// Empty indices: ParseBoundedIntSlice returns nil/empty for "e" with no args
		assert.Equal(t, "exchange empty result", c.Exec("e"))
	})
}

func TestOasisPokerCuiController_Stand(t *testing.T) {
	m := newMockOasisPokerInteractor()
	c := controller.NewOasisPokerCuiController(m)

	assert.Equal(t, "stand result", c.Exec("s"))
	assert.Equal(t, "stand result", c.Exec("stand"))
}

func TestOasisPokerCuiController_Play(t *testing.T) {
	m := newMockOasisPokerInteractor()
	c := controller.NewOasisPokerCuiController(m)

	assert.Equal(t, "play result", c.Exec("p"))
	assert.Equal(t, "play result", c.Exec("play"))
}

func TestOasisPokerCuiController_Fold(t *testing.T) {
	m := newMockOasisPokerInteractor()
	c := controller.NewOasisPokerCuiController(m)

	assert.Equal(t, "fold result", c.Exec("f"))
	assert.Equal(t, "fold result", c.Exec("fold"))
}

func TestOasisPokerCuiController_ActionLog(t *testing.T) {
	m := newMockOasisPokerInteractor()
	c := controller.NewOasisPokerCuiController(m)

	assert.Equal(t, "action log result", c.Exec("log"))
	assert.Equal(t, "action log result", c.Exec("l"))
}

func TestOasisPokerCuiController_Unknown(t *testing.T) {
	m := newMockOasisPokerInteractor()
	c := controller.NewOasisPokerCuiController(m)

	result := c.Exec("xyz")
	assert.Contains(t, result, "コマンドが不明です")
}

func TestOasisPokerCuiController_Empty(t *testing.T) {
	m := newMockOasisPokerInteractor()
	c := controller.NewOasisPokerCuiController(m)

	result := c.Exec("")
	assert.Contains(t, result, "'help' でコマンド一覧を表示します。")
}

func TestOasisPokerCuiController_Hint(t *testing.T) {
	m := newMockOasisPokerInteractor()
	c := controller.NewOasisPokerCuiController(m)
	assert.Equal(t, "hint result", c.Exec("h"))
	assert.Equal(t, "hint result", c.Exec("hint"))
	m.AssertCalled(t, "Hint")
}
