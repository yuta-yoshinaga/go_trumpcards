//go:build test

package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// newSpoonsGameMock returns a mock whose advanceCpu loop short-circuits
// immediately: gameEnd=false so action guards pass, but phase=GameEnd so the
// advanceCpu switch returns on the default branch without further calls.
func newSpoonsGameMock() *interfaces.MockSpoonsGame {
	g := new(interfaces.MockSpoonsGame)
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.SpoonsPhaseGameEnd)
	return g
}

func TestNewSpoonsInteractor_NilGuards(t *testing.T) {
	spMock := new(presenter.MockSpoonsPresenter)
	assert.PanicsWithValue(t, "SpoonsInteractor: g must not be nil", func() {
		usecase.NewSpoonsInteractor(nil, spMock)
	})
	assert.PanicsWithValue(t, "SpoonsInteractor: sp must not be nil", func() {
		usecase.NewSpoonsInteractor(domain.NewDefaultSpoons(), nil)
	})
}

func TestSpoonsInteractor_Reset(t *testing.T) {
	out := `{"phase":0}`
	spMock := new(presenter.MockSpoonsPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(out)
	g := newSpoonsGameMock()
	g.On("Reset").Return()

	si := usecase.NewSpoonsInteractor(g, spMock)
	assert.Equal(t, out, si.Reset())
	g.AssertCalled(t, "Reset")
}

func TestSpoonsInteractor_ResetWithConfig(t *testing.T) {
	out := `{"phase":0}`

	t.Run("valid", func(t *testing.T) {
		spMock := new(presenter.MockSpoonsPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(out)
		g := newSpoonsGameMock()
		cfg := domain.SpoonsConfig{CpuDifficulty: domain.SpoonsCpuHard}
		g.On("SetConfig", cfg).Return()
		g.On("Reset").Return()

		si := usecase.NewSpoonsInteractor(g, spMock)
		assert.Equal(t, out, si.ResetWithConfig(cfg))
		g.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config returns error", func(t *testing.T) {
		errOut := `{"error":"invalid"}`
		spMock := new(presenter.MockSpoonsPresenter)
		spMock.On("Output", mock.Anything, mock.MatchedBy(func(e error) bool { return e != nil })).Return(errOut)
		g := newSpoonsGameMock()

		si := usecase.NewSpoonsInteractor(g, spMock)
		assert.Equal(t, errOut, si.ResetWithConfig(domain.SpoonsConfig{CpuDifficulty: 99}))
	})
}

func TestSpoonsInteractor_Pass(t *testing.T) {
	out := `{"phase":0}`

	t.Run("success", func(t *testing.T) {
		spMock := new(presenter.MockSpoonsPresenter)
		spMock.On("Output", mock.Anything, nil).Return(out)
		g := newSpoonsGameMock()
		g.On("IsHumanTurn").Return(true)
		g.On("PlayerPass", 1).Return(nil)

		si := usecase.NewSpoonsInteractor(g, spMock)
		assert.Equal(t, out, si.Pass(1))
		g.AssertCalled(t, "PlayerPass", 1)
	})

	t.Run("error from domain", func(t *testing.T) {
		errOut := `{"error":"x"}`
		spMock := new(presenter.MockSpoonsPresenter)
		spMock.On("Output", mock.Anything, mock.MatchedBy(func(e error) bool { return e != nil })).Return(errOut)
		g := newSpoonsGameMock()
		g.On("IsHumanTurn").Return(true)
		g.On("PlayerPass", 0).Return(domain.ErrInvalidCard)

		si := usecase.NewSpoonsInteractor(g, spMock)
		assert.Equal(t, errOut, si.Pass(0))
	})

	t.Run("blocked when not playable", func(t *testing.T) {
		blockedOut := `{"blocked":true}`
		spMock := new(presenter.MockSpoonsPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(blockedOut)
		g := new(interfaces.MockSpoonsGame)
		g.On("GetGameEndFlag").Return(false)
		g.On("IsHumanTurn").Return(false)

		si := usecase.NewSpoonsInteractor(g, spMock)
		assert.Equal(t, blockedOut, si.Pass(0))
		g.AssertNotCalled(t, "PlayerPass", mock.Anything)
	})
}

func TestSpoonsInteractor_Grab(t *testing.T) {
	out := `{"phase":1}`

	t.Run("success", func(t *testing.T) {
		spMock := new(presenter.MockSpoonsPresenter)
		spMock.On("Output", mock.Anything, nil).Return(out)
		g := newSpoonsGameMock()
		g.On("PlayerGrabSpoon").Return(nil)

		si := usecase.NewSpoonsInteractor(g, spMock)
		assert.Equal(t, out, si.Grab())
		g.AssertCalled(t, "PlayerGrabSpoon")
	})

	t.Run("error from domain", func(t *testing.T) {
		errOut := `{"error":"x"}`
		spMock := new(presenter.MockSpoonsPresenter)
		spMock.On("Output", mock.Anything, mock.MatchedBy(func(e error) bool { return e != nil })).Return(errOut)
		g := newSpoonsGameMock()
		g.On("PlayerGrabSpoon").Return(domain.ErrInvalidPlay)

		si := usecase.NewSpoonsInteractor(g, spMock)
		assert.Equal(t, errOut, si.Grab())
	})

	t.Run("blocked when game ended", func(t *testing.T) {
		endOut := `{"gameEnd":true}`
		spMock := new(presenter.MockSpoonsPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(endOut)
		g := new(interfaces.MockSpoonsGame)
		g.On("GetGameEndFlag").Return(true)

		si := usecase.NewSpoonsInteractor(g, spMock)
		assert.Equal(t, endOut, si.Grab())
		g.AssertNotCalled(t, "PlayerGrabSpoon")
	})
}

func TestSpoonsInteractor_NextRound(t *testing.T) {
	out := `{"phase":0}`

	t.Run("success", func(t *testing.T) {
		spMock := new(presenter.MockSpoonsPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(out)
		g := newSpoonsGameMock()
		g.On("NextRound").Return()

		si := usecase.NewSpoonsInteractor(g, spMock)
		assert.Equal(t, out, si.NextRound())
		g.AssertCalled(t, "NextRound")
	})

	t.Run("blocked when game ended", func(t *testing.T) {
		endOut := `{"gameEnd":true}`
		spMock := new(presenter.MockSpoonsPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(endOut)
		g := new(interfaces.MockSpoonsGame)
		g.On("GetGameEndFlag").Return(true)

		si := usecase.NewSpoonsInteractor(g, spMock)
		assert.Equal(t, endOut, si.NextRound())
		g.AssertNotCalled(t, "NextRound")
	})
}

func TestSpoonsInteractor_GetConfigAndActionLog(t *testing.T) {
	cfg := domain.SpoonsConfig{CpuDifficulty: domain.SpoonsCpuHard}
	g := new(interfaces.MockSpoonsGame)
	g.On("GetConfig").Return(cfg)
	spMock := new(presenter.MockSpoonsPresenter)
	spMock.On("ActionLogOutput", mock.Anything).Return(`{"log":[]}`)

	si := usecase.NewSpoonsInteractor(g, spMock)
	assert.Equal(t, cfg, si.GetConfig())
	assert.Equal(t, `{"log":[]}`, si.ActionLog())
}

// TestSpoonsInteractor_FullGameThroughInteractor drives a real domain game
// through the interactor's advanceCpu loop until completion, exercising the
// pass/grab/round-end auto-advance path.
func TestSpoonsInteractor_FullGameThroughInteractor(t *testing.T) {
	for _, diff := range []domain.SpoonsCpuDifficulty{
		domain.SpoonsCpuEasy, domain.SpoonsCpuNormal, domain.SpoonsCpuHard,
	} {
		// 全員 CPU にすることで advanceCpu が単独で最後まで進む。
		players := []*domain.SpoonsPlayer{
			domain.NewSpoonsPlayer(false),
			domain.NewSpoonsPlayer(false),
			domain.NewSpoonsPlayer(false),
			domain.NewSpoonsPlayer(false),
		}
		game := domain.NewSpoons(domain.NewTrumpCards(0), players, domain.SpoonsConfig{CpuDifficulty: diff})
		sp := &spoonsCountingPresenter{}
		si := usecase.NewSpoonsInteractor(game, sp)
		si.Reset() // advanceCpu runs the entire game to completion
		assert.True(t, game.GetGameEndFlag())
		require.GreaterOrEqual(t, game.GetWinnerIdx(), 0)
	}
}

func TestSpoonsInteractor_SnapshotAndRestore(t *testing.T) {
	game := domain.NewDefaultSpoons()
	game.Reset()
	spMock := new(presenter.MockSpoonsPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return("{}")

	si := usecase.NewSpoonsInteractor(game, spMock)
	data, err := si.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	restored, err := usecase.RestoreSpoonsInteractor(data, spMock)
	require.NoError(t, err)
	assert.NotNil(t, restored)

	_, err = usecase.RestoreSpoonsInteractor([]byte("not-json"), spMock)
	assert.Error(t, err)
}

// spoonsCountingPresenter is a no-op presenter for driving the full game.
type spoonsCountingPresenter struct{}

func (p *spoonsCountingPresenter) Output(_ interfaces.SpoonsGame, _ error) string { return "" }
func (p *spoonsCountingPresenter) ActionLogOutput(_ interfaces.SpoonsGame) string { return "" }
