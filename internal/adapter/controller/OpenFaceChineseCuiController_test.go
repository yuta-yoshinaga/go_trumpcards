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

func TestOpenFaceChineseCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockOpenFaceChineseInteractor {
		m := new(mockUsecases.MockOpenFaceChineseInteractor)
		m.On("GetConfig").Return(domain.DefaultOpenFaceChineseConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Place", mock.Anything).Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewOpenFaceChineseCuiController(newMock()).Exec("q"))
		assert.Equal(t, "bye.", controller.NewOpenFaceChineseCuiController(newMock()).Exec("quit"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewOpenFaceChineseCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultOpenFaceChineseConfig())
	})

	t.Run("place with arg", func(t *testing.T) {
		m := newMock()
		c := controller.NewOpenFaceChineseCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("place 2"))
		m.AssertCalled(t, "Place", 2)
	})

	t.Run("place no args", func(t *testing.T) {
		result := controller.NewOpenFaceChineseCuiController(newMock()).Exec("place")
		assert.Contains(t, result, "Row is required")
	})

	t.Run("place invalid", func(t *testing.T) {
		result := controller.NewOpenFaceChineseCuiController(newMock()).Exec("place 9")
		assert.Contains(t, result, "Invalid row")
	})

	t.Run("front/middle/back shortcuts", func(t *testing.T) {
		m := newMock()
		c := controller.NewOpenFaceChineseCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("front"))
		assert.Equal(t, mockOutput, c.Exec("m"))
		assert.Equal(t, mockOutput, c.Exec("back"))
		m.AssertCalled(t, "Place", domain.OpenFaceChineseRowFront)
		m.AssertCalled(t, "Place", domain.OpenFaceChineseRowMiddle)
		m.AssertCalled(t, "Place", domain.OpenFaceChineseRowBack)
	})

	t.Run("nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewOpenFaceChineseCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewOpenFaceChineseCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultOpenFaceChineseConfig()
		expected.CpuDifficulty = domain.OpenFaceChineseCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty invalid", func(t *testing.T) {
		result := controller.NewOpenFaceChineseCuiController(newMock()).Exec("sd 9")
		assert.Contains(t, result, msgInvalidCpuDifficultyPrefix())
	})

	t.Run("setplayers", func(t *testing.T) {
		m := newMock()
		c := controller.NewOpenFaceChineseCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sp 4"))
		expected := domain.DefaultOpenFaceChineseConfig()
		expected.PlayerCount = 4
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setplayers invalid", func(t *testing.T) {
		result := controller.NewOpenFaceChineseCuiController(newMock()).Exec("sp 9")
		assert.Contains(t, result, msgInvalidPlayerCountPrefix())
	})

	t.Run("hint / log", func(t *testing.T) {
		m := newMock()
		c := controller.NewOpenFaceChineseCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "Hint")
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		result := controller.NewOpenFaceChineseCuiController(newMock()).Exec("zzz")
		assert.Contains(t, result, "コマンドが不明です")
	})
}
