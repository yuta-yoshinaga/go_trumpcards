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

func TestNewIndianPokerInteractor(t *testing.T) {
	mg := new(interfaces.MockIndianPokerGame)
	mp := new(presenter.MockIndianPokerPresenter)
	ipi := NewIndianPokerInteractor(mg, mp)
	assert.NotNil(t, ipi)
}

func TestNewIndianPokerInteractor_NilGame(t *testing.T) {
	mp := new(presenter.MockIndianPokerPresenter)
	assert.Panics(t, func() {
		NewIndianPokerInteractor(nil, mp)
	})
}

func TestNewIndianPokerInteractor_NilPresenter(t *testing.T) {
	mg := new(interfaces.MockIndianPokerGame)
	assert.Panics(t, func() {
		NewIndianPokerInteractor(mg, nil)
	})
}

func TestIndianPokerInteractor_Reset(t *testing.T) {
	mg := new(interfaces.MockIndianPokerGame)
	mp := new(presenter.MockIndianPokerPresenter)
	ipi := NewIndianPokerInteractor(mg, mp)

	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("reset output")

	result := ipi.Reset()
	assert.Equal(t, "reset output", result)
	mg.AssertCalled(t, "Reset")
}

func TestIndianPokerInteractor_Reset_Error(t *testing.T) {
	mg := new(interfaces.MockIndianPokerGame)
	mp := new(presenter.MockIndianPokerPresenter)
	ipi := NewIndianPokerInteractor(mg, mp)

	err := errors.New("reset failed")
	mg.On("Reset").Return(err)
	mp.On("Output", mg, err).Return("error output")

	result := ipi.Reset()
	assert.Equal(t, "error output", result)
}

func TestIndianPokerInteractor_ResetWithConfig(t *testing.T) {
	mg := new(interfaces.MockIndianPokerGame)
	mp := new(presenter.MockIndianPokerPresenter)
	ipi := NewIndianPokerInteractor(mg, mp)

	cfg := domain.DefaultIndianPokerConfig()
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("reset with config output")

	result := ipi.ResetWithConfig(cfg, nil)
	assert.Equal(t, "reset with config output", result)
	mg.AssertCalled(t, "SetConfig", cfg)
	mg.AssertCalled(t, "Reset")
}

func TestIndianPokerInteractor_ResetWithConfig_Error(t *testing.T) {
	mg := new(interfaces.MockIndianPokerGame)
	mp := new(presenter.MockIndianPokerPresenter)
	ipi := NewIndianPokerInteractor(mg, mp)

	cfg := domain.DefaultIndianPokerConfig()
	err := errors.New("reset failed")
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(err)
	mp.On("Output", mg, err).Return("error output")

	result := ipi.ResetWithConfig(cfg, nil)
	assert.Equal(t, "error output", result)
}

func TestIndianPokerInteractor_ResetWithConfig_ValidationError(t *testing.T) {
	mg := new(interfaces.MockIndianPokerGame)
	mp := new(presenter.MockIndianPokerPresenter)
	ipi := NewIndianPokerInteractor(mg, mp)

	mp.On("Output", mg, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")
	cfg := domain.IndianPokerConfig{Ante: 0, InitChips: 1000, BettingLimit: domain.BettingLimitNoLimit}
	result := ipi.ResetWithConfig(cfg, nil)
	assert.Equal(t, "validation error", result)
	mg.AssertNotCalled(t, "SetConfig", mock.Anything)
}

func TestIndianPokerInteractor_ResetWithConfig_WithProfile(t *testing.T) {
	mg := new(interfaces.MockIndianPokerGame)
	mp := new(presenter.MockIndianPokerPresenter)
	ipi := NewIndianPokerInteractor(mg, mp)

	cfg := domain.DefaultIndianPokerConfig()
	profileData := []byte(`{"gamesPlayed":3}`)
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(nil)
	mg.On("ImportProfile", profileData).Return(nil)
	mp.On("Output", mg, mock.Anything).Return("with profile output")

	result := ipi.ResetWithConfig(cfg, profileData)
	assert.Equal(t, "with profile output", result)
	mg.AssertCalled(t, "ImportProfile", profileData)
}

func TestIndianPokerInteractor_Action(t *testing.T) {
	mg := new(interfaces.MockIndianPokerGame)
	mp := new(presenter.MockIndianPokerPresenter)
	ipi := NewIndianPokerInteractor(mg, mp)

	mg.On("PlayerAction", domain.IndianPokerActionCheck, 0, 0).Return(nil)
	mp.On("Output", mg, mock.Anything).Return("action output")

	result := ipi.Action(domain.IndianPokerActionCheck, 0, 0)
	assert.Equal(t, "action output", result)
	mg.AssertCalled(t, "PlayerAction", domain.IndianPokerActionCheck, 0, 0)
}

func TestIndianPokerInteractor_Action_Error(t *testing.T) {
	mg := new(interfaces.MockIndianPokerGame)
	mp := new(presenter.MockIndianPokerPresenter)
	ipi := NewIndianPokerInteractor(mg, mp)

	err := errors.New("test error")
	mg.On("PlayerAction", domain.IndianPokerActionBet, 50, 0).Return(err)
	mp.On("Output", mg, err).Return("error output")

	result := ipi.Action(domain.IndianPokerActionBet, 50, 0)
	assert.Equal(t, "error output", result)
}

func TestIndianPokerInteractor_GetConfig(t *testing.T) {
	mg := new(interfaces.MockIndianPokerGame)
	mp := new(presenter.MockIndianPokerPresenter)
	ipi := NewIndianPokerInteractor(mg, mp)

	cfg := domain.DefaultIndianPokerConfig()
	mg.On("GetConfig").Return(cfg)

	result := ipi.GetConfig()
	assert.Equal(t, cfg, result)
	mg.AssertCalled(t, "GetConfig")
}

func TestIndianPokerInteractor_ActionLog(t *testing.T) {
	mg := new(interfaces.MockIndianPokerGame)
	mp := new(presenter.MockIndianPokerPresenter)
	mp.On("ActionLogOutput", mg).Return(`{"entries":[]}`)

	ipi := NewIndianPokerInteractor(mg, mp)
	result := ipi.ActionLog()
	assert.Equal(t, `{"entries":[]}`, result)
	mp.AssertExpectations(t)
}
