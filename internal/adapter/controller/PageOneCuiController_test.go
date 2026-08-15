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

func TestPageOneCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockPageOneInteractor {
		m := new(mockUsecases.MockPageOneInteractor)
		m.On("GetConfig").Return(domain.DefaultPageOneConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("Draw").Return(mockOutput)
		m.On("Declare").Return(mockOutput)
		m.On("SkipDeclare").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		return m
	}

	t.Run("hint h", func(t *testing.T) {
		m := newMock()
		c := controller.NewPageOneCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("hint"))
		m.AssertCalled(t, "Hint")
	})

	t.Run("quit q", func(t *testing.T) {
		c := controller.NewPageOneCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})
	t.Run("reset r", func(t *testing.T) {
		m := newMock()
		c := controller.NewPageOneCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultPageOneConfig())
	})
	t.Run("play p 3", func(t *testing.T) {
		m := newMock()
		c := controller.NewPageOneCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 3"))
		m.AssertCalled(t, "Play", 3)
	})
	t.Run("play no args", func(t *testing.T) {
		c := controller.NewPageOneCuiController(newMock())
		assert.Contains(t, c.Exec("p"), msgCardIndexRequired())
	})
	t.Run("play invalid", func(t *testing.T) {
		c := controller.NewPageOneCuiController(newMock())
		assert.Contains(t, c.Exec("p abc"), msgInvalidCardIndexPrefix())
	})
	t.Run("draw d", func(t *testing.T) {
		m := newMock()
		c := controller.NewPageOneCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("d"))
		m.AssertCalled(t, "Draw")
	})
	t.Run("declare dc", func(t *testing.T) {
		m := newMock()
		c := controller.NewPageOneCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("dc"))
		m.AssertCalled(t, "Declare")
	})
	t.Run("declare long form", func(t *testing.T) {
		m := newMock()
		c := controller.NewPageOneCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("declare"))
		m.AssertCalled(t, "Declare")
	})
	t.Run("skip sk", func(t *testing.T) {
		m := newMock()
		c := controller.NewPageOneCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sk"))
		m.AssertCalled(t, "SkipDeclare")
	})
	t.Run("skip long form", func(t *testing.T) {
		m := newMock()
		c := controller.NewPageOneCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("skip"))
		m.AssertCalled(t, "SkipDeclare")
	})
	t.Run("nextround nr", func(t *testing.T) {
		m := newMock()
		c := controller.NewPageOneCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextRound")
	})
	t.Run("setdifficulty valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewPageOneCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultPageOneConfig()
		expected.CpuDifficulty = domain.PageOneCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})
	t.Run("setdifficulty no args", func(t *testing.T) {
		c := controller.NewPageOneCuiController(newMock())
		assert.Contains(t, c.Exec("sd"), "required")
	})
	t.Run("setdifficulty out of range", func(t *testing.T) {
		c := controller.NewPageOneCuiController(newMock())
		assert.Contains(t, c.Exec("sd 3"), "Invalid CPU difficulty")
	})
	t.Run("setlimit valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewPageOneCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sl 300"))
		expected := domain.DefaultPageOneConfig()
		expected.PointLimit = 300
		m.AssertCalled(t, "ResetWithConfig", expected)
	})
	t.Run("setlimit zero", func(t *testing.T) {
		c := controller.NewPageOneCuiController(newMock())
		assert.Contains(t, c.Exec("sl 0"), "Invalid point limit")
	})
	t.Run("log", func(t *testing.T) {
		m := newMock()
		c := controller.NewPageOneCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "ActionLog")
	})
	t.Run("unknown", func(t *testing.T) {
		c := controller.NewPageOneCuiController(newMock())
		assert.Contains(t, c.Exec("unknown"), "コマンドが不明です")
	})
}
