package controller_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPigsTailCuiController_Method(t *testing.T) {
	mockOutput := "==========\nPig's Tail (ぶたのしっぽ)\n=========="
	ptiMock := new(usecase.MockPigsTailInteractor)
	ptiMock.On("GetConfig").Return(domain.DefaultPigsTailConfig())
	ptiMock.On("Reset", mock.Anything).Return(mockOutput)
	ptiMock.On("Action", 0).Return(mockOutput)
	ptiMock.On("ActionLog").Return(`[]`)

	tptc := controller.NewPigsTailCuiController(ptiMock)

	t.Run("success Exec q", func(t *testing.T) {
		assert.Equal(t, "bye.", tptc.Exec("q"))
	})
	t.Run("success Exec quit", func(t *testing.T) {
		assert.Equal(t, "bye.", tptc.Exec("quit"))
	})
	t.Run("success Exec r preserves config", func(t *testing.T) {
		assert.Equal(t, mockOutput, tptc.Exec("r"))
		ptiMock.AssertCalled(t, "GetConfig")
	})
	t.Run("success Exec reset preserves config", func(t *testing.T) {
		assert.Equal(t, mockOutput, tptc.Exec("reset"))
		ptiMock.AssertCalled(t, "GetConfig")
	})
	t.Run("success Exec d", func(t *testing.T) {
		assert.Equal(t, mockOutput, tptc.Exec("d"))
		ptiMock.AssertCalled(t, "Action", 0)
	})
	t.Run("success Exec draw", func(t *testing.T) {
		assert.Equal(t, mockOutput, tptc.Exec("draw"))
		ptiMock.AssertCalled(t, "Action", 0)
	})
	t.Run("success Exec log", func(t *testing.T) {
		assert.Equal(t, `[]`, tptc.Exec("log"))
		ptiMock.AssertCalled(t, "ActionLog")
	})
	t.Run("success Exec other returns unknown", func(t *testing.T) {
		result := tptc.Exec("xyz")
		assert.Contains(t, result, "コマンドが不明です")
	})
}
