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

func TestSheepsheadCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockSheepsheadInteractor {
		m := new(mockUsecases.MockSheepsheadInteractor)
		m.On("GetConfig").Return(domain.DefaultSheepsheadConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Pick", mock.Anything).Return(mockOutput)
		m.On("Bury", mock.Anything).Return(mockOutput)
		m.On("Call", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewSheepsheadCuiController(newMock()).Exec("q"))
		assert.Equal(t, "bye.", controller.NewSheepsheadCuiController(newMock()).Exec("quit"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewSheepsheadCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultSheepsheadConfig())
	})

	t.Run("pick", func(t *testing.T) {
		m := newMock()
		c := controller.NewSheepsheadCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("pick"))
		m.AssertCalled(t, "Pick", true)
	})

	t.Run("pick shorthand p", func(t *testing.T) {
		m := newMock()
		c := controller.NewSheepsheadCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p"))
		m.AssertCalled(t, "Pick", true)
	})

	t.Run("pass", func(t *testing.T) {
		m := newMock()
		c := controller.NewSheepsheadCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("pass"))
		m.AssertCalled(t, "Pick", false)
	})

	t.Run("bury two cards", func(t *testing.T) {
		m := newMock()
		c := controller.NewSheepsheadCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("b 0 2"))
		m.AssertCalled(t, "Bury", []int{0, 2})
	})

	t.Run("bury alias", func(t *testing.T) {
		m := newMock()
		c := controller.NewSheepsheadCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("bury 1 3"))
		m.AssertCalled(t, "Bury", []int{1, 3})
	})

	t.Run("bury missing args", func(t *testing.T) {
		result := controller.NewSheepsheadCuiController(newMock()).Exec("b 0")
		assert.Contains(t, result, "Usage")
	})

	t.Run("bury no args", func(t *testing.T) {
		result := controller.NewSheepsheadCuiController(newMock()).Exec("b")
		assert.Contains(t, result, "Usage")
	})

	t.Run("call suit", func(t *testing.T) {
		m := newMock()
		c := controller.NewSheepsheadCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("c 2"))
		m.AssertCalled(t, "Call", 2)
	})

	t.Run("call alias", func(t *testing.T) {
		m := newMock()
		c := controller.NewSheepsheadCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("call 1"))
		m.AssertCalled(t, "Call", 1)
	})

	t.Run("call no args", func(t *testing.T) {
		result := controller.NewSheepsheadCuiController(newMock()).Exec("c")
		assert.Contains(t, result, msgStem("suitRequiredThree"))
	})

	t.Run("call invalid suit", func(t *testing.T) {
		result := controller.NewSheepsheadCuiController(newMock()).Exec("c 9")
		assert.Contains(t, result, msgStem("invalidSuitThree"))
	})

	t.Run("play card", func(t *testing.T) {
		m := newMock()
		c := controller.NewSheepsheadCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("play 3"))
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("play no args", func(t *testing.T) {
		result := controller.NewSheepsheadCuiController(newMock()).Exec("play")
		assert.Contains(t, result, msgCardIndexRequired())
	})

	t.Run("next / nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewSheepsheadCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextTrick")
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewSheepsheadCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultSheepsheadConfig()
		expected.CpuDifficulty = domain.SheepsheadCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty invalid", func(t *testing.T) {
		result := controller.NewSheepsheadCuiController(newMock()).Exec("sd 9")
		assert.Contains(t, result, msgInvalidCpuDifficultyPrefix())
	})

	t.Run("setchips", func(t *testing.T) {
		m := newMock()
		c := controller.NewSheepsheadCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sb 5"))
		expected := domain.DefaultSheepsheadConfig()
		expected.BaseChips = 5
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setchips invalid", func(t *testing.T) {
		result := controller.NewSheepsheadCuiController(newMock()).Exec("sb 0")
		assert.Contains(t, result, "Invalid base chips")
	})

	t.Run("hint / log", func(t *testing.T) {
		m := newMock()
		c := controller.NewSheepsheadCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "Hint")
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		result := controller.NewSheepsheadCuiController(newMock()).Exec("zzz")
		assert.Contains(t, result, "コマンドが不明です")
	})
}
