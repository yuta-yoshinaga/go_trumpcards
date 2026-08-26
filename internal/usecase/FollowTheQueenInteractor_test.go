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

func TestNewFollowTheQueenInteractor(t *testing.T) {
	mg := new(interfaces.MockFollowTheQueenGame)
	mp := new(presenter.MockFollowTheQueenPresenter)
	si := NewFollowTheQueenInteractor(mg, mp)
	assert.NotNil(t, si)
}

func TestNewFollowTheQueenInteractor_NilGame(t *testing.T) {
	mp := new(presenter.MockFollowTheQueenPresenter)
	assert.Panics(t, func() {
		NewFollowTheQueenInteractor(nil, mp)
	})
}

func TestNewFollowTheQueenInteractor_NilPresenter(t *testing.T) {
	mg := new(interfaces.MockFollowTheQueenGame)
	assert.Panics(t, func() {
		NewFollowTheQueenInteractor(mg, nil)
	})
}

func TestFollowTheQueenInteractor_Reset(t *testing.T) {
	mg := new(interfaces.MockFollowTheQueenGame)
	mp := new(presenter.MockFollowTheQueenPresenter)
	si := NewFollowTheQueenInteractor(mg, mp)

	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("reset output")

	result := si.Reset()
	assert.Equal(t, "reset output", result)
	mg.AssertCalled(t, "Reset")
}

func TestFollowTheQueenInteractor_Reset_Error(t *testing.T) {
	mg := new(interfaces.MockFollowTheQueenGame)
	mp := new(presenter.MockFollowTheQueenPresenter)
	si := NewFollowTheQueenInteractor(mg, mp)

	err := errors.New("reset failed")
	mg.On("Reset").Return(err)
	mp.On("Output", mg, err).Return("error output")

	result := si.Reset()
	assert.Equal(t, "error output", result)
}

func TestFollowTheQueenInteractor_ResetWithConfig(t *testing.T) {
	mg := new(interfaces.MockFollowTheQueenGame)
	mp := new(presenter.MockFollowTheQueenPresenter)
	si := NewFollowTheQueenInteractor(mg, mp)

	cfg := domain.DefaultFollowTheQueenConfig()
	mg.On("GetPlayerCnt").Return(cfg.TableSize)
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("reset output")

	result := si.ResetWithConfig(cfg, nil)
	assert.Equal(t, "reset output", result)
}

func TestFollowTheQueenInteractor_ResetWithConfig_ValidationError(t *testing.T) {
	mg := new(interfaces.MockFollowTheQueenGame)
	mp := new(presenter.MockFollowTheQueenPresenter)
	si := NewFollowTheQueenInteractor(mg, mp)

	cfg := domain.FollowTheQueenConfig{Ante: 0} // invalid
	mp.On("Output", mg, mock.Anything).Return("validation error")

	result := si.ResetWithConfig(cfg, nil)
	assert.Equal(t, "validation error", result)
}

