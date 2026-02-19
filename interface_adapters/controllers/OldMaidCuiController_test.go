package controllers_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/controllers"
	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/controllers/usecases"

	"github.com/stretchr/testify/assert"
)

func TestOldMaidCuiController_Method(t *testing.T) {
	mockOutput := "==========\nOld Maid (ババ抜き)\n==========\n[You]: 0枚\n\nCPU 1: 0枚\nCPU 2: 0枚\nCPU 3: 0枚\n----------\n手番: あなた\n==========\n"
	omiMock := new(usecases.MockOldMaidInteractor)
	omiMock.On("Reset").Return(mockOutput)
	omiMock.On("Draw").Return(mockOutput)

	tomc := controllers.NewOldMaidCuiController(omiMock)

	t.Run("success Exec q", func(t *testing.T) {
		assert.Equal(t, "bye.", tomc.Exec("q"))
	})
	t.Run("success Exec quit", func(t *testing.T) {
		assert.Equal(t, "bye.", tomc.Exec("quit"))
	})
	t.Run("success Exec r", func(t *testing.T) {
		assert.Equal(t, mockOutput, tomc.Exec("r"))
	})
	t.Run("success Exec reset", func(t *testing.T) {
		assert.Equal(t, mockOutput, tomc.Exec("reset"))
	})
	t.Run("success Exec d", func(t *testing.T) {
		assert.Equal(t, mockOutput, tomc.Exec("d"))
	})
	t.Run("success Exec draw", func(t *testing.T) {
		assert.Equal(t, mockOutput, tomc.Exec("draw"))
	})
	t.Run("success Exec other", func(t *testing.T) {
		assert.Equal(t, "コマンドが不明です: other", tomc.Exec("other"))
	})
	t.Run("success Exec empty", func(t *testing.T) {
		assert.Equal(t, "コマンドが不明です: ", tomc.Exec(""))
	})
}
