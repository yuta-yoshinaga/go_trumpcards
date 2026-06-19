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

func TestNewFiveCardStudInteractor(t *testing.T) {
	mg := new(interfaces.MockFiveCardStudGame)
	mp := new(presenter.MockFiveCardStudPresenter)
	si := NewFiveCardStudInteractor(mg, mp)
	assert.NotNil(t, si)
}

func TestNewFiveCardStudInteractor_NilGame(t *testing.T) {
	mp := new(presenter.MockFiveCardStudPresenter)
	assert.Panics(t, func() {
		NewFiveCardStudInteractor(nil, mp)
	})
}

func TestNewFiveCardStudInteractor_NilPresenter(t *testing.T) {
	mg := new(interfaces.MockFiveCardStudGame)
	assert.Panics(t, func() {
		NewFiveCardStudInteractor(mg, nil)
	})
}

func TestFiveCardStudInteractor_Reset(t *testing.T) {
	mg := new(interfaces.MockFiveCardStudGame)
	mp := new(presenter.MockFiveCardStudPresenter)
	si := NewFiveCardStudInteractor(mg, mp)

	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("reset output")

	result := si.Reset()
	assert.Equal(t, "reset output", result)
	mg.AssertCalled(t, "Reset")
}

func TestFiveCardStudInteractor_Reset_Error(t *testing.T) {
	mg := new(interfaces.MockFiveCardStudGame)
	mp := new(presenter.MockFiveCardStudPresenter)
	si := NewFiveCardStudInteractor(mg, mp)

	err := errors.New("reset failed")
	mg.On("Reset").Return(err)
	mp.On("Output", mg, err).Return("error output")

	result := si.Reset()
	assert.Equal(t, "error output", result)
}

func TestFiveCardStudInteractor_ResetWithConfig(t *testing.T) {
	mg := new(interfaces.MockFiveCardStudGame)
	mp := new(presenter.MockFiveCardStudPresenter)
	si := NewFiveCardStudInteractor(mg, mp)

	cfg := domain.DefaultFiveCardStudConfig()
	mg.On("GetPlayerCnt").Return(cfg.TableSize)
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("reset output")

	result := si.ResetWithConfig(cfg, nil)
	assert.Equal(t, "reset output", result)
}

func TestFiveCardStudInteractor_ResetWithConfig_ValidationError(t *testing.T) {
	mg := new(interfaces.MockFiveCardStudGame)
	mp := new(presenter.MockFiveCardStudPresenter)
	si := NewFiveCardStudInteractor(mg, mp)

	cfg := domain.FiveCardStudConfig{Ante: 0} // invalid
	mp.On("Output", mg, mock.Anything).Return("validation error")

	result := si.ResetWithConfig(cfg, nil)
	assert.Equal(t, "validation error", result)
}

func TestFiveCardStudInteractor_ResetWithConfig_Resize(t *testing.T) {
	mg := new(interfaces.MockFiveCardStudGame)
	mp := new(presenter.MockFiveCardStudPresenter)
	si := NewFiveCardStudInteractor(mg, mp)

	cfg := domain.DefaultFiveCardStudConfig()
	cfg.TableSize = 4
	mg.On("GetPlayerCnt").Return(6) // different from cfg.TableSize
	mg.On("Resize", mock.Anything).Return()
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("resized output")

	result := si.ResetWithConfig(cfg, nil)
	assert.Equal(t, "resized output", result)
	mg.AssertCalled(t, "Resize", mock.Anything)
}

func TestFiveCardStudInteractor_ResetWithConfig_Error(t *testing.T) {
	mg := new(interfaces.MockFiveCardStudGame)
	mp := new(presenter.MockFiveCardStudPresenter)
	si := NewFiveCardStudInteractor(mg, mp)

	cfg := domain.DefaultFiveCardStudConfig()
	err := errors.New("reset failed")
	mg.On("GetPlayerCnt").Return(cfg.TableSize)
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(err)
	mp.On("Output", mg, err).Return("error output")

	result := si.ResetWithConfig(cfg, nil)
	assert.Equal(t, "error output", result)
}

func TestFiveCardStudInteractor_Action(t *testing.T) {
	mg := new(interfaces.MockFiveCardStudGame)
	mp := new(presenter.MockFiveCardStudPresenter)
	si := NewFiveCardStudInteractor(mg, mp)

	mg.On("PlayerAction", 0, 0, 100).Return(nil)
	mp.On("Output", mg, mock.Anything).Return("action output")

	result := si.Action(0, 0, 100)
	assert.Equal(t, "action output", result)
}

func TestFiveCardStudInteractor_GetConfig(t *testing.T) {
	mg := new(interfaces.MockFiveCardStudGame)
	mp := new(presenter.MockFiveCardStudPresenter)
	si := NewFiveCardStudInteractor(mg, mp)

	cfg := domain.DefaultFiveCardStudConfig()
	mg.On("GetConfig").Return(cfg)

	result := si.GetConfig()
	assert.Equal(t, cfg, result)
}

func TestFiveCardStudInteractor_ActionLog(t *testing.T) {
	mg := new(interfaces.MockFiveCardStudGame)
	mp := new(presenter.MockFiveCardStudPresenter)
	si := NewFiveCardStudInteractor(mg, mp)

	mp.On("ActionLogOutput", mg).Return("log output")

	result := si.ActionLog()
	assert.Equal(t, "log output", result)
}

func TestFiveCardStudInteractor_Snapshot(t *testing.T) {
	cfg := domain.DefaultFiveCardStudConfig()
	cfg.TableSize = 4
	players := domain.NewFiveCardStudPlayersForTable(cfg.TableSize)
	tc := domain.NewTrumpCards(0)
	s := domain.NewFiveCardStud(tc, players, cfg)
	mp := new(presenter.MockFiveCardStudPresenter)
	si := NewFiveCardStudInteractor(s, mp)

	data, err := si.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreFiveCardStudInteractor(t *testing.T) {
	cfg := domain.DefaultFiveCardStudConfig()
	cfg.TableSize = 4
	players := domain.NewFiveCardStudPlayersForTable(cfg.TableSize)
	tc := domain.NewTrumpCards(0)
	s := domain.NewFiveCardStud(tc, players, cfg)
	mp := new(presenter.MockFiveCardStudPresenter)
	si := NewFiveCardStudInteractor(s, mp)

	data, err := si.Snapshot()
	assert.NoError(t, err)

	restored, err := RestoreFiveCardStudInteractor(data, mp)
	assert.NoError(t, err)
	assert.NotNil(t, restored)
}

func TestRestoreFiveCardStudInteractor_InvalidJSON(t *testing.T) {
	mp := new(presenter.MockFiveCardStudPresenter)
	_, err := RestoreFiveCardStudInteractor([]byte("invalid"), mp)
	assert.Error(t, err)
}

func TestFiveCardStudInteractor_ResetWithConfig_WithProfile(t *testing.T) {
	mg := new(interfaces.MockFiveCardStudGame)
	mp := new(presenter.MockFiveCardStudPresenter)
	si := NewFiveCardStudInteractor(mg, mp)

	cfg := domain.DefaultFiveCardStudConfig()
	mg.On("GetPlayerCnt").Return(cfg.TableSize)
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(nil)
	mg.On("ImportProfile", mock.Anything).Return(nil)
	mp.On("Output", mg, mock.Anything).Return("profile output")

	result := si.ResetWithConfig(cfg, []byte(`{"gp":1}`))
	assert.Equal(t, "profile output", result)
	mg.AssertCalled(t, "ImportProfile", mock.Anything)
}
