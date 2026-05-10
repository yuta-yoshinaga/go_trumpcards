//go:build test

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

func TestNewBlackJackSwitchInteractor(t *testing.T) {
	mg := new(interfaces.MockBlackJackSwitchGame)
	mp := new(presenter.MockBlackJackSwitchPresenter)
	bi := NewBlackJackSwitchInteractor(mg, mp)
	assert.NotNil(t, bi)
}

func TestNewBlackJackSwitchInteractor_NilPanics(t *testing.T) {
	mp := new(presenter.MockBlackJackSwitchPresenter)
	assert.Panics(t, func() { NewBlackJackSwitchInteractor(nil, mp) })

	mg := new(interfaces.MockBlackJackSwitchGame)
	assert.Panics(t, func() { NewBlackJackSwitchInteractor(mg, nil) })
}

func TestBlackJackSwitchInteractor_Reset(t *testing.T) {
	mg := new(interfaces.MockBlackJackSwitchGame)
	mp := new(presenter.MockBlackJackSwitchPresenter)
	bi := NewBlackJackSwitchInteractor(mg, mp)
	mg.On("Reset").Return()
	mp.On("Output", mg, nil).Return("reset")
	assert.Equal(t, "reset", bi.Reset())
}

func TestBlackJackSwitchInteractor_Bet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mg := new(interfaces.MockBlackJackSwitchGame)
		mp := new(presenter.MockBlackJackSwitchPresenter)
		bi := NewBlackJackSwitchInteractor(mg, mp)
		mg.On("PlayerBet", 100).Return(nil)
		mp.On("Output", mg, nil).Return("bet ok")
		assert.Equal(t, "bet ok", bi.Bet(100))
	})

	t.Run("error propagates", func(t *testing.T) {
		mg := new(interfaces.MockBlackJackSwitchGame)
		mp := new(presenter.MockBlackJackSwitchPresenter)
		bi := NewBlackJackSwitchInteractor(mg, mp)
		mg.On("PlayerBet", 50).Return(errors.New("nope"))
		mp.On("Output", mg, mock.MatchedBy(func(e error) bool { return e != nil })).Return("err")
		assert.Equal(t, "err", bi.Bet(50))
	})
}

func TestBlackJackSwitchInteractor_ActionsForward(t *testing.T) {
	cases := []struct {
		name     string
		action   string
		callMock func(mg *interfaces.MockBlackJackSwitchGame)
		invoke   func(bi *BlackJackSwitchInteractor) string
	}{
		{
			"switch", "PlayerSwitch",
			func(mg *interfaces.MockBlackJackSwitchGame) { mg.On("PlayerSwitch").Return(nil) },
			func(bi *BlackJackSwitchInteractor) string { return bi.Switch() },
		},
		{
			"keep", "PlayerKeep",
			func(mg *interfaces.MockBlackJackSwitchGame) { mg.On("PlayerKeep").Return(nil) },
			func(bi *BlackJackSwitchInteractor) string { return bi.Keep() },
		},
		{
			"hit", "PlayerHit",
			func(mg *interfaces.MockBlackJackSwitchGame) { mg.On("PlayerHit").Return(nil) },
			func(bi *BlackJackSwitchInteractor) string { return bi.Hit() },
		},
		{
			"stand", "PlayerStand",
			func(mg *interfaces.MockBlackJackSwitchGame) { mg.On("PlayerStand").Return(nil) },
			func(bi *BlackJackSwitchInteractor) string { return bi.Stand() },
		},
		{
			"doubledown", "PlayerDoubleDown",
			func(mg *interfaces.MockBlackJackSwitchGame) { mg.On("PlayerDoubleDown").Return(nil) },
			func(bi *BlackJackSwitchInteractor) string { return bi.DoubleDown() },
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mg := new(interfaces.MockBlackJackSwitchGame)
			mp := new(presenter.MockBlackJackSwitchPresenter)
			bi := NewBlackJackSwitchInteractor(mg, mp)
			c.callMock(mg)
			mp.On("Output", mg, nil).Return("ok")
			assert.Equal(t, "ok", c.invoke(bi))
			mg.AssertCalled(t, c.action)
		})
	}
}

func TestBlackJackSwitchInteractor_ActionLog(t *testing.T) {
	mg := new(interfaces.MockBlackJackSwitchGame)
	mp := new(presenter.MockBlackJackSwitchPresenter)
	bi := NewBlackJackSwitchInteractor(mg, mp)
	mp.On("ActionLogOutput", mg).Return("log")
	assert.Equal(t, "log", bi.ActionLog())
}

func TestRestoreBlackJackSwitchInteractor(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		_, err := RestoreBlackJackSwitchInteractor([]byte("not json"), nil)
		assert.Error(t, err)
	})

	t.Run("snapshot round-trip", func(t *testing.T) {
		mp := new(presenter.MockBlackJackSwitchPresenter)
		bi := NewBlackJackSwitchInteractor(domain.NewDefaultBlackJackSwitch(), mp)
		bi.Game.Reset()
		require.NoError(t, bi.Game.PlayerBet(domain.BJSwitchMinBet))
		data, err := bi.Snapshot()
		require.NoError(t, err)
		restored, err := RestoreBlackJackSwitchInteractor(data, mp)
		require.NoError(t, err)
		require.NotNil(t, restored)
		assert.Equal(t, bi.Game.GetPhase(), restored.Game.GetPhase())
		assert.Equal(t, bi.Game.GetPlayer().GetChips(), restored.Game.GetPlayer().GetChips())
	})
}
