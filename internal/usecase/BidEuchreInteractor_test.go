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

const bidEuchreMockOutput = `{"phase":0}`

// bePlayable stubs a game sitting on the human's turn.
func bePlayable() (*interfaces.MockBidEuchreGame, *presenter.MockBidEuchrePresenter) {
	pMock := new(presenter.MockBidEuchrePresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(bidEuchreMockOutput)
	gameMock := new(interfaces.MockBidEuchreGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("GetPhase").Return(domain.BidEuchrePhaseHandEnd)
	return gameMock, pMock
}

func TestNewBidEuchreInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockBidEuchrePresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "BidEuchreInteractor: g must not be nil", func() {
			usecase.NewBidEuchreInteractor(nil, pMock)
		})
	})

	t.Run("panics when gp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockBidEuchreGame)
		assert.PanicsWithValue(t, "BidEuchreInteractor: gp must not be nil", func() {
			usecase.NewBidEuchreInteractor(gameMock, nil)
		})
	})
}

func TestBidEuchreInteractor_Reset(t *testing.T) {
	pMock := new(presenter.MockBidEuchrePresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(bidEuchreMockOutput)
	gameMock := new(interfaces.MockBidEuchreGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.BidEuchrePhaseBid)
	gameMock.On("IsHumanTurn").Return(true)

	bi := usecase.NewBidEuchreInteractor(gameMock, pMock)
	assert.Equal(t, bidEuchreMockOutput, bi.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestBidEuchreInteractor_ResetWithConfig(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockBidEuchrePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(bidEuchreMockOutput)
		gameMock := new(interfaces.MockBidEuchreGame)
		cfg := domain.DefaultBidEuchreConfig()
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.BidEuchrePhaseBid)
		gameMock.On("IsHumanTurn").Return(true)

		bi := usecase.NewBidEuchreInteractor(gameMock, pMock)
		assert.Equal(t, bidEuchreMockOutput, bi.ResetWithConfig(cfg))
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config never reaches the game", func(t *testing.T) {
		pMock := new(presenter.MockBidEuchrePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(bidEuchreMockOutput)
		gameMock := new(interfaces.MockBidEuchreGame)

		bi := usecase.NewBidEuchreInteractor(gameMock, pMock)
		assert.Equal(t, bidEuchreMockOutput, bi.ResetWithConfig(domain.BidEuchreConfig{CpuDifficulty: 9}))
		gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
		gameMock.AssertNotCalled(t, "Reset")
	})
}

// **宣言は bidIdx、切札指定は declarerIdx、プレイは currentIdx。**
// 取り違えると常に別の席で弾かれる。
func TestBidEuchreInteractor_UsesTheRightSeatPerPhase(t *testing.T) {
	t.Run("bidding uses the bid seat", func(t *testing.T) {
		gameMock, pMock := bePlayable()
		gameMock.On("GetBidPlayerIdx").Return(2)
		gameMock.On("Bid", 2, 4).Return(nil)
		gameMock.On("PassBid", 2).Return(nil)

		bi := usecase.NewBidEuchreInteractor(gameMock, pMock)
		bi.Bid(4)
		bi.PassBid()

		gameMock.AssertCalled(t, "Bid", 2, 4)
		gameMock.AssertCalled(t, "PassBid", 2)
	})

	// **切札を名乗るのは落札者であって手番ではない。**
	t.Run("naming trump uses the declarer seat", func(t *testing.T) {
		gameMock, pMock := bePlayable()
		gameMock.On("GetDeclarerIdx").Return(3)
		gameMock.On("ChooseTrump", 3, domain.BidEuchreTrumpNoLow).Return(nil)

		bi := usecase.NewBidEuchreInteractor(gameMock, pMock)
		bi.ChooseTrump(int(domain.BidEuchreTrumpNoLow))

		gameMock.AssertCalled(t, "ChooseTrump", 3, domain.BidEuchreTrumpNoLow)
	})

	t.Run("playing uses the current seat", func(t *testing.T) {
		gameMock, pMock := bePlayable()
		gameMock.On("GetCurrentPlayerIdx").Return(0)
		gameMock.On("PlayCard", 0, 5).Return(nil)

		bi := usecase.NewBidEuchreInteractor(gameMock, pMock)
		assert.Equal(t, bidEuchreMockOutput, bi.PlayCard(5))
		gameMock.AssertCalled(t, "PlayCard", 0, 5)
	})
}

func TestBidEuchreInteractor_BlockedWhenNotTheHumansTurn(t *testing.T) {
	pMock := new(presenter.MockBidEuchrePresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(bidEuchreMockOutput)
	gameMock := new(interfaces.MockBidEuchreGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	bi := usecase.NewBidEuchreInteractor(gameMock, pMock)
	assert.Equal(t, bidEuchreMockOutput, bi.PlayCard(0))
	assert.Equal(t, bidEuchreMockOutput, bi.Bid(3))
	assert.Equal(t, bidEuchreMockOutput, bi.ChooseTrump(0))
	gameMock.AssertNotCalled(t, "PlayCard", mock.Anything, mock.Anything)
	gameMock.AssertNotCalled(t, "Bid", mock.Anything, mock.Anything)
	gameMock.AssertNotCalled(t, "ChooseTrump", mock.Anything, mock.Anything)
}

func TestBidEuchreInteractor_DomainErrorIsPresented(t *testing.T) {
	wantErr := errors.New("a bid must beat the standing 4")
	pMock := new(presenter.MockBidEuchrePresenter)
	pMock.On("Output", mock.Anything, wantErr).Return(bidEuchreMockOutput)
	gameMock := new(interfaces.MockBidEuchreGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("GetBidPlayerIdx").Return(0)
	gameMock.On("Bid", 0, 3).Return(wantErr)

	bi := usecase.NewBidEuchreInteractor(gameMock, pMock)
	assert.Equal(t, bidEuchreMockOutput, bi.Bid(3))
	pMock.AssertCalled(t, "Output", mock.Anything, wantErr)
	gameMock.AssertNotCalled(t, "CpuPlay")
}

func TestBidEuchreInteractor_NextHand(t *testing.T) {
	t.Run("deals again", func(t *testing.T) {
		pMock := new(presenter.MockBidEuchrePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(bidEuchreMockOutput)
		gameMock := new(interfaces.MockBidEuchreGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextHand").Return(nil)
		gameMock.On("GetPhase").Return(domain.BidEuchrePhaseBid)
		gameMock.On("IsHumanTurn").Return(true)

		bi := usecase.NewBidEuchreInteractor(gameMock, pMock)
		assert.Equal(t, bidEuchreMockOutput, bi.NextHand())
		gameMock.AssertCalled(t, "NextHand")
	})

	t.Run("blocked after the game", func(t *testing.T) {
		pMock := new(presenter.MockBidEuchrePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(bidEuchreMockOutput)
		gameMock := new(interfaces.MockBidEuchreGame)
		gameMock.On("GetGameEndFlag").Return(true)

		bi := usecase.NewBidEuchreInteractor(gameMock, pMock)
		assert.Equal(t, bidEuchreMockOutput, bi.NextHand())
		gameMock.AssertNotCalled(t, "NextHand")
	})

	t.Run("an error is presented", func(t *testing.T) {
		wantErr := errors.New("the hand is still in progress")
		pMock := new(presenter.MockBidEuchrePresenter)
		pMock.On("Output", mock.Anything, wantErr).Return(bidEuchreMockOutput)
		gameMock := new(interfaces.MockBidEuchreGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextHand").Return(wantErr)

		bi := usecase.NewBidEuchreInteractor(gameMock, pMock)
		assert.Equal(t, bidEuchreMockOutput, bi.NextHand())
		pMock.AssertCalled(t, "Output", mock.Anything, wantErr)
	})
}

// **CPU ループは人間の手番と精算で止まる。**
func TestBidEuchreInteractor_RunCpuTurnsStops(t *testing.T) {
	t.Run("at the human's turn", func(t *testing.T) {
		pMock := new(presenter.MockBidEuchrePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(bidEuchreMockOutput)
		gameMock := new(interfaces.MockBidEuchreGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.BidEuchrePhasePlay)
		gameMock.On("IsHumanTurn").Return(false).Times(3)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("CpuPlay").Return()

		bi := usecase.NewBidEuchreInteractor(gameMock, pMock)
		bi.Reset()
		gameMock.AssertNumberOfCalls(t, "CpuPlay", 3)
	})

	t.Run("at the settlement", func(t *testing.T) {
		pMock := new(presenter.MockBidEuchrePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(bidEuchreMockOutput)
		gameMock := new(interfaces.MockBidEuchreGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.BidEuchrePhaseHandEnd)

		bi := usecase.NewBidEuchreInteractor(gameMock, pMock)
		bi.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})
}

func TestBidEuchreInteractor_GetConfigAndActionLog(t *testing.T) {
	cfg := domain.DefaultBidEuchreConfig()
	pMock := new(presenter.MockBidEuchrePresenter)
	pMock.On("ActionLogOutput", mock.Anything).Return(`[]`)
	gameMock := new(interfaces.MockBidEuchreGame)
	gameMock.On("GetConfig").Return(cfg)

	bi := usecase.NewBidEuchreInteractor(gameMock, pMock)
	assert.Equal(t, cfg, bi.GetConfig())
	assert.Equal(t, `[]`, bi.ActionLog())
}

func TestBidEuchreInteractor_SnapshotAndRestore(t *testing.T) {
	pMock := new(presenter.MockBidEuchrePresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(bidEuchreMockOutput)

	g := domain.NewDefaultBidEuchre()
	g.Reset()
	bi := usecase.NewBidEuchreInteractor(g, pMock)
	data, err := bi.Snapshot()
	assert.NoError(t, err)

	restored, err := usecase.RestoreBidEuchreInteractor(data, pMock)
	assert.NoError(t, err)
	assert.Equal(t, g.GetConfig(), restored.GetConfig())

	_, err = usecase.RestoreBidEuchreInteractor([]byte(`{`), pMock)
	assert.Error(t, err)
}

// **CUI のヒントはここを通る。**presenter の HintOutput に素通しする (#5730)。
func TestBidEuchreInteractor_Hint(t *testing.T) {
	pMock := new(presenter.MockBidEuchrePresenter)
	pMock.On("HintOutput", mock.Anything).Return("advice")
	gameMock := new(interfaces.MockBidEuchreGame)

	bi := usecase.NewBidEuchreInteractor(gameMock, pMock)
	assert.Equal(t, "advice", bi.Hint())
	pMock.AssertCalled(t, "HintOutput", gameMock)
}
