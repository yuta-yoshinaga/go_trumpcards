package controller_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestCassinoCuiController_Exec(t *testing.T) {
	mockOutput := `{"players":[]}`

	newMock := func() *mockUsecases.MockCassinoInteractor {
		m := new(mockUsecases.MockCassinoInteractor)
		m.On("Reset").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Take", mock.Anything, mock.Anything, mock.Anything).Return(mockOutput)
		m.On("Build", mock.Anything, mock.Anything, mock.Anything).Return(mockOutput)
		m.On("Trail", mock.Anything).Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("GetConfig").Return(domain.DefaultCassinoConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("ActionLog").Return("log")
		return m
	}

	t.Run("quit command", func(t *testing.T) {
		c := controller.NewCassinoCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewCassinoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("next round", func(t *testing.T) {
		m := newMock()
		c := controller.NewCassinoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("hint command", func(t *testing.T) {
		m := newMock()
		c := controller.NewCassinoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("hint"))
		m.AssertCalled(t, "Hint")
	})

	t.Run("take command", func(t *testing.T) {
		m := newMock()
		c := controller.NewCassinoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("t 0 1 2"))
		m.AssertCalled(t, "Take", 0, []int{1, 2}, []int{})
	})

	t.Run("take with build capture", func(t *testing.T) {
		m := newMock()
		c := controller.NewCassinoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("t 0 b 1"))
		m.AssertCalled(t, "Take", 0, []int{}, []int{1})
	})

	t.Run("take missing hand", func(t *testing.T) {
		m := newMock()
		c := controller.NewCassinoCuiController(m)
		out := c.Exec("t")
		assert.Contains(t, out, "Usage:")
	})

	t.Run("build command", func(t *testing.T) {
		m := newMock()
		c := controller.NewCassinoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("b 0 8 1"))
		m.AssertCalled(t, "Build", 0, []int{1}, 8)
	})

	t.Run("build bad value", func(t *testing.T) {
		m := newMock()
		c := controller.NewCassinoCuiController(m)
		out := c.Exec("b 0 xyz 1")
		assert.Contains(t, out, "Invalid")
	})

	t.Run("trail command", func(t *testing.T) {
		m := newMock()
		c := controller.NewCassinoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("tr 2"))
		m.AssertCalled(t, "Trail", 2)
	})

	t.Run("trail missing arg", func(t *testing.T) {
		m := newMock()
		c := controller.NewCassinoCuiController(m)
		out := c.Exec("tr")
		assert.Contains(t, out, "Usage:")
	})

	t.Run("sd (difficulty)", func(t *testing.T) {
		m := newMock()
		c := controller.NewCassinoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		m.AssertCalled(t, "ResetWithConfig", mock.MatchedBy(func(cfg domain.CassinoConfig) bool {
			return cfg.CpuDifficulty == domain.CassinoDifficultyHard
		}))
	})

	t.Run("sr list", func(t *testing.T) {
		m := newMock()
		c := controller.NewCassinoCuiController(m)
		out := c.Exec("sr list")
		assert.True(t, strings.Contains(out, "multibuild") || strings.Contains(out, "sweepbonus"))
	})

	t.Run("sr toggle", func(t *testing.T) {
		m := newMock()
		c := controller.NewCassinoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sr multibuild 0"))
		m.AssertCalled(t, "ResetWithConfig", mock.MatchedBy(func(cfg domain.CassinoConfig) bool {
			return !cfg.MultiBuildEnabled
		}))
	})

	t.Run("sr usage missing args", func(t *testing.T) {
		m := newMock()
		c := controller.NewCassinoCuiController(m)
		out := c.Exec("sr")
		assert.Contains(t, out, "Usage:")
	})

	t.Run("sr unknown rule", func(t *testing.T) {
		m := newMock()
		c := controller.NewCassinoCuiController(m)
		out := c.Exec("sr garbage 1")
		assert.Contains(t, out, "Unknown rule")
	})

	t.Run("sr bad value", func(t *testing.T) {
		m := newMock()
		c := controller.NewCassinoCuiController(m)
		out := c.Exec("sr multibuild 2")
		assert.Contains(t, out, "Invalid value")
	})

	t.Run("log command", func(t *testing.T) {
		m := newMock()
		c := controller.NewCassinoCuiController(m)
		assert.Equal(t, "log", c.Exec("log"))
	})

	// **落として残りで実行しない。** 打ち間違いを捨てると、プレイヤーが
	// 選んでいない組み合わせが実行される (issue #5390)。
	t.Run("refuses a mistyped index", func(t *testing.T) {
		m := newMock()
		c := controller.NewCassinoCuiController(m)
		assert.Contains(t, c.Exec("t 0 zz"), msgInvalidCardIndexPrefix(),
			"a mistyped index must be refused, not dropped")
		assert.Contains(t, c.Exec("t 0 b zz"), msgInvalidCardIndexPrefix(),
			"a mistyped index must be refused, not dropped")
	})
}
