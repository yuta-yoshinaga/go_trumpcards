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

const sjOut = `{"phase":0}`

func sjMocks() (*interfaces.MockSjavsGame, *presenter.MockSjavsPresenter) {
	g := new(interfaces.MockSjavsGame)
	cp := new(presenter.MockSjavsPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(sjOut)
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.SjavsPhasePlay)
	g.On("GetCurrentPlayerIdx").Return(0)
	return g, cp
}

func TestNewSjavsInteractor_NilGuards(t *testing.T) {
	cp := new(presenter.MockSjavsPresenter)
	assert.PanicsWithValue(t, "SjavsInteractor: c must not be nil", func() {
		usecase.NewSjavsInteractor(nil, cp)
	})
	assert.PanicsWithValue(t, "SjavsInteractor: cp must not be nil", func() {
		usecase.NewSjavsInteractor(new(interfaces.MockSjavsGame), nil)
	})
}

func TestSjavsInteractor_CommandsUseTheHumanSeat(t *testing.T) {
	g, cp := sjMocks()
	g.On("Reset").Return()
	g.On("Bid", 0, 6).Return(nil)
	g.On("PlayCard", 0, 3).Return(nil)

	si := usecase.NewSjavsInteractor(g, cp)
	assert.Equal(t, sjOut, si.Reset())
	si.Bid(6)
	si.Play(3)

	g.AssertCalled(t, "Bid", 0, 6)
	g.AssertCalled(t, "PlayCard", 0, 3)
}

func TestSjavsInteractor_SurfacesDomainErrors(t *testing.T) {
	for name, tc := range map[string]struct {
		method string
		args   []any
		call   func(*usecase.SjavsInteractor)
	}{
		"bid":  {"Bid", []any{0, 4}, func(si *usecase.SjavsInteractor) { si.Bid(4) }},
		"play": {"PlayCard", []any{0, 9}, func(si *usecase.SjavsInteractor) { si.Play(9) }},
	} {
		t.Run(name, func(t *testing.T) {
			g := new(interfaces.MockSjavsGame)
			cp := new(presenter.MockSjavsPresenter)
			wantErr := errors.New("a bid must be at least 5 cards")
			g.On("GetGameEndFlag").Return(false)
			g.On(tc.method, tc.args...).Return(wantErr)
			cp.On("Output", mock.Anything, wantErr).Return(sjOut)

			tc.call(usecase.NewSjavsInteractor(g, cp))
			cp.AssertCalled(t, "Output", mock.Anything, wantErr)
		})
	}
}

func TestSjavsInteractor_CommandsAreInertOnceTheRubberIsOver(t *testing.T) {
	g := new(interfaces.MockSjavsGame)
	cp := new(presenter.MockSjavsPresenter)
	g.On("GetGameEndFlag").Return(true)
	cp.On("Output", mock.Anything, mock.Anything).Return(sjOut)

	si := usecase.NewSjavsInteractor(g, cp)
	si.Bid(6)
	si.Play(0)

	g.AssertNotCalled(t, "Bid", mock.Anything, mock.Anything)
	g.AssertNotCalled(t, "PlayCard", mock.Anything, mock.Anything)
}

func TestSjavsInteractor_CpuBidsDuringTheBiddingPhase(t *testing.T) {
	// ビッドとプレイで送るコマンドが違う。フェーズを見ずに PlayCard を送ると、
	// ビッド中に「手番ではない」で止まる。
	g := new(interfaces.MockSjavsGame)
	cp := new(presenter.MockSjavsPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(sjOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.SjavsPhaseBid)
	g.On("GetCurrentPlayerIdx").Return(1).Once()
	g.On("GetCurrentPlayerIdx").Return(0)
	g.On("SjavsCpuDecide", 1).Return(domain.SjavsCpuAction{BidLength: 6, HandIdx: -1})
	g.On("Bid", 1, 6).Return(nil)

	usecase.NewSjavsInteractor(g, cp).Reset()
	g.AssertCalled(t, "Bid", 1, 6)
	g.AssertNotCalled(t, "PlayCard", mock.Anything, mock.Anything)
}

func TestSjavsInteractor_CpuLoopStopsAtTheEndOfAHand(t *testing.T) {
	// ハンド終了で止めないと、精算を読む間もなく次のハンドが配られる。
	g := new(interfaces.MockSjavsGame)
	cp := new(presenter.MockSjavsPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(sjOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.SjavsPhaseHandEnd)

	usecase.NewSjavsInteractor(g, cp).Reset()
	g.AssertNotCalled(t, "SjavsCpuDecide", mock.Anything)
	g.AssertNotCalled(t, "NextHand")
}

func TestSjavsInteractor_CpuLoopStopsWhenAMoveIsRejected(t *testing.T) {
	g := new(interfaces.MockSjavsGame)
	cp := new(presenter.MockSjavsPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(sjOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.SjavsPhasePlay)
	g.On("GetCurrentPlayerIdx").Return(1)
	g.On("SjavsCpuDecide", 1).Return(domain.SjavsCpuAction{HandIdx: 0})
	g.On("PlayCard", 1, 0).Return(errors.New("illegal"))

	usecase.NewSjavsInteractor(g, cp).Reset()
	g.AssertNumberOfCalls(t, "PlayCard", 1)
}

func TestSjavsInteractor_NextHandSurfacesItsError(t *testing.T) {
	g := new(interfaces.MockSjavsGame)
	cp := new(presenter.MockSjavsPresenter)
	wantErr := errors.New("the hand is still in progress")
	g.On("NextHand").Return(wantErr)
	cp.On("Output", mock.Anything, wantErr).Return(sjOut)

	usecase.NewSjavsInteractor(g, cp).NextHand()
	cp.AssertCalled(t, "Output", mock.Anything, wantErr)
}

func TestSjavsInteractor_ResetWithConfigAndAccessors(t *testing.T) {
	g, cp := sjMocks()
	g.On("Reset").Return()
	g.On("SetConfig", mock.Anything).Return()
	g.On("GetConfig").Return(domain.DefaultSjavsConfig())
	cp.On("HintOutput", mock.Anything).Return("hint")
	cp.On("ActionLogOutput", mock.Anything).Return("log")

	si := usecase.NewSjavsInteractor(g, cp)
	assert.NotEmpty(t, si.ResetWithConfig(domain.DefaultSjavsConfig()))
	assert.Equal(t, domain.DefaultSjavsConfig(), si.GetConfig())
	assert.Equal(t, "hint", si.Hint())
	assert.Equal(t, "log", si.ActionLog())
}

func TestRestoreSjavsInteractor(t *testing.T) {
	g := domain.NewDefaultSjavs()
	g.Reset()
	data, err := g.MarshalJSON()
	require.NoError(t, err)

	si, err := usecase.RestoreSjavsInteractor(data, new(presenter.MockSjavsPresenter))
	require.NoError(t, err)
	assert.NotNil(t, si)

	_, err = usecase.RestoreSjavsInteractor([]byte("{"), new(presenter.MockSjavsPresenter))
	assert.Error(t, err)
}
