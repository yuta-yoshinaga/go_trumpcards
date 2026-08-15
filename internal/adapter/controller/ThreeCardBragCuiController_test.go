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

func TestThreeCardBragCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockThreeCardBragInteractor {
		m := new(mockUsecases.MockThreeCardBragInteractor)
		m.On("GetConfig").Return(domain.DefaultThreeCardBragConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("See").Return(mockOutput)
		m.On("Bet").Return(mockOutput)
		m.On("Raise", mock.Anything).Return(mockOutput)
		m.On("Fold").Return(mockOutput)
		m.On("Show").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewThreeCardBragCuiController(newMock()).Exec("q"))
		assert.Equal(t, "bye.", controller.NewThreeCardBragCuiController(newMock()).Exec("quit"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewThreeCardBragCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultThreeCardBragConfig())
	})

	t.Run("see and alias", func(t *testing.T) {
		m := newMock()
		c := controller.NewThreeCardBragCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("s"))
		assert.Equal(t, mockOutput, c.Exec("see"))
		m.AssertCalled(t, "See")
	})

	t.Run("bet and alias", func(t *testing.T) {
		m := newMock()
		c := controller.NewThreeCardBragCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("b"))
		assert.Equal(t, mockOutput, c.Exec("bet"))
		m.AssertCalled(t, "Bet")
	})

	t.Run("raise with amount", func(t *testing.T) {
		m := newMock()
		c := controller.NewThreeCardBragCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("rs 4"))
		m.AssertCalled(t, "Raise", 4)
	})

	t.Run("raise alias", func(t *testing.T) {
		m := newMock()
		c := controller.NewThreeCardBragCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("raise 6"))
		m.AssertCalled(t, "Raise", 6)
	})

	t.Run("raise missing amount", func(t *testing.T) {
		result := controller.NewThreeCardBragCuiController(newMock()).Exec("rs")
		assert.Contains(t, result, "Stake is required")
	})

	t.Run("fold and alias", func(t *testing.T) {
		m := newMock()
		c := controller.NewThreeCardBragCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("f"))
		assert.Equal(t, mockOutput, c.Exec("fold"))
		m.AssertCalled(t, "Fold")
	})

	t.Run("show and alias", func(t *testing.T) {
		m := newMock()
		c := controller.NewThreeCardBragCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sh"))
		assert.Equal(t, mockOutput, c.Exec("show"))
		m.AssertCalled(t, "Show")
	})

	t.Run("next and aliases", func(t *testing.T) {
		m := newMock()
		c := controller.NewThreeCardBragCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		assert.Equal(t, mockOutput, c.Exec("next"))
		assert.Equal(t, mockOutput, c.Exec("nextround"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewThreeCardBragCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultThreeCardBragConfig()
		expected.CpuDifficulty = domain.ThreeCardBragCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty invalid", func(t *testing.T) {
		result := controller.NewThreeCardBragCuiController(newMock()).Exec("sd 9")
		assert.Contains(t, result, msgInvalidCpuDifficultyPrefix())
	})

	t.Run("setante", func(t *testing.T) {
		m := newMock()
		c := controller.NewThreeCardBragCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sa 3"))
		expected := domain.DefaultThreeCardBragConfig()
		expected.Ante = 3
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setchips", func(t *testing.T) {
		m := newMock()
		c := controller.NewThreeCardBragCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sc 50"))
		expected := domain.DefaultThreeCardBragConfig()
		expected.StartingChips = 50
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("hint / log", func(t *testing.T) {
		m := newMock()
		c := controller.NewThreeCardBragCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("l"))
		m.AssertCalled(t, "Hint")
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		result := controller.NewThreeCardBragCuiController(newMock()).Exec("zzz")
		assert.Contains(t, result, "コマンドが不明です")
	})
}
