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

const txOut = `{"phase":0}`

func txMocks() (*interfaces.MockTrexGame, *presenter.MockTrexPresenter) {
	g := new(interfaces.MockTrexGame)
	cp := new(presenter.MockTrexPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(txOut)
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.TrexPhasePlay)
	g.On("GetCurrentPlayerIdx").Return(0)
	return g, cp
}

func TestNewTrexInteractor_NilGuards(t *testing.T) {
	cp := new(presenter.MockTrexPresenter)
	assert.PanicsWithValue(t, "TrexInteractor: c must not be nil", func() {
		usecase.NewTrexInteractor(nil, cp)
	})
	assert.PanicsWithValue(t, "TrexInteractor: cp must not be nil", func() {
		usecase.NewTrexInteractor(new(interfaces.MockTrexGame), nil)
	})
}

func TestTrexInteractor_CommandsUseTheHumanSeat(t *testing.T) {
	g, cp := txMocks()
	g.On("Reset").Return()
	g.On("ChooseContract", 0, domain.TrexContractQueens).Return(nil)
	g.On("PlayCard", 0, 3).Return(nil)
	g.On("Pass", 0).Return(nil)

	ti := usecase.NewTrexInteractor(g, cp)
	assert.Equal(t, txOut, ti.Reset())
	ti.Choose(int(domain.TrexContractQueens))
	ti.Play(3)
	ti.Pass()

	g.AssertCalled(t, "ChooseContract", 0, domain.TrexContractQueens)
	g.AssertCalled(t, "PlayCard", 0, 3)
	g.AssertCalled(t, "Pass", 0)
}

func TestTrexInteractor_SurfacesDomainErrors(t *testing.T) {
	g := new(interfaces.MockTrexGame)
	cp := new(presenter.MockTrexPresenter)
	wantErr := errors.New("that contract has already been played in this kingdom")
	g.On("GetGameEndFlag").Return(false)
	g.On("ChooseContract", 0, domain.TrexContractQueens).Return(wantErr)
	cp.On("Output", mock.Anything, wantErr).Return(txOut)

	usecase.NewTrexInteractor(g, cp).Choose(int(domain.TrexContractQueens))
	cp.AssertCalled(t, "Output", mock.Anything, wantErr)
}

func TestTrexInteractor_CommandsAreInertOnceTheGameIsOver(t *testing.T) {
	g := new(interfaces.MockTrexGame)
	cp := new(presenter.MockTrexPresenter)
	g.On("GetGameEndFlag").Return(true)
	cp.On("Output", mock.Anything, mock.Anything).Return(txOut)

	ti := usecase.NewTrexInteractor(g, cp)
	ti.Choose(0)
	ti.Play(0)
	ti.Pass()

	g.AssertNotCalled(t, "ChooseContract", mock.Anything, mock.Anything)
	g.AssertNotCalled(t, "PlayCard", mock.Anything, mock.Anything)
	g.AssertNotCalled(t, "Pass", mock.Anything)
}

