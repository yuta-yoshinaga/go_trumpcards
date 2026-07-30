//go:build test

package usecase_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

const toepenOut = `{"phase":0}`

func toepenMocks() (*interfaces.MockToepenGame, *presenter.MockToepenPresenter) {
	g := new(interfaces.MockToepenGame)
	tp := new(presenter.MockToepenPresenter)
	tp.On("Output", mock.Anything, mock.Anything).Return(toepenOut)
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.ToepenPhasePlay)
	g.On("GetCurrentPlayerIdx").Return(0)
	return g, tp
}

func TestNewToepenInteractor_NilGuards(t *testing.T) {
	tp := new(presenter.MockToepenPresenter)
	assert.PanicsWithValue(t, "ToepenInteractor: t must not be nil", func() {
		usecase.NewToepenInteractor(nil, tp)
	})
	assert.PanicsWithValue(t, "ToepenInteractor: tp must not be nil", func() {
		usecase.NewToepenInteractor(new(interfaces.MockToepenGame), nil)
	})
}

func TestToepenInteractor_CommandsUseTheHumanSeat(t *testing.T) {
	g, tp := toepenMocks()
	g.On("Reset").Return()
	g.On("PlayCard", 0, 2).Return(nil)
	g.On("Toep", 0).Return(nil)
	g.On("Respond", 0, true).Return(nil)
	g.On("NextHand").Return(nil)

	ti := usecase.NewToepenInteractor(g, tp)
	assert.Equal(t, toepenOut, ti.Reset())
	ti.Play(2)
	ti.Toep()
	ti.Respond(true)
	ti.NextHand()

	g.AssertCalled(t, "PlayCard", 0, 2)
	g.AssertCalled(t, "Toep", 0)
	g.AssertCalled(t, "Respond", 0, true)
	g.AssertCalled(t, "NextHand")
}

func TestToepenInteractor_SurfacesDomainErrors(t *testing.T) {
	for name, tc := range map[string]struct {
		setup func(*interfaces.MockToepenGame, error)
		call  func(*usecase.ToepenInteractor)
	}{
		"play": {
			func(g *interfaces.MockToepenGame, e error) { g.On("PlayCard", 0, 9).Return(e) },
			func(ti *usecase.ToepenInteractor) { ti.Play(9) },
		},
		"toep": {
			func(g *interfaces.MockToepenGame, e error) { g.On("Toep", 0).Return(e) },
			func(ti *usecase.ToepenInteractor) { ti.Toep() },
		},
		"respond": {
			func(g *interfaces.MockToepenGame, e error) { g.On("Respond", 0, false).Return(e) },
			func(ti *usecase.ToepenInteractor) { ti.Respond(false) },
		},
		"next": {
			func(g *interfaces.MockToepenGame, e error) { g.On("NextHand").Return(e) },
			func(ti *usecase.ToepenInteractor) { ti.NextHand() },
		},
	} {
		t.Run(name, func(t *testing.T) {
			g := new(interfaces.MockToepenGame)
			tp := new(presenter.MockToepenPresenter)
			wantErr := errors.New("nope")
			g.On("GetGameEndFlag").Return(false)
			tc.setup(g, wantErr)
			tp.On("Output", mock.Anything, wantErr).Return(toepenOut)

			tc.call(usecase.NewToepenInteractor(g, tp))
			tp.AssertCalled(t, "Output", mock.Anything, wantErr)
		})
	}
}

func TestToepenInteractor_CommandsAreInertOnceTheGameIsOver(t *testing.T) {
	g := new(interfaces.MockToepenGame)
	tp := new(presenter.MockToepenPresenter)
	g.On("GetGameEndFlag").Return(true)
	tp.On("Output", mock.Anything, mock.Anything).Return(toepenOut)

	ti := usecase.NewToepenInteractor(g, tp)
	ti.Play(0)
	ti.Toep()
	ti.Respond(true)

	g.AssertNotCalled(t, "PlayCard", mock.Anything, mock.Anything)
	g.AssertNotCalled(t, "Toep", mock.Anything)
	g.AssertNotCalled(t, "Respond", mock.Anything, mock.Anything)
}

func TestToepenInteractor_CpuAnswersItsOwnToep(t *testing.T) {
	// A response phase belonging to a CPU must be answered by that CPU.
	// Returning early would leave the human pressing an opponent's stay/fold.
	g := new(interfaces.MockToepenGame)
	tp := new(presenter.MockToepenPresenter)
	tp.On("Output", mock.Anything, mock.Anything).Return(toepenOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.ToepenPhaseRespond).Once()
	g.On("GetPhase").Return(domain.ToepenPhasePlay)
	g.On("GetPendingRespondent").Return(2)
	g.On("GetCurrentPlayerIdx").Return(0)
	g.On("ToepenCpuDecide", 2).Return(domain.ToepenCpuAction{HandIdx: -1, Fold: true})
	g.On("Respond", 2, false).Return(nil)

	usecase.NewToepenInteractor(g, tp).Reset()
	g.AssertCalled(t, "Respond", 2, false)
}

