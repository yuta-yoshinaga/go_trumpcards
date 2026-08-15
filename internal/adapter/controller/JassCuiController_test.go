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

func TestJassCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockJassInteractor {
		m := new(mockUsecases.MockJassInteractor)
		m.On("GetConfig").Return(domain.DefaultJassConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("ChooseTrump", mock.Anything).Return(mockOutput)
		m.On("Schieben").Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit q", func(t *testing.T) {
		c := controller.NewJassCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("reset r preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewJassCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "GetConfig")
	})

	t.Run("call c with suit", func(t *testing.T) {
		m := newMock()
		c := controller.NewJassCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("c 2"))
		m.AssertCalled(t, "ChooseTrump", 2)
	})

	t.Run("call long form", func(t *testing.T) {
		m := newMock()
		c := controller.NewJassCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("call 3"))
		m.AssertCalled(t, "ChooseTrump", 3)
	})

	t.Run("call no args", func(t *testing.T) {
		c := controller.NewJassCuiController(newMock())
		assert.Contains(t, c.Exec("c"), "Suit is required")
	})

	t.Run("call out of range", func(t *testing.T) {
		c := controller.NewJassCuiController(newMock())
		assert.Contains(t, c.Exec("c 5"), "Invalid suit: 5")
	})

	t.Run("schieben sc", func(t *testing.T) {
		m := newMock()
		c := controller.NewJassCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sc"))
		m.AssertCalled(t, "Schieben")
	})

	t.Run("schieben long", func(t *testing.T) {
		m := newMock()
		c := controller.NewJassCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("schieben"))
		m.AssertCalled(t, "Schieben")
	})

	t.Run("play p with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewJassCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 3"))
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("play no args", func(t *testing.T) {
		c := controller.NewJassCuiController(newMock())
		assert.Contains(t, c.Exec("p"), msgCardIndexRequired())
	})

	t.Run("next n", func(t *testing.T) {
		m := newMock()
		c := controller.NewJassCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		m.AssertCalled(t, "NextTrick")
	})

	t.Run("nextround nr", func(t *testing.T) {
		m := newMock()
		c := controller.NewJassCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewJassCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultJassConfig()
		expected.CpuDifficulty = domain.JassCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("settarget valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewJassCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("st 500"))
		expected := domain.DefaultJassConfig()
		expected.TargetScore = 500
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("settarget invalid", func(t *testing.T) {
		c := controller.NewJassCuiController(newMock())
		assert.Contains(t, c.Exec("st 0"), "Invalid target score: 0")
	})

	t.Run("hint h", func(t *testing.T) {
		m := newMock()
		c := controller.NewJassCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		m.AssertCalled(t, "Hint")
	})

	t.Run("log", func(t *testing.T) {
		m := newMock()
		c := controller.NewJassCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewJassCuiController(newMock())
		assert.Contains(t, c.Exec("unknown"), "コマンドが不明")
	})

	t.Run("empty command", func(t *testing.T) {
		c := controller.NewJassCuiController(newMock())
		assert.Contains(t, c.Exec(""), "'help'")
	})
}
