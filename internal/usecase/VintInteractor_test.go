//go:build test

package usecase_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

const vintMockOutput = `{"phase":0}`

// vtPlayable stubs a game sitting on the human's turn.
func vtPlayable() (*interfaces.MockVintGame, *presenter.MockVintPresenter) {
	pMock := new(presenter.MockVintPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(vintMockOutput)
	gameMock := new(interfaces.MockVintGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("GetPhase").Return(domain.VintPhaseHandEnd)
	return gameMock, pMock
}

func TestNewVintInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockVintPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "VintInteractor: g must not be nil", func() {
			usecase.NewVintInteractor(nil, pMock)
		})
	})

	t.Run("panics when gp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockVintGame)
		assert.PanicsWithValue(t, "VintInteractor: gp must not be nil", func() {
			usecase.NewVintInteractor(gameMock, nil)
		})
	})
}

func TestVintInteractor_Reset(t *testing.T) {
	pMock := new(presenter.MockVintPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(vintMockOutput)
	gameMock := new(interfaces.MockVintGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.VintPhaseBid)
	gameMock.On("IsHumanTurn").Return(true)

	vi := usecase.NewVintInteractor(gameMock, pMock)
	assert.Equal(t, vintMockOutput, vi.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestVintInteractor_ResetWithConfig(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockVintPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(vintMockOutput)
		gameMock := new(interfaces.MockVintGame)
		cfg := domain.DefaultVintConfig()
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.VintPhaseBid)
		gameMock.On("IsHumanTurn").Return(true)

		vi := usecase.NewVintInteractor(gameMock, pMock)
		assert.Equal(t, vintMockOutput, vi.ResetWithConfig(cfg))
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config never reaches the game", func(t *testing.T) {
		pMock := new(presenter.MockVintPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(vintMockOutput)
		gameMock := new(interfaces.MockVintGame)

		vi := usecase.NewVintInteractor(gameMock, pMock)
		assert.Equal(t, vintMockOutput, vi.ResetWithConfig(domain.VintConfig{CpuDifficulty: 9}))
		gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
		gameMock.AssertNotCalled(t, "Reset")
	})
}

// **宣言は bidIdx、プレイは currentIdx。**取り違えると常に別の席で弾かれる。
func TestVintInteractor_UsesTheRightSeatPerPhase(t *testing.T) {
	t.Run("bidding uses the bid seat", func(t *testing.T) {
		gameMock, pMock := vtPlayable()
		gameMock.On("GetBidPlayerIdx").Return(2)
		gameMock.On("Bid", 2, 3, domain.VintDenomHeart).Return(nil)
		gameMock.On("PassBid", 2).Return(nil)

		vi := usecase.NewVintInteractor(gameMock, pMock)
		vi.Bid(3, domain.VintDenomHeart)
		vi.PassBid()

		gameMock.AssertCalled(t, "Bid", 2, 3, domain.VintDenomHeart)
		gameMock.AssertCalled(t, "PassBid", 2)
	})

	t.Run("playing uses the current seat", func(t *testing.T) {
		gameMock, pMock := vtPlayable()
		gameMock.On("GetCurrentPlayerIdx").Return(0)
		gameMock.On("PlayCard", 0, 7).Return(nil)

		vi := usecase.NewVintInteractor(gameMock, pMock)
		assert.Equal(t, vintMockOutput, vi.PlayCard(7))
		gameMock.AssertCalled(t, "PlayCard", 0, 7)
	})
}

func TestVintInteractor_BlockedWhenNotTheHumansTurn(t *testing.T) {
	pMock := new(presenter.MockVintPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(vintMockOutput)
	gameMock := new(interfaces.MockVintGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	vi := usecase.NewVintInteractor(gameMock, pMock)
	assert.Equal(t, vintMockOutput, vi.PlayCard(0))
	assert.Equal(t, vintMockOutput, vi.Bid(3, domain.VintDenomHeart))
	gameMock.AssertNotCalled(t, "PlayCard", mock.Anything, mock.Anything)
	gameMock.AssertNotCalled(t, "Bid", mock.Anything, mock.Anything, mock.Anything)
}

func TestVintInteractor_DomainErrorIsPresented(t *testing.T) {
	wantErr := errors.New("a bid must beat the standing 3")
	pMock := new(presenter.MockVintPresenter)
	pMock.On("Output", mock.Anything, wantErr).Return(vintMockOutput)
	gameMock := new(interfaces.MockVintGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("GetBidPlayerIdx").Return(0)
	gameMock.On("Bid", 0, 2, domain.VintDenomSpade).Return(wantErr)

	vi := usecase.NewVintInteractor(gameMock, pMock)
	assert.Equal(t, vintMockOutput, vi.Bid(2, domain.VintDenomSpade))
	pMock.AssertCalled(t, "Output", mock.Anything, wantErr)
	gameMock.AssertNotCalled(t, "CpuPlay")
}

func TestVintInteractor_NextHand(t *testing.T) {
	t.Run("deals again", func(t *testing.T) {
		pMock := new(presenter.MockVintPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(vintMockOutput)
		gameMock := new(interfaces.MockVintGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextHand").Return(nil)
		gameMock.On("GetPhase").Return(domain.VintPhaseBid)
		gameMock.On("IsHumanTurn").Return(true)

		vi := usecase.NewVintInteractor(gameMock, pMock)
		assert.Equal(t, vintMockOutput, vi.NextHand())
		gameMock.AssertCalled(t, "NextHand")
	})

	t.Run("blocked after the rubber", func(t *testing.T) {
		pMock := new(presenter.MockVintPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(vintMockOutput)
		gameMock := new(interfaces.MockVintGame)
		gameMock.On("GetGameEndFlag").Return(true)

		vi := usecase.NewVintInteractor(gameMock, pMock)
		assert.Equal(t, vintMockOutput, vi.NextHand())
		gameMock.AssertNotCalled(t, "NextHand")
	})

	t.Run("an error is presented", func(t *testing.T) {
		wantErr := errors.New("the hand is still in progress")
		pMock := new(presenter.MockVintPresenter)
		pMock.On("Output", mock.Anything, wantErr).Return(vintMockOutput)
		gameMock := new(interfaces.MockVintGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextHand").Return(wantErr)

		vi := usecase.NewVintInteractor(gameMock, pMock)
		assert.Equal(t, vintMockOutput, vi.NextHand())
		pMock.AssertCalled(t, "Output", mock.Anything, wantErr)
	})
}

// **CPU ループは人間の手番と精算で止まる。**
func TestVintInteractor_RunCpuTurnsStops(t *testing.T) {
	t.Run("at the human's turn", func(t *testing.T) {
		pMock := new(presenter.MockVintPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(vintMockOutput)
		gameMock := new(interfaces.MockVintGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.VintPhasePlay)
		gameMock.On("IsHumanTurn").Return(false).Times(3)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("CpuPlay").Return()

		vi := usecase.NewVintInteractor(gameMock, pMock)
		vi.Reset()
		gameMock.AssertNumberOfCalls(t, "CpuPlay", 3)
	})

	t.Run("at the settlement", func(t *testing.T) {
		pMock := new(presenter.MockVintPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(vintMockOutput)
		gameMock := new(interfaces.MockVintGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.VintPhaseHandEnd)

		vi := usecase.NewVintInteractor(gameMock, pMock)
		vi.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})
}

func TestVintInteractor_GetConfigAndActionLog(t *testing.T) {
	cfg := domain.DefaultVintConfig()
	pMock := new(presenter.MockVintPresenter)
	pMock.On("ActionLogOutput", mock.Anything).Return(`[]`)
	gameMock := new(interfaces.MockVintGame)
	gameMock.On("GetConfig").Return(cfg)

	vi := usecase.NewVintInteractor(gameMock, pMock)
	assert.Equal(t, cfg, vi.GetConfig())
	assert.Equal(t, `[]`, vi.ActionLog())
}

func TestVintInteractor_SnapshotAndRestore(t *testing.T) {
	pMock := new(presenter.MockVintPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(vintMockOutput)

	g := domain.NewDefaultVint()
	g.Reset()
	vi := usecase.NewVintInteractor(g, pMock)
	data, err := vi.Snapshot()
	assert.NoError(t, err)

	restored, err := usecase.RestoreVintInteractor(data, pMock)
	assert.NoError(t, err)
	assert.Equal(t, g.GetConfig(), restored.GetConfig())

	_, err = usecase.RestoreVintInteractor([]byte(`{`), pMock)
	assert.Error(t, err)
}
