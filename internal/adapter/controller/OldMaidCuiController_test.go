package controller_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOldMaidCuiController_Method(t *testing.T) {
	mockOutput := "==========\nOld Maid (ババ抜き)\n==========\n[You]: 0枚\n\nCPU 1: 0枚\nCPU 2: 0枚\nCPU 3: 0枚\n----------\n手番: あなた\n==========\n"
	omiMock := new(usecase.MockOldMaidInteractor)
	omiMock.On("Reset", mock.Anything).Return(mockOutput)
	omiMock.On("Draw", -1).Return(mockOutput)
	omiMock.On("Draw", 0).Return(mockOutput)
	omiMock.On("Draw", 2).Return(mockOutput)
	omiMock.On("Shuffle").Return(mockOutput)

	tomc := controller.NewOldMaidCuiController(omiMock)

	t.Run("success Exec q", func(t *testing.T) {
		assert.Equal(t, "bye.", tomc.Exec("q"))
	})
	t.Run("success Exec quit", func(t *testing.T) {
		assert.Equal(t, "bye.", tomc.Exec("quit"))
	})
	t.Run("success Exec r calls Reset with DefaultOldMaidConfig", func(t *testing.T) {
		assert.Equal(t, mockOutput, tomc.Exec("r"))
		omiMock.AssertCalled(t, "Reset", domain.DefaultOldMaidConfig())
	})
	t.Run("success Exec reset calls Reset with DefaultOldMaidConfig", func(t *testing.T) {
		assert.Equal(t, mockOutput, tomc.Exec("reset"))
		omiMock.AssertCalled(t, "Reset", domain.DefaultOldMaidConfig())
	})
	t.Run("success Exec d", func(t *testing.T) {
		assert.Equal(t, mockOutput, tomc.Exec("d"))
	})
	t.Run("success Exec draw", func(t *testing.T) {
		assert.Equal(t, mockOutput, tomc.Exec("draw"))
	})
	t.Run("success Exec d 0", func(t *testing.T) {
		assert.Equal(t, mockOutput, tomc.Exec("d 0"))
	})
	t.Run("success Exec draw 2", func(t *testing.T) {
		assert.Equal(t, mockOutput, tomc.Exec("draw 2"))
	})
	t.Run("success Exec s (shuffle)", func(t *testing.T) {
		assert.Equal(t, mockOutput, tomc.Exec("s"))
		omiMock.AssertCalled(t, "Shuffle")
	})
	t.Run("success Exec shuffle", func(t *testing.T) {
		assert.Equal(t, mockOutput, tomc.Exec("shuffle"))
		omiMock.AssertCalled(t, "Shuffle")
	})
	t.Run("success Exec other", func(t *testing.T) {
		assert.Equal(t, "コマンドが不明です: other", tomc.Exec("other"))
	})
	t.Run("success Exec empty", func(t *testing.T) {
		assert.Equal(t, "コマンドが不明です: ", tomc.Exec(""))
	})
}
