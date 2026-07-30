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

const sgOut = `{"phase":0}`

func sgMocks() (*interfaces.MockSkitgubbeGame, *presenter.MockSkitgubbePresenter) {
	g := new(interfaces.MockSkitgubbeGame)
	cp := new(presenter.MockSkitgubbePresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(sgOut)
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.SkitgubbePhaseCollect)
	g.On("GetCurrentPlayerIdx").Return(0)
	return g, cp
}

func TestNewSkitgubbeInteractor_NilGuards(t *testing.T) {
	cp := new(presenter.MockSkitgubbePresenter)
	assert.PanicsWithValue(t, "SkitgubbeInteractor: c must not be nil", func() {
		usecase.NewSkitgubbeInteractor(nil, cp)
	})
	assert.PanicsWithValue(t, "SkitgubbeInteractor: cp must not be nil", func() {
		usecase.NewSkitgubbeInteractor(new(interfaces.MockSkitgubbeGame), nil)
	})
}

func TestSkitgubbeInteractor_CommandsUseTheHumanSeat(t *testing.T) {
	g, cp := sgMocks()
	g.On("Reset").Return()
	g.On("PlayCard", 0, 2).Return(nil)
	g.On("PickUp", 0).Return(nil)

	si := usecase.NewSkitgubbeInteractor(g, cp)
	assert.Equal(t, sgOut, si.Reset())
	si.Play(2)
	si.PickUp()

	g.AssertCalled(t, "PlayCard", 0, 2)
	g.AssertCalled(t, "PickUp", 0)
}

func TestSkitgubbeInteractor_SurfacesDomainErrors(t *testing.T) {
	t.Run("play", func(t *testing.T) {
		g := new(interfaces.MockSkitgubbeGame)
		cp := new(presenter.MockSkitgubbePresenter)
		wantErr := errors.New("out of range")
		g.On("GetGameEndFlag").Return(false)
		g.On("PlayCard", 0, 9).Return(wantErr)
		cp.On("Output", mock.Anything, wantErr).Return(sgOut)

		usecase.NewSkitgubbeInteractor(g, cp).Play(9)
		cp.AssertCalled(t, "Output", mock.Anything, wantErr)
	})

	t.Run("pickup", func(t *testing.T) {
		// 出せる札があるのに引き取ろうとしたときの拒否も、そのまま返す。
		g := new(interfaces.MockSkitgubbeGame)
		cp := new(presenter.MockSkitgubbePresenter)
		wantErr := errors.New("you can beat the pile")
		g.On("GetGameEndFlag").Return(false)
		g.On("PickUp", 0).Return(wantErr)
		cp.On("Output", mock.Anything, wantErr).Return(sgOut)

		usecase.NewSkitgubbeInteractor(g, cp).PickUp()
		cp.AssertCalled(t, "Output", mock.Anything, wantErr)
	})
}

func TestSkitgubbeInteractor_CommandsAreInertOnceTheGameIsOver(t *testing.T) {
	g := new(interfaces.MockSkitgubbeGame)
	cp := new(presenter.MockSkitgubbePresenter)
	g.On("GetGameEndFlag").Return(true)
	cp.On("Output", mock.Anything, mock.Anything).Return(sgOut)

	si := usecase.NewSkitgubbeInteractor(g, cp)
	si.Play(0)
	si.PickUp()

	g.AssertNotCalled(t, "PlayCard", mock.Anything, mock.Anything)
	g.AssertNotCalled(t, "PickUp", mock.Anything)
}

func TestSkitgubbeInteractor_CpuPicksThePileUpItself(t *testing.T) {
	// 引き取りは CPU も行う。ここで戻ると、人間が相手の引き取りを代わりに
	// 押すことになる (第2フェーズは引き取りが頻繁に起きる)。
	g := new(interfaces.MockSkitgubbeGame)
	cp := new(presenter.MockSkitgubbePresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(sgOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetCurrentPlayerIdx").Return(1).Once()
	g.On("GetCurrentPlayerIdx").Return(0)
	g.On("SkitgubbeCpuDecide", 1).Return(domain.SkitgubbeCpuAction{PickUp: true, HandIdx: -1})
	g.On("PickUp", 1).Return(nil)

	usecase.NewSkitgubbeInteractor(g, cp).Reset()
	g.AssertCalled(t, "PickUp", 1)
	g.AssertNotCalled(t, "PlayCard", mock.Anything, mock.Anything)
}

func TestSkitgubbeInteractor_CpuLoopStopsWhenAMoveIsRejected(t *testing.T) {
	// この短絡がないと、拒否され続けるドメインが毎リクエストで上限まで回る。
	g := new(interfaces.MockSkitgubbeGame)
	cp := new(presenter.MockSkitgubbePresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(sgOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetCurrentPlayerIdx").Return(1)
	g.On("SkitgubbeCpuDecide", 1).Return(domain.SkitgubbeCpuAction{HandIdx: 0})
	g.On("PlayCard", 1, 0).Return(errors.New("illegal"))

	usecase.NewSkitgubbeInteractor(g, cp).Reset()
	g.AssertNumberOfCalls(t, "PlayCard", 1)
}

func TestSkitgubbeInteractor_CpuLoopStopsAtTheEnd(t *testing.T) {
	g := new(interfaces.MockSkitgubbeGame)
	cp := new(presenter.MockSkitgubbePresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(sgOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(true)

	usecase.NewSkitgubbeInteractor(g, cp).Reset()
	g.AssertNotCalled(t, "SkitgubbeCpuDecide", mock.Anything)
}

func TestSkitgubbeInteractor_ResetWithConfigAndAccessors(t *testing.T) {
	g, cp := sgMocks()
	g.On("Reset").Return()
	g.On("SetConfig", mock.Anything).Return()
	g.On("GetConfig").Return(domain.DefaultSkitgubbeConfig())
	cp.On("HintOutput", mock.Anything).Return("hint")
	cp.On("ActionLogOutput", mock.Anything).Return("log")

	si := usecase.NewSkitgubbeInteractor(g, cp)
	assert.NotEmpty(t, si.ResetWithConfig(domain.DefaultSkitgubbeConfig()))
	assert.Equal(t, domain.DefaultSkitgubbeConfig(), si.GetConfig())
	assert.Equal(t, "hint", si.Hint())
	assert.Equal(t, "log", si.ActionLog())
}

func TestRestoreSkitgubbeInteractor(t *testing.T) {
	g := domain.NewDefaultSkitgubbe()
	g.Reset()
	data, err := g.MarshalJSON()
	require.NoError(t, err)

	si, err := usecase.RestoreSkitgubbeInteractor(data, new(presenter.MockSkitgubbePresenter))
	require.NoError(t, err)
	assert.NotNil(t, si)

	_, err = usecase.RestoreSkitgubbeInteractor([]byte("{"), new(presenter.MockSkitgubbePresenter))
	assert.Error(t, err)
}
