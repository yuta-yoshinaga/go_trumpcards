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

const michiganMockOutput = `{"phase":0}`

func newMichiganMocks() (*interfaces.MockMichiganGame, *presenter.MockMichiganPresenter) {
	return new(interfaces.MockMichiganGame), new(presenter.MockMichiganPresenter)
}

func TestNewMichiganInteractor_NilGuards(t *testing.T) {
	spMock := new(presenter.MockMichiganPresenter)
	assert.PanicsWithValue(t, "MichiganInteractor: g must not be nil", func() {
		usecase.NewMichiganInteractor(nil, spMock)
	})
	gameMock := new(interfaces.MockMichiganGame)
	assert.PanicsWithValue(t, "MichiganInteractor: sp must not be nil", func() {
		usecase.NewMichiganInteractor(gameMock, nil)
	})
}

func TestMichiganInteractor_Bet(t *testing.T) {
	gm, sp := newMichiganMocks()
	sp.On("Output", mock.Anything, mock.Anything).Return(michiganMockOutput)
	gm.On("GetGameEndFlag").Return(false)
	gm.On("PlaceHumanBet", mock.Anything).Return(nil)

	ti := usecase.NewMichiganInteractor(gm, sp)
	assert.Equal(t, michiganMockOutput, ti.Bet([]int{2, 2, 2, 2}))
	gm.AssertCalled(t, "PlaceHumanBet", []int{2, 2, 2, 2})
}

func TestMichiganInteractor_Bet_GameEnded(t *testing.T) {
	gm, sp := newMichiganMocks()
	sp.On("Output", mock.Anything, mock.Anything).Return(michiganMockOutput)
	gm.On("GetGameEndFlag").Return(true)

	ti := usecase.NewMichiganInteractor(gm, sp)
	assert.Equal(t, michiganMockOutput, ti.Bet([]int{2, 2, 2, 2}))
	gm.AssertNotCalled(t, "PlaceHumanBet")
}

func TestMichiganInteractor_Play(t *testing.T) {
	gm, sp := newMichiganMocks()
	sp.On("Output", mock.Anything, mock.Anything).Return(michiganMockOutput)
	gm.On("GetGameEndFlag").Return(false)
	gm.On("PlayCard", 3).Return(nil)

	ti := usecase.NewMichiganInteractor(gm, sp)
	assert.Equal(t, michiganMockOutput, ti.Play(3))
	gm.AssertCalled(t, "PlayCard", 3)
}

func TestMichiganInteractor_Play_GameEnded(t *testing.T) {
	gm, sp := newMichiganMocks()
	sp.On("Output", mock.Anything, mock.Anything).Return(michiganMockOutput)
	gm.On("GetGameEndFlag").Return(true)

	ti := usecase.NewMichiganInteractor(gm, sp)
	assert.Equal(t, michiganMockOutput, ti.Play(0))
	gm.AssertNotCalled(t, "PlayCard")
}

func TestMichiganInteractor_ResetAndNextRound(t *testing.T) {
	gm, sp := newMichiganMocks()
	sp.On("Output", mock.Anything, mock.Anything).Return(michiganMockOutput)
	gm.On("Reset").Return()
	gm.On("NextRound").Return()

	ti := usecase.NewMichiganInteractor(gm, sp)
	assert.Equal(t, michiganMockOutput, ti.Reset())
	assert.Equal(t, michiganMockOutput, ti.NextRound())
	gm.AssertCalled(t, "Reset")
	gm.AssertCalled(t, "NextRound")
}

func TestMichiganInteractor_ResetWithConfigInvalid(t *testing.T) {
	gm, sp := newMichiganMocks()
	sp.On("Output", mock.Anything, mock.Anything).Return(michiganMockOutput)
	ti := usecase.NewMichiganInteractor(gm, sp)
	out := ti.ResetWithConfig(domain.MichiganConfig{PlayerCount: 99})
	assert.Equal(t, michiganMockOutput, out)
	gm.AssertNotCalled(t, "Reset")
}

func TestMichiganInteractor_ResetWithConfigValid(t *testing.T) {
	gm, sp := newMichiganMocks()
	sp.On("Output", mock.Anything, mock.Anything).Return(michiganMockOutput)
	gm.On("SetConfig", mock.Anything).Return()
	gm.On("Reset").Return()
	ti := usecase.NewMichiganInteractor(gm, sp)
	out := ti.ResetWithConfig(domain.DefaultMichiganConfig())
	assert.Equal(t, michiganMockOutput, out)
	gm.AssertCalled(t, "Reset")
}

func TestMichiganInteractor_HintAndLog(t *testing.T) {
	gm, sp := newMichiganMocks()
	sp.On("HintOutput", mock.Anything).Return("hint")
	sp.On("ActionLogOutput", mock.Anything).Return("log")

	ti := usecase.NewMichiganInteractor(gm, sp)
	assert.Equal(t, "hint", ti.Hint())
	assert.Equal(t, "log", ti.ActionLog())
}

func TestMichiganInteractor_GetConfig(t *testing.T) {
	gm, sp := newMichiganMocks()
	cfg := domain.MichiganConfig{PlayerCount: 3, Ante: 8, StartingChips: 100, TargetRounds: 5}
	gm.On("GetConfig").Return(cfg)
	ti := usecase.NewMichiganInteractor(gm, sp)
	assert.Equal(t, cfg, ti.GetConfig())
}

func TestMichiganInteractor_RealFlow(t *testing.T) {
	sp := new(presenter.MockMichiganPresenter)
	sp.On("Output", mock.Anything, mock.Anything).Return(michiganMockOutput)

	g := domain.NewDefaultMichigan()
	ti := usecase.NewMichiganInteractor(g, sp)

	// Place the human bet, then play through until the round resolves.
	budget := g.GetBetBudget()
	dist := make([]int, domain.MichiganBoodleCount)
	dist[0] = budget
	ti.Bet(dist)
	for i := 0; i < 300 && g.GetPhase() == domain.MichiganPhasePlay && g.IsHumanTurn(); i++ {
		pi := g.GetPlayableIndices()
		require.NotEmpty(t, pi)
		ti.Play(pi[0])
	}
	assert.Equal(t, domain.MichiganPhaseResult, g.GetPhase())
}

func TestMichiganInteractor_SnapshotAndRestore(t *testing.T) {
	sp := new(presenter.MockMichiganPresenter)
	sp.On("Output", mock.Anything, mock.Anything).Return(michiganMockOutput)

	real := usecase.NewMichiganInteractor(domain.NewDefaultMichigan(), sp)
	real.Reset()
	data, err := real.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	restored, err := usecase.RestoreMichiganInteractor(data, sp)
	require.NoError(t, err)
	assert.NotNil(t, restored)

	_, err = usecase.RestoreMichiganInteractor([]byte("not json"), sp)
	assert.Error(t, err)

	var g domain.Michigan
	require.NoError(t, json.Unmarshal(data, &g))
}
