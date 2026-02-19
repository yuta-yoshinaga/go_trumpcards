package usecases_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/usecases"
	"github.com/yuta-yoshinaga/go_trumpcards/usecases/presenters"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPokerInteractor_Method(t *testing.T) {
	mockOutput := `{"dealer":{"handRank":0,"handName":"High Card","cards":[]},"player":{"handRank":0,"handName":"High Card","cards":[]},"phase":1,"message":""}`
	ppMock := new(presenters.MockPokerPresenter)
	ppMock.On("Output", mock.AnythingOfType("string")).Return(mockOutput)
	tpi := usecases.NewPokerInteractor(ppMock)

	t.Run("success Reset", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpi.Reset())
	})
	t.Run("success Exchange", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpi.Exchange([]int{0, 1}))
	})
	t.Run("success Exchange empty indices", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpi.Exchange([]int{}))
	})
	t.Run("success Stand", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpi.Stand())
	})
}
