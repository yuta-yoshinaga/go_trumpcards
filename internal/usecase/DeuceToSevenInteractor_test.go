package usecase_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newDeuceToSevenMocks() (*interfaces.MockDeuceToSevenGame, *presenter.MockDeuceToSevenPresenter) {
	return new(interfaces.MockDeuceToSevenGame), new(presenter.MockDeuceToSevenPresenter)
}

func TestNewDeuceToSevenInteractor(t *testing.T) {
	mg, mp := newDeuceToSevenMocks()
	di := usecase.NewDeuceToSevenInteractor(mg, mp)
	assert.NotNil(t, di)
}

func TestNewDeuceToSevenInteractor_NilGame(t *testing.T) {
	_, mp := newDeuceToSevenMocks()
	assert.PanicsWithValue(t, "DeuceToSevenInteractor: g must not be nil", func() {
		usecase.NewDeuceToSevenInteractor(nil, mp)
	})
}

func TestNewDeuceToSevenInteractor_NilPresenter(t *testing.T) {
	mg, _ := newDeuceToSevenMocks()
	assert.PanicsWithValue(t, "DeuceToSevenInteractor: pp must not be nil", func() {
		usecase.NewDeuceToSevenInteractor(mg, nil)
	})
}

func TestDeuceToSevenInteractor_Reset(t *testing.T) {
	mg, mp := newDeuceToSevenMocks()
	di := usecase.NewDeuceToSevenInteractor(mg, mp)

	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("reset output")

	assert.Equal(t, "reset output", di.Reset())
	mg.AssertCalled(t, "Reset")
}

func TestDeuceToSevenInteractor_Hint(t *testing.T) {
	mg, mp := newDeuceToSevenMocks()
	di := usecase.NewDeuceToSevenInteractor(mg, mp)

	mp.On("HintOutput", mg).Return("hint output")

	assert.Equal(t, "hint output", di.Hint())
	mp.AssertCalled(t, "HintOutput", mg)
}

func TestDeuceToSevenInteractor_Reset_Error(t *testing.T) {
	mg, mp := newDeuceToSevenMocks()
	di := usecase.NewDeuceToSevenInteractor(mg, mp)

	err := errors.New("reset failed")
	mg.On("Reset").Return(err)
	mp.On("Output", mg, err).Return("error output")

	assert.Equal(t, "error output", di.Reset())
}

func TestDeuceToSevenInteractor_GetConfig(t *testing.T) {
	mg, mp := newDeuceToSevenMocks()
	di := usecase.NewDeuceToSevenInteractor(mg, mp)

	want := domain.DeuceToSevenConfig{InitChips: 1000, Ante: 10, MinBet: 10, CpuCount: 2}
	mg.On("GetConfig").Return(want)

	assert.Equal(t, want, di.GetConfig())
}

func TestDeuceToSevenInteractor_ResetWithConfig(t *testing.T) {
	mg, mp := newDeuceToSevenMocks()
	di := usecase.NewDeuceToSevenInteractor(mg, mp)

	cfg := domain.DeuceToSevenConfig{InitChips: 2000, Ante: 20, MinBet: 20, CpuCount: 2}
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("reset with config output")

	assert.Equal(t, "reset with config output", di.ResetWithConfig(cfg, nil))
	mg.AssertCalled(t, "SetConfig", cfg)
	mg.AssertCalled(t, "Reset")
}

func TestDeuceToSevenInteractor_ResetWithConfig_InvalidConfig(t *testing.T) {
	mg, mp := newDeuceToSevenMocks()
	di := usecase.NewDeuceToSevenInteractor(mg, mp)

	cfg := domain.DeuceToSevenConfig{InitChips: 1000, Ante: 10, MinBet: 10, CpuCount: 0}
	mp.On("Output", mg, mock.Anything).Return("invalid config output")

	assert.Equal(t, "invalid config output", di.ResetWithConfig(cfg, nil))
	mg.AssertNotCalled(t, "SetConfig")
	mg.AssertNotCalled(t, "Reset")
}

