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

func TestWattenCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockWattenInteractor {
		m := new(mockUsecases.MockWattenInteractor)
		m.On("GetConfig").Return(domain.DefaultWattenConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Declare", mock.Anything, mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("Raise").Return(mockOutput)
		m.On("Respond", mock.Anything).Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit q", func(t *testing.T) {
		c := controller.NewWattenCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("reset r preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewWattenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "GetConfig")
	})

	t.Run("declare d rank suit", func(t *testing.T) {
		m := newMock()
		c := controller.NewWattenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("d 10 3"))
		m.AssertCalled(t, "Declare", 10, 3)
	})

	t.Run("declare long form", func(t *testing.T) {
		m := newMock()
		c := controller.NewWattenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("declare 1 2"))
		m.AssertCalled(t, "Declare", 1, 2)
	})

	t.Run("declare missing args", func(t *testing.T) {
		c := controller.NewWattenCuiController(newMock())
		assert.Contains(t, c.Exec("d"), msgStem("rankAndSuitRequired"))
	})

	t.Run("declare bad suit", func(t *testing.T) {
		c := controller.NewWattenCuiController(newMock())
		assert.Contains(t, c.Exec("d 10 9"), msgStem("invalidSuit"))
	})

	t.Run("play p index", func(t *testing.T) {
		m := newMock()
		c := controller.NewWattenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 3"))
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("play no args", func(t *testing.T) {
		c := controller.NewWattenCuiController(newMock())
		assert.Contains(t, c.Exec("p"), msgCardIndexRequired())
	})

	t.Run("raise rz", func(t *testing.T) {
		m := newMock()
		c := controller.NewWattenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("rz"))
		m.AssertCalled(t, "Raise")
	})

	t.Run("respond hold", func(t *testing.T) {
		m := newMock()
		c := controller.NewWattenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("resp h"))
		m.AssertCalled(t, "Respond", true)
	})

	t.Run("respond fold", func(t *testing.T) {
		m := newMock()
		c := controller.NewWattenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("respond f"))
		m.AssertCalled(t, "Respond", false)
	})

	t.Run("respond missing", func(t *testing.T) {
		c := controller.NewWattenCuiController(newMock())
		assert.Contains(t, c.Exec("resp"), msgStem("responseRequiredHoldFold"))
	})

	t.Run("respond invalid", func(t *testing.T) {
		c := controller.NewWattenCuiController(newMock())
		assert.Contains(t, c.Exec("resp x"), "Invalid response")
	})

	t.Run("nextround nr", func(t *testing.T) {
		m := newMock()
		c := controller.NewWattenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewWattenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultWattenConfig()
		expected.CpuDifficulty = domain.WattenCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("settarget valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewWattenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("st 21"))
		expected := domain.DefaultWattenConfig()
		expected.TargetScore = 21
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("settarget invalid", func(t *testing.T) {
		c := controller.NewWattenCuiController(newMock())
		assert.Contains(t, c.Exec("st 0"), msgInvalidTargetScore("0"))
	})

	t.Run("hint h", func(t *testing.T) {
		m := newMock()
		c := controller.NewWattenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		m.AssertCalled(t, "Hint")
	})

	t.Run("log", func(t *testing.T) {
		m := newMock()
		c := controller.NewWattenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewWattenCuiController(newMock())
		assert.Contains(t, c.Exec("unknown"), "コマンドが不明")
	})

	t.Run("empty command", func(t *testing.T) {
		c := controller.NewWattenCuiController(newMock())
		assert.Contains(t, c.Exec(""), "'help'")
	})
}
