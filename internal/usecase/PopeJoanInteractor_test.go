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

const pjOut = `{"phase":0}`

func pjMocks() (*interfaces.MockPopeJoanGame, *presenter.MockPopeJoanPresenter) {
	g := new(interfaces.MockPopeJoanGame)
	cp := new(presenter.MockPopeJoanPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(pjOut)
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.PopeJoanPhasePlay)
	g.On("GetCurrentPlayerIdx").Return(0)
	return g, cp
}

func TestNewPopeJoanInteractor_NilGuards(t *testing.T) {
	cp := new(presenter.MockPopeJoanPresenter)
	assert.PanicsWithValue(t, "PopeJoanInteractor: c must not be nil", func() {
		usecase.NewPopeJoanInteractor(nil, cp)
	})
	assert.PanicsWithValue(t, "PopeJoanInteractor: cp must not be nil", func() {
		usecase.NewPopeJoanInteractor(new(interfaces.MockPopeJoanGame), nil)
	})
}

func TestPopeJoanInteractor_PlayUsesTheHumanSeat(t *testing.T) {
	g, cp := pjMocks()
	g.On("Reset").Return()
	g.On("Play", 0, 3).Return(nil)

	pi := usecase.NewPopeJoanInteractor(g, cp)
	assert.Equal(t, pjOut, pi.Reset())
	pi.Play(3)

	g.AssertCalled(t, "Play", 0, 3)
}

func TestPopeJoanInteractor_SurfacesDomainErrors(t *testing.T) {
	g := new(interfaces.MockPopeJoanGame)
	cp := new(presenter.MockPopeJoanPresenter)
	wantErr := errors.New("a new run must be led with your lowest card")
	g.On("GetGameEndFlag").Return(false)
	g.On("Play", 0, 1).Return(wantErr)
	cp.On("Output", mock.Anything, wantErr).Return(pjOut)

	usecase.NewPopeJoanInteractor(g, cp).Play(1)
	cp.AssertCalled(t, "Output", mock.Anything, wantErr)
}

func TestPopeJoanInteractor_PlayIsInertOnceTheGameIsOver(t *testing.T) {
	g := new(interfaces.MockPopeJoanGame)
	cp := new(presenter.MockPopeJoanPresenter)
	g.On("GetGameEndFlag").Return(true)
	cp.On("Output", mock.Anything, mock.Anything).Return(pjOut)

	usecase.NewPopeJoanInteractor(g, cp).Play(0)
	g.AssertNotCalled(t, "Play", mock.Anything, mock.Anything)
}

func TestPopeJoanInteractor_CpuPlaysUntilTheHumansTurn(t *testing.T) {
	g := new(interfaces.MockPopeJoanGame)
	cp := new(presenter.MockPopeJoanPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(pjOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.PopeJoanPhasePlay)
	g.On("GetCurrentPlayerIdx").Return(2).Once()
	g.On("GetCurrentPlayerIdx").Return(0)
	g.On("PopeJoanCpuDecide", 2).Return(1)
	g.On("Play", 2, 1).Return(nil)

	usecase.NewPopeJoanInteractor(g, cp).Reset()
	g.AssertCalled(t, "Play", 2, 1)
	g.AssertNumberOfCalls(t, "Play", 1)
}

func TestPopeJoanInteractor_CpuLoopStopsAtTheEndOfADeal(t *testing.T) {
	// ディール終了で止めないと、8 区画の精算を読む間もなく次が配られる。
	g := new(interfaces.MockPopeJoanGame)
	cp := new(presenter.MockPopeJoanPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(pjOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.PopeJoanPhaseDealEnd)

	usecase.NewPopeJoanInteractor(g, cp).Reset()
	g.AssertNotCalled(t, "PopeJoanCpuDecide", mock.Anything)
	g.AssertNotCalled(t, "NextDeal")
}

func TestPopeJoanInteractor_CpuLoopStopsWhenAMoveIsRejected(t *testing.T) {
	g := new(interfaces.MockPopeJoanGame)
	cp := new(presenter.MockPopeJoanPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(pjOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.PopeJoanPhasePlay)
	g.On("GetCurrentPlayerIdx").Return(1)
	g.On("PopeJoanCpuDecide", 1).Return(0)
	g.On("Play", 1, 0).Return(errors.New("illegal"))

	usecase.NewPopeJoanInteractor(g, cp).Reset()
	g.AssertNumberOfCalls(t, "Play", 1)
}

func TestPopeJoanInteractor_CpuLoopStopsWhenThereIsNothingToPlay(t *testing.T) {
	g := new(interfaces.MockPopeJoanGame)
	cp := new(presenter.MockPopeJoanPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(pjOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.PopeJoanPhasePlay)
	g.On("GetCurrentPlayerIdx").Return(1)
	g.On("PopeJoanCpuDecide", 1).Return(-1)

	usecase.NewPopeJoanInteractor(g, cp).Reset()
	g.AssertNotCalled(t, "Play", mock.Anything, mock.Anything)
}

func TestPopeJoanInteractor_NextDealSurfacesItsError(t *testing.T) {
	g := new(interfaces.MockPopeJoanGame)
	cp := new(presenter.MockPopeJoanPresenter)
	wantErr := errors.New("the deal is still in progress")
	g.On("NextDeal").Return(wantErr)
	cp.On("Output", mock.Anything, wantErr).Return(pjOut)

	usecase.NewPopeJoanInteractor(g, cp).NextDeal()
	cp.AssertCalled(t, "Output", mock.Anything, wantErr)
}

func TestPopeJoanInteractor_ResetWithConfigAndAccessors(t *testing.T) {
	g, cp := pjMocks()
	g.On("Reset").Return()
	g.On("SetConfig", mock.Anything).Return()
	g.On("GetConfig").Return(domain.DefaultPopeJoanConfig())
	cp.On("HintOutput", mock.Anything).Return("hint")
	cp.On("ActionLogOutput", mock.Anything).Return("log")

	pi := usecase.NewPopeJoanInteractor(g, cp)
	assert.NotEmpty(t, pi.ResetWithConfig(domain.DefaultPopeJoanConfig()))
	assert.Equal(t, domain.DefaultPopeJoanConfig(), pi.GetConfig())
	assert.Equal(t, "hint", pi.Hint())
	assert.Equal(t, "log", pi.ActionLog())
}

func TestRestorePopeJoanInteractor(t *testing.T) {
	g := domain.NewDefaultPopeJoan()
	g.Reset()
	data, err := g.MarshalJSON()
	require.NoError(t, err)

	pi, err := usecase.RestorePopeJoanInteractor(data, new(presenter.MockPopeJoanPresenter))
	require.NoError(t, err)
	assert.NotNil(t, pi)

	_, err = usecase.RestorePopeJoanInteractor([]byte("{"), new(presenter.MockPopeJoanPresenter))
	assert.Error(t, err)
}
