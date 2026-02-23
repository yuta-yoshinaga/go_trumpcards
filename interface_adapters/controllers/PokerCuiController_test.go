package controllers_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/controllers"
	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/controllers/usecases"

	"github.com/stretchr/testify/assert"
)

func TestPokerCuiController_Method(t *testing.T) {
	mockOutput := "----------\nplayer hand\n[0]SPADE 5\n----------\ndealer hand\n----------\n"
	piMock := new(usecases.MockPokerInteractor)
	piMock.On("Reset").Return(mockOutput)
	piMock.On("Exchange", []int{0, 1, 2}).Return(mockOutput)
	piMock.On("Exchange", []int{0}).Return(mockOutput)
	piMock.On("Exchange", []int{}).Return(mockOutput)
	piMock.On("Stand").Return(mockOutput)
	piMock.On("Bet", 20).Return(mockOutput)
	piMock.On("Bet", 0).Return(mockOutput)
	piMock.On("Call").Return(mockOutput)
	piMock.On("Raise", 30).Return(mockOutput)
	piMock.On("Raise", 0).Return(mockOutput)
	piMock.On("Fold").Return(mockOutput)
	piMock.On("Check").Return(mockOutput)

	tpc := controllers.NewPokerCuiController(piMock)

	t.Run("success Exec q", func(t *testing.T) {
		assert.Equal(t, "bye.", tpc.Exec("q"))
	})
	t.Run("success Exec quit", func(t *testing.T) {
		assert.Equal(t, "bye.", tpc.Exec("quit"))
	})
	t.Run("success Exec r", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpc.Exec("r"))
	})
	t.Run("success Exec reset", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpc.Exec("reset"))
	})
	t.Run("success Exec e with indices", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpc.Exec("e 0 1 2"))
	})
	t.Run("success Exec exchange with index", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpc.Exec("exchange 0"))
	})
	t.Run("success Exec e no indices", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpc.Exec("e"))
	})
	t.Run("success Exec s", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpc.Exec("s"))
	})
	t.Run("success Exec stand", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpc.Exec("stand"))
	})
	t.Run("success Exec b with amount", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpc.Exec("b 20"))
	})
	t.Run("success Exec bet with amount", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpc.Exec("bet 20"))
	})
	t.Run("success Exec b no amount", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpc.Exec("b"))
	})
	t.Run("success Exec c", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpc.Exec("c"))
	})
	t.Run("success Exec call", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpc.Exec("call"))
	})
	t.Run("success Exec ra with amount", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpc.Exec("ra 30"))
	})
	t.Run("success Exec raise with amount", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpc.Exec("raise 30"))
	})
	t.Run("success Exec ra no amount", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpc.Exec("ra"))
	})
	t.Run("success Exec f", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpc.Exec("f"))
	})
	t.Run("success Exec fold", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpc.Exec("fold"))
	})
	t.Run("success Exec ck", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpc.Exec("ck"))
	})
	t.Run("success Exec check", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpc.Exec("check"))
	})
	t.Run("success Exec other", func(t *testing.T) {
		assert.Equal(t, "Unsupported command.", tpc.Exec("other"))
	})
	t.Run("success Exec empty", func(t *testing.T) {
		assert.Equal(t, "Unsupported command.", tpc.Exec(""))
	})
}
