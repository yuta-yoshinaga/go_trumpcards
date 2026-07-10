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

func TestNewOichoKabuInteractor(t *testing.T) {
	mg := new(interfaces.MockOichoKabuGame)
	mp := new(presenter.MockOichoKabuPresenter)
	oi := NewOichoKabuInteractor(mg, mp)
	assert.NotNil(t, oi)
}

func TestNewOichoKabuInteractor_NilPanics(t *testing.T) {
	mp := new(presenter.MockOichoKabuPresenter)
	assert.Panics(t, func() { NewOichoKabuInteractor(nil, mp) })

	mg := new(interfaces.MockOichoKabuGame)
	assert.Panics(t, func() { NewOichoKabuInteractor(mg, nil) })
}

func TestOichoKabuInteractor_Reset(t *testing.T) {
	mg := new(interfaces.MockOichoKabuGame)
	mp := new(presenter.MockOichoKabuPresenter)
	oi := NewOichoKabuInteractor(mg, mp)

	mg.On("Reset").Return()
	mp.On("Output", mg, nil).Return("reset output")

	assert.Equal(t, "reset output", oi.Reset())
	mg.AssertCalled(t, "Reset")
}

func TestOichoKabuInteractor_Bet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mg := new(interfaces.MockOichoKabuGame)
		mp := new(presenter.MockOichoKabuPresenter)
		oi := NewOichoKabuInteractor(mg, mp)

		mg.On("Bet", 100).Return(nil)
		mp.On("Output", mg, nil).Return("bet output")

		assert.Equal(t, "bet output", oi.Bet(100))
		mg.AssertCalled(t, "Bet", 100)
	})

	t.Run("error", func(t *testing.T) {
		mg := new(interfaces.MockOichoKabuGame)
		mp := new(presenter.MockOichoKabuPresenter)
		oi := NewOichoKabuInteractor(mg, mp)

		mg.On("Bet", 100).Return(errors.New("bet error"))
		mp.On("Output", mg, mock.MatchedBy(func(e error) bool { return e != nil })).Return("error output")

		assert.Equal(t, "error output", oi.Bet(100))
	})
}

func TestOichoKabuInteractor_Draw(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mg := new(interfaces.MockOichoKabuGame)
		mp := new(presenter.MockOichoKabuPresenter)
		oi := NewOichoKabuInteractor(mg, mp)

		mg.On("Draw").Return(nil)
		mp.On("Output", mg, nil).Return("draw ok")

		assert.Equal(t, "draw ok", oi.Draw())
	})

	t.Run("error", func(t *testing.T) {
		mg := new(interfaces.MockOichoKabuGame)
		mp := new(presenter.MockOichoKabuPresenter)
		oi := NewOichoKabuInteractor(mg, mp)

		mg.On("Draw").Return(errors.New("nope"))
		mp.On("Output", mg, mock.MatchedBy(func(e error) bool { return e != nil })).Return("draw err")

		assert.Equal(t, "draw err", oi.Draw())
	})
}

func TestOichoKabuInteractor_Stand(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mg := new(interfaces.MockOichoKabuGame)
		mp := new(presenter.MockOichoKabuPresenter)
		oi := NewOichoKabuInteractor(mg, mp)

		mg.On("Stand").Return(nil)
		mp.On("Output", mg, nil).Return("stand ok")

		assert.Equal(t, "stand ok", oi.Stand())
	})

	t.Run("error", func(t *testing.T) {
		mg := new(interfaces.MockOichoKabuGame)
		mp := new(presenter.MockOichoKabuPresenter)
		oi := NewOichoKabuInteractor(mg, mp)

		mg.On("Stand").Return(errors.New("nope"))
		mp.On("Output", mg, mock.MatchedBy(func(e error) bool { return e != nil })).Return("stand err")

		assert.Equal(t, "stand err", oi.Stand())
	})
}

func TestOichoKabuInteractor_ActionLog(t *testing.T) {
	mg := new(interfaces.MockOichoKabuGame)
	mp := new(presenter.MockOichoKabuPresenter)
	oi := NewOichoKabuInteractor(mg, mp)

	mp.On("ActionLogOutput", mg).Return("log output")
	assert.Equal(t, "log output", oi.ActionLog())
}

func TestRestoreOichoKabuInteractor(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		_, err := RestoreOichoKabuInteractor([]byte("not json"), nil)
		assert.Error(t, err)
	})

	t.Run("snapshot round-trip preserves phase, chips, bet", func(t *testing.T) {
		mp := new(presenter.MockOichoKabuPresenter)
		oi := NewOichoKabuInteractor(domain.NewDefaultOichoKabu(), mp)

		oi.Game.Reset()
		require.NoError(t, oi.Game.Bet(domain.OichoKabuMinBet))

		data, err := oi.Snapshot()
		require.NoError(t, err)

		restored, err := RestoreOichoKabuInteractor(data, mp)
		require.NoError(t, err)
		require.NotNil(t, restored)
		assert.Equal(t, oi.Game.GetPhase(), restored.Game.GetPhase())
		assert.Equal(t, oi.Game.GetChips(), restored.Game.GetChips())
		assert.Equal(t, oi.Game.GetBet(), restored.Game.GetBet())
	})
}
