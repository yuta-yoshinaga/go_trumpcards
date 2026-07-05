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

const anacondaMockOutput = `{"phase":0}`

func newAnacondaMocks() (*interfaces.MockAnacondaGame, *presenter.MockAnacondaPresenter) {
	return new(interfaces.MockAnacondaGame), new(presenter.MockAnacondaPresenter)
}

func TestNewAnacondaInteractor_NilGuards(t *testing.T) {
	spMock := new(presenter.MockAnacondaPresenter)
	assert.PanicsWithValue(t, "AnacondaInteractor: g must not be nil", func() {
		usecase.NewAnacondaInteractor(nil, spMock)
	})
	gameMock := new(interfaces.MockAnacondaGame)
	assert.PanicsWithValue(t, "AnacondaInteractor: sp must not be nil", func() {
		usecase.NewAnacondaInteractor(gameMock, nil)
	})
}

func TestAnacondaInteractor_Pass(t *testing.T) {
	gm, sp := newAnacondaMocks()
	sp.On("Output", mock.Anything, mock.Anything).Return(anacondaMockOutput)
	gm.On("GetGameEndFlag").Return(false)
	gm.On("Pass", []int{0, 1, 2}).Return(nil)

	ti := usecase.NewAnacondaInteractor(gm, sp)
	assert.Equal(t, anacondaMockOutput, ti.Pass([]int{0, 1, 2}))
	gm.AssertCalled(t, "Pass", []int{0, 1, 2})
}

func TestAnacondaInteractor_Pass_GameEnded(t *testing.T) {
	gm, sp := newAnacondaMocks()
	sp.On("Output", mock.Anything, mock.Anything).Return(anacondaMockOutput)
	gm.On("GetGameEndFlag").Return(true)

	ti := usecase.NewAnacondaInteractor(gm, sp)
	assert.Equal(t, anacondaMockOutput, ti.Pass([]int{0, 1, 2}))
	gm.AssertNotCalled(t, "Pass", mock.Anything)
}

func TestAnacondaInteractor_KeepAndBets(t *testing.T) {
	gm, sp := newAnacondaMocks()
	sp.On("Output", mock.Anything, mock.Anything).Return(anacondaMockOutput)
	gm.On("GetGameEndFlag").Return(false)
	gm.On("Keep", []int{0, 1, 2, 3, 4}).Return(nil)
	gm.On("PlayerCall").Return(nil)
	gm.On("PlayerRaise").Return(assert.AnError)
	gm.On("PlayerFold").Return(nil)

	ti := usecase.NewAnacondaInteractor(gm, sp)
	assert.Equal(t, anacondaMockOutput, ti.Keep([]int{0, 1, 2, 3, 4}))
	assert.Equal(t, anacondaMockOutput, ti.Call())
	assert.Equal(t, anacondaMockOutput, ti.Raise())
	assert.Equal(t, anacondaMockOutput, ti.Fold())
	gm.AssertCalled(t, "PlayerCall")
	gm.AssertCalled(t, "PlayerRaise")
	gm.AssertCalled(t, "PlayerFold")
}

func TestAnacondaInteractor_Keep_GameEnded(t *testing.T) {
	gm, sp := newAnacondaMocks()
	sp.On("Output", mock.Anything, mock.Anything).Return(anacondaMockOutput)
	gm.On("GetGameEndFlag").Return(true)

	ti := usecase.NewAnacondaInteractor(gm, sp)
	assert.Equal(t, anacondaMockOutput, ti.Keep([]int{0, 1, 2, 3, 4}))
	assert.Equal(t, anacondaMockOutput, ti.Call())
	gm.AssertNotCalled(t, "Keep", mock.Anything)
	gm.AssertNotCalled(t, "PlayerCall")
}

func TestAnacondaInteractor_ResetAndNextRound(t *testing.T) {
	gm, sp := newAnacondaMocks()
	sp.On("Output", mock.Anything, mock.Anything).Return(anacondaMockOutput)
	gm.On("Reset").Return()
	gm.On("NextRound").Return()

	ti := usecase.NewAnacondaInteractor(gm, sp)
	assert.Equal(t, anacondaMockOutput, ti.Reset())
	assert.Equal(t, anacondaMockOutput, ti.NextRound())
	gm.AssertCalled(t, "Reset")
	gm.AssertCalled(t, "NextRound")
}

func TestAnacondaInteractor_ResetWithConfigInvalid(t *testing.T) {
	gm, sp := newAnacondaMocks()
	sp.On("Output", mock.Anything, mock.Anything).Return(anacondaMockOutput)
	ti := usecase.NewAnacondaInteractor(gm, sp)
	out := ti.ResetWithConfig(domain.AnacondaConfig{PlayerCount: 99})
	assert.Equal(t, anacondaMockOutput, out)
	gm.AssertNotCalled(t, "Reset")
}

func TestAnacondaInteractor_ResetWithConfigValid(t *testing.T) {
	gm, sp := newAnacondaMocks()
	sp.On("Output", mock.Anything, mock.Anything).Return(anacondaMockOutput)
	gm.On("SetConfig", mock.Anything).Return()
	gm.On("Reset").Return()
	ti := usecase.NewAnacondaInteractor(gm, sp)
	out := ti.ResetWithConfig(domain.DefaultAnacondaConfig())
	assert.Equal(t, anacondaMockOutput, out)
	gm.AssertCalled(t, "Reset")
}

func TestAnacondaInteractor_HintAndLog(t *testing.T) {
	gm, sp := newAnacondaMocks()
	sp.On("HintOutput", mock.Anything).Return("hint")
	sp.On("ActionLogOutput", mock.Anything).Return("log")

	ti := usecase.NewAnacondaInteractor(gm, sp)
	assert.Equal(t, "hint", ti.Hint())
	assert.Equal(t, "log", ti.ActionLog())
}

func TestAnacondaInteractor_GetConfig(t *testing.T) {
	gm, sp := newAnacondaMocks()
	cfg := domain.AnacondaConfig{PlayerCount: 3, Ante: 5, StartingChips: 100, TargetRounds: 5}
	gm.On("GetConfig").Return(cfg)
	ti := usecase.NewAnacondaInteractor(gm, sp)
	assert.Equal(t, cfg, ti.GetConfig())
}

func TestAnacondaInteractor_RealFlow(t *testing.T) {
	sp := new(presenter.MockAnacondaPresenter)
	sp.On("Output", mock.Anything, mock.Anything).Return(anacondaMockOutput)

	g := domain.NewDefaultAnaconda()
	ti := usecase.NewAnacondaInteractor(g, sp)

	ti.Reset()
	assert.Equal(t, domain.AnacondaPhasePass, g.GetPhase())
	ti.Pass([]int{0, 1, 2})
	assert.Equal(t, 2, g.GetPassCount())
}

func TestAnacondaInteractor_SnapshotAndRestore(t *testing.T) {
	sp := new(presenter.MockAnacondaPresenter)
	sp.On("Output", mock.Anything, mock.Anything).Return(anacondaMockOutput)

	real := usecase.NewAnacondaInteractor(domain.NewDefaultAnaconda(), sp)
	real.Reset()
	data, err := real.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	restored, err := usecase.RestoreAnacondaInteractor(data, sp)
	require.NoError(t, err)
	assert.NotNil(t, restored)

	_, err = usecase.RestoreAnacondaInteractor([]byte("not json"), sp)
	assert.Error(t, err)

	var g domain.Anaconda
	require.NoError(t, json.Unmarshal(data, &g))
}
