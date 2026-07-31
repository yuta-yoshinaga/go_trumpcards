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

const pcOut = `{"phase":0}`

func pcMocks() (*interfaces.MockPochGame, *presenter.MockPochPresenter) {
	g := new(interfaces.MockPochGame)
	cp := new(presenter.MockPochPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(pcOut)
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.PochPhasePochen)
	g.On("GetCurrentPlayerIdx").Return(0)
	g.On("GetPlayer", 0).Return(domain.NewPochPlayer(true))
	return g, cp
}

func TestNewPochInteractor_NilGuards(t *testing.T) {
	cp := new(presenter.MockPochPresenter)
	assert.PanicsWithValue(t, "PochInteractor: c must not be nil", func() {
		usecase.NewPochInteractor(nil, cp)
	})
	assert.PanicsWithValue(t, "PochInteractor: cp must not be nil", func() {
		usecase.NewPochInteractor(new(interfaces.MockPochGame), nil)
	})
}

func TestPochInteractor_CommandsUseTheHumanSeat(t *testing.T) {
	g, cp := pcMocks()
	g.On("Reset").Return()
	g.On("Bet", 0).Return(nil)
	g.On("Fold", 0).Return(nil)
	g.On("Play", 0, 3).Return(nil)

	pi := usecase.NewPochInteractor(g, cp)
	assert.Equal(t, pcOut, pi.Reset())
	pi.Bet()
	pi.Fold()
	pi.Play(3)

	g.AssertCalled(t, "Bet", 0)
	g.AssertCalled(t, "Fold", 0)
	g.AssertCalled(t, "Play", 0, 3)
}

func TestPochInteractor_SurfacesDomainErrors(t *testing.T) {
	g := new(interfaces.MockPochGame)
	cp := new(presenter.MockPochPresenter)
	wantErr := errors.New("you must play the next higher card of the same suit")
	g.On("GetGameEndFlag").Return(false)
	g.On("Play", 0, 1).Return(wantErr)
	cp.On("Output", mock.Anything, wantErr).Return(pcOut)

	usecase.NewPochInteractor(g, cp).Play(1)
	cp.AssertCalled(t, "Output", mock.Anything, wantErr)
}

func TestPochInteractor_CommandsAreInertOnceTheGameIsOver(t *testing.T) {
	g := new(interfaces.MockPochGame)
	cp := new(presenter.MockPochPresenter)
	g.On("GetGameEndFlag").Return(true)
	cp.On("Output", mock.Anything, mock.Anything).Return(pcOut)

	pi := usecase.NewPochInteractor(g, cp)
	pi.Bet()
	pi.Fold()
	pi.Play(0)

	g.AssertNotCalled(t, "Bet", mock.Anything)
	g.AssertNotCalled(t, "Fold", mock.Anything)
	g.AssertNotCalled(t, "Play", mock.Anything, mock.Anything)
}

func TestPochInteractor_CpuBetsAndFolds(t *testing.T) {
	for name, tc := range map[string]struct {
		action domain.PochCpuAction
		call   string
	}{
		"bets":  {domain.PochCpuAction{Type: "bet", HandIdx: -1}, "Bet"},
		"folds": {domain.PochCpuAction{Type: "fold", HandIdx: -1}, "Fold"},
	} {
		t.Run(name, func(t *testing.T) {
			g := new(interfaces.MockPochGame)
			cp := new(presenter.MockPochPresenter)
			cp.On("Output", mock.Anything, mock.Anything).Return(pcOut)
			g.On("Reset").Return()
			g.On("GetGameEndFlag").Return(false)
			g.On("GetPhase").Return(domain.PochPhasePochen).Once()
			g.On("GetPhase").Return(domain.PochPhaseDealEnd)
			g.On("GetCurrentPlayerIdx").Return(1)
			g.On("PochCpuDecide", 1).Return(tc.action)
			g.On(tc.call, 1).Return(nil)

			usecase.NewPochInteractor(g, cp).Reset()
			g.AssertCalled(t, tc.call, 1)
		})
	}
}

