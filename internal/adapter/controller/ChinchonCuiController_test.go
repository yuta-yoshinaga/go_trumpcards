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

func TestChinchonCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockChinchonInteractor {
		m := new(mockUsecases.MockChinchonInteractor)
		m.On("GetConfig").Return(domain.DefaultChinchonConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("DrawFromStock").Return(mockOutput)
		m.On("DrawFromDiscard").Return(mockOutput)
		m.On("Discard", mock.Anything).Return(mockOutput)
		m.On("Knock", mock.Anything).Return(mockOutput)
		m.On("Layoff", mock.Anything).Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit q", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewChinchonCuiController(newMock()).Exec("q"))
	})

	t.Run("reset r preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewChinchonCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultChinchonConfig())
	})

	t.Run("drawstock ds", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewChinchonCuiController(m).Exec("ds"))
		m.AssertCalled(t, "DrawFromStock")
	})

	t.Run("drawdiscard dd", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewChinchonCuiController(m).Exec("dd"))
		m.AssertCalled(t, "DrawFromDiscard")
	})

	t.Run("discard d with index", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewChinchonCuiController(m).Exec("d 3"))
		m.AssertCalled(t, "Discard", 3)
	})

	t.Run("discard d no args", func(t *testing.T) {
		result := controller.NewChinchonCuiController(newMock()).Exec("d")
		assert.Contains(t, result, msgCardIndexRequired())
	})

	t.Run("discard d invalid", func(t *testing.T) {
		result := controller.NewChinchonCuiController(newMock()).Exec("d abc")
		assert.Contains(t, result, msgInvalidCardIndexPrefix())
	})

	t.Run("knock k with index", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewChinchonCuiController(m).Exec("k 0"))
		m.AssertCalled(t, "Knock", 0)
	})

	t.Run("knock k no args", func(t *testing.T) {
		result := controller.NewChinchonCuiController(newMock()).Exec("k")
		assert.Contains(t, result, msgCardIndexRequired())
	})

	t.Run("layoff lo with indices", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewChinchonCuiController(m).Exec("lo 0,1"))
		m.AssertCalled(t, "Layoff", []int{0, 1})
	})

	t.Run("layoff lo no args", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewChinchonCuiController(m).Exec("lo"))
		m.AssertCalled(t, "Layoff", ([]int)(nil))
	})

	t.Run("nextround nr", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewChinchonCuiController(m).Exec("nr"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty sd valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewChinchonCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultChinchonConfig()
		expected.CpuDifficulty = domain.ChinchonCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty sd over 2", func(t *testing.T) {
		result := controller.NewChinchonCuiController(newMock()).Exec("sd 3")
		assert.Equal(t, msgInvalidCpuDifficulty("3"), result)
	})

	t.Run("setplayers sp valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewChinchonCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sp 2"))
		expected := domain.DefaultChinchonConfig()
		expected.PlayerCount = 2
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setplayers sp out of range", func(t *testing.T) {
		result := controller.NewChinchonCuiController(newMock()).Exec("sp 5")
		assert.Equal(t, msgInvalidPlayerCount("5"), result)
	})

	t.Run("log", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewChinchonCuiController(m).Exec("log"))
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown", func(t *testing.T) {
		result := controller.NewChinchonCuiController(newMock()).Exec("xyz")
		assert.Contains(t, result, "コマンドが不明です")
	})

	t.Run("empty", func(t *testing.T) {
		result := controller.NewChinchonCuiController(newMock()).Exec("")
		assert.Contains(t, result, "'help' でコマンド一覧を表示します。")
	})
}
