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

const dsOut = `{"phase":0}`

func dsMocks() (*interfaces.MockDesmocheGame, *presenter.MockDesmochePresenter) {
	g := new(interfaces.MockDesmocheGame)
	cp := new(presenter.MockDesmochePresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(dsOut)
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.DesmochePhaseAct)
	g.On("GetCurrentPlayerIdx").Return(0)
	return g, cp
}

func TestNewDesmocheInteractor_NilGuards(t *testing.T) {
	cp := new(presenter.MockDesmochePresenter)
	assert.PanicsWithValue(t, "DesmocheInteractor: c must not be nil", func() {
		usecase.NewDesmocheInteractor(nil, cp)
	})
	assert.PanicsWithValue(t, "DesmocheInteractor: cp must not be nil", func() {
		usecase.NewDesmocheInteractor(new(interfaces.MockDesmocheGame), nil)
	})
}

func TestDesmocheInteractor_CommandsUseTheHumanSeat(t *testing.T) {
	g, cp := dsMocks()
	g.On("Reset").Return()
	g.On("DrawFromStock", 0).Return(nil)
	g.On("DrawFromDiscard", 0).Return(nil)
	g.On("Meld", 0, []int{1, 2, 3}).Return(nil)
	g.On("LayOff", 0, 4, 1).Return(nil)
	g.On("Desmoche", 0, 1, 2, 3).Return(nil)
	g.On("Discard", 0, 5).Return(nil)

	di := usecase.NewDesmocheInteractor(g, cp)
	assert.Equal(t, dsOut, di.Reset())
	di.DrawStock()
	di.DrawDiscard()
	di.Meld([]int{1, 2, 3})
	di.LayOff(4, 1)
	di.Desmoche(1, 2, 3)
	di.Discard(5)

	g.AssertCalled(t, "DrawFromStock", 0)
	g.AssertCalled(t, "DrawFromDiscard", 0)
	g.AssertCalled(t, "Meld", 0, []int{1, 2, 3})
	// カード添字とメルド添字を取り違えると、別のメルドに付けてしまう。
	g.AssertCalled(t, "LayOff", 0, 4, 1)
	// desmoche は from / card / to の順。取り違えると別のメルドを崩す。
	g.AssertCalled(t, "Desmoche", 0, 1, 2, 3)
	g.AssertCalled(t, "Discard", 0, 5)
}

func TestDesmocheInteractor_SurfacesDomainErrors(t *testing.T) {
	g := new(interfaces.MockDesmocheGame)
	cp := new(presenter.MockDesmochePresenter)
	wantErr := errors.New("those cards form neither a set nor a run")
	g.On("GetGameEndFlag").Return(false)
	g.On("Meld", 0, []int{0, 1, 2}).Return(wantErr)
	cp.On("Output", mock.Anything, wantErr).Return(dsOut)

	usecase.NewDesmocheInteractor(g, cp).Meld([]int{0, 1, 2})
	cp.AssertCalled(t, "Output", mock.Anything, wantErr)
}

