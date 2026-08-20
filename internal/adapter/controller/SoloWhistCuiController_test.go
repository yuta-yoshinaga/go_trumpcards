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

func TestSoloWhistCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockSoloWhistInteractor {
		m := new(mockUsecases.MockSoloWhistInteractor)
		m.On("GetConfig").Return(domain.DefaultSoloWhistConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Bid", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewSoloWhistCuiController(newMock()).Exec("q"))
		assert.Equal(t, "bye.", controller.NewSoloWhistCuiController(newMock()).Exec("quit"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewSoloWhistCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultSoloWhistConfig())
	})

	t.Run("bid solo", func(t *testing.T) {
		m := newMock()
		c := controller.NewSoloWhistCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("bid 1"))
		m.AssertCalled(t, "Bid", 1)
	})

	t.Run("bid no args", func(t *testing.T) {
		result := controller.NewSoloWhistCuiController(newMock()).Exec("bid")
		assert.Contains(t, result, msgStem("bidRequiredAbundance"))
	})

	t.Run("bid invalid", func(t *testing.T) {
		result := controller.NewSoloWhistCuiController(newMock()).Exec("bid 9")
		assert.Contains(t, result, msgStem("invalidBid03"))
	})

	t.Run("pass maps to bid 0", func(t *testing.T) {
		m := newMock()
		c := controller.NewSoloWhistCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("pass"))
		m.AssertCalled(t, "Bid", 0)
	})

	t.Run("play card", func(t *testing.T) {
		m := newMock()
		c := controller.NewSoloWhistCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("play 3"))
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("play no args", func(t *testing.T) {
		result := controller.NewSoloWhistCuiController(newMock()).Exec("play")
		assert.Contains(t, result, msgCardIndexRequired())
	})

	t.Run("next / nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewSoloWhistCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextTrick")
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewSoloWhistCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultSoloWhistConfig()
		expected.CpuDifficulty = domain.SoloWhistCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty invalid", func(t *testing.T) {
		result := controller.NewSoloWhistCuiController(newMock()).Exec("sd 9")
		assert.Contains(t, result, msgInvalidCpuDifficultyPrefix())
	})

	t.Run("hint / log", func(t *testing.T) {
		m := newMock()
		c := controller.NewSoloWhistCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "Hint")
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		result := controller.NewSoloWhistCuiController(newMock()).Exec("zzz")
		assert.Contains(t, result, "コマンドが不明です")
	})
}
