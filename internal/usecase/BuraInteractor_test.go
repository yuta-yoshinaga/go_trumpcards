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

const buraOut = `{"phase":0}`

// buraMocks wires a presenter that always renders and a game whose turn is
// already the human's, so runCpuTurns returns immediately unless a test says
// otherwise.
func buraMocks() (*interfaces.MockBuraGame, *presenter.MockBuraPresenter) {
	gameMock := new(interfaces.MockBuraGame)
	bpMock := new(presenter.MockBuraPresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(buraOut)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetCurrentPlayerIdx").Return(0)
	return gameMock, bpMock
}

func TestNewBuraInteractor_NilGuards(t *testing.T) {
	bpMock := new(presenter.MockBuraPresenter)

	t.Run("panics when b is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "BuraInteractor: b must not be nil", func() {
			usecase.NewBuraInteractor(nil, bpMock)
		})
	})

	t.Run("panics when bp is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "BuraInteractor: bp must not be nil", func() {
			usecase.NewBuraInteractor(new(interfaces.MockBuraGame), nil)
		})
	})
}

func TestBuraInteractor_Reset(t *testing.T) {
	gameMock, bpMock := buraMocks()
	gameMock.On("Reset").Return()

	assert.Equal(t, buraOut, usecase.NewBuraInteractor(gameMock, bpMock).Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestBuraInteractor_PlayForwardsTheIndices(t *testing.T) {
	gameMock, bpMock := buraMocks()
	gameMock.On("PlayCards", 0, []int{0, 1}).Return(nil)

	usecase.NewBuraInteractor(gameMock, bpMock).Play([]int{0, 1})
	// The whole selection has to reach the domain: a lead is up to three cards
	// and dropping any of them changes which trick is being played.
	gameMock.AssertCalled(t, "PlayCards", 0, []int{0, 1})
}

func TestBuraInteractor_PlaySurfacesADomainError(t *testing.T) {
	gameMock := new(interfaces.MockBuraGame)
	bpMock := new(presenter.MockBuraPresenter)
	wantErr := errors.New("card index 99 out of range")
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayCards", 0, []int{99}).Return(wantErr)
	bpMock.On("Output", mock.Anything, wantErr).Return(buraOut)

	assert.Equal(t, buraOut, usecase.NewBuraInteractor(gameMock, bpMock).Play([]int{99}))
	bpMock.AssertCalled(t, "Output", mock.Anything, wantErr)
}

func TestBuraInteractor_ClaimAndDeclareUseTheHumanSeat(t *testing.T) {
	gameMock, bpMock := buraMocks()
	gameMock.On("Claim", 0).Return(nil)
	gameMock.On("DeclareCombination", 0).Return(nil)

	bi := usecase.NewBuraInteractor(gameMock, bpMock)
	bi.Claim()
	bi.Declare()

	gameMock.AssertCalled(t, "Claim", 0)
	gameMock.AssertCalled(t, "DeclareCombination", 0)
}

func TestBuraInteractor_DeclareSurfacesTheRejection(t *testing.T) {
	gameMock := new(interfaces.MockBuraGame)
	bpMock := new(presenter.MockBuraPresenter)
	wantErr := errors.New("no winning combination in hand")
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("DeclareCombination", 0).Return(wantErr)
	bpMock.On("Output", mock.Anything, wantErr).Return(buraOut)

	usecase.NewBuraInteractor(gameMock, bpMock).Declare()
	bpMock.AssertCalled(t, "Output", mock.Anything, wantErr)
}

func TestBuraInteractor_CommandsAreInertOnceTheRoundIsOver(t *testing.T) {
	gameMock := new(interfaces.MockBuraGame)
	bpMock := new(presenter.MockBuraPresenter)
	gameMock.On("GetGameEndFlag").Return(true)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(buraOut)

	bi := usecase.NewBuraInteractor(gameMock, bpMock)
	bi.Play([]int{0})
	bi.Claim()
	bi.Declare()

	gameMock.AssertNotCalled(t, "PlayCards", mock.Anything, mock.Anything)
	gameMock.AssertNotCalled(t, "Claim", mock.Anything)
	gameMock.AssertNotCalled(t, "DeclareCombination", mock.Anything)
}

func TestBuraInteractor_RunsCpuTurnsUntilItIsTheHumansTurn(t *testing.T) {
	gameMock := new(interfaces.MockBuraGame)
	bpMock := new(presenter.MockBuraPresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(buraOut)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)

	// Seat 1 acts twice, then control returns to the human. Sequenced with
	// Once() rather than a function return: testify hands Arguments.Int a raw
	// value, so returning a closure panics rather than being called.
	gameMock.On("GetCurrentPlayerIdx").Return(1).Once()
	gameMock.On("GetCurrentPlayerIdx").Return(1).Once()
	gameMock.On("GetCurrentPlayerIdx").Return(0)
	gameMock.On("BuraCpuDecide", 1).Return(domain.BuraCpuAction{Indices: []int{0}})
	gameMock.On("PlayCards", 1, []int{0}).Return(nil)

	usecase.NewBuraInteractor(gameMock, bpMock).Reset()
	gameMock.AssertNumberOfCalls(t, "PlayCards", 2)
}

func TestBuraInteractor_CpuLoopHonoursDeclareAndClaim(t *testing.T) {
	for _, tc := range []struct {
		name   string
		action domain.BuraCpuAction
		method string
	}{
		{"declare", domain.BuraCpuAction{Declare: true}, "DeclareCombination"},
		{"claim", domain.BuraCpuAction{Claim: true}, "Claim"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gameMock := new(interfaces.MockBuraGame)
			bpMock := new(presenter.MockBuraPresenter)
			bpMock.On("Output", mock.Anything, mock.Anything).Return(buraOut)
			gameMock.On("Reset").Return()
			gameMock.On("GetGameEndFlag").Return(false)
			gameMock.On("GetCurrentPlayerIdx").Return(1)
			gameMock.On("BuraCpuDecide", 1).Return(tc.action)
			// Returning an error stops the loop, which is what keeps this test
			// from spinning to the turn cap.
			gameMock.On(tc.method, 1).Return(errors.New("stop"))

			usecase.NewBuraInteractor(gameMock, bpMock).Reset()
			gameMock.AssertCalled(t, tc.method, 1)
		})
	}
}

func TestBuraInteractor_CpuLoopIsBounded(t *testing.T) {
	// A domain that never yields the turn must not hang the request. The cap
	// exists because a hung Worker request is worse than a truncated one.
	gameMock := new(interfaces.MockBuraGame)
	bpMock := new(presenter.MockBuraPresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(buraOut)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetCurrentPlayerIdx").Return(1)
	gameMock.On("BuraCpuDecide", 1).Return(domain.BuraCpuAction{Indices: []int{0}})
	gameMock.On("PlayCards", 1, []int{0}).Return(nil)

	usecase.NewBuraInteractor(gameMock, bpMock).Reset()
	calls := 0
	for _, c := range gameMock.Calls {
		if c.Method == "PlayCards" {
			calls++
		}
	}
	assert.Less(t, calls, 100, "the CPU loop must terminate on its own")
}

func TestBuraInteractor_ResetWithConfigValidates(t *testing.T) {
	gameMock, bpMock := buraMocks()
	gameMock.On("Reset").Return()
	gameMock.On("SetConfig", mock.Anything).Return()
	gameMock.On("GetConfig").Return(domain.DefaultBuraConfig())

	bi := usecase.NewBuraInteractor(gameMock, bpMock)
	assert.NotEmpty(t, bi.ResetWithConfig(domain.DefaultBuraConfig()))
	assert.Equal(t, domain.DefaultBuraConfig(), bi.GetConfig())
}

func TestBuraInteractor_HintAndActionLogDelegate(t *testing.T) {
	gameMock := new(interfaces.MockBuraGame)
	bpMock := new(presenter.MockBuraPresenter)
	bpMock.On("HintOutput", mock.Anything).Return("hint")
	bpMock.On("ActionLogOutput", mock.Anything).Return("log")

	bi := usecase.NewBuraInteractor(gameMock, bpMock)
	assert.Equal(t, "hint", bi.Hint())
	assert.Equal(t, "log", bi.ActionLog())
}

func TestRestoreBuraInteractor(t *testing.T) {
	g := domain.NewDefaultBura()
	g.Reset()
	data, err := g.MarshalJSON()
	require.NoError(t, err)

	bi, err := usecase.RestoreBuraInteractor(data, new(presenter.MockBuraPresenter))
	require.NoError(t, err)
	assert.NotNil(t, bi)

	_, err = usecase.RestoreBuraInteractor([]byte("{"), new(presenter.MockBuraPresenter))
	assert.Error(t, err)
}

func TestBuraInteractor_CpuLoopStopsWhenAPlayIsRejected(t *testing.T) {
	// The Declare and Claim branches' error paths are covered above; this is
	// the default branch, which is the one that actually runs on almost every
	// turn. Without the short-circuit a domain that keeps rejecting the CPU's
	// choice would spin to buraCpuTurnCap on every request.
	gameMock := new(interfaces.MockBuraGame)
	bpMock := new(presenter.MockBuraPresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(buraOut)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetCurrentPlayerIdx").Return(1)
	gameMock.On("BuraCpuDecide", 1).Return(domain.BuraCpuAction{Indices: []int{0}})
	gameMock.On("PlayCards", 1, []int{0}).Return(errors.New("illegal"))

	usecase.NewBuraInteractor(gameMock, bpMock).Reset()

	gameMock.AssertNumberOfCalls(t, "PlayCards", 1)
}