func TestToepenInteractor_CpuLoopStopsAtTheHumansResponse(t *testing.T) {
	g := new(interfaces.MockToepenGame)
	tp := new(presenter.MockToepenPresenter)
	tp.On("Output", mock.Anything, mock.Anything).Return(toepenOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.ToepenPhaseRespond)
	g.On("GetPendingRespondent").Return(0)

	usecase.NewToepenInteractor(g, tp).Reset()
	g.AssertNotCalled(t, "ToepenCpuDecide", mock.Anything)
}

func TestToepenInteractor_CpuLoopStopsWhenAPlayIsRejected(t *testing.T) {
	// Without the short-circuit a domain that keeps rejecting the CPU's choice
	// would burn the whole turn cap on every request.
	g := new(interfaces.MockToepenGame)
	tp := new(presenter.MockToepenPresenter)
	tp.On("Output", mock.Anything, mock.Anything).Return(toepenOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.ToepenPhasePlay)
	g.On("GetCurrentPlayerIdx").Return(1)
	g.On("ToepenCpuDecide", 1).Return(domain.ToepenCpuAction{HandIdx: 0})
	g.On("PlayCard", 1, 0).Return(errors.New("illegal"))

	usecase.NewToepenInteractor(g, tp).Reset()
	g.AssertNumberOfCalls(t, "PlayCard", 1)
}

func TestToepenInteractor_CpuLoopStopsAtAHandEnd(t *testing.T) {
	g := new(interfaces.MockToepenGame)
	tp := new(presenter.MockToepenPresenter)
	tp.On("Output", mock.Anything, mock.Anything).Return(toepenOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.ToepenPhaseHandEnd)

	usecase.NewToepenInteractor(g, tp).Reset()
	g.AssertNotCalled(t, "ToepenCpuDecide", mock.Anything)
}

func TestToepenInteractor_ResetWithConfigAndAccessors(t *testing.T) {
	g, tp := toepenMocks()
	g.On("Reset").Return()
	g.On("SetConfig", mock.Anything).Return()
	g.On("GetConfig").Return(domain.DefaultToepenConfig())
	tp.On("HintOutput", mock.Anything).Return("hint")
	tp.On("ActionLogOutput", mock.Anything).Return("log")

	ti := usecase.NewToepenInteractor(g, tp)
	assert.NotEmpty(t, ti.ResetWithConfig(domain.DefaultToepenConfig()))
	assert.Equal(t, domain.DefaultToepenConfig(), ti.GetConfig())
	assert.Equal(t, "hint", ti.Hint())
	assert.Equal(t, "log", ti.ActionLog())
}

func TestRestoreToepenInteractor(t *testing.T) {
	g := domain.NewDefaultToepen()
	g.Reset()
	data, err := g.MarshalJSON()
	require.NoError(t, err)

	ti, err := usecase.RestoreToepenInteractor(data, new(presenter.MockToepenPresenter))
	require.NoError(t, err)
	assert.NotNil(t, ti)

	_, err = usecase.RestoreToepenInteractor([]byte("{"), new(presenter.MockToepenPresenter))
	assert.Error(t, err)
}

func TestToepenInteractor_RedealUsesTheHumanSeatAndSurfacesRejection(t *testing.T) {
	g, tp := toepenMocks()
	g.On("Redeal", 0).Return(nil)
	usecase.NewToepenInteractor(g, tp).Redeal()
	g.AssertCalled(t, "Redeal", 0)

	g2 := new(interfaces.MockToepenGame)
	tp2 := new(presenter.MockToepenPresenter)
	wantErr := errors.New("a redeal needs a hand of nothing but A, K, Q and J")
	g2.On("GetGameEndFlag").Return(false)
	g2.On("Redeal", 0).Return(wantErr)
	tp2.On("Output", mock.Anything, wantErr).Return(toepenOut)

	usecase.NewToepenInteractor(g2, tp2).Redeal()
	tp2.AssertCalled(t, "Output", mock.Anything, wantErr)
}

func TestToepenInteractor_RedealIsInertOnceTheGameIsOver(t *testing.T) {
	g := new(interfaces.MockToepenGame)
	tp := new(presenter.MockToepenPresenter)
	g.On("GetGameEndFlag").Return(true)
	tp.On("Output", mock.Anything, mock.Anything).Return(toepenOut)

	usecase.NewToepenInteractor(g, tp).Redeal()
	g.AssertNotCalled(t, "Redeal", mock.Anything)
}
