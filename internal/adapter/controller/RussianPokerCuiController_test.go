//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockRussianPokerInteractor() *usecase.MockRussianPokerInteractor {
	m := new(usecase.MockRussianPokerInteractor)
	m.On("Reset").Return("reset result")
	m.On("Bet", 100).Return("bet result")
	m.On("Exchange", []int{}).Return("exchange empty result")
	m.On("Exchange", []int{0}).Return("exchange 0 result")
	m.On("Exchange", []int{0, 2, 4}).Return("exchange multi result")
	m.On("Buy6th").Return("buy6th result")
	m.On("Select", 3).Return("select result")
	m.On("Play").Return("play result")
	m.On("Fold").Return("fold result")
	m.On("ForceExchange").Return("force result")
	m.On("Decline").Return("decline result")
	m.On("ActionLog").Return("action log result")
	return m
}

func TestRussianPokerCuiController_Quit(t *testing.T) {
	m := newMockRussianPokerInteractor()
	c := controller.NewRussianPokerCuiController(m)

	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestRussianPokerCuiController_Reset(t *testing.T) {
	m := newMockRussianPokerInteractor()
	c := controller.NewRussianPokerCuiController(m)

	assert.Equal(t, "reset result", c.Exec("r"))
	assert.Equal(t, "reset result", c.Exec("reset"))
}

func TestRussianPokerCuiController_Bet(t *testing.T) {
	m := newMockRussianPokerInteractor()
	c := controller.NewRussianPokerCuiController(m)

	t.Run("ante", func(t *testing.T) {
		assert.Equal(t, "bet result", c.Exec("b 100"))
	})

	t.Run("bet long form", func(t *testing.T) {
		assert.Equal(t, "bet result", c.Exec("bet 100"))
	})
}

func TestRussianPokerCuiController_Bet_Errors(t *testing.T) {
	m := newMockRussianPokerInteractor()
	c := controller.NewRussianPokerCuiController(m)

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
}

func TestRussianPokerCuiController_Exchange(t *testing.T) {
	m := newMockRussianPokerInteractor()
	c := controller.NewRussianPokerCuiController(m)

	t.Run("single index", func(t *testing.T) {
		assert.Equal(t, "exchange 0 result", c.Exec("e 0"))
	})

	t.Run("multiple indices", func(t *testing.T) {
		assert.Equal(t, "exchange multi result", c.Exec("exchange 0 2 4"))
	})

	t.Run("no indices", func(t *testing.T) {
		assert.Equal(t, "exchange empty result", c.Exec("e"))
	})
}

func TestRussianPokerCuiController_Buy6th(t *testing.T) {
	m := newMockRussianPokerInteractor()
	c := controller.NewRussianPokerCuiController(m)

	assert.Equal(t, "buy6th result", c.Exec("6"))
	assert.Equal(t, "buy6th result", c.Exec("buy6th"))
}

func TestRussianPokerCuiController_Select(t *testing.T) {
	m := newMockRussianPokerInteractor()
	c := controller.NewRussianPokerCuiController(m)

	t.Run("select with index", func(t *testing.T) {
		assert.Equal(t, "select result", c.Exec("sel 3"))
	})

	t.Run("select long form", func(t *testing.T) {
		assert.Equal(t, "select result", c.Exec("select 3"))
	})
}

func TestRussianPokerCuiController_Select_Errors(t *testing.T) {
	m := newMockRussianPokerInteractor()
	c := controller.NewRussianPokerCuiController(m)

	t.Run("missing args", func(t *testing.T) {
		result := c.Exec("sel")
		assert.Contains(t, result, "Discard index is required")
	})

	t.Run("invalid index", func(t *testing.T) {
		result := c.Exec("sel abc")
		assert.Contains(t, result, "Invalid index")
	})
}

func TestRussianPokerCuiController_Play(t *testing.T) {
	m := newMockRussianPokerInteractor()
	c := controller.NewRussianPokerCuiController(m)

	assert.Equal(t, "play result", c.Exec("p"))
	assert.Equal(t, "play result", c.Exec("play"))
}

func TestRussianPokerCuiController_Fold(t *testing.T) {
	m := newMockRussianPokerInteractor()
	c := controller.NewRussianPokerCuiController(m)

	assert.Equal(t, "fold result", c.Exec("f"))
	assert.Equal(t, "fold result", c.Exec("fold"))
}

func TestRussianPokerCuiController_ForceExchange(t *testing.T) {
	m := newMockRussianPokerInteractor()
	c := controller.NewRussianPokerCuiController(m)

	assert.Equal(t, "force result", c.Exec("force"))
	assert.Equal(t, "force result", c.Exec("fe"))
}

func TestRussianPokerCuiController_Decline(t *testing.T) {
	m := newMockRussianPokerInteractor()
	c := controller.NewRussianPokerCuiController(m)

	assert.Equal(t, "decline result", c.Exec("d"))
	assert.Equal(t, "decline result", c.Exec("decline"))
}

func TestRussianPokerCuiController_ActionLog(t *testing.T) {
	m := newMockRussianPokerInteractor()
	c := controller.NewRussianPokerCuiController(m)

	assert.Equal(t, "action log result", c.Exec("log"))
	assert.Equal(t, "action log result", c.Exec("l"))
}

func TestRussianPokerCuiController_Unknown(t *testing.T) {
	m := newMockRussianPokerInteractor()
	c := controller.NewRussianPokerCuiController(m)

	result := c.Exec("xyz")
	assert.Contains(t, result, "コマンドが不明です")
}

func TestRussianPokerCuiController_Empty(t *testing.T) {
	m := newMockRussianPokerInteractor()
	c := controller.NewRussianPokerCuiController(m)

	result := c.Exec("")
	assert.Contains(t, result, "'help' でコマンド一覧を表示します。")
}

// **落として残りで実行しない。** 打ち間違いを捨てると、プレイヤーが選んで
// いない組み合わせが実行される (issue #5390)。
func TestRussianPokerCuiController_RefusesMistypedIndex(t *testing.T) {
	m := newMockRussianPokerInteractor()
	c := controller.NewRussianPokerCuiController(m)
	assert.Contains(t, c.Exec("e 0 zz"), msgInvalidCardIndexPrefix(),
		"a mistyped index must be refused, not dropped")
}
