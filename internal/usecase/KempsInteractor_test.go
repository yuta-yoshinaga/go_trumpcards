//go:build test

package usecase_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// newKempsGameMock returns a mock whose advanceCpu loop short-circuits
// immediately: gameEnd=false so action guards pass, but phase=GameEnd so the
// advanceCpu switch returns on the default branch without further calls.
func newKempsGameMock() *interfaces.MockKempsGame {
	g := new(interfaces.MockKempsGame)
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.KempsPhaseGameEnd)
	return g
}

func TestNewKempsInteractor_NilGuards(t *testing.T) {
	spMock := new(presenter.MockKempsPresenter)
	assert.PanicsWithValue(t, "KempsInteractor: g must not be nil", func() {
		usecase.NewKempsInteractor(nil, spMock)
	})
	assert.PanicsWithValue(t, "KempsInteractor: sp must not be nil", func() {
		usecase.NewKempsInteractor(domain.NewDefaultKemps(), nil)
	})
}

func TestKempsInteractor_Reset(t *testing.T) {
	out := `{"phase":0}`
	spMock := new(presenter.MockKempsPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(out)
	g := newKempsGameMock()
	g.On("Reset").Return()

	ki := usecase.NewKempsInteractor(g, spMock)
	assert.Equal(t, out, ki.Reset())
	g.AssertCalled(t, "Reset")
}

func TestKempsInteractor_ResetWithConfig(t *testing.T) {
	out := `{"phase":0}`

	t.Run("valid", func(t *testing.T) {
		spMock := new(presenter.MockKempsPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(out)
		g := newKempsGameMock()
		cfg := domain.KempsConfig{CpuDifficulty: domain.KempsCpuHard, TargetScore: 5}
		g.On("SetConfig", cfg).Return()
		g.On("Reset").Return()

		ki := usecase.NewKempsInteractor(g, spMock)
		assert.Equal(t, out, ki.ResetWithConfig(cfg))
		g.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config returns error", func(t *testing.T) {
		errOut := `{"error":"invalid"}`
		spMock := new(presenter.MockKempsPresenter)
		spMock.On("Output", mock.Anything, mock.MatchedBy(func(e error) bool { return e != nil })).Return(errOut)
		g := newKempsGameMock()

		ki := usecase.NewKempsInteractor(g, spMock)
		assert.Equal(t, errOut, ki.ResetWithConfig(domain.KempsConfig{CpuDifficulty: 99, TargetScore: 5}))
	})
}

func TestKempsInteractor_Swap(t *testing.T) {
	out := `{"phase":0}`

	t.Run("success", func(t *testing.T) {
		spMock := new(presenter.MockKempsPresenter)
		spMock.On("Output", mock.Anything, nil).Return(out)
		g := newKempsGameMock()
		g.On("IsHumanTurn").Return(true)
		g.On("PlayerSwap", 1, 2).Return(nil)

		ki := usecase.NewKempsInteractor(g, spMock)
		assert.Equal(t, out, ki.Swap(1, 2))
		g.AssertCalled(t, "PlayerSwap", 1, 2)
	})

	t.Run("error from domain", func(t *testing.T) {
		errOut := `{"error":"x"}`
		spMock := new(presenter.MockKempsPresenter)
		spMock.On("Output", mock.Anything, mock.MatchedBy(func(e error) bool { return e != nil })).Return(errOut)
		g := newKempsGameMock()
		g.On("IsHumanTurn").Return(true)
		g.On("PlayerSwap", 0, 0).Return(domain.ErrInvalidCard)

		ki := usecase.NewKempsInteractor(g, spMock)
		assert.Equal(t, errOut, ki.Swap(0, 0))
	})

	t.Run("blocked when not playable", func(t *testing.T) {
		blockedOut := `{"blocked":true}`
		spMock := new(presenter.MockKempsPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(blockedOut)
		g := new(interfaces.MockKempsGame)
		g.On("GetGameEndFlag").Return(false)
		g.On("IsHumanTurn").Return(false)

		ki := usecase.NewKempsInteractor(g, spMock)
		assert.Equal(t, blockedOut, ki.Swap(0, 0))
		g.AssertNotCalled(t, "PlayerSwap", mock.Anything, mock.Anything)
	})
}

func TestKempsInteractor_Pass(t *testing.T) {
	out := `{"phase":0}`
	spMock := new(presenter.MockKempsPresenter)
	spMock.On("Output", mock.Anything, nil).Return(out)
	g := newKempsGameMock()
	g.On("IsHumanTurn").Return(true)
	g.On("PlayerPass").Return(nil)

	ki := usecase.NewKempsInteractor(g, spMock)
	assert.Equal(t, out, ki.Pass())
	g.AssertCalled(t, "PlayerPass")
}

func TestKempsInteractor_SetSignal(t *testing.T) {
	out := `{"phase":0}`
	spMock := new(presenter.MockKempsPresenter)
	spMock.On("Output", mock.Anything, nil).Return(out)
	g := newKempsGameMock()
	g.On("PlayerSetSignal", 1).Return()

	ki := usecase.NewKempsInteractor(g, spMock)
	assert.Equal(t, out, ki.SetSignal(1))
	g.AssertCalled(t, "PlayerSetSignal", 1)
}

func TestKempsInteractor_DeclareKemps(t *testing.T) {
	out := `{"phase":2}`
	spMock := new(presenter.MockKempsPresenter)
	spMock.On("Output", mock.Anything, nil).Return(out)
	g := newKempsGameMock()
	g.On("PlayerDeclareKemps").Return(nil)

	ki := usecase.NewKempsInteractor(g, spMock)
	assert.Equal(t, out, ki.DeclareKemps())
	g.AssertCalled(t, "PlayerDeclareKemps")
}

func TestKempsInteractor_DeclareCounterKemps(t *testing.T) {
	out := `{"phase":2}`
	spMock := new(presenter.MockKempsPresenter)
	spMock.On("Output", mock.Anything, nil).Return(out)
	g := newKempsGameMock()
	g.On("PlayerDeclareCounterKemps", 1).Return(nil)

	ki := usecase.NewKempsInteractor(g, spMock)
	assert.Equal(t, out, ki.DeclareCounterKemps(1))
	g.AssertCalled(t, "PlayerDeclareCounterKemps", 1)
}

func TestKempsInteractor_DeclareErrorAndGameEnd(t *testing.T) {
	errOut := `{"error":"x"}`
	spMock := new(presenter.MockKempsPresenter)
	spMock.On("Output", mock.Anything, mock.MatchedBy(func(e error) bool { return e != nil })).Return(errOut)
	g := newKempsGameMock()
	g.On("PlayerDeclareKemps").Return(domain.ErrWrongPhase)
	ki := usecase.NewKempsInteractor(g, spMock)
	assert.Equal(t, errOut, ki.DeclareKemps())

	// game ended blocks declare.
	endOut := `{"gameEnd":true}`
	spMock2 := new(presenter.MockKempsPresenter)
	spMock2.On("Output", mock.Anything, mock.Anything).Return(endOut)
	g2 := new(interfaces.MockKempsGame)
	g2.On("GetGameEndFlag").Return(true)
	ki2 := usecase.NewKempsInteractor(g2, spMock2)
	assert.Equal(t, endOut, ki2.DeclareCounterKemps(1))
	g2.AssertNotCalled(t, "PlayerDeclareCounterKemps", mock.Anything)
}

func TestKempsInteractor_NextRound(t *testing.T) {
	out := `{"phase":0}`

	t.Run("success", func(t *testing.T) {
		spMock := new(presenter.MockKempsPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(out)
		g := newKempsGameMock()
		g.On("NextRound").Return()

		ki := usecase.NewKempsInteractor(g, spMock)
		assert.Equal(t, out, ki.NextRound())
		g.AssertCalled(t, "NextRound")
	})

	t.Run("blocked when game ended", func(t *testing.T) {
		endOut := `{"gameEnd":true}`
		spMock := new(presenter.MockKempsPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(endOut)
		g := new(interfaces.MockKempsGame)
		g.On("GetGameEndFlag").Return(true)

		ki := usecase.NewKempsInteractor(g, spMock)
		assert.Equal(t, endOut, ki.NextRound())
		g.AssertNotCalled(t, "NextRound")
	})
}

func TestKempsInteractor_GetConfigAndActionLog(t *testing.T) {
	cfg := domain.KempsConfig{CpuDifficulty: domain.KempsCpuHard, TargetScore: 5}
	g := new(interfaces.MockKempsGame)
	g.On("GetConfig").Return(cfg)
	spMock := new(presenter.MockKempsPresenter)
	spMock.On("ActionLogOutput", mock.Anything).Return(`{"log":[]}`)

	ki := usecase.NewKempsInteractor(g, spMock)
	assert.Equal(t, cfg, ki.GetConfig())
	assert.Equal(t, `{"log":[]}`, ki.ActionLog())
}

// TestKempsInteractor_FullGameThroughInteractor drives a real domain game
// through the interactor's advanceCpu loop until completion across difficulties.
func TestKempsInteractor_FullGameThroughInteractor(t *testing.T) {
	for _, diff := range []domain.KempsCpuDifficulty{
		domain.KempsCpuEasy, domain.KempsCpuNormal, domain.KempsCpuHard,
	} {
		players := []*domain.KempsPlayer{
			domain.NewKempsPlayer(false),
			domain.NewKempsPlayer(false),
			domain.NewKempsPlayer(false),
			domain.NewKempsPlayer(false),
		}
		game := domain.NewKemps(domain.NewTrumpCards(0), players,
			domain.KempsConfig{CpuDifficulty: diff, TargetScore: 3})
		game.SetRand(rand.New(rand.NewSource(11)))
		sp := &kempsCountingPresenter{}
		ki := usecase.NewKempsInteractor(game, sp)
		ki.Reset() // advanceCpu runs the entire game to completion
		assert.True(t, game.GetGameEndFlag())
		require.GreaterOrEqual(t, game.GetWinnerTeam(), 0)
	}
}

func TestKempsInteractor_SnapshotAndRestore(t *testing.T) {
	game := domain.NewDefaultKemps()
	game.Reset()
	spMock := new(presenter.MockKempsPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return("{}")

	ki := usecase.NewKempsInteractor(game, spMock)
	data, err := ki.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	restored, err := usecase.RestoreKempsInteractor(data, spMock)
	require.NoError(t, err)
	assert.NotNil(t, restored)

	_, err = usecase.RestoreKempsInteractor([]byte("not-json"), spMock)
	assert.Error(t, err)
}

// kempsCountingPresenter is a no-op presenter for driving the full game.
type kempsCountingPresenter struct{}

func (p *kempsCountingPresenter) Output(_ interfaces.KempsGame, _ error) string { return "" }
func (p *kempsCountingPresenter) ActionLogOutput(_ interfaces.KempsGame) string { return "" }
