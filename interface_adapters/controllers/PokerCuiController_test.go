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
	t.Run("success Exec other", func(t *testing.T) {
		assert.Equal(t, "Unsupported command.", tpc.Exec("other"))
	})
	t.Run("success Exec empty", func(t *testing.T) {
		assert.Equal(t, "Unsupported command.", tpc.Exec(""))
	})
}
