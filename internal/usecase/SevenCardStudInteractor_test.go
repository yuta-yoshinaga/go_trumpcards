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

func TestNewSevenCardStudInteractor(t *testing.T) {
	mg := new(interfaces.MockSevenCardStudGame)
	mp := new(presenter.MockSevenCardStudPresenter)
	si := NewSevenCardStudInteractor(mg, mp)
	assert.NotNil(t, si)
}

func TestNewSevenCardStudInteractor_NilGame(t *testing.T) {
	mp := new(presenter.MockSevenCardStudPresenter)
	assert.Panics(t, func() {
		NewSevenCardStudInteractor(nil, mp)
	})
}

func TestNewSevenCardStudInteractor_NilPresenter(t *testing.T) {
	mg := new(interfaces.MockSevenCardStudGame)
	assert.Panics(t, func() {
		NewSevenCardStudInteractor(mg, nil)
	})
}

func TestSevenCardStudInteractor_Reset(t *testing.T) {
	mg := new(interfaces.MockSevenCardStudGame)
	mp := new(presenter.MockSevenCardStudPresenter)
	si := NewSevenCardStudInteractor(mg, mp)

	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("reset output")

	result := si.Reset()
	assert.Equal(t, "reset output", result)
	mg.AssertCalled(t, "Reset")
}

func TestSevenCardStudInteractor_Reset_Error(t *testing.T) {
	mg := new(interfaces.MockSevenCardStudGame)
	mp := new(presenter.MockSevenCardStudPresenter)
	si := NewSevenCardStudInteractor(mg, mp)

	err := errors.New("reset failed")
	mg.On("Reset").Return(err)
	mp.On("Output", mg, err).Return("error output")

	result := si.Reset()
	assert.Equal(t, "error output", result)
}

func TestSevenCardStudInteractor_ResetWithConfig(t *testing.T) {
	mg := new(interfaces.MockSevenCardStudGame)
	mp := new(presenter.MockSevenCardStudPresenter)
	si := NewSevenCardStudInteractor(mg, mp)

	cfg := domain.DefaultSevenCardStudConfig()
	mg.On("GetPlayerCnt").Return(cfg.TableSize)
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("reset output")

	result := si.ResetWithConfig(cfg, nil)
	assert.Equal(t, "reset output", result)
}

func TestSevenCardStudInteractor_ResetWithConfig_ValidationError(t *testing.T) {
	mg := new(interfaces.MockSevenCardStudGame)
	mp := new(presenter.MockSevenCardStudPresenter)
	si := NewSevenCardStudInteractor(mg, mp)

	cfg := domain.SevenCardStudConfig{Ante: 0} // invalid
	mp.On("Output", mg, mock.Anything).Return("validation error")

	result := si.ResetWithConfig(cfg, nil)
	assert.Equal(t, "validation error", result)
}

func TestSevenCardStudInteractor_ResetWithConfig_Resize(t *testing.T) {
	mg := new(interfaces.MockSevenCardStudGame)
	mp := new(presenter.MockSevenCardStudPresenter)
	si := NewSevenCardStudInteractor(mg, mp)

	cfg := domain.DefaultSevenCardStudConfig()
	cfg.TableSize = 4
	mg.On("GetPlayerCnt").Return(7) // different from cfg.TableSize
	mg.On("Resize", mock.Anything).Return()
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("resized output")

	result := si.ResetWithConfig(cfg, nil)
	assert.Equal(t, "resized output", result)
	mg.AssertCalled(t, "Resize", mock.Anything)
}

func TestSevenCardStudInteractor_ResetWithConfig_Error(t *testing.T) {
	mg := new(interfaces.MockSevenCardStudGame)
	mp := new(presenter.MockSevenCardStudPresenter)
	si := NewSevenCardStudInteractor(mg, mp)

	cfg := domain.DefaultSevenCardStudConfig()
	err := errors.New("reset failed")
	mg.On("GetPlayerCnt").Return(cfg.TableSize)
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(err)
	mp.On("Output", mg, err).Return("error output")

	result := si.ResetWithConfig(cfg, nil)
	assert.Equal(t, "error output", result)
}

