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

const zwOut = `{"phase":0}`

func zwMocks() (*interfaces.MockZwickerGame, *presenter.MockZwickerPresenter) {
	g := new(interfaces.MockZwickerGame)
	cp := new(presenter.MockZwickerPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(zwOut)
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.ZwickerPhasePlay)
	g.On("GetCurrentPlayerIdx").Return(0)
	return g, cp
}

func TestNewZwickerInteractor_NilGuards(t *testing.T) {
	cp := new(presenter.MockZwickerPresenter)
	assert.PanicsWithValue(t, "ZwickerInteractor: c must not be nil", func() {
		usecase.NewZwickerInteractor(nil, cp)
	})
	assert.PanicsWithValue(t, "ZwickerInteractor: cp must not be nil", func() {
		usecase.NewZwickerInteractor(new(interfaces.MockZwickerGame), nil)
	})
}

func TestZwickerInteractor_CommandsUseTheHumanSeat(t *testing.T) {
	g, cp := zwMocks()
	g.On("Reset").Return()
	g.On("Take", 0, 1, 7, []int{2, 3}, []int{0}).Return(nil)
	g.On("Build", 0, 1, []int{2}, 9).Return(nil)
	g.On("Trail", 0, 4).Return(nil)

	zi := usecase.NewZwickerInteractor(g, cp)
	assert.Equal(t, zwOut, zi.Reset())
	zi.Take(1, 7, []int{2, 3}, []int{0})
	zi.Build(1, []int{2}, 9)
	zi.Trail(4)

	// **値は札とは別の引数。**A と絵札は 2 択を持つので、取り違えると
	// まったく別の捕獲になる。
	g.AssertCalled(t, "Take", 0, 1, 7, []int{2, 3}, []int{0})
	g.AssertCalled(t, "Build", 0, 1, []int{2}, 9)
	g.AssertCalled(t, "Trail", 0, 4)
}

func TestZwickerInteractor_SurfacesDomainErrors(t *testing.T) {
	g := new(interfaces.MockZwickerGame)
	cp := new(presenter.MockZwickerPresenter)
	wantErr := errors.New("those table cards do not add up to 7")
	g.On("GetGameEndFlag").Return(false)
	g.On("Take", 0, 0, 7, []int{0}, []int(nil)).Return(wantErr)
	cp.On("Output", mock.Anything, wantErr).Return(zwOut)

	usecase.NewZwickerInteractor(g, cp).Take(0, 7, []int{0}, nil)
	cp.AssertCalled(t, "Output", mock.Anything, wantErr)
}

