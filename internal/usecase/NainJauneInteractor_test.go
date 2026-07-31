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

const njOut = `{"phase":0}`

func njMocks() (*interfaces.MockNainJauneGame, *presenter.MockNainJaunePresenter) {
	g := new(interfaces.MockNainJauneGame)
	cp := new(presenter.MockNainJaunePresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(njOut)
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.NainJaunePhasePlay)
	g.On("GetCurrentPlayerIdx").Return(0)
	return g, cp
}

func TestNewNainJauneInteractor_NilGuards(t *testing.T) {
	cp := new(presenter.MockNainJaunePresenter)
	assert.PanicsWithValue(t, "NainJauneInteractor: c must not be nil", func() {
		usecase.NewNainJauneInteractor(nil, cp)
	})
	assert.PanicsWithValue(t, "NainJauneInteractor: cp must not be nil", func() {
		usecase.NewNainJauneInteractor(new(interfaces.MockNainJauneGame), nil)
	})
}

func TestNainJauneInteractor_PlayUsesTheHumanSeat(t *testing.T) {
	g, cp := njMocks()
	g.On("Reset").Return()
	g.On("Play", 0, 3).Return(nil)

	pi := usecase.NewNainJauneInteractor(g, cp)
	assert.Equal(t, njOut, pi.Reset())
	pi.Play(3)

	g.AssertCalled(t, "Play", 0, 3)
}

func TestNainJauneInteractor_SurfacesDomainErrors(t *testing.T) {
	g := new(interfaces.MockNainJauneGame)
	cp := new(presenter.MockNainJaunePresenter)
	wantErr := errors.New("a new run must be led with your lowest card")
	g.On("GetGameEndFlag").Return(false)
	g.On("Play", 0, 1).Return(wantErr)
	cp.On("Output", mock.Anything, wantErr).Return(njOut)

	usecase.NewNainJauneInteractor(g, cp).Play(1)
	cp.AssertCalled(t, "Output", mock.Anything, wantErr)
}

func TestNainJauneInteractor_PlayIsInertOnceTheGameIsOver(t *testing.T) {
	g := new(interfaces.MockNainJauneGame)
	cp := new(presenter.MockNainJaunePresenter)
	g.On("GetGameEndFlag").Return(true)
	cp.On("Output", mock.Anything, mock.Anything).Return(njOut)

	usecase.NewNainJauneInteractor(g, cp).Play(0)
	g.AssertNotCalled(t, "Play", mock.Anything, mock.Anything)
}

func TestNainJauneInteractor_CpuPlaysUntilTheHumansTurn(t *testing.T) {
	g := new(interfaces.MockNainJauneGame)
	cp := new(presenter.MockNainJaunePresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(njOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.NainJaunePhasePlay)
	g.On("GetCurrentPlayerIdx").Return(2).Once()
	g.On("GetCurrentPlayerIdx").Return(0)
	g.On("NainJauneCpuDecide", 2).Return(1)
	g.On("Play", 2, 1).Return(nil)

	usecase.NewNainJauneInteractor(g, cp).Reset()
	g.AssertCalled(t, "Play", 2, 1)
	g.AssertNumberOfCalls(t, "Play", 1)
}

func TestNainJauneInteractor_CpuLoopStopsAtTheEndOfADeal(t *testing.T) {
	// ディール終了で止めないと、5 区画の精算を読む間もなく次が配られる。
	g := new(interfaces.MockNainJauneGame)
	cp := new(presenter.MockNainJaunePresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(njOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.NainJaunePhaseDealEnd)

	usecase.NewNainJauneInteractor(g, cp).Reset()
	g.AssertNotCalled(t, "NainJauneCpuDecide", mock.Anything)
	g.AssertNotCalled(t, "NextDeal")
}

func TestNainJauneInteractor_CpuLoopStopsWhenAMoveIsRejected(t *testing.T) {
	g := new(interfaces.MockNainJauneGame)
	cp := new(presenter.MockNainJaunePresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(njOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.NainJaunePhasePlay)
	g.On("GetCurrentPlayerIdx").Return(1)
	g.On("NainJauneCpuDecide", 1).Return(0)
	g.On("Play", 1, 0).Return(errors.New("illegal"))

	usecase.NewNainJauneInteractor(g, cp).Reset()
	g.AssertNumberOfCalls(t, "Play", 1)
}

func TestNainJauneInteractor_CpuLoopStopsWhenThereIsNothingToPlay(t *testing.T) {
	g := new(interfaces.MockNainJauneGame)
	cp := new(presenter.MockNainJaunePresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(njOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.NainJaunePhasePlay)
	g.On("GetCurrentPlayerIdx").Return(1)
	g.On("NainJauneCpuDecide", 1).Return(-1)

	usecase.NewNainJauneInteractor(g, cp).Reset()
	g.AssertNotCalled(t, "Play", mock.Anything, mock.Anything)
}

func TestNainJauneInteractor_NextDealSurfacesItsError(t *testing.T) {
	g := new(interfaces.MockNainJauneGame)
	cp := new(presenter.MockNainJaunePresenter)
	wantErr := errors.New("the deal is still in progress")
	g.On("NextDeal").Return(wantErr)
	cp.On("Output", mock.Anything, wantErr).Return(njOut)

	usecase.NewNainJauneInteractor(g, cp).NextDeal()
	cp.AssertCalled(t, "Output", mock.Anything, wantErr)
}

func TestNainJauneInteractor_ResetWithConfigAndAccessors(t *testing.T) {
	g, cp := njMocks()
	g.On("Reset").Return()
	g.On("SetConfig", mock.Anything).Return()
	g.On("GetConfig").Return(domain.DefaultNainJauneConfig())
	cp.On("HintOutput", mock.Anything).Return("hint")
	cp.On("ActionLogOutput", mock.Anything).Return("log")

	pi := usecase.NewNainJauneInteractor(g, cp)
	assert.NotEmpty(t, pi.ResetWithConfig(domain.DefaultNainJauneConfig()))
	assert.Equal(t, domain.DefaultNainJauneConfig(), pi.GetConfig())
	assert.Equal(t, "hint", pi.Hint())
	assert.Equal(t, "log", pi.ActionLog())
}

func TestRestoreNainJauneInteractor(t *testing.T) {
	g := domain.NewDefaultNainJaune()
	g.Reset()
	data, err := g.MarshalJSON()
	require.NoError(t, err)

	pi, err := usecase.RestoreNainJauneInteractor(data, new(presenter.MockNainJaunePresenter))
	require.NoError(t, err)
	assert.NotNil(t, pi)

	_, err = usecase.RestoreNainJauneInteractor([]byte("{"), new(presenter.MockNainJaunePresenter))
	assert.Error(t, err)
}