func TestSevenCardStudInteractor_Action(t *testing.T) {
	mg := new(interfaces.MockSevenCardStudGame)
	mp := new(presenter.MockSevenCardStudPresenter)
	si := NewSevenCardStudInteractor(mg, mp)

	mg.On("PlayerAction", 0, 0, 100).Return(nil)
	mp.On("Output", mg, mock.Anything).Return("action output")

	result := si.Action(0, 0, 100)
	assert.Equal(t, "action output", result)
}

func TestSevenCardStudInteractor_GetConfig(t *testing.T) {
	mg := new(interfaces.MockSevenCardStudGame)
	mp := new(presenter.MockSevenCardStudPresenter)
	si := NewSevenCardStudInteractor(mg, mp)

	cfg := domain.DefaultSevenCardStudConfig()
	mg.On("GetConfig").Return(cfg)

	result := si.GetConfig()
	assert.Equal(t, cfg, result)
}

func TestSevenCardStudInteractor_ActionLog(t *testing.T) {
	mg := new(interfaces.MockSevenCardStudGame)
	mp := new(presenter.MockSevenCardStudPresenter)
	si := NewSevenCardStudInteractor(mg, mp)

	mp.On("ActionLogOutput", mg).Return("log output")

	result := si.ActionLog()
	assert.Equal(t, "log output", result)
}

func TestSevenCardStudInteractor_Snapshot(t *testing.T) {
	cfg := domain.DefaultSevenCardStudConfig()
	cfg.TableSize = 4
	players := domain.NewSevenCardStudPlayersForTable(cfg.TableSize)
	tc := domain.NewTrumpCards(0)
	s := domain.NewSevenCardStud(tc, players, cfg)
	mp := new(presenter.MockSevenCardStudPresenter)
	si := NewSevenCardStudInteractor(s, mp)

	data, err := si.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreSevenCardStudInteractor(t *testing.T) {
	cfg := domain.DefaultSevenCardStudConfig()
	cfg.TableSize = 4
	players := domain.NewSevenCardStudPlayersForTable(cfg.TableSize)
	tc := domain.NewTrumpCards(0)
	s := domain.NewSevenCardStud(tc, players, cfg)
	mp := new(presenter.MockSevenCardStudPresenter)
	si := NewSevenCardStudInteractor(s, mp)

	data, err := si.Snapshot()
	assert.NoError(t, err)

	restored, err := RestoreSevenCardStudInteractor(data, mp)
	assert.NoError(t, err)
	assert.NotNil(t, restored)
}

func TestRestoreSevenCardStudInteractor_InvalidJSON(t *testing.T) {
	mp := new(presenter.MockSevenCardStudPresenter)
	_, err := RestoreSevenCardStudInteractor([]byte("invalid"), mp)
	assert.Error(t, err)
}

func TestSevenCardStudInteractor_ResetWithConfig_WithProfile(t *testing.T) {
	mg := new(interfaces.MockSevenCardStudGame)
	mp := new(presenter.MockSevenCardStudPresenter)
	si := NewSevenCardStudInteractor(mg, mp)

	cfg := domain.DefaultSevenCardStudConfig()
	mg.On("GetPlayerCnt").Return(cfg.TableSize)
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(nil)
	mg.On("ImportProfile", mock.Anything).Return(nil)
	mp.On("Output", mg, mock.Anything).Return("profile output")

	result := si.ResetWithConfig(cfg, []byte(`{"gp":1}`))
	assert.Equal(t, "profile output", result)
	mg.AssertCalled(t, "ImportProfile", mock.Anything)
}

func TestSevenCardStudInteractor_Hint(t *testing.T) {
	mg := new(interfaces.MockSevenCardStudGame)
	mp := new(presenter.MockSevenCardStudPresenter)
	si := NewSevenCardStudInteractor(mg, mp)

	mp.On("HintOutput", mg).Return("hint output")

	assert.Equal(t, "hint output", si.Hint())
}
