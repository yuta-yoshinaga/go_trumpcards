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

func TestConquianCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockConquianInteractor {
		m := new(mockUsecases.MockConquianInteractor)
		m.On("GetConfig").Return(domain.DefaultConquianConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("DrawFromStock").Return(mockOutput)
		m.On("DrawFromDiscard").Return(mockOutput)
		m.On("Meld", mock.Anything).Return(mockOutput)
		m.On("Discard", mock.Anything).Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit q", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewConquianCuiController(newMock()).Exec("q"))
	})

	t.Run("reset r preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewConquianCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultConquianConfig())
	})

	t.Run("drawstock ds", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewConquianCuiController(m).Exec("ds"))
		m.AssertCalled(t, "DrawFromStock")
	})

	t.Run("drawdiscard dd", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewConquianCuiController(m).Exec("dd"))
		m.AssertCalled(t, "DrawFromDiscard")
	})

	t.Run("meld m with groups", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewConquianCuiController(m).Exec("m 0,1,2;3"))
		m.AssertCalled(t, "Meld", [][]int{{0, 1, 2}, {3}})
	})

	t.Run("meld m no args", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewConquianCuiController(m).Exec("m"))
		m.AssertCalled(t, "Meld", ([][]int)(nil))
	})

	t.Run("discard d with index", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewConquianCuiController(m).Exec("d 3"))
		m.AssertCalled(t, "Discard", 3)
	})

	t.Run("discard d no args", func(t *testing.T) {
		result := controller.NewConquianCuiController(newMock()).Exec("d")
		assert.Contains(t, result, msgCardIndexRequired())
	})

	t.Run("discard d invalid", func(t *testing.T) {
		result := controller.NewConquianCuiController(newMock()).Exec("d abc")
		assert.Contains(t, result, msgInvalidCardIndexPrefix())
	})

	t.Run("nextround nr", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewConquianCuiController(m).Exec("nr"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty sd valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewConquianCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultConquianConfig()
		expected.CpuDifficulty = domain.ConquianCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty sd no args", func(t *testing.T) {
		result := controller.NewConquianCuiController(newMock()).Exec("sd")
		assert.Contains(t, result, msgCpuDifficultyRequired())
	})

	t.Run("setdifficulty sd over 2", func(t *testing.T) {
		result := controller.NewConquianCuiController(newMock()).Exec("sd 3")
		assert.Equal(t, msgInvalidCpuDifficulty("3"), result)
	})

	t.Run("setwins sw valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewConquianCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sw 3"))
		expected := domain.DefaultConquianConfig()
		expected.TargetWins = 3
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setwins sw zero", func(t *testing.T) {
		result := controller.NewConquianCuiController(newMock()).Exec("sw 0")
		assert.Equal(t, msgKey("invalidTargetWins1OrMore", "val", "0"), result)
	})

	t.Run("log", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewConquianCuiController(m).Exec("log"))
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown", func(t *testing.T) {
		result := controller.NewConquianCuiController(newMock()).Exec("xyz")
		assert.Contains(t, result, "コマンドが不明です")
	})

	t.Run("empty", func(t *testing.T) {
		result := controller.NewConquianCuiController(newMock()).Exec("")
		assert.Contains(t, result, "'help' でコマンド一覧を表示します。")
	})
}
