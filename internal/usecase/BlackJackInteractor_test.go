package usecase_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewBlackJackInteractor_NilGuards(t *testing.T) {
	bjpMock := new(presenter.MockBlackJackPresenter)
	t.Run("panics when bj is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "BlackJackInteractor: bj must not be nil", func() {
			usecase.NewBlackJackInteractor(nil, bjpMock)
		})
	})
	t.Run("panics when bjp is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "BlackJackInteractor: bjp must not be nil", func() {
			usecase.NewBlackJackInteractor(domain.NewDefaultBlackJack(), nil)
		})
	})
}

func TestBlackJackInteractor_Method(t *testing.T) {
	bjpMock := new(presenter.MockBlackJackPresenter)
	bjpMock.On("Output", mock.Anything, mock.Anything).Return("----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n")
	tbj := usecase.NewBlackJackInteractor(domain.NewDefaultBlackJack(), bjpMock)
	t.Run("success Reset", func(t *testing.T) {
		assert.Equal(t, "----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n", tbj.Reset())
	})
	t.Run("success Hit", func(t *testing.T) {
		assert.Equal(t, "----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n", tbj.Hit())
	})
	t.Run("success Stand", func(t *testing.T) {
		assert.Equal(t, "----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n", tbj.Stand())
	})
	t.Run("success Bet", func(t *testing.T) {
		assert.Equal(t, "----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n", tbj.Bet(100))
	})
	t.Run("success DoubleDown", func(t *testing.T) {
		assert.Equal(t, "----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n", tbj.DoubleDown())
	})
	t.Run("success Split", func(t *testing.T) {
		assert.Equal(t, "----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n", tbj.Split())
	})
	t.Run("success Insurance", func(t *testing.T) {
		assert.Equal(t, "----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n", tbj.Insurance())
	})
	t.Run("success DeclineInsurance", func(t *testing.T) {
		assert.Equal(t, "----------\ndealer score \nCLOVER 2,\n----------\nplayer score 22\nSPADE 2,SPADE 10,SPADE 11\n----------\n", tbj.DeclineInsurance())
	})
}

func TestBlackJackInteractor_NewMethods(t *testing.T) {
	bjpMock := new(presenter.MockBlackJackPresenter)
	bjpMock.On("Output", mock.Anything, mock.Anything).Return("ok")
	tbj := usecase.NewBlackJackInteractor(domain.NewDefaultBlackJack(), bjpMock)
	t.Run("Surrender delegates to presenter", func(t *testing.T) {
		assert.Equal(t, "ok", tbj.Surrender())
	})
	t.Run("SetDeckCount delegates to presenter", func(t *testing.T) {
		assert.Equal(t, "ok", tbj.SetDeckCount(2))
	})
	t.Run("ToggleHint delegates to presenter", func(t *testing.T) {
		assert.Equal(t, "ok", tbj.ToggleHint())
	})
}
