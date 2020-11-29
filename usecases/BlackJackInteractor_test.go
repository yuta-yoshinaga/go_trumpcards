package usecases_test

import (
	"testing"
	"go_trumpcards/usecases"
	"go_trumpcards/usecases/presenters"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBlackJackInteractor_Method(t *testing.T) {
	bjpMock := new(presenters.MockBlackJackPresenter)
	bjpMock.On("Output", mock.AnythingOfType("string")).Return("----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n")
	tbj := usecases.NewBlackJackInteractor(bjpMock)
	t.Run("success Reset", func(t *testing.T) {
		assert.Equal(t, "----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n", tbj.Reset())
	})
	t.Run("success Hit", func(t *testing.T) {
		assert.Equal(t, "----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n", tbj.Hit())
	})
	t.Run("success Stand", func(t *testing.T) {
		assert.Equal(t, "----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n", tbj.Stand())
	})
}
