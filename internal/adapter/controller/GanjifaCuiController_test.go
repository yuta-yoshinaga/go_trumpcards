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

func TestGanjifaCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockGanjifaInteractor {
		m := new(mockUsecases.MockGanjifaInteractor)
		m.On("GetConfig").Return(domain.DefaultGanjifaConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewGanjifaCuiController(newMock()).Exec("q"))
		assert.Equal(t, "bye.", controller.NewGanjifaCuiController(newMock()).Exec("quit"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewGanjifaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultGanjifaConfig())
	})

	t.Run("play card", func(t *testing.T) {
		m := newMock()
		c := controller.NewGanjifaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("play 3"))
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("play no args", func(t *testing.T) {
		result := controller.NewGanjifaCuiController(newMock()).Exec("play")
		assert.Contains(t, result, msgCardIndexRequired())
	})

	// Ganjifa has no bidding, so bid/pass must not silently resolve to something.
	t.Run("bid and pass are not commands here", func(t *testing.T) {
		for _, cmd := range []string{"bid 1", "pass"} {
			result := controller.NewGanjifaCuiController(newMock()).Exec(cmd)
			assert.Contains(t, result, "コマンドが不明です", "%q must not be accepted", cmd)
		}
	})

	t.Run("next / nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewGanjifaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextTrick")
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewGanjifaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultGanjifaConfig()
		expected.CpuDifficulty = domain.GanjifaCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty invalid", func(t *testing.T) {
		result := controller.NewGanjifaCuiController(newMock()).Exec("sd 9")
		assert.Contains(t, result, msgInvalidCpuDifficultyPrefix())
	})

	t.Run("hint / log", func(t *testing.T) {
		m := newMock()
		c := controller.NewGanjifaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "Hint")
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		result := controller.NewGanjifaCuiController(newMock()).Exec("zzz")
		assert.Contains(t, result, "コマンドが不明です")
	})
}
