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

func TestNewDragonTigerInteractor(t *testing.T) {
	mg := new(interfaces.MockDragonTigerGame)
	mp := new(presenter.MockDragonTigerPresenter)
	di := NewDragonTigerInteractor(mg, mp)
	assert.NotNil(t, di)
}

func TestNewDragonTigerInteractor_NilPanics(t *testing.T) {
	mp := new(presenter.MockDragonTigerPresenter)
	assert.Panics(t, func() { NewDragonTigerInteractor(nil, mp) })

	mg := new(interfaces.MockDragonTigerGame)
	assert.Panics(t, func() { NewDragonTigerInteractor(mg, nil) })
}

func TestDragonTigerInteractor_Reset(t *testing.T) {
	mg := new(interfaces.MockDragonTigerGame)
	mp := new(presenter.MockDragonTigerPresenter)
	di := NewDragonTigerInteractor(mg, mp)

	mg.On("Reset").Return()
	mp.On("Output", mg, nil).Return("reset output")

	assert.Equal(t, "reset output", di.Reset())
	mg.AssertCalled(t, "Reset")
}

func TestDragonTigerInteractor_Bet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mg := new(interfaces.MockDragonTigerGame)
		mp := new(presenter.MockDragonTigerPresenter)
		di := NewDragonTigerInteractor(mg, mp)

		mg.On("Bet", 100, domain.DragonTigerBetDragon).Return(nil)
		mp.On("Output", mg, nil).Return("bet output")

		assert.Equal(t, "bet output", di.Bet(100, domain.DragonTigerBetDragon))
		mg.AssertCalled(t, "Bet", 100, domain.DragonTigerBetDragon)
	})

	t.Run("error propagates to presenter", func(t *testing.T) {
		mg := new(interfaces.MockDragonTigerGame)
		mp := new(presenter.MockDragonTigerPresenter)
		di := NewDragonTigerInteractor(mg, mp)

		mg.On("Bet", 100, domain.DragonTigerBetDragon).Return(errors.New("bet error"))
		mp.On("Output", mg, mock.MatchedBy(func(e error) bool { return e != nil })).Return("error output")

		assert.Equal(t, "error output", di.Bet(100, domain.DragonTigerBetDragon))
	})
}

func TestDragonTigerInteractor_ClearHistory(t *testing.T) {
	mg := new(interfaces.MockDragonTigerGame)
	mp := new(presenter.MockDragonTigerPresenter)
	di := NewDragonTigerInteractor(mg, mp)

	mg.On("ClearHistory").Return()
	mp.On("Output", mg, nil).Return("cleared")

	assert.Equal(t, "cleared", di.ClearHistory())
	mg.AssertCalled(t, "ClearHistory")
}

func TestDragonTigerInteractor_ActionLog(t *testing.T) {
	mg := new(interfaces.MockDragonTigerGame)
	mp := new(presenter.MockDragonTigerPresenter)
	di := NewDragonTigerInteractor(mg, mp)

	mp.On("ActionLogOutput", mg).Return("log output")
	assert.Equal(t, "log output", di.ActionLog())
}

func TestRestoreDragonTigerInteractor(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		_, err := RestoreDragonTigerInteractor([]byte("not json"), nil)
		assert.Error(t, err)
	})

	t.Run("snapshot round-trip preserves phase, chips, bet", func(t *testing.T) {
		mp := new(presenter.MockDragonTigerPresenter)
		di := NewDragonTigerInteractor(domain.NewDefaultDragonTiger(), mp)

		// Drive the game past Reset so the phase advances and chip total is set.
		di.Game.Reset()
		require.NoError(t, di.Game.Bet(domain.DragonTigerMinBet, domain.DragonTigerBetDragon))

		data, err := di.Snapshot()
		require.NoError(t, err)

		restored, err := RestoreDragonTigerInteractor(data, mp)
		require.NoError(t, err)
		require.NotNil(t, restored)
		assert.Equal(t, di.Game.GetPhase(), restored.Game.GetPhase())
		assert.Equal(t, di.Game.GetChips(), restored.Game.GetChips())
		assert.Equal(t, di.Game.GetBetAmount(), restored.Game.GetBetAmount())
	})
}