func TestDesmocheInteractor_CommandsAreInertOnceTheGameIsOver(t *testing.T) {
	g := new(interfaces.MockDesmocheGame)
	cp := new(presenter.MockDesmochePresenter)
	g.On("GetGameEndFlag").Return(true)
	cp.On("Output", mock.Anything, mock.Anything).Return(dsOut)

	di := usecase.NewDesmocheInteractor(g, cp)
	di.DrawStock()
	di.Meld([]int{0, 1, 2})
	di.LayOff(0, 0)
	di.Desmoche(0, 0, 1)
	di.Discard(0)

	g.AssertNotCalled(t, "DrawFromStock", mock.Anything)
	g.AssertNotCalled(t, "Meld", mock.Anything, mock.Anything)
	g.AssertNotCalled(t, "LayOff", mock.Anything, mock.Anything, mock.Anything)
	g.AssertNotCalled(t, "Desmoche", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	g.AssertNotCalled(t, "Discard", mock.Anything, mock.Anything)
}

func TestDesmocheInteractor_CpuDrawsBeforeItActs(t *testing.T) {
	// 引く前に捨てさせると「まず引け」で必ず弾かれる。
	g := new(interfaces.MockDesmocheGame)
	cp := new(presenter.MockDesmochePresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(dsOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.DesmochePhaseDraw).Once()
	g.On("GetPhase").Return(domain.DesmochePhaseRoundEnd)
	g.On("GetCurrentPlayerIdx").Return(1)
	g.On("DesmocheCpuDecide", 1).Return(domain.DesmocheCpuAction{DiscardIdx: -1})
	g.On("DrawFromStock", 1).Return(nil)

	usecase.NewDesmocheInteractor(g, cp).Reset()
	g.AssertCalled(t, "DrawFromStock", 1)
	g.AssertNotCalled(t, "Discard", mock.Anything, mock.Anything)
}

func TestDesmocheInteractor_CpuMeldsBeforeDiscarding(t *testing.T) {
	g := new(interfaces.MockDesmocheGame)
	cp := new(presenter.MockDesmochePresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(dsOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.DesmochePhaseAct).Once()
	g.On("GetPhase").Return(domain.DesmochePhaseRoundEnd)
	g.On("GetCurrentPlayerIdx").Return(1)
	g.On("DesmocheCpuDecide", 1).Return(domain.DesmocheCpuAction{MeldIdxs: []int{0, 1, 2}, DiscardIdx: -1})
	g.On("Meld", 1, []int{0, 1, 2}).Return(nil)

	usecase.NewDesmocheInteractor(g, cp).Reset()
	g.AssertCalled(t, "Meld", 1, []int{0, 1, 2})
}

func TestDesmocheInteractor_CpuLoopStopsAtTheEndOfARound(t *testing.T) {
	// ラウンド終了で止めないと、ポットが持ち越されたのかを読む間もなく次が配られる。
	g := new(interfaces.MockDesmocheGame)
	cp := new(presenter.MockDesmochePresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(dsOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.DesmochePhaseRoundEnd)

	usecase.NewDesmocheInteractor(g, cp).Reset()
	g.AssertNotCalled(t, "DesmocheCpuDecide", mock.Anything)
	g.AssertNotCalled(t, "NextRound")
}

func TestDesmocheInteractor_CpuLoopStopsWhenAMoveIsRejected(t *testing.T) {
	g := new(interfaces.MockDesmocheGame)
	cp := new(presenter.MockDesmochePresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(dsOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.DesmochePhaseAct)
	g.On("GetCurrentPlayerIdx").Return(1)
	g.On("DesmocheCpuDecide", 1).Return(domain.DesmocheCpuAction{DiscardIdx: 0})
	g.On("Discard", 1, 0).Return(errors.New("illegal"))

	usecase.NewDesmocheInteractor(g, cp).Reset()
	g.AssertNumberOfCalls(t, "Discard", 1)
}

func TestDesmocheInteractor_CpuLoopStopsWhenThereIsNothingToDo(t *testing.T) {
	g := new(interfaces.MockDesmocheGame)
	cp := new(presenter.MockDesmochePresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(dsOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.DesmochePhaseAct)
	g.On("GetCurrentPlayerIdx").Return(1)
	g.On("DesmocheCpuDecide", 1).Return(domain.DesmocheCpuAction{DiscardIdx: -1})

	usecase.NewDesmocheInteractor(g, cp).Reset()
	g.AssertNotCalled(t, "Discard", mock.Anything, mock.Anything)
	g.AssertNotCalled(t, "Meld", mock.Anything, mock.Anything)
}

func TestDesmocheInteractor_NextRoundSurfacesItsError(t *testing.T) {
	g := new(interfaces.MockDesmocheGame)
	cp := new(presenter.MockDesmochePresenter)
	wantErr := errors.New("the round is still in progress")
	g.On("NextRound").Return(wantErr)
	cp.On("Output", mock.Anything, wantErr).Return(dsOut)

	usecase.NewDesmocheInteractor(g, cp).NextRound()
	cp.AssertCalled(t, "Output", mock.Anything, wantErr)
}

func TestDesmocheInteractor_ResetWithConfigAndAccessors(t *testing.T) {
	g, cp := dsMocks()
	g.On("Reset").Return()
	g.On("SetConfig", mock.Anything).Return()
	g.On("GetConfig").Return(domain.DefaultDesmocheConfig())
	cp.On("HintOutput", mock.Anything).Return("hint")
	cp.On("ActionLogOutput", mock.Anything).Return("log")

	di := usecase.NewDesmocheInteractor(g, cp)
	assert.NotEmpty(t, di.ResetWithConfig(domain.DefaultDesmocheConfig()))
	assert.Equal(t, domain.DefaultDesmocheConfig(), di.GetConfig())
	assert.Equal(t, "hint", di.Hint())
	assert.Equal(t, "log", di.ActionLog())
}

func TestRestoreDesmocheInteractor(t *testing.T) {
	g := domain.NewDefaultDesmoche()
	g.Reset()
	data, err := g.MarshalJSON()
	require.NoError(t, err)

	di, err := usecase.RestoreDesmocheInteractor(data, new(presenter.MockDesmochePresenter))
	require.NoError(t, err)
	assert.NotNil(t, di)

	_, err = usecase.RestoreDesmocheInteractor([]byte("{"), new(presenter.MockDesmochePresenter))
	assert.Error(t, err)
}
