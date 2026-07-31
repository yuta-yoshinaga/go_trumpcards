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

const lbOut = `{"phase":0}`

func lbMocks() (*interfaces.MockLobaGame, *presenter.MockLobaPresenter) {
	g := new(interfaces.MockLobaGame)
	cp := new(presenter.MockLobaPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(lbOut)
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.LobaPhaseAct)
	g.On("GetCurrentPlayerIdx").Return(0)
	return g, cp
}

func TestNewLobaInteractor_NilGuards(t *testing.T) {
	cp := new(presenter.MockLobaPresenter)
	assert.PanicsWithValue(t, "LobaInteractor: c must not be nil", func() {
		usecase.NewLobaInteractor(nil, cp)
	})
	assert.PanicsWithValue(t, "LobaInteractor: cp must not be nil", func() {
		usecase.NewLobaInteractor(new(interfaces.MockLobaGame), nil)
	})
}

func TestLobaInteractor_CommandsUseTheHumanSeat(t *testing.T) {
	g, cp := lbMocks()
	g.On("Reset").Return()
	g.On("DrawFromStock", 0).Return(nil)
	g.On("DrawFromDiscard", 0).Return(nil)
	g.On("Meld", 0, []int{1, 2, 3}).Return(nil)
	g.On("LayOff", 0, 4, 1).Return(nil)
	g.On("Discard", 0, 5).Return(nil)

	li := usecase.NewLobaInteractor(g, cp)
	assert.Equal(t, lbOut, li.Reset())
	li.DrawStock()
	li.DrawDiscard()
	li.Meld([]int{1, 2, 3})
	li.LayOff(4, 1)
	li.Discard(5)

	g.AssertCalled(t, "DrawFromStock", 0)
	g.AssertCalled(t, "DrawFromDiscard", 0)
	g.AssertCalled(t, "Meld", 0, []int{1, 2, 3})
	// カード添字とメルド添字を取り違えると、別のメルドに付けてしまう。
	g.AssertCalled(t, "LayOff", 0, 4, 1)
	g.AssertCalled(t, "Discard", 0, 5)
}

func TestLobaInteractor_SurfacesDomainErrors(t *testing.T) {
	g := new(interfaces.MockLobaGame)
	cp := new(presenter.MockLobaPresenter)
	wantErr := errors.New("a pierna needs three different suits")
	g.On("GetGameEndFlag").Return(false)
	g.On("Meld", 0, []int{0, 1, 2}).Return(wantErr)
	cp.On("Output", mock.Anything, wantErr).Return(lbOut)

	usecase.NewLobaInteractor(g, cp).Meld([]int{0, 1, 2})
	cp.AssertCalled(t, "Output", mock.Anything, wantErr)
}

func TestLobaInteractor_CommandsAreInertOnceTheGameIsOver(t *testing.T) {
	g := new(interfaces.MockLobaGame)
	cp := new(presenter.MockLobaPresenter)
	g.On("GetGameEndFlag").Return(true)
	cp.On("Output", mock.Anything, mock.Anything).Return(lbOut)

	li := usecase.NewLobaInteractor(g, cp)
	li.DrawStock()
	li.Meld([]int{0, 1, 2})
	li.Discard(0)

	g.AssertNotCalled(t, "DrawFromStock", mock.Anything)
	g.AssertNotCalled(t, "Meld", mock.Anything, mock.Anything)
	g.AssertNotCalled(t, "Discard", mock.Anything, mock.Anything)
}

