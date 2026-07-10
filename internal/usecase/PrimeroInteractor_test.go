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

const primeroMockOutput = `{"phase":0}`

func newPrimeroMocks() (*interfaces.MockPrimeroGame, *presenter.MockPrimeroPresenter) {
	return new(interfaces.MockPrimeroGame), new(presenter.MockPrimeroPresenter)
}

func TestNewPrimeroInteractor_NilGuards(t *testing.T) {
	spMock := new(presenter.MockPrimeroPresenter)
	assert.PanicsWithValue(t, "PrimeroInteractor: g must not be nil", func() {
		usecase.NewPrimeroInteractor(nil, spMock)
	})
	gameMock := new(interfaces.MockPrimeroGame)
	assert.PanicsWithValue(t, "PrimeroInteractor: sp must not be nil", func() {
		usecase.NewPrimeroInteractor(gameMock, nil)
	})
}

func TestPrimeroInteractor_Bet(t *testing.T) {
	for _, action := range []string{"call", "raise", "fold"} {
		t.Run(action, func(t *testing.T) {
			gm, sp := newPrimeroMocks()
			sp.On("Output", mock.Anything, mock.Anything).Return(primeroMockOutput)
			gm.On("GetGameEndFlag").Return(false)
			gm.On("PlayerCall").Return(nil)
			gm.On("PlayerRaise").Return(nil)
			gm.On("PlayerFold").Return(nil)

			ti := usecase.NewPrimeroInteractor(gm, sp)
			assert.Equal(t, primeroMockOutput, ti.Bet(action))
		})
	}
}

func TestPrimeroInteractor_Bet_UnknownAction(t *testing.T) {
	gm, sp := newPrimeroMocks()
	sp.On("Output", mock.Anything, mock.Anything).Return(primeroMockOutput)
	gm.On("GetGameEndFlag").Return(false)

	ti := usecase.NewPrimeroInteractor(gm, sp)
	assert.Equal(t, primeroMockOutput, ti.Bet("zzz"))
	gm.AssertNotCalled(t, "PlayerCall")
}

func TestPrimeroInteractor_Bet_GameEnded(t *testing.T) {
	gm, sp := newPrimeroMocks()
	sp.On("Output", mock.Anything, mock.Anything).Return(primeroMockOutput)
	gm.On("GetGameEndFlag").Return(true)

	ti := usecase.NewPrimeroInteractor(gm, sp)
	assert.Equal(t, primeroMockOutput, ti.Bet("call"))
	gm.AssertNotCalled(t, "PlayerCall")
}

func TestPrimeroInteractor_ResetAndNextRound(t *testing.T) {
	gm, sp := newPrimeroMocks()
	sp.On("Output", mock.Anything, mock.Anything).Return(primeroMockOutput)
	gm.On("Reset").Return()
	gm.On("NextRound").Return()

	ti := usecase.NewPrimeroInteractor(gm, sp)
	assert.Equal(t, primeroMockOutput, ti.Reset())
	assert.Equal(t, primeroMockOutput, ti.NextRound())
	gm.AssertCalled(t, "Reset")
	gm.AssertCalled(t, "NextRound")
}

func TestPrimeroInteractor_ResetWithConfigInvalid(t *testing.T) {
	gm, sp := newPrimeroMocks()
	sp.On("Output", mock.Anything, mock.Anything).Return(primeroMockOutput)
	ti := usecase.NewPrimeroInteractor(gm, sp)
	out := ti.ResetWithConfig(domain.PrimeroConfig{PlayerCount: 99})
	assert.Equal(t, primeroMockOutput, out)
	gm.AssertNotCalled(t, "Reset")
}

func TestPrimeroInteractor_ResetWithConfigValid(t *testing.T) {
	gm, sp := newPrimeroMocks()
	sp.On("Output", mock.Anything, mock.Anything).Return(primeroMockOutput)
	gm.On("SetConfig", mock.Anything).Return()
	gm.On("Reset").Return()
	ti := usecase.NewPrimeroInteractor(gm, sp)
	out := ti.ResetWithConfig(domain.DefaultPrimeroConfig())
	assert.Equal(t, primeroMockOutput, out)
	gm.AssertCalled(t, "Reset")
}

func TestPrimeroInteractor_HintAndLog(t *testing.T) {
	gm, sp := newPrimeroMocks()
	sp.On("HintOutput", mock.Anything).Return("hint")
	sp.On("ActionLogOutput", mock.Anything).Return("log")

	ti := usecase.NewPrimeroInteractor(gm, sp)
	assert.Equal(t, "hint", ti.Hint())
	assert.Equal(t, "log", ti.ActionLog())
}

func TestPrimeroInteractor_GetConfig(t *testing.T) {
	gm, sp := newPrimeroMocks()
	cfg := domain.PrimeroConfig{PlayerCount: 3, Ante: 5, StartingChips: 100, TargetRounds: 5}
	gm.On("GetConfig").Return(cfg)
	ti := usecase.NewPrimeroInteractor(gm, sp)
	assert.Equal(t, cfg, ti.GetConfig())
}

func TestPrimeroInteractor_RealFlow(t *testing.T) {
	sp := new(presenter.MockPrimeroPresenter)
	sp.On("Output", mock.Anything, mock.Anything).Return(primeroMockOutput)

	g := domain.NewDefaultPrimero()
	ti := usecase.NewPrimeroInteractor(g, sp)

	ti.Reset()
	for i := 0; i < 100 && g.GetPhase() == domain.PrimeroPhaseBetting && g.IsHumanTurn(); i++ {
		ti.Bet("call")
	}
	assert.Equal(t, domain.PrimeroPhaseResult, g.GetPhase())
}

func TestPrimeroInteractor_SnapshotAndRestore(t *testing.T) {
	sp := new(presenter.MockPrimeroPresenter)
	sp.On("Output", mock.Anything, mock.Anything).Return(primeroMockOutput)

	real := usecase.NewPrimeroInteractor(domain.NewDefaultPrimero(), sp)
	real.Reset()
	data, err := real.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	restored, err := usecase.RestorePrimeroInteractor(data, sp)
	require.NoError(t, err)
	assert.NotNil(t, restored)

	_, err = usecase.RestorePrimeroInteractor([]byte("not json"), sp)
	assert.Error(t, err)

	var g domain.Primero
	require.NoError(t, json.Unmarshal(data, &g))
}
