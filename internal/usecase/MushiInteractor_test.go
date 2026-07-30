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

const mushiOut = `{"phase":0}`

// mushiMocks wires a presenter that always renders and a game already on the
// human's turn, so runCpuTurns returns at once unless a test says otherwise.
func mushiMocks() (*interfaces.MockMushiGame, *presenter.MockMushiPresenter) {
	gameMock := new(interfaces.MockMushiGame)
	mpMock := new(presenter.MockMushiPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(mushiOut)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.MushiPhasePlay)
	gameMock.On("GetCurrentPlayerIdx").Return(0)
	return gameMock, mpMock
}

func TestNewMushiInteractor_NilGuards(t *testing.T) {
	mpMock := new(presenter.MockMushiPresenter)

	assert.PanicsWithValue(t, "MushiInteractor: m must not be nil", func() {
		usecase.NewMushiInteractor(nil, mpMock)
	})
	assert.PanicsWithValue(t, "MushiInteractor: mp must not be nil", func() {
		usecase.NewMushiInteractor(new(interfaces.MockMushiGame), nil)
	})
}

func TestMushiInteractor_ResetAndPlay(t *testing.T) {
	gameMock, mpMock := mushiMocks()
	gameMock.On("Reset").Return()
	gameMock.On("PlayCard", 0, 3).Return(nil)

	mi := usecase.NewMushiInteractor(gameMock, mpMock)
	assert.Equal(t, mushiOut, mi.Reset())
	mi.Play(3)

	gameMock.AssertCalled(t, "Reset")
	gameMock.AssertCalled(t, "PlayCard", 0, 3)
}

func TestMushiInteractor_SelectUsesTheHumanSeat(t *testing.T) {
	gameMock, mpMock := mushiMocks()
	gameMock.On("SelectCapture", 0, 2).Return(nil)

	usecase.NewMushiInteractor(gameMock, mpMock).Select(2)
	gameMock.AssertCalled(t, "SelectCapture", 0, 2)
}

func TestMushiInteractor_SurfacesDomainErrors(t *testing.T) {
	for name, tc := range map[string]struct {
		method string
		call   func(*usecase.MushiInteractor)
	}{
		"play":   {"PlayCard", func(mi *usecase.MushiInteractor) { mi.Play(9) }},
		"select": {"SelectCapture", func(mi *usecase.MushiInteractor) { mi.Select(9) }},
	} {
		t.Run(name, func(t *testing.T) {
			gameMock := new(interfaces.MockMushiGame)
			mpMock := new(presenter.MockMushiPresenter)
			wantErr := errors.New("out of range")
			gameMock.On("GetGameEndFlag").Return(false)
			gameMock.On(tc.method, 0, 9).Return(wantErr)
			mpMock.On("Output", mock.Anything, wantErr).Return(mushiOut)

			tc.call(usecase.NewMushiInteractor(gameMock, mpMock))
			mpMock.AssertCalled(t, "Output", mock.Anything, wantErr)
		})
	}
}

func TestMushiInteractor_NextRoundSurfacesItsRejection(t *testing.T) {
	gameMock := new(interfaces.MockMushiGame)
	mpMock := new(presenter.MockMushiPresenter)
	wantErr := errors.New("the round is still in progress")
	gameMock.On("NextRound").Return(wantErr)
	mpMock.On("Output", mock.Anything, wantErr).Return(mushiOut)

	usecase.NewMushiInteractor(gameMock, mpMock).NextRound()
	mpMock.AssertCalled(t, "Output", mock.Anything, wantErr)
}