func TestDeuceToSevenInteractor_ResetWithConfig_ResetError(t *testing.T) {
	mg, mp := newDeuceToSevenMocks()
	di := usecase.NewDeuceToSevenInteractor(mg, mp)

	cfg := domain.DeuceToSevenConfig{InitChips: 1000, Ante: 10, MinBet: 10, CpuCount: 3}
	err := errors.New("reset failed")
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(err)
	mp.On("Output", mg, err).Return("error output")

	assert.Equal(t, "error output", di.ResetWithConfig(cfg, nil))
}

func TestDeuceToSevenInteractor_ResetWithConfig_ImportsProfile(t *testing.T) {
	mg, mp := newDeuceToSevenMocks()
	di := usecase.NewDeuceToSevenInteractor(mg, mp)

	cfg := domain.DeuceToSevenConfig{InitChips: 2000, Ante: 20, MinBet: 20, CpuCount: 2}
	profileData := []byte(`{"gamesPlayed":3}`)
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(nil)
	mg.On("ImportProfile", profileData).Return(nil)
	mp.On("Output", mg, mock.Anything).Return("with profile output")

	assert.Equal(t, "with profile output", di.ResetWithConfig(cfg, profileData))
	mg.AssertCalled(t, "ImportProfile", profileData)
}

func TestDeuceToSevenInteractor_Action(t *testing.T) {
	mg, mp := newDeuceToSevenMocks()
	di := usecase.NewDeuceToSevenInteractor(mg, mp)

	mg.On("PlayerAction", domain.DeuceToSevenActionCheck, 0, 0).Return(nil)
	mp.On("Output", mg, mock.Anything).Return("action output")

	assert.Equal(t, "action output", di.Action(domain.DeuceToSevenActionCheck, 0, 0))
	mg.AssertCalled(t, "PlayerAction", domain.DeuceToSevenActionCheck, 0, 0)
}

func TestDeuceToSevenInteractor_Action_Error(t *testing.T) {
	mg, mp := newDeuceToSevenMocks()
	di := usecase.NewDeuceToSevenInteractor(mg, mp)

	err := errors.New("action failed")
	mg.On("PlayerAction", domain.DeuceToSevenActionBet, 50, 0).Return(err)
	mp.On("Output", mg, err).Return("error output")

	assert.Equal(t, "error output", di.Action(domain.DeuceToSevenActionBet, 50, 0))
}

func TestDeuceToSevenInteractor_Exchange(t *testing.T) {
	mg, mp := newDeuceToSevenMocks()
	di := usecase.NewDeuceToSevenInteractor(mg, mp)

	indices := []int{0, 2}
	mg.On("PlayerExchange", indices).Return(nil)
	mp.On("Output", mg, mock.Anything).Return("exchange output")

	assert.Equal(t, "exchange output", di.Exchange(indices))
	mg.AssertCalled(t, "PlayerExchange", indices)
}

func TestDeuceToSevenInteractor_Stand(t *testing.T) {
	mg, mp := newDeuceToSevenMocks()
	di := usecase.NewDeuceToSevenInteractor(mg, mp)

	mg.On("PlayerStand").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("stand output")

	assert.Equal(t, "stand output", di.Stand())
	mg.AssertCalled(t, "PlayerStand")
}

func TestDeuceToSevenInteractor_ActionLog(t *testing.T) {
	mg, mp := newDeuceToSevenMocks()
	mp.On("ActionLogOutput", mg).Return(`{"entries":[]}`)
	di := usecase.NewDeuceToSevenInteractor(mg, mp)

	assert.Equal(t, `{"entries":[]}`, di.ActionLog())
	mp.AssertExpectations(t)
}

func TestDeuceToSevenInteractor_Snapshot(t *testing.T) {
	g := domain.NewDefaultDeuceToSeven()
	di := usecase.NewDeuceToSevenInteractor(g, new(presenter.MockDeuceToSevenPresenter))
	data, err := di.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreDeuceToSevenInteractor(t *testing.T) {
	g := domain.NewDefaultDeuceToSeven()
	assert.NoError(t, g.Reset())
	di := usecase.NewDeuceToSevenInteractor(g, new(presenter.MockDeuceToSevenPresenter))
	data, err := di.Snapshot()
	assert.NoError(t, err)

	restored, err := usecase.RestoreDeuceToSevenInteractor(data, new(presenter.MockDeuceToSevenPresenter))
	assert.NoError(t, err)
	assert.NotNil(t, restored)
	assert.Equal(t, g.GetConfig(), restored.GetConfig())
}
