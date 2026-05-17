package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func TestNewFourCardPokerInteractor(t *testing.T) {
	mg := new(interfaces.MockFourCardPokerGame)
	mp := new(presenter.MockFourCardPokerPresenter)
	fi := NewFourCardPokerInteractor(mg, mp)
	assert.NotNil(t, fi)
}

func TestNewFourCardPokerInteractor_NilPanics(t *testing.T) {
	mp := new(presenter.MockFourCardPokerPresenter)
	assert.Panics(t, func() { NewFourCardPokerInteractor(nil, mp) })

	mg := new(interfaces.MockFourCardPokerGame)
	assert.Panics(t, func() { NewFourCardPokerInteractor(mg, nil) })
}

func TestFourCardPokerInteractor_Reset(t *testing.T) {
	mg := new(interfaces.MockFourCardPokerGame)
	mp := new(presenter.MockFourCardPokerPresenter)
	fi := NewFourCardPokerInteractor(mg, mp)

	mg.On("Reset").Return()
	mp.On("Output", mg, nil).Return("reset out")

	assert.Equal(t, "reset out", fi.Reset())
	mg.AssertCalled(t, "Reset")
}

func TestFourCardPokerInteractor_Bet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mg := new(interfaces.MockFourCardPokerGame)
		mp := new(presenter.MockFourCardPokerPresenter)
		fi := NewFourCardPokerInteractor(mg, mp)

		mg.On("Bet", 100, 50).Return(nil)
		mp.On("Output", mg, nil).Return("bet out")

		assert.Equal(t, "bet out", fi.Bet(100, 50))
	})

	t.Run("error", func(t *testing.T) {
		mg := new(interfaces.MockFourCardPokerGame)
		mp := new(presenter.MockFourCardPokerPresenter)
		fi := NewFourCardPokerInteractor(mg, mp)

		mg.On("Bet", 100, 0).Return(errors.New("boom"))
		mp.On("Output", mg, mock.MatchedBy(func(e error) bool { return e != nil && e.Error() == "boom" })).Return("err")

		assert.Equal(t, "err", fi.Bet(100, 0))
	})
}

func TestFourCardPokerInteractor_Play(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mg := new(interfaces.MockFourCardPokerGame)
		mp := new(presenter.MockFourCardPokerPresenter)
		fi := NewFourCardPokerInteractor(mg, mp)

		mg.On("Play", 2).Return(nil)
		mp.On("Output", mg, nil).Return("play out")

		assert.Equal(t, "play out", fi.Play(2))
	})

	t.Run("error", func(t *testing.T) {
		mg := new(interfaces.MockFourCardPokerGame)
		mp := new(presenter.MockFourCardPokerPresenter)
		fi := NewFourCardPokerInteractor(mg, mp)

		mg.On("Play", 1).Return(errors.New("nope"))
		mp.On("Output", mg, mock.MatchedBy(func(e error) bool { return e != nil })).Return("err")

		assert.Equal(t, "err", fi.Play(1))
	})
}

func TestFourCardPokerInteractor_Fold(t *testing.T) {
	mg := new(interfaces.MockFourCardPokerGame)
	mp := new(presenter.MockFourCardPokerPresenter)
	fi := NewFourCardPokerInteractor(mg, mp)

	mg.On("Fold").Return(nil)
	mp.On("Output", mg, nil).Return("fold out")

	assert.Equal(t, "fold out", fi.Fold())
}

func TestFourCardPokerInteractor_ActionLog(t *testing.T) {
	mg := new(interfaces.MockFourCardPokerGame)
	mp := new(presenter.MockFourCardPokerPresenter)
	fi := NewFourCardPokerInteractor(mg, mp)

	mp.On("ActionLogOutput", mg).Return("log out")

	assert.Equal(t, "log out", fi.ActionLog())
}