func TestPochInteractor_CpuPlays(t *testing.T) {
	g := new(interfaces.MockPochGame)
	cp := new(presenter.MockPochPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(pcOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.PochPhaseStops).Once()
	g.On("GetPhase").Return(domain.PochPhaseDealEnd)
	g.On("GetCurrentPlayerIdx").Return(2)
	g.On("PochCpuDecide", 2).Return(domain.PochCpuAction{Type: "play", HandIdx: 1})
	g.On("Play", 2, 1).Return(nil)

	usecase.NewPochInteractor(g, cp).Reset()
	g.AssertCalled(t, "Play", 2, 1)
}

// TestPochInteractor_CpuKeepsGoingWhileTheHumanHasFolded は、人間が pochen で
// 降りていても CPU 同士で決着させることを確かめる。止めるとストップスに
// 進めないまま固まる。
func TestPochInteractor_CpuKeepsGoingWhileTheHumanHasFolded(t *testing.T) {
	folded := domain.NewPochPlayer(true)
	folded.Fold()

	g := new(interfaces.MockPochGame)
	cp := new(presenter.MockPochPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(pcOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.PochPhasePochen).Once()
	g.On("GetPhase").Return(domain.PochPhaseDealEnd)
	g.On("GetCurrentPlayerIdx").Return(0)
	g.On("GetPlayer", 0).Return(folded)
	g.On("PochCpuDecide", 0).Return(domain.PochCpuAction{Type: "fold", HandIdx: -1})
	g.On("Fold", 0).Return(nil)

	usecase.NewPochInteractor(g, cp).Reset()
	g.AssertCalled(t, "Fold", 0)
}

// TestPochInteractor_CpuStopsAtTheHumansTurn は逆方向 -- 降りていなければ
// 人間の手番で必ず止まる。
func TestPochInteractor_CpuStopsAtTheHumansTurn(t *testing.T) {
	g, cp := pcMocks()
	g.On("Reset").Return()
	g.On("PochCpuDecide", mock.Anything).Return(domain.PochCpuAction{Type: "fold", HandIdx: -1})

	usecase.NewPochInteractor(g, cp).Reset()
	g.AssertNotCalled(t, "Fold", mock.Anything)
	g.AssertNotCalled(t, "Bet", mock.Anything)
}

func TestPochInteractor_CpuLoopStopsAtTheEndOfADeal(t *testing.T) {
	// ディール終了で止めないと、9 区画の精算を読む間もなく次が配られる。
	g := new(interfaces.MockPochGame)
	cp := new(presenter.MockPochPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(pcOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.PochPhaseDealEnd)

	usecase.NewPochInteractor(g, cp).Reset()
	g.AssertNotCalled(t, "PochCpuDecide", mock.Anything)
	g.AssertNotCalled(t, "NextDeal")
}

func TestPochInteractor_CpuLoopStopsWhenAMoveIsRejected(t *testing.T) {
	g := new(interfaces.MockPochGame)
	cp := new(presenter.MockPochPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(pcOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.PochPhaseStops)
	g.On("GetCurrentPlayerIdx").Return(1)
	g.On("PochCpuDecide", 1).Return(domain.PochCpuAction{Type: "play", HandIdx: 0})
	g.On("Play", 1, 0).Return(errors.New("illegal"))

	usecase.NewPochInteractor(g, cp).Reset()
	g.AssertNumberOfCalls(t, "Play", 1)
}

func TestPochInteractor_CpuLoopStopsWhenThereIsNothingToPlay(t *testing.T) {
	g := new(interfaces.MockPochGame)
	cp := new(presenter.MockPochPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(pcOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.PochPhaseStops)
	g.On("GetCurrentPlayerIdx").Return(1)
	g.On("PochCpuDecide", 1).Return(domain.PochCpuAction{Type: "play", HandIdx: -1})

	usecase.NewPochInteractor(g, cp).Reset()
	g.AssertNotCalled(t, "Play", mock.Anything, mock.Anything)
}

func TestPochInteractor_NextDealSurfacesItsError(t *testing.T) {
	g := new(interfaces.MockPochGame)
	cp := new(presenter.MockPochPresenter)
	wantErr := errors.New("the deal is still in progress")
	g.On("NextDeal").Return(wantErr)
	cp.On("Output", mock.Anything, wantErr).Return(pcOut)

	usecase.NewPochInteractor(g, cp).NextDeal()
	cp.AssertCalled(t, "Output", mock.Anything, wantErr)
}

func TestPochInteractor_ResetWithConfigAndAccessors(t *testing.T) {
	g, cp := pcMocks()
	g.On("Reset").Return()
	g.On("SetConfig", mock.Anything).Return()
	g.On("GetConfig").Return(domain.DefaultPochConfig())
	cp.On("HintOutput", mock.Anything).Return("hint")
	cp.On("ActionLogOutput", mock.Anything).Return("log")

	pi := usecase.NewPochInteractor(g, cp)
	assert.NotEmpty(t, pi.ResetWithConfig(domain.DefaultPochConfig()))
	assert.Equal(t, domain.DefaultPochConfig(), pi.GetConfig())
	assert.Equal(t, "hint", pi.Hint())
	assert.Equal(t, "log", pi.ActionLog())
}

func TestRestorePochInteractor(t *testing.T) {
	g := domain.NewDefaultPoch()
	g.Reset()
	data, err := g.MarshalJSON()
	require.NoError(t, err)

	pi, err := usecase.RestorePochInteractor(data, new(presenter.MockPochPresenter))
	require.NoError(t, err)
	assert.NotNil(t, pi)

	_, err = usecase.RestorePochInteractor([]byte("{"), new(presenter.MockPochPresenter))
	assert.Error(t, err)
}
