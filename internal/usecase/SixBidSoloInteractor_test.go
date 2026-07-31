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

const sixBidSoloMockOutput = `{"phase":0}`

// sbsPlayable stubs a game sitting on the human's turn.
func sbsPlayable() (*interfaces.MockSixBidSoloGame, *presenter.MockSixBidSoloPresenter) {
	pMock := new(presenter.MockSixBidSoloPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(sixBidSoloMockOutput)
	gameMock := new(interfaces.MockSixBidSoloGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("GetPhase").Return(domain.SixBidSoloPhaseHandEnd)
	return gameMock, pMock
}

func TestNewSixBidSoloInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockSixBidSoloPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "SixBidSoloInteractor: g must not be nil", func() {
			usecase.NewSixBidSoloInteractor(nil, pMock)
		})
	})

	t.Run("panics when gp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockSixBidSoloGame)
		assert.PanicsWithValue(t, "SixBidSoloInteractor: gp must not be nil", func() {
			usecase.NewSixBidSoloInteractor(gameMock, nil)
		})
	})
}

func TestSixBidSoloInteractor_Reset(t *testing.T) {
	pMock := new(presenter.MockSixBidSoloPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(sixBidSoloMockOutput)
	gameMock := new(interfaces.MockSixBidSoloGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SixBidSoloPhaseBid)
	gameMock.On("IsHumanTurn").Return(true)

	si := usecase.NewSixBidSoloInteractor(gameMock, pMock)
	assert.Equal(t, sixBidSoloMockOutput, si.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestSixBidSoloInteractor_ResetWithConfig(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockSixBidSoloPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(sixBidSoloMockOutput)
		gameMock := new(interfaces.MockSixBidSoloGame)
		cfg := domain.DefaultSixBidSoloConfig()
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.SixBidSoloPhaseBid)
		gameMock.On("IsHumanTurn").Return(true)

		si := usecase.NewSixBidSoloInteractor(gameMock, pMock)
		assert.Equal(t, sixBidSoloMockOutput, si.ResetWithConfig(cfg))
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config never reaches the game", func(t *testing.T) {
		pMock := new(presenter.MockSixBidSoloPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(sixBidSoloMockOutput)
		gameMock := new(interfaces.MockSixBidSoloGame)

		si := usecase.NewSixBidSoloInteractor(gameMock, pMock)
		assert.Equal(t, sixBidSoloMockOutput, si.ResetWithConfig(domain.SixBidSoloConfig{TargetHands: 99}))
		gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
		gameMock.AssertNotCalled(t, "Reset")
	})
}

// **宣言は bidIdx、切札指定は declarerIdx、プレイは currentIdx。**
func TestSixBidSoloInteractor_UsesTheRightSeatPerPhase(t *testing.T) {
	t.Run("bidding uses the bid seat", func(t *testing.T) {
		gameMock, pMock := sbsPlayable()
		gameMock.On("GetBidPlayerIdx").Return(2)
		gameMock.On("Bid", 2, domain.SixBidSoloBidGuarantee).Return(nil)
		gameMock.On("PassBid", 2).Return(nil)

		si := usecase.NewSixBidSoloInteractor(gameMock, pMock)
		si.Bid(int(domain.SixBidSoloBidGuarantee))
		si.PassBid()

		gameMock.AssertCalled(t, "Bid", 2, domain.SixBidSoloBidGuarantee)
		gameMock.AssertCalled(t, "PassBid", 2)
	})

	// **切札を名乗るのは落札者。**手番ではない。
	t.Run("declaring uses the declarer seat", func(t *testing.T) {
		gameMock, pMock := sbsPlayable()
		gameMock.On("GetDeclarerIdx").Return(1)
		gameMock.On("Declare", 1, domain.CardDesignHeart, mock.Anything).Return(nil)

		si := usecase.NewSixBidSoloInteractor(gameMock, pMock)
		si.Declare(domain.CardDesignHeart, 0, 0)

		gameMock.AssertCalled(t, "Declare", 1, domain.CardDesignHeart, mock.Anything)
	})

	t.Run("playing uses the current seat", func(t *testing.T) {
		gameMock, pMock := sbsPlayable()
		gameMock.On("GetCurrentPlayerIdx").Return(0)
		gameMock.On("PlayCard", 0, 7).Return(nil)

		si := usecase.NewSixBidSoloInteractor(gameMock, pMock)
		assert.Equal(t, sixBidSoloMockOutput, si.PlayCard(7))
		gameMock.AssertCalled(t, "PlayCard", 0, 7)
	})
}

// **指名札はコール・ソロだけに要る。**スートが 0 なら札は送られていない。
func TestSixBidSoloInteractor_DeclareBuildsTheCalledCardOnlyWhenGiven(t *testing.T) {
	t.Run("no called card", func(t *testing.T) {
		gameMock, pMock := sbsPlayable()
		gameMock.On("GetDeclarerIdx").Return(0)
		gameMock.On("Declare", 0, domain.CardDesignSpade, (*domain.Card)(nil)).Return(nil)

		si := usecase.NewSixBidSoloInteractor(gameMock, pMock)
		si.Declare(domain.CardDesignSpade, 0, 0)
		gameMock.AssertCalled(t, "Declare", 0, domain.CardDesignSpade, (*domain.Card)(nil))
	})

	t.Run("a called card is built", func(t *testing.T) {
		gameMock, pMock := sbsPlayable()
		gameMock.On("GetDeclarerIdx").Return(0)
		gameMock.On("Declare", 0, domain.CardDesignSpade, mock.MatchedBy(func(c *domain.Card) bool {
			return c != nil && c.GetDesign() == domain.CardDesignHeart && c.GetValue() == 1
		})).Return(nil)

		si := usecase.NewSixBidSoloInteractor(gameMock, pMock)
		si.Declare(domain.CardDesignSpade, domain.CardDesignHeart, 1)
		gameMock.AssertNumberOfCalls(t, "Declare", 1)
	})
}

func TestSixBidSoloInteractor_BlockedWhenNotTheHumansTurn(t *testing.T) {
	pMock := new(presenter.MockSixBidSoloPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(sixBidSoloMockOutput)
	gameMock := new(interfaces.MockSixBidSoloGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	si := usecase.NewSixBidSoloInteractor(gameMock, pMock)
	assert.Equal(t, sixBidSoloMockOutput, si.PlayCard(0))
	assert.Equal(t, sixBidSoloMockOutput, si.Bid(1))
	assert.Equal(t, sixBidSoloMockOutput, si.Declare(1, 0, 0))
	gameMock.AssertNotCalled(t, "PlayCard", mock.Anything, mock.Anything)
	gameMock.AssertNotCalled(t, "Bid", mock.Anything, mock.Anything)
	gameMock.AssertNotCalled(t, "Declare", mock.Anything, mock.Anything, mock.Anything)
}

func TestSixBidSoloInteractor_DomainErrorIsPresented(t *testing.T) {
	wantErr := errors.New("a bid must beat the standing one")
	pMock := new(presenter.MockSixBidSoloPresenter)
	pMock.On("Output", mock.Anything, wantErr).Return(sixBidSoloMockOutput)
	gameMock := new(interfaces.MockSixBidSoloGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("GetBidPlayerIdx").Return(0)
	gameMock.On("Bid", 0, domain.SixBidSoloBidSolo).Return(wantErr)

	si := usecase.NewSixBidSoloInteractor(gameMock, pMock)
	assert.Equal(t, sixBidSoloMockOutput, si.Bid(int(domain.SixBidSoloBidSolo)))
	pMock.AssertCalled(t, "Output", mock.Anything, wantErr)
	gameMock.AssertNotCalled(t, "CpuPlay")
}

func TestSixBidSoloInteractor_NextHand(t *testing.T) {
	t.Run("deals again", func(t *testing.T) {
		pMock := new(presenter.MockSixBidSoloPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(sixBidSoloMockOutput)
		gameMock := new(interfaces.MockSixBidSoloGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextHand").Return(nil)
		gameMock.On("GetPhase").Return(domain.SixBidSoloPhaseBid)
		gameMock.On("IsHumanTurn").Return(true)

		si := usecase.NewSixBidSoloInteractor(gameMock, pMock)
		assert.Equal(t, sixBidSoloMockOutput, si.NextHand())
		gameMock.AssertCalled(t, "NextHand")
	})

	t.Run("blocked after the game", func(t *testing.T) {
		pMock := new(presenter.MockSixBidSoloPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(sixBidSoloMockOutput)
		gameMock := new(interfaces.MockSixBidSoloGame)
		gameMock.On("GetGameEndFlag").Return(true)

		si := usecase.NewSixBidSoloInteractor(gameMock, pMock)
		assert.Equal(t, sixBidSoloMockOutput, si.NextHand())
		gameMock.AssertNotCalled(t, "NextHand")
	})

	t.Run("an error is presented", func(t *testing.T) {
		wantErr := errors.New("the hand is still in progress")
		pMock := new(presenter.MockSixBidSoloPresenter)
		pMock.On("Output", mock.Anything, wantErr).Return(sixBidSoloMockOutput)
		gameMock := new(interfaces.MockSixBidSoloGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextHand").Return(wantErr)

		si := usecase.NewSixBidSoloInteractor(gameMock, pMock)
		assert.Equal(t, sixBidSoloMockOutput, si.NextHand())
		pMock.AssertCalled(t, "Output", mock.Anything, wantErr)
	})
}

// **CPU ループは人間の手番と精算で止まる。**
func TestSixBidSoloInteractor_RunCpuTurnsStops(t *testing.T) {
	t.Run("at the human's turn", func(t *testing.T) {
		pMock := new(presenter.MockSixBidSoloPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(sixBidSoloMockOutput)
		gameMock := new(interfaces.MockSixBidSoloGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.SixBidSoloPhasePlay)
		gameMock.On("IsHumanTurn").Return(false).Times(3)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("CpuPlay").Return()

		si := usecase.NewSixBidSoloInteractor(gameMock, pMock)
		si.Reset()
		gameMock.AssertNumberOfCalls(t, "CpuPlay", 3)
	})

	t.Run("at the settlement", func(t *testing.T) {
		pMock := new(presenter.MockSixBidSoloPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(sixBidSoloMockOutput)
		gameMock := new(interfaces.MockSixBidSoloGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.SixBidSoloPhaseHandEnd)

		si := usecase.NewSixBidSoloInteractor(gameMock, pMock)
		si.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})
}

func TestSixBidSoloInteractor_GetConfigAndActionLog(t *testing.T) {
	cfg := domain.DefaultSixBidSoloConfig()
	pMock := new(presenter.MockSixBidSoloPresenter)
	pMock.On("ActionLogOutput", mock.Anything).Return(`[]`)
	gameMock := new(interfaces.MockSixBidSoloGame)
	gameMock.On("GetConfig").Return(cfg)

	si := usecase.NewSixBidSoloInteractor(gameMock, pMock)
	assert.Equal(t, cfg, si.GetConfig())
	assert.Equal(t, `[]`, si.ActionLog())
}

func TestSixBidSoloInteractor_SnapshotAndRestore(t *testing.T) {
	pMock := new(presenter.MockSixBidSoloPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(sixBidSoloMockOutput)

	g := domain.NewDefaultSixBidSolo()
	g.Reset()
	si := usecase.NewSixBidSoloInteractor(g, pMock)
	data, err := si.Snapshot()
	assert.NoError(t, err)

	restored, err := usecase.RestoreSixBidSoloInteractor(data, pMock)
	assert.NoError(t, err)
	assert.Equal(t, g.GetConfig(), restored.GetConfig())

	_, err = usecase.RestoreSixBidSoloInteractor([]byte(`{`), pMock)
	assert.Error(t, err)
}
