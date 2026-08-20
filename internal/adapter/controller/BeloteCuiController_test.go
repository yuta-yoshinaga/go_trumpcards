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

func TestBeloteCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockBeloteInteractor {
		m := new(mockUsecases.MockBeloteInteractor)
		m.On("GetConfig").Return(domain.DefaultBeloteConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("PickUp", mock.Anything).Return(mockOutput)
		m.On("CallTrump", mock.Anything).Return(mockOutput)
		m.On("Pass").Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit q", func(t *testing.T) {
		c := controller.NewBeloteCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("quit long", func(t *testing.T) {
		c := controller.NewBeloteCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	t.Run("reset r preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewBeloteCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "GetConfig")
	})

	t.Run("orderup o", func(t *testing.T) {
		m := newMock()
		c := controller.NewBeloteCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("o"))
		m.AssertCalled(t, "PickUp", true)
	})

	t.Run("orderup long form", func(t *testing.T) {
		m := newMock()
		c := controller.NewBeloteCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("orderup"))
		m.AssertCalled(t, "PickUp", true)
	})

	t.Run("pass pa", func(t *testing.T) {
		m := newMock()
		c := controller.NewBeloteCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("pa"))
		m.AssertCalled(t, "Pass")
	})

	t.Run("call c with suit", func(t *testing.T) {
		m := newMock()
		c := controller.NewBeloteCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("c 2"))
		m.AssertCalled(t, "CallTrump", 2)
	})

	t.Run("call long form", func(t *testing.T) {
		m := newMock()
		c := controller.NewBeloteCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("call 3"))
		m.AssertCalled(t, "CallTrump", 3)
	})

	t.Run("call no args", func(t *testing.T) {
		c := controller.NewBeloteCuiController(newMock())
		assert.Contains(t, c.Exec("c"), msgStem("suitRequiredRange"))
	})

	t.Run("call invalid arg", func(t *testing.T) {
		c := controller.NewBeloteCuiController(newMock())
		assert.Contains(t, c.Exec("c abc"), msgStem("invalidSuit"))
	})

	t.Run("call out of range", func(t *testing.T) {
		c := controller.NewBeloteCuiController(newMock())
		assert.Contains(t, c.Exec("c 5"), msgKey("invalidSuit", "val", "5"))
	})

	t.Run("play p with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewBeloteCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 3"))
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("play no args", func(t *testing.T) {
		c := controller.NewBeloteCuiController(newMock())
		assert.Contains(t, c.Exec("p"), msgCardIndexRequired())
	})

	t.Run("play invalid arg", func(t *testing.T) {
		c := controller.NewBeloteCuiController(newMock())
		assert.Contains(t, c.Exec("p abc"), msgInvalidCardIndexPrefix())
	})

	t.Run("next n", func(t *testing.T) {
		m := newMock()
		c := controller.NewBeloteCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		m.AssertCalled(t, "NextTrick")
	})

	t.Run("nextround nr", func(t *testing.T) {
		m := newMock()
		c := controller.NewBeloteCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewBeloteCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultBeloteConfig()
		expected.CpuDifficulty = domain.BeloteCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty invalid", func(t *testing.T) {
		c := controller.NewBeloteCuiController(newMock())
		assert.Contains(t, c.Exec("sd abc"), msgInvalidCpuDifficultyPrefix())
	})

	t.Run("setdifficulty no args", func(t *testing.T) {
		c := controller.NewBeloteCuiController(newMock())
		assert.Contains(t, c.Exec("sd"), msgCpuDifficultyRequired())
	})

	t.Run("settarget valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewBeloteCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("st 500"))
		expected := domain.DefaultBeloteConfig()
		expected.TargetScore = 500
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("settarget no args", func(t *testing.T) {
		c := controller.NewBeloteCuiController(newMock())
		assert.Contains(t, c.Exec("st"), msgTargetScoreRequired())
	})

	t.Run("settarget invalid", func(t *testing.T) {
		c := controller.NewBeloteCuiController(newMock())
		assert.Contains(t, c.Exec("st 0"), msgInvalidTargetScore("0"))
	})

	t.Run("hint h", func(t *testing.T) {
		m := newMock()
		c := controller.NewBeloteCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		m.AssertCalled(t, "Hint")
	})

	t.Run("log", func(t *testing.T) {
		m := newMock()
		c := controller.NewBeloteCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewBeloteCuiController(newMock())
		assert.Contains(t, c.Exec("unknown"), "コマンドが不明")
	})

	t.Run("empty command", func(t *testing.T) {
		c := controller.NewBeloteCuiController(newMock())
		assert.Contains(t, c.Exec(""), "'help'")
	})
}
