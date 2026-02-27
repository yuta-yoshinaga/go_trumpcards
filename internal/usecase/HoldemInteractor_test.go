package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func TestNewHoldemInteractor(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	mp := new(presenter.MockHoldemPresenter)
	hi := NewHoldemInteractor(mg, mp)
	assert.NotNil(t, hi)
}

func TestNewHoldemInteractor_NilGame(t *testing.T) {
	mp := new(presenter.MockHoldemPresenter)
	assert.Panics(t, func() {
		NewHoldemInteractor(nil, mp)
	})
}

func TestNewHoldemInteractor_NilPresenter(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	assert.Panics(t, func() {
		NewHoldemInteractor(mg, nil)
	})
}

func TestHoldemInteractor_Reset(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	mp := new(presenter.MockHoldemPresenter)
	hi := NewHoldemInteractor(mg, mp)

	mg.On("Reset").Return()
	mp.On("Output", mg, mock.Anything).Return("reset output")

	result := hi.Reset()
	assert.Equal(t, "reset output", result)
	mg.AssertCalled(t, "Reset")
}

func TestHoldemInteractor_ResetWithConfig(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	mp := new(presenter.MockHoldemPresenter)
	hi := NewHoldemInteractor(mg, mp)

	cfg := domain.HoldemConfig{SmallBlind: 10, BigBlind: 20, InitChips: 2000}
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return()
	mp.On("Output", mg, mock.Anything).Return("reset with config output")

	result := hi.ResetWithConfig(cfg)
	assert.Equal(t, "reset with config output", result)
	mg.AssertCalled(t, "SetConfig", cfg)
	mg.AssertCalled(t, "Reset")
}

func TestHoldemInteractor_Action(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	mp := new(presenter.MockHoldemPresenter)
	hi := NewHoldemInteractor(mg, mp)

	mg.On("GetGameEndFlag").Return(false)
	mg.On("IsHumanTurn").Return(true)
	mg.On("PlayerAction", domain.HoldemActionCheck, 0).Return(nil)
	mp.On("Output", mg, mock.Anything).Return("action output")

	result := hi.Action(domain.HoldemActionCheck, 0)
	assert.Equal(t, "action output", result)
	mg.AssertCalled(t, "PlayerAction", domain.HoldemActionCheck, 0)
}

func TestHoldemInteractor_Action_GameEnded(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	mp := new(presenter.MockHoldemPresenter)
	hi := NewHoldemInteractor(mg, mp)

	mg.On("GetGameEndFlag").Return(true)
	mp.On("Output", mg, mock.Anything).Return("game ended")

	result := hi.Action(domain.HoldemActionCheck, 0)
	assert.Equal(t, "game ended", result)
	mg.AssertNotCalled(t, "PlayerAction", mock.Anything, mock.Anything)
}

func TestHoldemInteractor_Action_NotHumanTurn(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	mp := new(presenter.MockHoldemPresenter)
	hi := NewHoldemInteractor(mg, mp)

	mg.On("GetGameEndFlag").Return(false)
	mg.On("IsHumanTurn").Return(false)
	mp.On("Output", mg, mock.Anything).Return("not your turn")

	result := hi.Action(domain.HoldemActionCheck, 0)
	assert.Equal(t, "not your turn", result)
	mg.AssertNotCalled(t, "PlayerAction", mock.Anything, mock.Anything)
}

func TestHoldemInteractor_Action_Error(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	mp := new(presenter.MockHoldemPresenter)
	hi := NewHoldemInteractor(mg, mp)

	err := errors.New("test error")
	mg.On("GetGameEndFlag").Return(false)
	mg.On("IsHumanTurn").Return(true)
	mg.On("PlayerAction", domain.HoldemActionBet, 50).Return(err)
	mp.On("Output", mg, err).Return("error output")

	result := hi.Action(domain.HoldemActionBet, 50)
	assert.Equal(t, "error output", result)
}
