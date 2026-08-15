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

func TestThirtyOneCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockThirtyOneInteractor {
		m := new(mockUsecases.MockThirtyOneInteractor)
		m.On("GetConfig").Return(domain.DefaultThirtyOneConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("DrawFromStock").Return(mockOutput)
		m.On("DrawFromDiscard").Return(mockOutput)
		m.On("Discard", mock.Anything).Return(mockOutput)
		m.On("Knock").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		return m
	}

	// **h/hint がコントローラまで通っていること (#4806)。**Barbu / Macau と同じ配線。
	t.Run("hint command", func(t *testing.T) {
		m := newMock()
		c := controller.NewThirtyOneCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("hint"))
		m.AssertCalled(t, "Hint")
	})

	t.Run("quit", func(t *testing.T) {
		c := controller.NewThirtyOneCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	t.Run("reset", func(t *testing.T) {
		m := newMock()
		c := controller.NewThirtyOneCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		assert.Equal(t, mockOutput, c.Exec("reset"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultThirtyOneConfig())
	})

	t.Run("draw and discard", func(t *testing.T) {
		m := newMock()
		c := controller.NewThirtyOneCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("ds"))
		assert.Equal(t, mockOutput, c.Exec("dd"))
		assert.Equal(t, mockOutput, c.Exec("d 1"))
		m.AssertCalled(t, "DrawFromStock")
		m.AssertCalled(t, "DrawFromDiscard")
		m.AssertCalled(t, "Discard", 1)
	})

	t.Run("discard no args", func(t *testing.T) {
		c := controller.NewThirtyOneCuiController(newMock())
		assert.Contains(t, c.Exec("d"), msgCardIndexRequired())
		assert.Contains(t, c.Exec("d abc"), msgInvalidCardIndexPrefix())
	})

	t.Run("knock takes no index", func(t *testing.T) {
		m := newMock()
		c := controller.NewThirtyOneCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("k"))
		assert.Equal(t, mockOutput, c.Exec("knock"))
		m.AssertCalled(t, "Knock")
	})

	t.Run("nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewThirtyOneCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("nr"))
		assert.Equal(t, mockOutput, c.Exec("nextround"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewThirtyOneCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultThirtyOneConfig()
		expected.CpuDifficulty = domain.ThirtyOneCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty errors", func(t *testing.T) {
		c := controller.NewThirtyOneCuiController(newMock())
		assert.Contains(t, c.Exec("sd"), "required")
		assert.Contains(t, c.Exec("sd abc"), "Invalid CPU difficulty")
		assert.Contains(t, c.Exec("sd 9"), "Invalid CPU difficulty")
	})

	t.Run("setlives valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewThirtyOneCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sv 5"))
		expected := domain.DefaultThirtyOneConfig()
		expected.InitialLives = 5
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setlives errors", func(t *testing.T) {
		c := controller.NewThirtyOneCuiController(newMock())
		assert.Contains(t, c.Exec("sv"), "required")
		assert.Contains(t, c.Exec("sv abc"), "Invalid lives")
		assert.Contains(t, c.Exec("sv 0"), "Invalid lives")
	})

	t.Run("log", func(t *testing.T) {
		m := newMock()
		c := controller.NewThirtyOneCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("log"))
		assert.Equal(t, mockOutput, c.Exec("l"))
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewThirtyOneCuiController(newMock())
		assert.Contains(t, c.Exec("unknown"), "コマンドが不明です")
	})

	t.Run("empty command", func(t *testing.T) {
		c := controller.NewThirtyOneCuiController(newMock())
		assert.Contains(t, c.Exec(""), "'help' でコマンド一覧を表示します。")
	})
}
