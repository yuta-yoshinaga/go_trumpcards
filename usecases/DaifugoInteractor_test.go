package usecases_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/usecases"
	"github.com/yuta-yoshinaga/go_trumpcards/usecases/presenters"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDaifugoInteractor_Method(t *testing.T) {
	mockOutput := `{"players":[],"currentTurn":0,"lastPlay":[],"isRevolution":false,"gameEndFlag":false,"message":""}`
	dpMock := new(presenters.MockDaifugoPresenter)
	dpMock.On("Output", mock.Anything).Return(mockOutput)
	di := usecases.NewDaifugoInteractor(dpMock)

	t.Run("success Reset", func(t *testing.T) {
		assert.Equal(t, mockOutput, di.Reset())
	})
	t.Run("success Play", func(t *testing.T) {
		assert.Equal(t, mockOutput, di.Play([]int{0}))
	})
	t.Run("success Pass", func(t *testing.T) {
		assert.Equal(t, mockOutput, di.Pass())
	})
}
