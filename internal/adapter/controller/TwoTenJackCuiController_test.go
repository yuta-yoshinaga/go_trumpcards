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

func TestTwoTenJackCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockTwoTenJackInteractor {
		m := new(mockUsecases.MockTwoTenJackInteractor)
		m.On("GetConfig").Return(domain.DefaultTwoTenJackConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("DeclareTrump", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		c := controller.NewTwoTenJackCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	t.Run("reset", func(t *testing.T) {
		m := newMock()
		c := controller.NewTwoTenJackCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultTwoTenJackConfig())
		assert.Equal(t, mockOutput, c.Exec("reset"))
	})

	t.Run("declare valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewTwoTenJackCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("d 1"))
		m.AssertCalled(t, "DeclareTrump", 1)
		assert.Equal(t, mockOutput, c.Exec("declare 3"))
		m.AssertCalled(t, "DeclareTrump", 3)
	})

	t.Run("declare missing arg", func(t *testing.T) {
		c := controller.NewTwoTenJackCuiController(newMock())
		assert.Contains(t, c.Exec("d"), "Trump suit is required")
	})

	t.Run("declare invalid arg", func(t *testing.T) {
		c := controller.NewTwoTenJackCuiController(newMock())
		assert.Contains(t, c.Exec("d abc"), "Invalid trump suit")
	})

	t.Run("declare out of range", func(t *testing.T) {
		c := controller.NewTwoTenJackCuiController(newMock())
		assert.Contains(t, c.Exec("d 0"), "Invalid trump suit")
		assert.Contains(t, c.Exec("d 5"), "Invalid trump suit")
	})

	t.Run("play", func(t *testing.T) {
		m := newMock()
		c := controller.NewTwoTenJackCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 3"))
		m.AssertCalled(t, "Play", 3)
		assert.Equal(t, mockOutput, c.Exec("play 5"))
		m.AssertCalled(t, "Play", 5)
	})

	t.Run("play missing arg", func(t *testing.T) {
		c := controller.NewTwoTenJackCuiController(newMock())
		assert.Contains(t, c.Exec("p"), msgCardIndexRequired())
	})

	t.Run("play invalid arg", func(t *testing.T) {
		c := controller.NewTwoTenJackCuiController(newMock())
		assert.Contains(t, c.Exec("p abc"), msgInvalidCardIndexPrefix())
	})

	t.Run("next/nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewTwoTenJackCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		m.AssertCalled(t, "NextTrick")
		assert.Equal(t, mockOutput, c.Exec("next"))
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextRound")
		assert.Equal(t, mockOutput, c.Exec("nextround"))
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewTwoTenJackCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultTwoTenJackConfig()
		expected.CpuDifficulty = domain.TwoTenJackCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
		assert.Contains(t, c.Exec("sd"), "required")
		assert.Contains(t, c.Exec("sd abc"), "Invalid CPU difficulty")
		assert.Contains(t, c.Exec("sd 3"), "Invalid CPU difficulty")
	})

	t.Run("setlimit", func(t *testing.T) {
		m := newMock()
		c := controller.NewTwoTenJackCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sl 100"))
		expected := domain.DefaultTwoTenJackConfig()
		expected.PointLimit = 100
		m.AssertCalled(t, "ResetWithConfig", expected)
		assert.Contains(t, c.Exec("sl"), "required")
		assert.Contains(t, c.Exec("sl abc"), "Invalid point limit")
		assert.Contains(t, c.Exec("sl 0"), "Invalid point limit")
	})

	t.Run("hint and log", func(t *testing.T) {
		m := newMock()
		c := controller.NewTwoTenJackCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("hint"))
		assert.Equal(t, mockOutput, c.Exec("log"))
		assert.Equal(t, mockOutput, c.Exec("l"))
	})

	t.Run("unknown and empty", func(t *testing.T) {
		c := controller.NewTwoTenJackCuiController(newMock())
		assert.Contains(t, c.Exec("unknown"), "コマンドが不明です")
		assert.Contains(t, c.Exec(""), "'help'")
	})
}
