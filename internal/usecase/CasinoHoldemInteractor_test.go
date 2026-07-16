package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func TestNewCasinoHoldemInteractor(t *testing.T) {
	mockGame := new(interfaces.MockCasinoHoldemGame)
	mockPresenter := new(presenter.MockCasinoHoldemPresenter)
	ci := NewCasinoHoldemInteractor(mockGame, mockPresenter)
	assert.NotNil(t, ci)
}

func TestNewCasinoHoldemInteractor_NilPanics(t *testing.T) {
	mockPresenter := new(presenter.MockCasinoHoldemPresenter)
	assert.Panics(t, func() { NewCasinoHoldemInteractor(nil, mockPresenter) })

	mockGame := new(interfaces.MockCasinoHoldemGame)
	assert.Panics(t, func() { NewCasinoHoldemInteractor(mockGame, nil) })
}

func TestCasinoHoldemInteractor_Reset(t *testing.T) {
	mockGame := new(interfaces.MockCasinoHoldemGame)
	mockPresenter := new(presenter.MockCasinoHoldemPresenter)
	ci := NewCasinoHoldemInteractor(mockGame, mockPresenter)

	mockGame.On("Reset").Return()
	mockPresenter.On("Output", mockGame, nil).Return("reset output")

	result := ci.Reset()
	assert.Equal(t, "reset output", result)
	mockGame.AssertCalled(t, "Reset")
}

func TestCasinoHoldemInteractor_Bet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockGame := new(interfaces.MockCasinoHoldemGame)
		mockPresenter := new(presenter.MockCasinoHoldemPresenter)
		ci := NewCasinoHoldemInteractor(mockGame, mockPresenter)

		mockGame.On("Bet", 100, 10).Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("bet output")

		result := ci.Bet(100, 10)
		assert.Equal(t, "bet output", result)
	})

	t.Run("error", func(t *testing.T) {
		mockGame := new(interfaces.MockCasinoHoldemGame)
		mockPresenter := new(presenter.MockCasinoHoldemPresenter)
		ci := NewCasinoHoldemInteractor(mockGame, mockPresenter)

		err := errors.New("test error")
		mockGame.On("Bet", 100, 0).Return(err)
		mockPresenter.On("Output", mockGame, mock.MatchedBy(func(e error) bool {
			return e != nil && e.Error() == "test error"
		})).Return("error output")

		result := ci.Bet(100, 0)
		assert.Equal(t, "error output", result)
	})
}

func TestCasinoHoldemInteractor_Call(t *testing.T) {
	mockGame := new(interfaces.MockCasinoHoldemGame)
	mockPresenter := new(presenter.MockCasinoHoldemPresenter)
	ci := NewCasinoHoldemInteractor(mockGame, mockPresenter)

	mockGame.On("Call").Return(nil)
	mockPresenter.On("Output", mockGame, nil).Return("call output")

	assert.Equal(t, "call output", ci.Call())
}

func TestCasinoHoldemInteractor_Fold(t *testing.T) {
	mockGame := new(interfaces.MockCasinoHoldemGame)
	mockPresenter := new(presenter.MockCasinoHoldemPresenter)
	ci := NewCasinoHoldemInteractor(mockGame, mockPresenter)

	mockGame.On("Fold").Return(nil)
	mockPresenter.On("Output", mockGame, nil).Return("fold output")

	assert.Equal(t, "fold output", ci.Fold())
}

func TestCasinoHoldemInteractor_ActionLog(t *testing.T) {
	mockGame := new(interfaces.MockCasinoHoldemGame)
	mockPresenter := new(presenter.MockCasinoHoldemPresenter)
	ci := NewCasinoHoldemInteractor(mockGame, mockPresenter)

	mockPresenter.On("ActionLogOutput", mockGame).Return("log output")

	assert.Equal(t, "log output", ci.ActionLog())
}

func TestRestoreCasinoHoldemInteractor(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		_, err := RestoreCasinoHoldemInteractor([]byte("not json"), nil)
		assert.Error(t, err)
	})

	t.Run("snapshot round-trip preserves phase, chips, ante, bonus", func(t *testing.T) {
		mp := new(presenter.MockCasinoHoldemPresenter)
		ci := NewCasinoHoldemInteractor(domain.NewDefaultCasinoHoldem(), mp)

		// Drive the game past Bet so the phase advances and the bets are set.
		require.NoError(t, ci.Game.Bet(100, 10))

		data, err := ci.Snapshot()
		require.NoError(t, err)

		restored, err := RestoreCasinoHoldemInteractor(data, mp)
		require.NoError(t, err)
		require.NotNil(t, restored)
		assert.Equal(t, ci.Game.GetPhase(), restored.Game.GetPhase())
		assert.Equal(t, ci.Game.GetChips(), restored.Game.GetChips())
		assert.Equal(t, ci.Game.GetAnteBet(), restored.Game.GetAnteBet())
		assert.Equal(t, ci.Game.GetBonusBet(), restored.Game.GetBonusBet())
	})
}

func TestCasinoHoldemInteractor_Hint(t *testing.T) {
	mockGame := new(interfaces.MockCasinoHoldemGame)
	mockPresenter := new(presenter.MockCasinoHoldemPresenter)
	ci := NewCasinoHoldemInteractor(mockGame, mockPresenter)

	mockPresenter.On("HintOutput", mockGame).Return("hint output")
	assert.Equal(t, "hint output", ci.Hint())
}