func TestMushiInteractor_NextRoundRunsTheCpu(t *testing.T) {
	gameMock, mpMock := mushiMocks()
	gameMock.On("NextRound").Return(nil)

	assert.Equal(t, mushiOut, usecase.NewMushiInteractor(gameMock, mpMock).NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestMushiInteractor_CommandsAreInertOnceTheGameIsOver(t *testing.T) {
	gameMock := new(interfaces.MockMushiGame)
	mpMock := new(presenter.MockMushiPresenter)
	gameMock.On("GetGameEndFlag").Return(true)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(mushiOut)

	mi := usecase.NewMushiInteractor(gameMock, mpMock)
	mi.Play(0)
	mi.Select(0)

	gameMock.AssertNotCalled(t, "PlayCard", mock.Anything, mock.Anything)
	gameMock.AssertNotCalled(t, "SelectCapture", mock.Anything, mock.Anything)
}

func TestMushiInteractor_CpuLoopResolvesItsOwnSelections(t *testing.T) {
	// A selection phase on the CPU's turn must be resolved by the CPU. Returning
	// early would leave the human pressing the opponent's choice.
	gameMock := new(interfaces.MockMushiGame)
	mpMock := new(presenter.MockMushiPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(mushiOut)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.MushiPhaseSelect).Once()
	gameMock.On("GetPhase").Return(domain.MushiPhasePlay)
	gameMock.On("GetCurrentPlayerIdx").Return(1).Once()
	gameMock.On("GetCurrentPlayerIdx").Return(0)
	gameMock.On("MushiCpuDecide", 1).Return(domain.MushiCpuAction{HandIdx: -1, FieldIdx: 2})
	gameMock.On("SelectCapture", 1, 2).Return(nil)

	usecase.NewMushiInteractor(gameMock, mpMock).Reset()
	gameMock.AssertCalled(t, "SelectCapture", 1, 2)
}

func TestMushiInteractor_CpuLoopStopsWhenAMoveIsRejected(t *testing.T) {
	// Without the short-circuit a domain that keeps rejecting the CPU's choice
	// would burn the whole turn cap on every request.
	gameMock := new(interfaces.MockMushiGame)
	mpMock := new(presenter.MockMushiPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(mushiOut)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.MushiPhasePlay)
	gameMock.On("GetCurrentPlayerIdx").Return(1)
	gameMock.On("MushiCpuDecide", 1).Return(domain.MushiCpuAction{HandIdx: 0, FieldIdx: -1})
	gameMock.On("PlayCard", 1, 0).Return(errors.New("illegal"))

	usecase.NewMushiInteractor(gameMock, mpMock).Reset()
	gameMock.AssertNumberOfCalls(t, "PlayCard", 1)
}

func TestMushiInteractor_CpuLoopStopsAtARoundEnd(t *testing.T) {
	gameMock := new(interfaces.MockMushiGame)
	mpMock := new(presenter.MockMushiPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(mushiOut)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.MushiPhaseRoundEnd)

	usecase.NewMushiInteractor(gameMock, mpMock).Reset()
	gameMock.AssertNotCalled(t, "MushiCpuDecide", mock.Anything)
}

func TestMushiInteractor_ResetWithConfigAndAccessors(t *testing.T) {
	gameMock, mpMock := mushiMocks()
	gameMock.On("Reset").Return()
	gameMock.On("SetConfig", mock.Anything).Return()
	gameMock.On("GetConfig").Return(domain.DefaultMushiConfig())
	mpMock.On("HintOutput", mock.Anything).Return("hint")
	mpMock.On("ActionLogOutput", mock.Anything).Return("log")

	mi := usecase.NewMushiInteractor(gameMock, mpMock)
	assert.NotEmpty(t, mi.ResetWithConfig(domain.DefaultMushiConfig()))
	assert.Equal(t, domain.DefaultMushiConfig(), mi.GetConfig())
	assert.Equal(t, "hint", mi.Hint())
	assert.Equal(t, "log", mi.ActionLog())
}

func TestRestoreMushiInteractor(t *testing.T) {
	g := domain.NewDefaultMushi()
	g.Reset()
	data, err := g.MarshalJSON()
	require.NoError(t, err)

	mi, err := usecase.RestoreMushiInteractor(data, new(presenter.MockMushiPresenter))
	require.NoError(t, err)
	assert.NotNil(t, mi)

	_, err = usecase.RestoreMushiInteractor([]byte("{"), new(presenter.MockMushiPresenter))
	assert.Error(t, err)
}
