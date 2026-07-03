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

const bouillotteMockOutput = `{"phase":0}`

func newBouillotteMocks() (*interfaces.MockBouillotteGame, *presenter.MockBouillottePresenter) {
	return new(interfaces.MockBouillotteGame), new(presenter.MockBouillottePresenter)
}

func TestNewBouillotteInteractor_NilGuards(t *testing.T) {
	spMock := new(presenter.MockBouillottePresenter)
	assert.PanicsWithValue(t, "BouillotteInteractor: g must not be nil", func() {
		usecase.NewBouillotteInteractor(nil, spMock)
	})
	gameMock := new(interfaces.MockBouillotteGame)
	assert.PanicsWithValue(t, "BouillotteInteractor: sp must not be nil", func() {
		usecase.NewBouillotteInteractor(gameMock, nil)
	})
}

func TestBouillotteInteractor_Bet(t *testing.T) {
	for _, action := range []string{"call", "raise", "fold"} {
		t.Run(action, func(t *testing.T) {
			gm, sp := newBouillotteMocks()
			sp.On("Output", mock.Anything, mock.Anything).Return(bouillotteMockOutput)
			gm.On("GetGameEndFlag").Return(false)
			gm.On("PlayerCall").Return(nil)
			gm.On("PlayerRaise").Return(nil)
			gm.On("PlayerFold").Return(nil)

			ti := usecase.NewBouillotteInteractor(gm, sp)
			assert.Equal(t, bouillotteMockOutput, ti.Bet(action))
		})
	}
}

func TestBouillotteInteractor_Bet_UnknownAction(t *testing.T) {
	gm, sp := newBouillotteMocks()
	sp.On("Output", mock.Anything, mock.Anything).Return(bouillotteMockOutput)
	gm.On("GetGameEndFlag").Return(false)

	ti := usecase.NewBouillotteInteractor(gm, sp)
	assert.Equal(t, bouillotteMockOutput, ti.Bet("zzz"))
	gm.AssertNotCalled(t, "PlayerCall")
}

func TestBouillotteInteractor_Bet_GameEnded(t *testing.T) {
	gm, sp := newBouillotteMocks()
	sp.On("Output", mock.Anything, mock.Anything).Return(bouillotteMockOutput)
	gm.On("GetGameEndFlag").Return(true)

	ti := usecase.NewBouillotteInteractor(gm, sp)
	assert.Equal(t, bouillotteMockOutput, ti.Bet("call"))
	gm.AssertNotCalled(t, "PlayerCall")
}

func TestBouillotteInteractor_ResetAndNextRound(t *testing.T) {
	gm, sp := newBouillotteMocks()
	sp.On("Output", mock.Anything, mock.Anything).Return(bouillotteMockOutput)
	gm.On("Reset").Return()
	gm.On("NextRound").Return()

	ti := usecase.NewBouillotteInteractor(gm, sp)
	assert.Equal(t, bouillotteMockOutput, ti.Reset())
	assert.Equal(t, bouillotteMockOutput, ti.NextRound())
	gm.AssertCalled(t, "Reset")
	gm.AssertCalled(t, "NextRound")
}

func TestBouillotteInteractor_ResetWithConfigInvalid(t *testing.T) {
	gm, sp := newBouillotteMocks()
	sp.On("Output", mock.Anything, mock.Anything).Return(bouillotteMockOutput)
	ti := usecase.NewBouillotteInteractor(gm, sp)
	out := ti.ResetWithConfig(domain.BouillotteConfig{PlayerCount: 99})
	assert.Equal(t, bouillotteMockOutput, out)
	gm.AssertNotCalled(t, "Reset")
}

func TestBouillotteInteractor_ResetWithConfigValid(t *testing.T) {
	gm, sp := newBouillotteMocks()
	sp.On("Output", mock.Anything, mock.Anything).Return(bouillotteMockOutput)
	gm.On("SetConfig", mock.Anything).Return()
	gm.On("Reset").Return()
	ti := usecase.NewBouillotteInteractor(gm, sp)
	out := ti.ResetWithConfig(domain.DefaultBouillotteConfig())
	assert.Equal(t, bouillotteMockOutput, out)
	gm.AssertCalled(t, "Reset")
}

func TestBouillotteInteractor_HintAndLog(t *testing.T) {
	gm, sp := newBouillotteMocks()
	sp.On("HintOutput", mock.Anything).Return("hint")
	sp.On("ActionLogOutput", mock.Anything).Return("log")

	ti := usecase.NewBouillotteInteractor(gm, sp)
	assert.Equal(t, "hint", ti.Hint())
	assert.Equal(t, "log", ti.ActionLog())
}

func TestBouillotteInteractor_GetConfig(t *testing.T) {
	gm, sp := newBouillotteMocks()
	cfg := domain.BouillotteConfig{PlayerCount: 3, Ante: 5, StartingChips: 100, TargetRounds: 5}
	gm.On("GetConfig").Return(cfg)
	ti := usecase.NewBouillotteInteractor(gm, sp)
	assert.Equal(t, cfg, ti.GetConfig())
}

func TestBouillotteInteractor_RealFlow(t *testing.T) {
	sp := new(presenter.MockBouillottePresenter)
	sp.On("Output", mock.Anything, mock.Anything).Return(bouillotteMockOutput)

	g := domain.NewDefaultBouillotte()
	ti := usecase.NewBouillotteInteractor(g, sp)

	ti.Reset()
	for i := 0; i < 100 && g.GetPhase() == domain.BouillottePhaseBetting && g.IsHumanTurn(); i++ {
		ti.Bet("call")
	}
	assert.Equal(t, domain.BouillottePhaseResult, g.GetPhase())
}

func TestBouillotteInteractor_SnapshotAndRestore(t *testing.T) {
	sp := new(presenter.MockBouillottePresenter)
	sp.On("Output", mock.Anything, mock.Anything).Return(bouillotteMockOutput)

	real := usecase.NewBouillotteInteractor(domain.NewDefaultBouillotte(), sp)
	real.Reset()
	data, err := real.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	restored, err := usecase.RestoreBouillotteInteractor(data, sp)
	require.NoError(t, err)
	assert.NotNil(t, restored)

	_, err = usecase.RestoreBouillotteInteractor([]byte("not json"), sp)
	assert.Error(t, err)

	var g domain.Bouillotte
	require.NoError(t, json.Unmarshal(data, &g))
}
