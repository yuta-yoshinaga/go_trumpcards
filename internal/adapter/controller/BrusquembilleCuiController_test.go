//go:build test && (!js || !wasm || classic)

package controller_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestBrusquembilleCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`
	newMock := func() *mockUsecases.MockBrusquembilleInteractor {
		m := new(mockUsecases.MockBrusquembilleInteractor)
		m.On("GetConfig").Return(domain.DefaultBrusquembilleConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit short", func(t *testing.T) {
		c := controller.NewBrusquembilleCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("quit long", func(t *testing.T) {
		c := controller.NewBrusquembilleCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewBrusquembilleCuiController(m)
		got := c.Exec("r")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultBrusquembilleConfig())
	})

	t.Run("play with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewBrusquembilleCuiController(m)
		got := c.Exec("p 1")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "Play", 1)
	})

	t.Run("play missing index", func(t *testing.T) {
		c := controller.NewBrusquembilleCuiController(newMock())
		got := c.Exec("p")
		assert.Contains(t, got, msgCardIndexRequired())
	})

	t.Run("play invalid index", func(t *testing.T) {
		c := controller.NewBrusquembilleCuiController(newMock())
		got := c.Exec("p abc")
		assert.Contains(t, got, msgInvalidCardIndexPrefix())
	})

	t.Run("next short", func(t *testing.T) {
		m := newMock()
		c := controller.NewBrusquembilleCuiController(m)
		got := c.Exec("n")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "NextTrick")
	})

	t.Run("hint short", func(t *testing.T) {
		m := newMock()
		c := controller.NewBrusquembilleCuiController(m)
		got := c.Exec("h")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "Hint")
	})

	t.Run("log short", func(t *testing.T) {
		m := newMock()
		c := controller.NewBrusquembilleCuiController(m)
		got := c.Exec("l")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewBrusquembilleCuiController(newMock())
		got := c.Exec("xyz")
		assert.NotEqual(t, "bye.", got)
		assert.NotEmpty(t, got)
	})
}

// TestBrusquembilleCuiController_SetPlayers は、席数を変えるコマンドが
// **実際にドメインまで届く**ことを見る。
//
// **入口が無ければ 2〜5 人卓は誰にも届かない。** ドメインを席数可変にしても、
// コマンドもフォームも無ければ既定の 2 人卓しか始まらず、可変にした意味が無い。
func TestBrusquembilleCuiController_SetPlayers(t *testing.T) {
	for _, n := range []int{2, 3, 4, 5} {
		m := new(mockUsecases.MockBrusquembilleInteractor)
		m.On("GetConfig").Return(domain.DefaultBrusquembilleConfig())
		m.On("ResetWithConfig", mock.MatchedBy(func(cfg domain.BrusquembilleConfig) bool {
			return cfg.PlayerCnt == n
		})).Return("ok")
		c := controller.NewBrusquembilleCuiController(m)

		assert.Equal(t, "ok", c.Exec("sp "+strconv.Itoa(n)), "sp %d が届かない", n)
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	}
}

// TestBrusquembilleCuiController_SetPlayersRejectsOutOfRange は負のコントロール。
// **どんな数でも通す実装だと、上のテストは通ってしまう。**
func TestBrusquembilleCuiController_SetPlayersRejectsOutOfRange(t *testing.T) {
	for _, n := range []string{"1", "6", "0", "-1"} {
		m := new(mockUsecases.MockBrusquembilleInteractor)
		m.On("GetConfig").Return(domain.DefaultBrusquembilleConfig()).Maybe()
		m.On("ResetWithConfig", mock.Anything).Return("ok").Maybe()
		c := controller.NewBrusquembilleCuiController(m)

		c.Exec("sp " + n)
		m.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
	}
}
