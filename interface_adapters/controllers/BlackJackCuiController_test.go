package controllers_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/controllers"
	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/controllers/usecases"

	"github.com/stretchr/testify/assert"
)

func TestBlackJackCuiController_Method(t *testing.T) {
	bjiMock := new(usecases.MockBlackJackInteractor)
	bjiMock.On("Reset").Return("----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n")
	bjiMock.On("Hit").Return("----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n")
	bjiMock.On("Stand").Return("----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n")
	tbc := controllers.NewBlackJackCuiController(bjiMock)
	t.Run("success Exec q", func(t *testing.T) {
		assert.Equal(t, "bye.", tbc.Exec("q"))
	})
	t.Run("success Exec quit", func(t *testing.T) {
		assert.Equal(t, "bye.", tbc.Exec("quit"))
	})
	t.Run("success Exec r", func(t *testing.T) {
		assert.Equal(t, "----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n", tbc.Exec("r"))
	})
	t.Run("success Exec reset", func(t *testing.T) {
		assert.Equal(t, "----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n", tbc.Exec("reset"))
	})
	t.Run("success Exec h", func(t *testing.T) {
		assert.Equal(t, "----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n", tbc.Exec("h"))
	})
	t.Run("success Exec hit", func(t *testing.T) {
		assert.Equal(t, "----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n", tbc.Exec("hit"))
	})
	t.Run("success Exec s", func(t *testing.T) {
		assert.Equal(t, "----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n", tbc.Exec("s"))
	})
	t.Run("success Exec stand", func(t *testing.T) {
		assert.Equal(t, "----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n", tbc.Exec("stand"))
	})
	t.Run("success Exec other", func(t *testing.T) {
		assert.Equal(t, "Unsupported command.", tbc.Exec("other"))
	})
}
