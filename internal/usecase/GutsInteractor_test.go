//go:build test

package usecase_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

const gutsMockOutput = `{"phase":0}`

func newGutsMocks() (*interfaces.MockGutsGame, *presenter.MockGutsPresenter) {
	return new(interfaces.MockGutsGame), new(presenter.MockGutsPresenter)
}

func TestNewGutsInteractor_NilGuards(t *testing.T) {
	spMock := new(presenter.MockGutsPresenter)
	assert.PanicsWithValue(t, "GutsInteractor: g must not be nil", func() {
		usecase.NewGutsInteractor(nil, spMock)
	})
	gameMock := new(interfaces.MockGutsGame)
	assert.PanicsWithValue(t, "GutsInteractor: sp must not be nil", func() {
		usecase.NewGutsInteractor(gameMock, nil)
	})
}

func TestGutsInteractor_Declare(t *testing.T) {
	gm, sp := newGutsMocks()
	sp.On("Output", mock.Anything, mock.Anything).Return(gutsMockOutput)
	gm.On("GetGameEndFlag").Return(false)
	gm.On("Declare", true).Return(nil)

	ti := usecase.NewGutsInteractor(gm, sp)
	assert.Equal(t, gutsMockOutput, ti.Declare(true))
	gm.AssertCalled(t, "Declare", true)
}

func TestGutsInteractor_Declare_Error(t *testing.T) {
	gm, sp := newGutsMocks()
	sp.On("Output", mock.Anything, mock.Anything).Return(gutsMockOutput)
	gm.On("GetGameEndFlag").Return(false)
	gm.On("Declare", false).Return(assert.AnError)

	ti := usecase.NewGutsInteractor(gm, sp)
	assert.Equal(t, gutsMockOutput, ti.Declare(false))
}

func TestGutsInteractor_Declare_GameEnded(t *testing.T) {
	gm, sp := newGutsMocks()
	sp.On("Output", mock.Anything, mock.Anything).Return(gutsMockOutput)
	gm.On("GetGameEndFlag").Return(true)

	ti := usecase.NewGutsInteractor(gm, sp)
	assert.Equal(t, gutsMockOutput, ti.Declare(true))
	gm.AssertNotCalled(t, "Declare", mock.Anything)
}

func TestGutsInteractor_ResetAndNextRound(t *testing.T) {
	gm, sp := newGutsMocks()
	sp.On("Output", mock.Anything, mock.Anything).Return(gutsMockOutput)
	gm.On("Reset").Return()
	gm.On("NextRound").Return()

	ti := usecase.NewGutsInteractor(gm, sp)
	assert.Equal(t, gutsMockOutput, ti.Reset())
	assert.Equal(t, gutsMockOutput, ti.NextRound())
	gm.AssertCalled(t, "Reset")
	gm.AssertCalled(t, "NextRound")
}

func TestGutsInteractor_ResetWithConfigInvalid(t *testing.T) {
	gm, sp := newGutsMocks()
	sp.On("Output", mock.Anything, mock.Anything).Return(gutsMockOutput)
	ti := usecase.NewGutsInteractor(gm, sp)
	out := ti.ResetWithConfig(domain.GutsConfig{PlayerCount: 99})
	assert.Equal(t, gutsMockOutput, out)
	gm.AssertNotCalled(t, "Reset")
}

func TestGutsInteractor_ResetWithConfigValid(t *testing.T) {
	gm, sp := newGutsMocks()
	sp.On("Output", mock.Anything, mock.Anything).Return(gutsMockOutput)
	gm.On("SetConfig", mock.Anything).Return()
	gm.On("Reset").Return()
	ti := usecase.NewGutsInteractor(gm, sp)
	out := ti.ResetWithConfig(domain.DefaultGutsConfig())
	assert.Equal(t, gutsMockOutput, out)
	gm.AssertCalled(t, "Reset")
}

func TestGutsInteractor_HintAndLog(t *testing.T) {
	gm, sp := newGutsMocks()
	sp.On("HintOutput", mock.Anything).Return("hint")
	sp.On("ActionLogOutput", mock.Anything).Return("log")

	ti := usecase.NewGutsInteractor(gm, sp)
	assert.Equal(t, "hint", ti.Hint())
	assert.Equal(t, "log", ti.ActionLog())
}

func TestGutsInteractor_GetConfig(t *testing.T) {
	gm, sp := newGutsMocks()
	cfg := domain.GutsConfig{PlayerCount: 3, Ante: 5, StartingChips: 100, TargetRounds: 5}
	gm.On("GetConfig").Return(cfg)
	ti := usecase.NewGutsInteractor(gm, sp)
	assert.Equal(t, cfg, ti.GetConfig())
}

func TestGutsInteractor_RealFlow(t *testing.T) {
	sp := new(presenter.MockGutsPresenter)
	sp.On("Output", mock.Anything, mock.Anything).Return(gutsMockOutput)

	g := domain.NewDefaultGuts()
	ti := usecase.NewGutsInteractor(g, sp)

	ti.Reset()
	assert.Equal(t, domain.GutsPhaseDeclare, g.GetPhase())
	ti.Declare(true)
	assert.Equal(t, domain.GutsPhaseResult, g.GetPhase())
}

func TestGutsInteractor_SnapshotAndRestore(t *testing.T) {
	sp := new(presenter.MockGutsPresenter)
	sp.On("Output", mock.Anything, mock.Anything).Return(gutsMockOutput)

	real := usecase.NewGutsInteractor(domain.NewDefaultGuts(), sp)
	real.Reset()
	data, err := real.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	restored, err := usecase.RestoreGutsInteractor(data, sp)
	require.NoError(t, err)
	assert.NotNil(t, restored)

	_, err = usecase.RestoreGutsInteractor([]byte("not json"), sp)
	assert.Error(t, err)

	var g domain.Guts
	require.NoError(t, json.Unmarshal(data, &g))
}
