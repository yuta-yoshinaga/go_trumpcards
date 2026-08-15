//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockUltimateTexasHoldemInteractor() *usecase.MockUltimateTexasHoldemInteractor {
	m := new(usecase.MockUltimateTexasHoldemInteractor)
	m.On("Reset").Return("reset result")
	m.On("Bet", 100, 0).Return("bet result")
	m.On("Bet", 100, 10).Return("bet with trips result")
	m.On("Play", 4).Return("play4 result")
	m.On("Play", 3).Return("play3 result")
	m.On("Play", 2).Return("play2 result")
	m.On("Play", 1).Return("play1 result")
	m.On("Check").Return("check result")
	m.On("Fold").Return("fold result")
	m.On("ActionLog").Return("action log result")
	m.On("Hint").Return("hint result")
	return m
}

func TestUltimateTexasHoldemCuiController_Quit(t *testing.T) {
	m := newMockUltimateTexasHoldemInteractor()
	c := controller.NewUltimateTexasHoldemCuiController(m)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestUltimateTexasHoldemCuiController_Reset(t *testing.T) {
	m := newMockUltimateTexasHoldemInteractor()
	c := controller.NewUltimateTexasHoldemCuiController(m)
	assert.Equal(t, "reset result", c.Exec("r"))
	assert.Equal(t, "reset result", c.Exec("reset"))
}

func TestUltimateTexasHoldemCuiController_Bet(t *testing.T) {
	m := newMockUltimateTexasHoldemInteractor()
	c := controller.NewUltimateTexasHoldemCuiController(m)

	t.Run("ante only", func(t *testing.T) {
		assert.Equal(t, "bet result", c.Exec("b 100"))
	})
	t.Run("ante with trips", func(t *testing.T) {
		assert.Equal(t, "bet with trips result", c.Exec("b 100 10"))
	})
	t.Run("bet long form", func(t *testing.T) {
		assert.Equal(t, "bet result", c.Exec("bet 100"))
	})
}

func TestUltimateTexasHoldemCuiController_Bet_Errors(t *testing.T) {
	m := newMockUltimateTexasHoldemInteractor()
	c := controller.NewUltimateTexasHoldemCuiController(m)

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
	t.Run("invalid trips", func(t *testing.T) {
		result := c.Exec("b 100 abc")
		assert.Contains(t, result, "Invalid trips amount")
	})
}

func TestUltimateTexasHoldemCuiController_Play(t *testing.T) {
	m := newMockUltimateTexasHoldemInteractor()
	c := controller.NewUltimateTexasHoldemCuiController(m)
	assert.Equal(t, "play4 result", c.Exec("p 4"))
	assert.Equal(t, "play3 result", c.Exec("p 3"))
	assert.Equal(t, "play2 result", c.Exec("play 2"))
	assert.Equal(t, "play1 result", c.Exec("p 1"))
}

func TestUltimateTexasHoldemCuiController_Play_Errors(t *testing.T) {
	m := newMockUltimateTexasHoldemInteractor()
	c := controller.NewUltimateTexasHoldemCuiController(m)
	t.Run("missing multiplier", func(t *testing.T) {
		result := c.Exec("p")
		assert.Contains(t, result, "Play multiplier is required")
	})
	t.Run("invalid multiplier", func(t *testing.T) {
		result := c.Exec("p abc")
		assert.Contains(t, result, "Invalid play multiplier")
	})
}

func TestUltimateTexasHoldemCuiController_Check(t *testing.T) {
	m := newMockUltimateTexasHoldemInteractor()
	c := controller.NewUltimateTexasHoldemCuiController(m)
	assert.Equal(t, "check result", c.Exec("c"))
	assert.Equal(t, "check result", c.Exec("check"))
}

func TestUltimateTexasHoldemCuiController_Fold(t *testing.T) {
	m := newMockUltimateTexasHoldemInteractor()
	c := controller.NewUltimateTexasHoldemCuiController(m)
	assert.Equal(t, "fold result", c.Exec("f"))
	assert.Equal(t, "fold result", c.Exec("fold"))
}

func TestUltimateTexasHoldemCuiController_ActionLog(t *testing.T) {
	m := newMockUltimateTexasHoldemInteractor()
	c := controller.NewUltimateTexasHoldemCuiController(m)
	assert.Equal(t, "action log result", c.Exec("log"))
	assert.Equal(t, "action log result", c.Exec("l"))
}

func TestUltimateTexasHoldemCuiController_Unknown(t *testing.T) {
	m := newMockUltimateTexasHoldemInteractor()
	c := controller.NewUltimateTexasHoldemCuiController(m)
	result := c.Exec("xyz")
	assert.Contains(t, result, "コマンドが不明です")
}

func TestUltimateTexasHoldemCuiController_Empty(t *testing.T) {
	m := newMockUltimateTexasHoldemInteractor()
	c := controller.NewUltimateTexasHoldemCuiController(m)
	result := c.Exec("")
	assert.Contains(t, result, "'help' でコマンド一覧を表示します。")
}

// **CUI には 4x/3x/2x/1x/check/fold を選ぶ材料が何も無かった (#4709)。**
func TestUltimateTexasHoldemCuiController_Hint(t *testing.T) {
	m := newMockUltimateTexasHoldemInteractor()
	c := controller.NewUltimateTexasHoldemCuiController(m)

	assert.Equal(t, "hint result", c.Exec("h"))
	assert.Equal(t, "hint result", c.Exec("hint"))
}

// **既存コマンドは何も変わらない。**
func TestUltimateTexasHoldemCuiController_HintDoesNotShadowOthers(t *testing.T) {
	m := newMockUltimateTexasHoldemInteractor()
	c := controller.NewUltimateTexasHoldemCuiController(m)

	assert.Equal(t, "check result", c.Exec("c"))
	m.AssertNotCalled(t, "Hint")
}