func TestLobaInteractor_CpuDrawsBeforeItActs(t *testing.T) {
	// 引く前に捨てさせると「まず引け」で必ず弾かれる。
	g := new(interfaces.MockLobaGame)
	cp := new(presenter.MockLobaPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(lbOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.LobaPhaseDraw).Once()
	g.On("GetPhase").Return(domain.LobaPhaseRoundEnd)
	g.On("GetCurrentPlayerIdx").Return(1)
	g.On("LobaCpuDecide", 1).Return(domain.LobaCpuAction{DiscardIdx: -1})
	g.On("DrawFromStock", 1).Return(nil)

	usecase.NewLobaInteractor(g, cp).Reset()
	g.AssertCalled(t, "DrawFromStock", 1)
	g.AssertNotCalled(t, "Discard", mock.Anything, mock.Anything)
}

func TestLobaInteractor_CpuMeldsBeforeDiscarding(t *testing.T) {
	g := new(interfaces.MockLobaGame)
	cp := new(presenter.MockLobaPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(lbOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.LobaPhaseAct).Once()
	g.On("GetPhase").Return(domain.LobaPhaseRoundEnd)
	g.On("GetCurrentPlayerIdx").Return(1)
	g.On("LobaCpuDecide", 1).Return(domain.LobaCpuAction{MeldIdxs: []int{0, 1, 2}, DiscardIdx: -1})
	g.On("Meld", 1, []int{0, 1, 2}).Return(nil)

	usecase.NewLobaInteractor(g, cp).Reset()
	g.AssertCalled(t, "Meld", 1, []int{0, 1, 2})
}

func TestLobaInteractor_CpuLoopStopsAtTheEndOfARound(t *testing.T) {
	// ラウンド終了で止めないと、精算を読む間もなく次が配られる。
	g := new(interfaces.MockLobaGame)
	cp := new(presenter.MockLobaPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(lbOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.LobaPhaseRoundEnd)

	usecase.NewLobaInteractor(g, cp).Reset()
	g.AssertNotCalled(t, "LobaCpuDecide", mock.Anything)
	g.AssertNotCalled(t, "NextRound")
}

func TestLobaInteractor_CpuLoopStopsWhenAMoveIsRejected(t *testing.T) {
	g := new(interfaces.MockLobaGame)
	cp := new(presenter.MockLobaPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(lbOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.LobaPhaseAct)
	g.On("GetCurrentPlayerIdx").Return(1)
	g.On("LobaCpuDecide", 1).Return(domain.LobaCpuAction{DiscardIdx: 0})
	g.On("Discard", 1, 0).Return(errors.New("illegal"))

	usecase.NewLobaInteractor(g, cp).Reset()
	g.AssertNumberOfCalls(t, "Discard", 1)
}

func TestLobaInteractor_NextRoundSurfacesItsError(t *testing.T) {
	g := new(interfaces.MockLobaGame)
	cp := new(presenter.MockLobaPresenter)
	wantErr := errors.New("the round is still in progress")
	g.On("NextRound").Return(wantErr)
	cp.On("Output", mock.Anything, wantErr).Return(lbOut)

	usecase.NewLobaInteractor(g, cp).NextRound()
	cp.AssertCalled(t, "Output", mock.Anything, wantErr)
}

func TestLobaInteractor_ResetWithConfigAndAccessors(t *testing.T) {
	g, cp := lbMocks()
	g.On("Reset").Return()
	g.On("SetConfig", mock.Anything).Return()
	g.On("GetConfig").Return(domain.DefaultLobaConfig())
	cp.On("HintOutput", mock.Anything).Return("hint")
	cp.On("ActionLogOutput", mock.Anything).Return("log")

	li := usecase.NewLobaInteractor(g, cp)
	assert.NotEmpty(t, li.ResetWithConfig(domain.DefaultLobaConfig()))
	assert.Equal(t, domain.DefaultLobaConfig(), li.GetConfig())
	assert.Equal(t, "hint", li.Hint())
	assert.Equal(t, "log", li.ActionLog())
}

func TestRestoreLobaInteractor(t *testing.T) {
	g := domain.NewDefaultLoba()
	g.Reset()
	data, err := g.MarshalJSON()
	require.NoError(t, err)

	li, err := usecase.RestoreLobaInteractor(data, new(presenter.MockLobaPresenter))
	require.NoError(t, err)
	assert.NotNil(t, li)

	_, err = usecase.RestoreLobaInteractor([]byte("{"), new(presenter.MockLobaPresenter))
	assert.Error(t, err)
}
