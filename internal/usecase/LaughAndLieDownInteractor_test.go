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

const lldOut = `{"phase":0}`

func lldMocks() (*interfaces.MockLaughAndLieDownGame, *presenter.MockLaughAndLieDownPresenter) {
	g := new(interfaces.MockLaughAndLieDownGame)
	cp := new(presenter.MockLaughAndLieDownPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(lldOut)
	g.On("GetGameEndFlag").Return(false)
	g.On("GetCurrentPlayerIdx").Return(0)
	return g, cp
}

func TestNewLaughAndLieDownInteractor_NilGuards(t *testing.T) {
	cp := new(presenter.MockLaughAndLieDownPresenter)
	assert.PanicsWithValue(t, "LaughAndLieDownInteractor: c must not be nil", func() {
		usecase.NewLaughAndLieDownInteractor(nil, cp)
	})
	assert.PanicsWithValue(t, "LaughAndLieDownInteractor: cp must not be nil", func() {
		usecase.NewLaughAndLieDownInteractor(new(interfaces.MockLaughAndLieDownGame), nil)
	})
}

func TestLaughAndLieDownInteractor_CarriesTheTakeCountThroughToTheDomain(t *testing.T) {
	// 1 枚取りと 3 枚取りは別の手なので、取得枚数を落とすと必ず 1 枚取りになる。
	g, cp := lldMocks()
	g.On("Reset").Return()
	g.On("PlayCard", 0, 2, 1).Return(nil)
	g.On("PlayCard", 0, 4, 3).Return(nil)

	li := usecase.NewLaughAndLieDownInteractor(g, cp)
	assert.Equal(t, lldOut, li.Reset())
	li.Play(2, 1)
	li.Play(4, 3)

	g.AssertCalled(t, "PlayCard", 0, 2, 1)
	g.AssertCalled(t, "PlayCard", 0, 4, 3)
}

func TestLaughAndLieDownInteractor_SurfacesDomainErrors(t *testing.T) {
	g := new(interfaces.MockLaughAndLieDownGame)
	cp := new(presenter.MockLaughAndLieDownPresenter)
	wantErr := errors.New("only 2 card(s) of that rank are on the table")
	g.On("GetGameEndFlag").Return(false)
	g.On("PlayCard", 0, 1, 3).Return(wantErr)
	cp.On("Output", mock.Anything, wantErr).Return(lldOut)

	usecase.NewLaughAndLieDownInteractor(g, cp).Play(1, 3)
	cp.AssertCalled(t, "Output", mock.Anything, wantErr)
}

func TestLaughAndLieDownInteractor_CommandsAreInertOnceTheGameIsOver(t *testing.T) {
	g := new(interfaces.MockLaughAndLieDownGame)
	cp := new(presenter.MockLaughAndLieDownPresenter)
	g.On("GetGameEndFlag").Return(true)
	cp.On("Output", mock.Anything, mock.Anything).Return(lldOut)

	usecase.NewLaughAndLieDownInteractor(g, cp).Play(0, 1)
	g.AssertNotCalled(t, "PlayCard", mock.Anything, mock.Anything, mock.Anything)
}

func TestLaughAndLieDownInteractor_CpuLoopStopsWhenTheCpuHasNoCapture(t *testing.T) {
	// 取れない CPU を降ろすのはドメインの手番送りの仕事。ここで無理に
	// PlayCard を投げると、-1 の添字でエラーを踏み続ける。
	g := new(interfaces.MockLaughAndLieDownGame)
	cp := new(presenter.MockLaughAndLieDownPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(lldOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetCurrentPlayerIdx").Return(1)
	g.On("LaughAndLieDownCpuDecide", 1).Return(domain.LaughAndLieDownCpuAction{HandIdx: -1, TakeCount: 1})

	usecase.NewLaughAndLieDownInteractor(g, cp).Reset()
	g.AssertNotCalled(t, "PlayCard", mock.Anything, mock.Anything, mock.Anything)
}

func TestLaughAndLieDownInteractor_CpuLoopStopsWhenAMoveIsRejected(t *testing.T) {
	g := new(interfaces.MockLaughAndLieDownGame)
	cp := new(presenter.MockLaughAndLieDownPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(lldOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetCurrentPlayerIdx").Return(1)
	g.On("LaughAndLieDownCpuDecide", 1).Return(domain.LaughAndLieDownCpuAction{HandIdx: 0, TakeCount: 1})
	g.On("PlayCard", 1, 0, 1).Return(errors.New("illegal"))

	usecase.NewLaughAndLieDownInteractor(g, cp).Reset()
	g.AssertNumberOfCalls(t, "PlayCard", 1)
}

func TestLaughAndLieDownInteractor_CpuLoopStopsAtTheEnd(t *testing.T) {
	g := new(interfaces.MockLaughAndLieDownGame)
	cp := new(presenter.MockLaughAndLieDownPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(lldOut)
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(true)

	usecase.NewLaughAndLieDownInteractor(g, cp).Reset()
	g.AssertNotCalled(t, "LaughAndLieDownCpuDecide", mock.Anything)
}

func TestLaughAndLieDownInteractor_ResetWithConfigAndAccessors(t *testing.T) {
	g, cp := lldMocks()
	g.On("Reset").Return()
	g.On("SetConfig", mock.Anything).Return()
	g.On("GetConfig").Return(domain.DefaultLaughAndLieDownConfig())
	cp.On("HintOutput", mock.Anything).Return("hint")
	cp.On("ActionLogOutput", mock.Anything).Return("log")

	li := usecase.NewLaughAndLieDownInteractor(g, cp)
	assert.NotEmpty(t, li.ResetWithConfig(domain.DefaultLaughAndLieDownConfig()))
	assert.Equal(t, domain.DefaultLaughAndLieDownConfig(), li.GetConfig())
	assert.Equal(t, "hint", li.Hint())
	assert.Equal(t, "log", li.ActionLog())
}

func TestRestoreLaughAndLieDownInteractor(t *testing.T) {
	g := domain.NewDefaultLaughAndLieDown()
	g.Reset()
	data, err := g.MarshalJSON()
	require.NoError(t, err)

	li, err := usecase.RestoreLaughAndLieDownInteractor(data, new(presenter.MockLaughAndLieDownPresenter))
	require.NoError(t, err)
	assert.NotNil(t, li)

	_, err = usecase.RestoreLaughAndLieDownInteractor([]byte("{"), new(presenter.MockLaughAndLieDownPresenter))
	assert.Error(t, err)
}
