//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockThreeCardInteractor() *usecase.MockThreeCardInteractor {
	m := new(usecase.MockThreeCardInteractor)
	m.On("Reset").Return("reset result")
	m.On("Bet", 100, 0).Return("bet result")
	m.On("Bet", 100, 50).Return("bet with pairplus result")
	m.On("Play").Return("play result")
	m.On("Fold").Return("fold result")
	m.On("ActionLog").Return("action log result")
	return m
}

func TestThreeCardCuiController_Quit(t *testing.T) {
	m := newMockThreeCardInteractor()
	c := controller.NewThreeCardCuiController(m)

	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestThreeCardCuiController_Reset(t *testing.T) {
	m := newMockThreeCardInteractor()
	c := controller.NewThreeCardCuiController(m)

	assert.Equal(t, "reset result", c.Exec("r"))
	assert.Equal(t, "reset result", c.Exec("reset"))
}

func TestThreeCardCuiController_Bet(t *testing.T) {
	m := newMockThreeCardInteractor()
	c := controller.NewThreeCardCuiController(m)

	t.Run("ante only", func(t *testing.T) {
		assert.Equal(t, "bet result", c.Exec("b 100"))
	})

	t.Run("ante with pair plus", func(t *testing.T) {
		assert.Equal(t, "bet with pairplus result", c.Exec("b 100 50"))
	})

	t.Run("bet long form", func(t *testing.T) {
		assert.Equal(t, "bet result", c.Exec("bet 100"))
	})
}

func TestThreeCardCuiController_Bet_Errors(t *testing.T) {
	m := newMockThreeCardInteractor()
	c := controller.NewThreeCardCuiController(m)

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

	t.Run("invalid pair plus", func(t *testing.T) {
		result := c.Exec("b 100 abc")
		assert.Contains(t, result, msgStem("invalidPairPlusAmount"))
	})
}

func TestThreeCardCuiController_Play(t *testing.T) {
	m := newMockThreeCardInteractor()
	c := controller.NewThreeCardCuiController(m)

	assert.Equal(t, "play result", c.Exec("p"))
	assert.Equal(t, "play result", c.Exec("play"))
}

func TestThreeCardCuiController_Fold(t *testing.T) {
	m := newMockThreeCardInteractor()
	c := controller.NewThreeCardCuiController(m)

	assert.Equal(t, "fold result", c.Exec("f"))
	assert.Equal(t, "fold result", c.Exec("fold"))
}

func TestThreeCardCuiController_ActionLog(t *testing.T) {
	m := newMockThreeCardInteractor()
	c := controller.NewThreeCardCuiController(m)

	assert.Equal(t, "action log result", c.Exec("log"))
	assert.Equal(t, "action log result", c.Exec("l"))
}

func TestThreeCardCuiController_Unknown(t *testing.T) {
	m := newMockThreeCardInteractor()
	c := controller.NewThreeCardCuiController(m)

	result := c.Exec("xyz")
	assert.Contains(t, result, "コマンドが不明です")
}

func TestThreeCardCuiController_Empty(t *testing.T) {
	m := newMockThreeCardInteractor()
	c := controller.NewThreeCardCuiController(m)

	result := c.Exec("")
	assert.Contains(t, result, "'help' でコマンド一覧を表示します。")
}

// #5513: Web はラウンド終了時にワンクリックで再ベットできるのに、CUI には同等の
// コマンドが無く毎ラウンド bet <ante> <pairPlus> を手打ちしていた。
func TestThreeCardCuiController_Rebet(t *testing.T) {
	for _, cmd := range []string{"rb", "rebet"} {
		m := new(usecase.MockThreeCardInteractor)
		m.On("Rebet").Return(`{"phase":1}`)
		c := controller.NewThreeCardCuiController(m)

		assert.Equal(t, `{"phase":1}`, c.Exec(cmd), cmd)
		m.AssertCalled(t, "Rebet")
	}
}

// **未知のコマンド扱いにしない。** 以前は rebet が「不明なコマンド」で弾かれていた。
func TestThreeCardCuiController_RebetIsARecognisedVerb(t *testing.T) {
	m := new(usecase.MockThreeCardInteractor)
	m.On("Rebet").Return(`{}`)
	c := controller.NewThreeCardCuiController(m)
	out := c.Exec("rebet")
	assert.NotContains(t, out, "Unsupported")
	assert.NotContains(t, out, "不明")
}