func TestTrexInteractor_TheCpuNeverChoosesForTheHumanKing(t *testing.T) {
	// ここで CPU が選んでしまうと、このゲームの肝である「王が契約を選ぶ」が
	// 人間の手から消える。
	g := new(interfaces.MockTrexGame)
	cp := new(presenter.MockTrexPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(txOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.TrexPhaseChoose)
	g.On("GetKingIdx").Return(0)

	usecase.NewTrexInteractor(g, cp).Reset()
	g.AssertNotCalled(t, "ChooseContract", mock.Anything, mock.Anything)
}

func TestTrexInteractor_TheCpuChoosesWhenItIsTheKing(t *testing.T) {
	g := new(interfaces.MockTrexGame)
	cp := new(presenter.MockTrexPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(txOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.TrexPhaseChoose).Once()
	g.On("GetPhase").Return(domain.TrexPhaseDealEnd)
	g.On("GetKingIdx").Return(2)
	g.On("TrexCpuDecide", 2).Return(domain.TrexCpuAction{Contract: domain.TrexContractTricks, HandIdx: -1})
	g.On("ChooseContract", 2, domain.TrexContractTricks).Return(nil)

	usecase.NewTrexInteractor(g, cp).Reset()
	g.AssertCalled(t, "ChooseContract", 2, domain.TrexContractTricks)
}

func TestTrexInteractor_TheCpuPassesInTheDominoes(t *testing.T) {
	// 詰まった CPU がパスできないと、ドミノが誰の手番でも進まなくなる。
	g := new(interfaces.MockTrexGame)
	cp := new(presenter.MockTrexPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(txOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.TrexPhasePlay)
	g.On("GetCurrentPlayerIdx").Return(1).Once()
	g.On("GetCurrentPlayerIdx").Return(0)
	g.On("TrexCpuDecide", 1).Return(domain.TrexCpuAction{HandIdx: -1, Pass: true})
	g.On("Pass", 1).Return(nil)

	usecase.NewTrexInteractor(g, cp).Reset()
	g.AssertCalled(t, "Pass", 1)
	g.AssertNotCalled(t, "PlayCard", mock.Anything, mock.Anything)
}

func TestTrexInteractor_CpuLoopStopsAtTheEndOfADeal(t *testing.T) {
	// ディール終了で止めないと、精算を読む間もなく次が配られる。
	g := new(interfaces.MockTrexGame)
	cp := new(presenter.MockTrexPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(txOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.TrexPhaseDealEnd)

	usecase.NewTrexInteractor(g, cp).Reset()
	g.AssertNotCalled(t, "TrexCpuDecide", mock.Anything)
	g.AssertNotCalled(t, "NextDeal")
}

func TestTrexInteractor_CpuLoopStopsWhenAMoveIsRejected(t *testing.T) {
	g := new(interfaces.MockTrexGame)
	cp := new(presenter.MockTrexPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(txOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.TrexPhasePlay)
	g.On("GetCurrentPlayerIdx").Return(1)
	g.On("TrexCpuDecide", 1).Return(domain.TrexCpuAction{HandIdx: 0})
	g.On("PlayCard", 1, 0).Return(errors.New("illegal"))

	usecase.NewTrexInteractor(g, cp).Reset()
	g.AssertNumberOfCalls(t, "PlayCard", 1)
}

func TestTrexInteractor_NextDealSurfacesItsError(t *testing.T) {
	g := new(interfaces.MockTrexGame)
	cp := new(presenter.MockTrexPresenter)
	wantErr := errors.New("the deal is still in progress")
	g.On("NextDeal").Return(wantErr)
	cp.On("Output", mock.Anything, wantErr).Return(txOut)

	usecase.NewTrexInteractor(g, cp).NextDeal()
	cp.AssertCalled(t, "Output", mock.Anything, wantErr)
}

func TestTrexInteractor_ResetWithConfigAndAccessors(t *testing.T) {
	g, cp := txMocks()
	g.On("Reset").Return()
	g.On("SetConfig", mock.Anything).Return()
	g.On("GetConfig").Return(domain.DefaultTrexConfig())
	cp.On("HintOutput", mock.Anything).Return("hint")
	cp.On("ActionLogOutput", mock.Anything).Return("log")

	ti := usecase.NewTrexInteractor(g, cp)
	assert.NotEmpty(t, ti.ResetWithConfig(domain.DefaultTrexConfig()))
	assert.Equal(t, domain.DefaultTrexConfig(), ti.GetConfig())
	assert.Equal(t, "hint", ti.Hint())
	assert.Equal(t, "log", ti.ActionLog())
}

func TestRestoreTrexInteractor(t *testing.T) {
	g := domain.NewDefaultTrex()
	g.Reset()
	data, err := g.MarshalJSON()
	require.NoError(t, err)

	ti, err := usecase.RestoreTrexInteractor(data, new(presenter.MockTrexPresenter))
	require.NoError(t, err)
	assert.NotNil(t, ti)

	_, err = usecase.RestoreTrexInteractor([]byte("{"), new(presenter.MockTrexPresenter))
	assert.Error(t, err)
}