func TestZwickerInteractor_CommandsAreInertOnceTheGameIsOver(t *testing.T) {
	g := new(interfaces.MockZwickerGame)
	cp := new(presenter.MockZwickerPresenter)
	g.On("GetGameEndFlag").Return(true)
	cp.On("Output", mock.Anything, mock.Anything).Return(zwOut)

	zi := usecase.NewZwickerInteractor(g, cp)
	zi.Take(0, 7, []int{0}, nil)
	zi.Build(0, []int{0}, 9)
	zi.Trail(0)

	g.AssertNotCalled(t, "Take", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	g.AssertNotCalled(t, "Build", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	g.AssertNotCalled(t, "Trail", mock.Anything, mock.Anything)
}

func TestZwickerInteractor_CpuTakesWhenItCan(t *testing.T) {
	g := new(interfaces.MockZwickerGame)
	cp := new(presenter.MockZwickerPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(zwOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.ZwickerPhasePlay).Once()
	g.On("GetPhase").Return(domain.ZwickerPhaseRoundEnd)
	g.On("GetCurrentPlayerIdx").Return(1)
	g.On("ZwickerCpuDecide", 1).Return(domain.ZwickerCpuAction{
		Type: "take", HandIdx: 2, Value: 7, TableIdxs: []int{0},
	})
	g.On("Take", 1, 2, 7, []int{0}, []int(nil)).Return(nil)

	usecase.NewZwickerInteractor(g, cp).Reset()
	g.AssertCalled(t, "Take", 1, 2, 7, []int{0}, []int(nil))
	g.AssertNotCalled(t, "Trail", mock.Anything, mock.Anything)
}

func TestZwickerInteractor_CpuTrailsOtherwise(t *testing.T) {
	g := new(interfaces.MockZwickerGame)
	cp := new(presenter.MockZwickerPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(zwOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.ZwickerPhasePlay).Once()
	g.On("GetPhase").Return(domain.ZwickerPhaseRoundEnd)
	g.On("GetCurrentPlayerIdx").Return(1)
	g.On("ZwickerCpuDecide", 1).Return(domain.ZwickerCpuAction{Type: "trail", HandIdx: 0})
	g.On("Trail", 1, 0).Return(nil)

	usecase.NewZwickerInteractor(g, cp).Reset()
	g.AssertCalled(t, "Trail", 1, 0)
}

func TestZwickerInteractor_CpuLoopStopsAtTheEndOfADeal(t *testing.T) {
	// ディール終了で止めないと、30 点の内訳を読む間もなく次が配られる。
	g := new(interfaces.MockZwickerGame)
	cp := new(presenter.MockZwickerPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(zwOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.ZwickerPhaseRoundEnd)

	usecase.NewZwickerInteractor(g, cp).Reset()
	g.AssertNotCalled(t, "ZwickerCpuDecide", mock.Anything)
	g.AssertNotCalled(t, "NextRound")
}

func TestZwickerInteractor_CpuLoopStopsWhenAMoveIsRejected(t *testing.T) {
	g := new(interfaces.MockZwickerGame)
	cp := new(presenter.MockZwickerPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(zwOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.ZwickerPhasePlay)
	g.On("GetCurrentPlayerIdx").Return(1)
	g.On("ZwickerCpuDecide", 1).Return(domain.ZwickerCpuAction{Type: "trail", HandIdx: 0})
	g.On("Trail", 1, 0).Return(errors.New("illegal"))

	usecase.NewZwickerInteractor(g, cp).Reset()
	g.AssertNumberOfCalls(t, "Trail", 1)
}

func TestZwickerInteractor_CpuLoopStopsWhenThereIsNothingToDo(t *testing.T) {
	g := new(interfaces.MockZwickerGame)
	cp := new(presenter.MockZwickerPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(zwOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.ZwickerPhasePlay)
	g.On("GetCurrentPlayerIdx").Return(1)
	g.On("ZwickerCpuDecide", 1).Return(domain.ZwickerCpuAction{Type: "trail", HandIdx: -1})

	usecase.NewZwickerInteractor(g, cp).Reset()
	g.AssertNotCalled(t, "Trail", mock.Anything, mock.Anything)
}

func TestZwickerInteractor_NextRoundSurfacesItsError(t *testing.T) {
	g := new(interfaces.MockZwickerGame)
	cp := new(presenter.MockZwickerPresenter)
	wantErr := errors.New("the deal is still in progress")
	g.On("NextRound").Return(wantErr)
	cp.On("Output", mock.Anything, wantErr).Return(zwOut)

	usecase.NewZwickerInteractor(g, cp).NextRound()
	cp.AssertCalled(t, "Output", mock.Anything, wantErr)
}

func TestZwickerInteractor_ResetWithConfigAndAccessors(t *testing.T) {
	g, cp := zwMocks()
	g.On("Reset").Return()
	g.On("SetConfig", mock.Anything).Return()
	g.On("GetConfig").Return(domain.DefaultZwickerConfig())
	cp.On("HintOutput", mock.Anything).Return("hint")
	cp.On("ActionLogOutput", mock.Anything).Return("log")

	zi := usecase.NewZwickerInteractor(g, cp)
	assert.NotEmpty(t, zi.ResetWithConfig(domain.DefaultZwickerConfig()))
	assert.Equal(t, domain.DefaultZwickerConfig(), zi.GetConfig())
	assert.Equal(t, "hint", zi.Hint())
	assert.Equal(t, "log", zi.ActionLog())
}

func TestRestoreZwickerInteractor(t *testing.T) {
	g := domain.NewDefaultZwicker()
	g.Reset()
	data, err := g.MarshalJSON()
	require.NoError(t, err)

	zi, err := usecase.RestoreZwickerInteractor(data, new(presenter.MockZwickerPresenter))
	require.NoError(t, err)
	assert.NotNil(t, zi)

	_, err = usecase.RestoreZwickerInteractor([]byte("{"), new(presenter.MockZwickerPresenter))
	assert.Error(t, err)
}
