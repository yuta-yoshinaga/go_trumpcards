package usecase_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPokerInteractor_Method(t *testing.T) {
	mockOutput := `{"dealer":{"handRank":0,"handName":"High Card","cards":[]},"player":{"handRank":0,"handName":"High Card","cards":[]},"phase":1,"message":"","pot":20,"ante":10}`
	ppMock := new(presenter.MockPokerPresenter)
	ppMock.On("Output", mock.AnythingOfType("string")).Return(mockOutput)
	tpi := usecase.NewPokerInteractor(ppMock)

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
	t.Run("success Bet", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpi.Bet(10))
	})
	t.Run("success Call", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpi.Call())
	})
	t.Run("success Raise", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpi.Raise(20))
	})
	t.Run("success Fold", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpi.Fold())
	})
	t.Run("success Check", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpi.Check())
	})
}
