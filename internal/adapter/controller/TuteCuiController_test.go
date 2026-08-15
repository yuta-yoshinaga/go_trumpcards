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

func TestTuteCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockTuteInteractor {
		m := new(mockUsecases.MockTuteInteractor)
		m.On("GetConfig").Return(domain.DefaultTuteConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("DeclareMarriage", mock.Anything).Return(mockOutput)
		m.On("DeclareTute").Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewTuteCuiController(newMock()).Exec("q"))
		assert.Equal(t, "bye.", controller.NewTuteCuiController(newMock()).Exec("quit"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewTuteCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultTuteConfig())
	})

	t.Run("play card", func(t *testing.T) {
		m := newMock()
		c := controller.NewTuteCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("play 3"))
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("play no args", func(t *testing.T) {
		result := controller.NewTuteCuiController(newMock()).Exec("play")
		assert.Contains(t, result, msgCardIndexRequired())
	})

	t.Run("marriage shorthand m", func(t *testing.T) {
		m := newMock()
		c := controller.NewTuteCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("m 1"))
		m.AssertCalled(t, "DeclareMarriage", 1)
	})

	t.Run("marriage full", func(t *testing.T) {
		m := newMock()
		c := controller.NewTuteCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("marriage 3"))
		m.AssertCalled(t, "DeclareMarriage", 3)
	})

	t.Run("marriage invalid suit", func(t *testing.T) {
		result := controller.NewTuteCuiController(newMock()).Exec("m 9")
		assert.Contains(t, result, msgStem("invalidSuitRange"))
	})

	t.Run("marriage no args", func(t *testing.T) {
		result := controller.NewTuteCuiController(newMock()).Exec("m")
		assert.Contains(t, result, msgStem("suitRequiredSymbolsPlain"))
	})

	t.Run("tute declaration", func(t *testing.T) {
		m := newMock()
		c := controller.NewTuteCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("tute"))
		m.AssertCalled(t, "DeclareTute")
	})

	t.Run("next / nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewTuteCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextTrick")
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewTuteCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultTuteConfig()
		expected.CpuDifficulty = domain.TuteCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty invalid", func(t *testing.T) {
		result := controller.NewTuteCuiController(newMock()).Exec("sd 9")
		assert.Contains(t, result, msgInvalidCpuDifficultyPrefix())
	})

	t.Run("hint / log", func(t *testing.T) {
		m := newMock()
		c := controller.NewTuteCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "Hint")
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		result := controller.NewTuteCuiController(newMock()).Exec("zzz")
		assert.Contains(t, result, "コマンドが不明です")
	})
}
