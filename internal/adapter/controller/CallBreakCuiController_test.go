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

func TestCallBreakCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockCallBreakInteractor {
		m := new(mockUsecases.MockCallBreakInteractor)
		m.On("GetConfig").Return(domain.DefaultCallBreakConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Bid", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit q", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewCallBreakCuiController(newMock()).Exec("q"))
	})
	t.Run("quit quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewCallBreakCuiController(newMock()).Exec("quit"))
	})

	t.Run("reset r", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCallBreakCuiController(m).Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultCallBreakConfig())
	})

	t.Run("bid valid", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCallBreakCuiController(m).Exec("b 3"))
		m.AssertCalled(t, "Bid", 3)
	})
	t.Run("bid long form", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCallBreakCuiController(m).Exec("bid 5"))
		m.AssertCalled(t, "Bid", 5)
	})
	t.Run("bid no args", func(t *testing.T) {
		assert.Contains(t, controller.NewCallBreakCuiController(newMock()).Exec("b"), "Bid value is required")
	})
	t.Run("bid invalid arg", func(t *testing.T) {
		assert.Contains(t, controller.NewCallBreakCuiController(newMock()).Exec("b abc"), "Invalid bid value")
	})
	t.Run("bid below min", func(t *testing.T) {
		assert.Contains(t, controller.NewCallBreakCuiController(newMock()).Exec("b 0"), "Invalid bid value: 0")
	})
	t.Run("bid above max", func(t *testing.T) {
		assert.Contains(t, controller.NewCallBreakCuiController(newMock()).Exec("b 14"), "Invalid bid value: 14")
	})

	t.Run("play valid", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCallBreakCuiController(m).Exec("p 2"))
		m.AssertCalled(t, "Play", 2)
	})
	t.Run("play long form", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCallBreakCuiController(m).Exec("play 5"))
		m.AssertCalled(t, "Play", 5)
	})
	t.Run("play no args", func(t *testing.T) {
		assert.Contains(t, controller.NewCallBreakCuiController(newMock()).Exec("p"), msgCardIndexRequired())
	})
	t.Run("play invalid arg", func(t *testing.T) {
		assert.Contains(t, controller.NewCallBreakCuiController(newMock()).Exec("p abc"), msgInvalidCardIndexPrefix())
	})

	t.Run("next n", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCallBreakCuiController(m).Exec("n"))
		m.AssertCalled(t, "NextTrick")
	})
	t.Run("nextround nr", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCallBreakCuiController(m).Exec("nr"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty valid", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCallBreakCuiController(m).Exec("sd 2"))
		expected := domain.DefaultCallBreakConfig()
		expected.CpuDifficulty = domain.CallBreakCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})
	t.Run("setdifficulty invalid", func(t *testing.T) {
		assert.Contains(t, controller.NewCallBreakCuiController(newMock()).Exec("sd 5"), "Invalid CPU difficulty")
	})
	t.Run("setdifficulty no args", func(t *testing.T) {
		assert.Contains(t, controller.NewCallBreakCuiController(newMock()).Exec("sd"), "required")
	})

	t.Run("setrounds valid", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCallBreakCuiController(m).Exec("sr 7"))
		expected := domain.DefaultCallBreakConfig()
		expected.MaxRounds = 7
		m.AssertCalled(t, "ResetWithConfig", expected)
	})
	t.Run("setrounds zero is invalid", func(t *testing.T) {
		assert.Contains(t, controller.NewCallBreakCuiController(newMock()).Exec("sr 0"), "Invalid round count")
	})
	t.Run("setrounds long form", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCallBreakCuiController(m).Exec("setrounds 10"))
		expected := domain.DefaultCallBreakConfig()
		expected.MaxRounds = 10
		m.AssertCalled(t, "ResetWithConfig", expected)
	})
	t.Run("setrounds no args", func(t *testing.T) {
		assert.Contains(t, controller.NewCallBreakCuiController(newMock()).Exec("sr"), "required")
	})

	t.Run("log", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCallBreakCuiController(m).Exec("log"))
		m.AssertCalled(t, "ActionLog")
	})
	t.Run("hint", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCallBreakCuiController(m).Exec("h"))
		m.AssertCalled(t, "Hint")
	})

	t.Run("unknown command", func(t *testing.T) {
		assert.Contains(t, controller.NewCallBreakCuiController(newMock()).Exec("unknown"), "コマンドが不明です")
	})
	t.Run("empty command", func(t *testing.T) {
		assert.Contains(t, controller.NewCallBreakCuiController(newMock()).Exec(""), "'help' でコマンド一覧を表示します。")
	})
}
