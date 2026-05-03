package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func TestNewCasinoWarInteractor(t *testing.T) {
	mg := new(interfaces.MockCasinoWarGame)
	mp := new(presenter.MockCasinoWarPresenter)
	ci := NewCasinoWarInteractor(mg, mp)
	assert.NotNil(t, ci)
}

func TestNewCasinoWarInteractor_NilPanics(t *testing.T) {
	mp := new(presenter.MockCasinoWarPresenter)
	assert.Panics(t, func() { NewCasinoWarInteractor(nil, mp) })

	mg := new(interfaces.MockCasinoWarGame)
	assert.Panics(t, func() { NewCasinoWarInteractor(mg, nil) })
}

func TestCasinoWarInteractor_Reset(t *testing.T) {
	mg := new(interfaces.MockCasinoWarGame)
	mp := new(presenter.MockCasinoWarPresenter)
	ci := NewCasinoWarInteractor(mg, mp)

	mg.On("Reset").Return()
	mp.On("Output", mg, nil).Return("reset output")

	assert.Equal(t, "reset output", ci.Reset())
	mg.AssertCalled(t, "Reset")
}

func TestCasinoWarInteractor_Bet(t *testing.T) {
	t.Run("success calls ResolveInitial", func(t *testing.T) {
		mg := new(interfaces.MockCasinoWarGame)
		mp := new(presenter.MockCasinoWarPresenter)
		ci := NewCasinoWarInteractor(mg, mp)

		mg.On("Bet", 100).Return(nil)
		mg.On("ResolveInitial").Return()
		mp.On("Output", mg, nil).Return("bet output")

		assert.Equal(t, "bet output", ci.Bet(100))
		mg.AssertCalled(t, "ResolveInitial")
	})

	t.Run("error skips ResolveInitial", func(t *testing.T) {
		mg := new(interfaces.MockCasinoWarGame)
		mp := new(presenter.MockCasinoWarPresenter)
		ci := NewCasinoWarInteractor(mg, mp)

		mg.On("Bet", 100).Return(errors.New("bet error"))
		mp.On("Output", mg, mock.MatchedBy(func(e error) bool { return e != nil })).Return("error output")

		assert.Equal(t, "error output", ci.Bet(100))
		mg.AssertNotCalled(t, "ResolveInitial")
	})
}

func TestCasinoWarInteractor_Surrender(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mg := new(interfaces.MockCasinoWarGame)
		mp := new(presenter.MockCasinoWarPresenter)
		ci := NewCasinoWarInteractor(mg, mp)

		mg.On("Surrender").Return(nil)
		mp.On("Output", mg, nil).Return("surrender ok")

		assert.Equal(t, "surrender ok", ci.Surrender())
	})

	t.Run("error", func(t *testing.T) {
		mg := new(interfaces.MockCasinoWarGame)
		mp := new(presenter.MockCasinoWarPresenter)
		ci := NewCasinoWarInteractor(mg, mp)

		mg.On("Surrender").Return(errors.New("nope"))
		mp.On("Output", mg, mock.MatchedBy(func(e error) bool { return e != nil })).Return("surrender err")

		assert.Equal(t, "surrender err", ci.Surrender())
	})
}

func TestCasinoWarInteractor_War(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mg := new(interfaces.MockCasinoWarGame)
		mp := new(presenter.MockCasinoWarPresenter)
		ci := NewCasinoWarInteractor(mg, mp)

		mg.On("GoToWar").Return(nil)
		mp.On("Output", mg, nil).Return("war ok")

		assert.Equal(t, "war ok", ci.War())
	})

	t.Run("error", func(t *testing.T) {
		mg := new(interfaces.MockCasinoWarGame)
		mp := new(presenter.MockCasinoWarPresenter)
		ci := NewCasinoWarInteractor(mg, mp)

		mg.On("GoToWar").Return(errors.New("nope"))
		mp.On("Output", mg, mock.MatchedBy(func(e error) bool { return e != nil })).Return("war err")

		assert.Equal(t, "war err", ci.War())
	})
}

func TestCasinoWarInteractor_ActionLog(t *testing.T) {
	mg := new(interfaces.MockCasinoWarGame)
	mp := new(presenter.MockCasinoWarPresenter)
	ci := NewCasinoWarInteractor(mg, mp)

	mp.On("ActionLogOutput", mg).Return("log output")
	assert.Equal(t, "log output", ci.ActionLog())
}

func TestRestoreCasinoWarInteractor(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		_, err := RestoreCasinoWarInteractor([]byte("not json"), nil)
		assert.Error(t, err)
	})
}
