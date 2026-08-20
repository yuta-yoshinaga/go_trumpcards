package controller_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBourreCuiController_Exec(t *testing.T) {
	mockOutput := `{"players":[]}`

	newMock := func() *mockUsecases.MockBourreInteractor {
		m := new(mockUsecases.MockBourreInteractor)
		m.On("Reset").Return(mockOutput)
		m.On("Decide", mock.Anything).Return(mockOutput)
		m.On("Draw", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextHand").Return(mockOutput)
		m.On("GetConfig").Return(domain.DefaultBourreConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("ActionLog").Return(`{"entries":[]}`)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		c := controller.NewBourreCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})
	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewBourreCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("decide play", func(t *testing.T) {
		m := newMock()
		c := controller.NewBourreCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("d 1"))
		m.AssertCalled(t, "Decide", true)
	})
	t.Run("decide fold", func(t *testing.T) {
		m := newMock()
		c := controller.NewBourreCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("decide 0"))
		m.AssertCalled(t, "Decide", false)
	})
	t.Run("decide invalid", func(t *testing.T) {
		c := controller.NewBourreCuiController(newMock())
		assert.Contains(t, c.Exec("d 9"), msgStem("invalidDecision0Or1"))
	})
	t.Run("draw indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewBourreCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("dr 0 2"))
		m.AssertCalled(t, "Draw", []int{0, 2})
	})
	t.Run("play index", func(t *testing.T) {
		m := newMock()
		c := controller.NewBourreCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 1"))
		m.AssertCalled(t, "Play", 1)
	})
	t.Run("next", func(t *testing.T) {
		m := newMock()
		c := controller.NewBourreCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		m.AssertCalled(t, "NextHand")
	})
	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewBourreCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 1"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("setdifficulty invalid", func(t *testing.T) {
		c := controller.NewBourreCuiController(newMock())
		assert.Contains(t, c.Exec("sd 9"), msgInvalidCpuDifficultyPrefix())
	})
	t.Run("log", func(t *testing.T) {
		m := newMock()
		c := controller.NewBourreCuiController(m)
		assert.Equal(t, `{"entries":[]}`, c.Exec("log"))
	})

	// **落として残りで実行しない。** 打ち間違いを捨てると、プレイヤーが
	// 選んでいない組み合わせが実行される (issue #5390)。
	t.Run("refuses a mistyped index", func(t *testing.T) {
		m := newMock()
		c := controller.NewBourreCuiController(m)
		assert.Contains(t, c.Exec("dr 0 zz"), msgInvalidCardIndexPrefix(),
			"a mistyped index must be refused, not dropped")
	})
}
