package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newFaroTestInteractor() (*FaroInteractor, *interfaces.MockFaroGame, *presenter.MockFaroPresenter) {
	mockGame := new(interfaces.MockFaroGame)
	mockPresenter := new(presenter.MockFaroPresenter)
	return NewFaroInteractor(mockGame, mockPresenter), mockGame, mockPresenter
}

func TestNewFaroInteractor(t *testing.T) {
	fi, _, _ := newFaroTestInteractor()
	assert.NotNil(t, fi)
}

func TestNewFaroInteractor_NilPanics(t *testing.T) {
	mp := new(presenter.MockFaroPresenter)
	assert.Panics(t, func() { NewFaroInteractor(nil, mp) })
	mg := new(interfaces.MockFaroGame)
	assert.Panics(t, func() { NewFaroInteractor(mg, nil) })
}

func TestFaroInteractor_Reset(t *testing.T) {
	fi, mg, mp := newFaroTestInteractor()
	mg.On("Reset").Return()
	mp.On("Output", mg, nil).Return("reset output")
	assert.Equal(t, "reset output", fi.Reset())
	mg.AssertCalled(t, "Reset")
}

func TestFaroInteractor_NextRound(t *testing.T) {
	fi, mg, mp := newFaroTestInteractor()
	mg.On("NextRound").Return()
	mp.On("Output", mg, nil).Return("next output")
	assert.Equal(t, "next output", fi.NextRound())
}

func TestFaroInteractor_PlaceBet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fi, mg, mp := newFaroTestInteractor()
		mg.On("PlayerPlaceBet", 7, 100, true).Return(nil)
		mp.On("Output", mg, nil).Return("bet output")
		assert.Equal(t, "bet output", fi.PlaceBet(7, 100, true))
	})
	t.Run("error", func(t *testing.T) {
		fi, mg, mp := newFaroTestInteractor()
		mg.On("PlayerPlaceBet", 7, 100, false).Return(errors.New("bad"))
		mp.On("Output", mg, mock.MatchedBy(func(e error) bool { return e != nil })).Return("err output")
		assert.Equal(t, "err output", fi.PlaceBet(7, 100, false))
	})
}

func TestFaroInteractor_ClearBet(t *testing.T) {
	fi, mg, mp := newFaroTestInteractor()
	mg.On("PlayerClearBet", 3).Return(nil)
	mp.On("Output", mg, nil).Return("clear output")
	assert.Equal(t, "clear output", fi.ClearBet(3))
}

func TestFaroInteractor_ClearAll(t *testing.T) {
	fi, mg, mp := newFaroTestInteractor()
	mg.On("PlayerClearAll").Return(nil)
	mp.On("Output", mg, nil).Return("clearall output")
	assert.Equal(t, "clearall output", fi.ClearAll())
}

func TestFaroInteractor_DealTurn(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fi, mg, mp := newFaroTestInteractor()
		mg.On("PlayerDealTurn").Return(nil)
		mp.On("Output", mg, nil).Return("deal output")
		assert.Equal(t, "deal output", fi.DealTurn())
	})
	t.Run("error", func(t *testing.T) {
		fi, mg, mp := newFaroTestInteractor()
		mg.On("PlayerDealTurn").Return(errors.New("exhausted"))
		mp.On("Output", mg, mock.MatchedBy(func(e error) bool { return e != nil })).Return("deal err")
		assert.Equal(t, "deal err", fi.DealTurn())
	})
}

func TestFaroInteractor_Call(t *testing.T) {
	fi, mg, mp := newFaroTestInteractor()
	order := []int{3, 9, 12}
	mg.On("PlayerCall", order).Return(nil)
	mp.On("Output", mg, nil).Return("call output")
	assert.Equal(t, "call output", fi.Call(order))
}

func TestFaroInteractor_ActionLog(t *testing.T) {
	fi, mg, mp := newFaroTestInteractor()
	mp.On("ActionLogOutput", mg).Return("log output")
	assert.Equal(t, "log output", fi.ActionLog())
}

func TestRestoreFaroInteractor(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		_, err := RestoreFaroInteractor([]byte("not json"), nil)
		assert.Error(t, err)
	})
}
