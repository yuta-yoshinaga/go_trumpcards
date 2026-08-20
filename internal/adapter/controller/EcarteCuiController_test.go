//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestEcarteCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`
	newMock := func() *mockUsecases.MockEcarteInteractor {
		m := new(mockUsecases.MockEcarteInteractor)
		m.On("GetConfig").Return(domain.DefaultEcarteConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Propose").Return(mockOutput)
		m.On("Stand").Return(mockOutput)
		m.On("Respond", mock.Anything).Return(mockOutput)
		m.On("Discard", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit short", func(t *testing.T) {
		c := controller.NewEcarteCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewEcarteCuiController(m)
		got := c.Exec("r")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultEcarteConfig())
	})

	t.Run("propose short", func(t *testing.T) {
		m := newMock()
		c := controller.NewEcarteCuiController(m)
		got := c.Exec("pr")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "Propose")
	})

	t.Run("propose long", func(t *testing.T) {
		m := newMock()
		c := controller.NewEcarteCuiController(m)
		got := c.Exec("propose")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "Propose")
	})

	t.Run("stand short", func(t *testing.T) {
		m := newMock()
		c := controller.NewEcarteCuiController(m)
		got := c.Exec("st")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "Stand")
	})

	t.Run("accept", func(t *testing.T) {
		m := newMock()
		c := controller.NewEcarteCuiController(m)
		got := c.Exec("a")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "Respond", true)
	})

	t.Run("refuse", func(t *testing.T) {
		m := newMock()
		c := controller.NewEcarteCuiController(m)
		got := c.Exec("rf")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "Respond", false)
	})

	t.Run("discard with indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewEcarteCuiController(m)
		got := c.Exec("d 0 2")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "Discard", []int{0, 2})
	})

	t.Run("discard with no indices stands pat", func(t *testing.T) {
		m := newMock()
		c := controller.NewEcarteCuiController(m)
		got := c.Exec("discard")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "Discard", []int{})
	})

	t.Run("play with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewEcarteCuiController(m)
		got := c.Exec("p 1")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "Play", 1)
	})

	t.Run("play missing index", func(t *testing.T) {
		c := controller.NewEcarteCuiController(newMock())
		assert.Contains(t, c.Exec("p"), msgCardIndexRequired())
	})

	t.Run("play invalid index", func(t *testing.T) {
		c := controller.NewEcarteCuiController(newMock())
		assert.Contains(t, c.Exec("p abc"), msgInvalidCardIndexPrefix())
	})

	t.Run("next short", func(t *testing.T) {
		m := newMock()
		c := controller.NewEcarteCuiController(m)
		got := c.Exec("n")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "NextRound")
	})

	t.Run("nextround long", func(t *testing.T) {
		m := newMock()
		c := controller.NewEcarteCuiController(m)
		got := c.Exec("nextround")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewEcarteCuiController(m)
		got := c.Exec("sd 2")
		assert.Equal(t, mockOutput, got)
		expected := domain.DefaultEcarteConfig()
		expected.CpuDifficulty = domain.EcarteCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("settarget", func(t *testing.T) {
		m := newMock()
		c := controller.NewEcarteCuiController(m)
		got := c.Exec("tg 10")
		assert.Equal(t, mockOutput, got)
		expected := domain.DefaultEcarteConfig()
		expected.TargetScore = 10
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("hint short", func(t *testing.T) {
		m := newMock()
		c := controller.NewEcarteCuiController(m)
		got := c.Exec("h")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "Hint")
	})

	t.Run("log short", func(t *testing.T) {
		m := newMock()
		c := controller.NewEcarteCuiController(m)
		got := c.Exec("l")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewEcarteCuiController(newMock())
		got := c.Exec("xyz")
		assert.NotEqual(t, "bye.", got)
		assert.NotEmpty(t, got)
	})

	// **落として残りで実行しない。** 打ち間違いを捨てると、プレイヤーが
	// 選んでいない組み合わせが実行される (issue #5390)。
	t.Run("refuses a mistyped index", func(t *testing.T) {
		m := newMock()
		c := controller.NewEcarteCuiController(m)
		assert.Contains(t, c.Exec("d 0 zz"), msgInvalidCardIndexPrefix(),
			"a mistyped index must be refused, not dropped")
	})
}