func TestFollowTheQueenInteractor_ResetWithConfig_Resize(t *testing.T) {
	mg := new(interfaces.MockFollowTheQueenGame)
	mp := new(presenter.MockFollowTheQueenPresenter)
	si := NewFollowTheQueenInteractor(mg, mp)

	cfg := domain.DefaultFollowTheQueenConfig()
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

func TestFollowTheQueenInteractor_ResetWithConfig_Error(t *testing.T) {
	mg := new(interfaces.MockFollowTheQueenGame)
	mp := new(presenter.MockFollowTheQueenPresenter)
	si := NewFollowTheQueenInteractor(mg, mp)

	cfg := domain.DefaultFollowTheQueenConfig()
	err := errors.New("reset failed")
	mg.On("GetPlayerCnt").Return(cfg.TableSize)
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(err)
	mp.On("Output", mg, err).Return("error output")

	result := si.ResetWithConfig(cfg, nil)
	assert.Equal(t, "error output", result)
}

func TestFollowTheQueenInteractor_Action(t *testing.T) {
	mg := new(interfaces.MockFollowTheQueenGame)
	mp := new(presenter.MockFollowTheQueenPresenter)
	si := NewFollowTheQueenInteractor(mg, mp)

	mg.On("PlayerAction", 0, 0, 100).Return(nil)
	mp.On("Output", mg, mock.Anything).Return("action output")

	result := si.Action(0, 0, 100)
	assert.Equal(t, "action output", result)
}

func TestFollowTheQueenInteractor_GetConfig(t *testing.T) {
	mg := new(interfaces.MockFollowTheQueenGame)
	mp := new(presenter.MockFollowTheQueenPresenter)
	si := NewFollowTheQueenInteractor(mg, mp)

	cfg := domain.DefaultFollowTheQueenConfig()
	mg.On("GetConfig").Return(cfg)

	result := si.GetConfig()
	assert.Equal(t, cfg, result)
}

func TestFollowTheQueenInteractor_ActionLog(t *testing.T) {
	mg := new(interfaces.MockFollowTheQueenGame)
	mp := new(presenter.MockFollowTheQueenPresenter)
	si := NewFollowTheQueenInteractor(mg, mp)

	mp.On("ActionLogOutput", mg).Return("log output")

	result := si.ActionLog()
	assert.Equal(t, "log output", result)
}

func TestFollowTheQueenInteractor_Snapshot(t *testing.T) {
	cfg := domain.DefaultFollowTheQueenConfig()
	cfg.TableSize = 4
	players := domain.NewFollowTheQueenPlayersForTable(cfg.TableSize)
	tc := domain.NewTrumpCards(0)
	s := domain.NewFollowTheQueen(tc, players, cfg)
	mp := new(presenter.MockFollowTheQueenPresenter)
	si := NewFollowTheQueenInteractor(s, mp)

	data, err := si.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreFollowTheQueenInteractor(t *testing.T) {
	cfg := domain.DefaultFollowTheQueenConfig()
	cfg.TableSize = 4
	players := domain.NewFollowTheQueenPlayersForTable(cfg.TableSize)
	tc := domain.NewTrumpCards(0)
	s := domain.NewFollowTheQueen(tc, players, cfg)
	mp := new(presenter.MockFollowTheQueenPresenter)
	si := NewFollowTheQueenInteractor(s, mp)

	data, err := si.Snapshot()
	assert.NoError(t, err)

	restored, err := RestoreFollowTheQueenInteractor(data, mp)
	assert.NoError(t, err)
	assert.NotNil(t, restored)
}

func TestRestoreFollowTheQueenInteractor_InvalidJSON(t *testing.T) {
	mp := new(presenter.MockFollowTheQueenPresenter)
	_, err := RestoreFollowTheQueenInteractor([]byte("invalid"), mp)
	assert.Error(t, err)
}

func TestFollowTheQueenInteractor_ResetWithConfig_WithProfile(t *testing.T) {
	mg := new(interfaces.MockFollowTheQueenGame)
	mp := new(presenter.MockFollowTheQueenPresenter)
	si := NewFollowTheQueenInteractor(mg, mp)

	cfg := domain.DefaultFollowTheQueenConfig()
	mg.On("GetPlayerCnt").Return(cfg.TableSize)
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(nil)
	mg.On("ImportProfile", mock.Anything).Return(nil)
	mp.On("Output", mg, mock.Anything).Return("profile output")

	result := si.ResetWithConfig(cfg, []byte(`{"gp":1}`))
	assert.Equal(t, "profile output", result)
	mg.AssertCalled(t, "ImportProfile", mock.Anything)
}

func TestFollowTheQueenInteractor_Hint(t *testing.T) {
	mg := new(interfaces.MockFollowTheQueenGame)
	mp := new(presenter.MockFollowTheQueenPresenter)
	si := NewFollowTheQueenInteractor(mg, mp)

	mp.On("HintOutput", mg).Return("hint output")

	assert.Equal(t, "hint output", si.Hint())
}
