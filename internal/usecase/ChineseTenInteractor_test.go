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

const ctOut = `{"phase":0}`

func ctMocks() (*interfaces.MockChineseTenGame, *presenter.MockChineseTenPresenter) {
	g := new(interfaces.MockChineseTenGame)
	cp := new(presenter.MockChineseTenPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(ctOut)
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.ChineseTenPhasePlay)
	g.On("GetCurrentPlayerIdx").Return(0)
	return g, cp
}

func TestNewChineseTenInteractor_NilGuards(t *testing.T) {
	cp := new(presenter.MockChineseTenPresenter)
	assert.PanicsWithValue(t, "ChineseTenInteractor: c must not be nil", func() {
		usecase.NewChineseTenInteractor(nil, cp)
	})
	assert.PanicsWithValue(t, "ChineseTenInteractor: cp must not be nil", func() {
		usecase.NewChineseTenInteractor(new(interfaces.MockChineseTenGame), nil)
	})
}

func TestChineseTenInteractor_CommandsUseTheHumanSeat(t *testing.T) {
	g, cp := ctMocks()
	g.On("Reset").Return()
	g.On("PlayCard", 0, 3).Return(nil)
	g.On("SelectCapture", 0, 2).Return(nil)

	ci := usecase.NewChineseTenInteractor(g, cp)
	assert.Equal(t, ctOut, ci.Reset())
	ci.Play(3)
	ci.Select(2)

	g.AssertCalled(t, "PlayCard", 0, 3)
	g.AssertCalled(t, "SelectCapture", 0, 2)
}

func TestChineseTenInteractor_SurfacesDomainErrors(t *testing.T) {
	for name, tc := range map[string]struct {
		method string
		call   func(*usecase.ChineseTenInteractor)
	}{
		"play":   {"PlayCard", func(ci *usecase.ChineseTenInteractor) { ci.Play(9) }},
		"select": {"SelectCapture", func(ci *usecase.ChineseTenInteractor) { ci.Select(9) }},
	} {
		t.Run(name, func(t *testing.T) {
			g := new(interfaces.MockChineseTenGame)
			cp := new(presenter.MockChineseTenPresenter)
			wantErr := errors.New("out of range")
			g.On("GetGameEndFlag").Return(false)
			g.On(tc.method, 0, 9).Return(wantErr)
			cp.On("Output", mock.Anything, wantErr).Return(ctOut)

			tc.call(usecase.NewChineseTenInteractor(g, cp))
			cp.AssertCalled(t, "Output", mock.Anything, wantErr)
		})
	}
}

func TestChineseTenInteractor_CommandsAreInertOnceTheGameIsOver(t *testing.T) {
	g := new(interfaces.MockChineseTenGame)
	cp := new(presenter.MockChineseTenPresenter)
	g.On("GetGameEndFlag").Return(true)
	cp.On("Output", mock.Anything, mock.Anything).Return(ctOut)

	ci := usecase.NewChineseTenInteractor(g, cp)
	ci.Play(0)
	ci.Select(0)

	g.AssertNotCalled(t, "PlayCard", mock.Anything, mock.Anything)
	g.AssertNotCalled(t, "SelectCapture", mock.Anything, mock.Anything)
}

func TestChineseTenInteractor_CpuResolvesItsOwnSelection(t *testing.T) {
	// A selection phase on the CPU's turn must be resolved by the CPU;
	// returning early would leave the human pressing the opponent's choice.
	g := new(interfaces.MockChineseTenGame)
	cp := new(presenter.MockChineseTenPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(ctOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetCurrentPlayerIdx").Return(1).Once()
	g.On("GetCurrentPlayerIdx").Return(0)
	g.On("GetPhase").Return(domain.ChineseTenPhaseSelect)
	g.On("ChineseTenCpuDecide", 1).Return(domain.ChineseTenCpuAction{HandIdx: -1, LayoutIdx: 2})
	g.On("SelectCapture", 1, 2).Return(nil)

	usecase.NewChineseTenInteractor(g, cp).Reset()
	g.AssertCalled(t, "SelectCapture", 1, 2)
}

func TestChineseTenInteractor_CpuLoopStopsWhenAMoveIsRejected(t *testing.T) {
	// Without the short-circuit a domain that keeps rejecting the CPU's choice
	// would burn the whole turn cap on every request.
	g := new(interfaces.MockChineseTenGame)
	cp := new(presenter.MockChineseTenPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(ctOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetCurrentPlayerIdx").Return(1)
	g.On("GetPhase").Return(domain.ChineseTenPhasePlay)
	g.On("ChineseTenCpuDecide", 1).Return(domain.ChineseTenCpuAction{HandIdx: 0, LayoutIdx: -1})
	g.On("PlayCard", 1, 0).Return(errors.New("illegal"))

	usecase.NewChineseTenInteractor(g, cp).Reset()
	g.AssertNumberOfCalls(t, "PlayCard", 1)
}

func TestChineseTenInteractor_CpuLoopStopsAtTheEnd(t *testing.T) {
	g := new(interfaces.MockChineseTenGame)
	cp := new(presenter.MockChineseTenPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(ctOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(true)

	usecase.NewChineseTenInteractor(g, cp).Reset()
	g.AssertNotCalled(t, "ChineseTenCpuDecide", mock.Anything)
}

func TestChineseTenInteractor_ResetWithConfigAndAccessors(t *testing.T) {
	g, cp := ctMocks()
	g.On("Reset").Return()
	g.On("SetConfig", mock.Anything).Return()
	g.On("GetConfig").Return(domain.DefaultChineseTenConfig())
	cp.On("HintOutput", mock.Anything).Return("hint")
	cp.On("ActionLogOutput", mock.Anything).Return("log")

	ci := usecase.NewChineseTenInteractor(g, cp)
	assert.NotEmpty(t, ci.ResetWithConfig(domain.DefaultChineseTenConfig()))
	assert.Equal(t, domain.DefaultChineseTenConfig(), ci.GetConfig())
	assert.Equal(t, "hint", ci.Hint())
	assert.Equal(t, "log", ci.ActionLog())
}

func TestRestoreChineseTenInteractor(t *testing.T) {
	g := domain.NewDefaultChineseTen()
	g.Reset()
	data, err := g.MarshalJSON()
	require.NoError(t, err)

	ci, err := usecase.RestoreChineseTenInteractor(data, new(presenter.MockChineseTenPresenter))
	require.NoError(t, err)
	assert.NotNil(t, ci)

	_, err = usecase.RestoreChineseTenInteractor([]byte("{"), new(presenter.MockChineseTenPresenter))
	assert.Error(t, err)
}
