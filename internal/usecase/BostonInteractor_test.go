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

const bostonMockOutput = `{"phase":0}`

// bsPlayable stubs a game sitting on the human's turn.
func bsPlayable() (*interfaces.MockBostonGame, *presenter.MockBostonPresenter) {
	pMock := new(presenter.MockBostonPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(bostonMockOutput)
	gameMock := new(interfaces.MockBostonGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("GetPhase").Return(domain.BostonPhaseHandEnd)
	return gameMock, pMock
}

func TestNewBostonInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockBostonPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "BostonInteractor: g must not be nil", func() {
			usecase.NewBostonInteractor(nil, pMock)
		})
	})

	t.Run("panics when gp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockBostonGame)
		assert.PanicsWithValue(t, "BostonInteractor: gp must not be nil", func() {
			usecase.NewBostonInteractor(gameMock, nil)
		})
	})
}

func TestBostonInteractor_Reset(t *testing.T) {
	pMock := new(presenter.MockBostonPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(bostonMockOutput)
	gameMock := new(interfaces.MockBostonGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.BostonPhaseBid)
	gameMock.On("IsHumanTurn").Return(true)

	bi := usecase.NewBostonInteractor(gameMock, pMock)
	assert.Equal(t, bostonMockOutput, bi.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestBostonInteractor_ResetWithConfig(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockBostonPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(bostonMockOutput)
		gameMock := new(interfaces.MockBostonGame)
		cfg := domain.BostonConfig{TargetHands: 5}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.BostonPhaseBid)
		gameMock.On("IsHumanTurn").Return(true)

		bi := usecase.NewBostonInteractor(gameMock, pMock)
		assert.Equal(t, bostonMockOutput, bi.ResetWithConfig(cfg))
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config never reaches the game", func(t *testing.T) {
		pMock := new(presenter.MockBostonPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(bostonMockOutput)
		gameMock := new(interfaces.MockBostonGame)

		bi := usecase.NewBostonInteractor(gameMock, pMock)
		assert.Equal(t, bostonMockOutput, bi.ResetWithConfig(domain.BostonConfig{TargetHands: 0}))
		gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
		gameMock.AssertNotCalled(t, "Reset")
	})
}

// **フェーズごとに使う席が違う。**宣言は bidIdx、パートナー指名は declarerIdx、
// プレイは currentIdx。取り違えると常に別の席で動こうとして弾かれる。
func TestBostonInteractor_UsesTheRightSeatPerPhase(t *testing.T) {
	t.Run("bidding uses the bid seat", func(t *testing.T) {
		gameMock, pMock := bsPlayable()
		gameMock.On("GetBidPlayerIdx").Return(2)
		gameMock.On("Bid", 2, domain.BostonBidSeven, domain.CardDesignHeart).Return(nil)
		gameMock.On("PassBid", 2).Return(nil)

		bi := usecase.NewBostonInteractor(gameMock, pMock)
		bi.Bid(domain.BostonBidSeven, domain.CardDesignHeart)
		bi.PassBid()

		gameMock.AssertCalled(t, "Bid", 2, domain.BostonBidSeven, domain.CardDesignHeart)
		gameMock.AssertCalled(t, "PassBid", 2)
	})

	// **パートナー指名は落札者の席。**宣言手番はもう進んでいる。
	t.Run("calling a partner uses the declarer seat", func(t *testing.T) {
		gameMock, pMock := bsPlayable()
		gameMock.On("GetDeclarerIdx").Return(3)
		gameMock.On("CallPartner", 3, 1).Return(nil)

		bi := usecase.NewBostonInteractor(gameMock, pMock)
		assert.Equal(t, bostonMockOutput, bi.CallPartner(1))
		gameMock.AssertCalled(t, "CallPartner", 3, 1)
		gameMock.AssertNotCalled(t, "GetBidPlayerIdx")
	})

	t.Run("playing uses the current seat", func(t *testing.T) {
		gameMock, pMock := bsPlayable()
		gameMock.On("GetCurrentPlayerIdx").Return(0)
		gameMock.On("PlayCard", 0, 5).Return(nil)

		bi := usecase.NewBostonInteractor(gameMock, pMock)
		assert.Equal(t, bostonMockOutput, bi.PlayCard(5))
		gameMock.AssertCalled(t, "PlayCard", 0, 5)
	})
}

// **-1 は「単独で戦う」という有効な選択。**弾いてはいけない。
func TestBostonInteractor_GoingAlonePassesMinusOne(t *testing.T) {
	gameMock, pMock := bsPlayable()
	gameMock.On("GetDeclarerIdx").Return(0)
	gameMock.On("CallPartner", 0, -1).Return(nil)

	bi := usecase.NewBostonInteractor(gameMock, pMock)
	assert.Equal(t, bostonMockOutput, bi.CallPartner(-1))
	gameMock.AssertCalled(t, "CallPartner", 0, -1)
}

func TestBostonInteractor_BlockedWhenNotTheHumansTurn(t *testing.T) {
	pMock := new(presenter.MockBostonPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(bostonMockOutput)
	gameMock := new(interfaces.MockBostonGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	bi := usecase.NewBostonInteractor(gameMock, pMock)
	assert.Equal(t, bostonMockOutput, bi.PlayCard(0))
	assert.Equal(t, bostonMockOutput, bi.Bid(domain.BostonBidFive, domain.CardDesignSpade))
	assert.Equal(t, bostonMockOutput, bi.CallPartner(1))
	gameMock.AssertNotCalled(t, "PlayCard", mock.Anything, mock.Anything)
	gameMock.AssertNotCalled(t, "Bid", mock.Anything, mock.Anything, mock.Anything)
	gameMock.AssertNotCalled(t, "CallPartner", mock.Anything, mock.Anything)
}

func TestBostonInteractor_DomainErrorIsPresented(t *testing.T) {
	wantErr := errors.New("a bid must beat the standing seven")
	pMock := new(presenter.MockBostonPresenter)
	pMock.On("Output", mock.Anything, wantErr).Return(bostonMockOutput)
	gameMock := new(interfaces.MockBostonGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("GetBidPlayerIdx").Return(0)
	gameMock.On("Bid", 0, domain.BostonBidLittleMisere, 0).Return(wantErr)

	bi := usecase.NewBostonInteractor(gameMock, pMock)
	assert.Equal(t, bostonMockOutput, bi.Bid(domain.BostonBidLittleMisere, 0))
	pMock.AssertCalled(t, "Output", mock.Anything, wantErr)
	gameMock.AssertNotCalled(t, "CpuPlay")
}

func TestBostonInteractor_NextHand(t *testing.T) {
	t.Run("deals again", func(t *testing.T) {
		pMock := new(presenter.MockBostonPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(bostonMockOutput)
		gameMock := new(interfaces.MockBostonGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextHand").Return(nil)
		gameMock.On("GetPhase").Return(domain.BostonPhaseBid)
		gameMock.On("IsHumanTurn").Return(true)

		bi := usecase.NewBostonInteractor(gameMock, pMock)
		assert.Equal(t, bostonMockOutput, bi.NextHand())
		gameMock.AssertCalled(t, "NextHand")
	})

	t.Run("blocked after the game ends", func(t *testing.T) {
		pMock := new(presenter.MockBostonPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(bostonMockOutput)
		gameMock := new(interfaces.MockBostonGame)
		gameMock.On("GetGameEndFlag").Return(true)

		bi := usecase.NewBostonInteractor(gameMock, pMock)
		assert.Equal(t, bostonMockOutput, bi.NextHand())
		gameMock.AssertNotCalled(t, "NextHand")
	})

	t.Run("an error is presented", func(t *testing.T) {
		wantErr := errors.New("the hand is still in progress")
		pMock := new(presenter.MockBostonPresenter)
		pMock.On("Output", mock.Anything, wantErr).Return(bostonMockOutput)
		gameMock := new(interfaces.MockBostonGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextHand").Return(wantErr)

		bi := usecase.NewBostonInteractor(gameMock, pMock)
		assert.Equal(t, bostonMockOutput, bi.NextHand())
		pMock.AssertCalled(t, "Output", mock.Anything, wantErr)
	})
}

// **CPU ループは人間の手番と精算で止まる。**止まらないと操作できない。
func TestBostonInteractor_RunCpuTurnsStops(t *testing.T) {
	t.Run("at the human's turn", func(t *testing.T) {
		pMock := new(presenter.MockBostonPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(bostonMockOutput)
		gameMock := new(interfaces.MockBostonGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.BostonPhasePlay)
		gameMock.On("IsHumanTurn").Return(false).Times(3)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("CpuPlay").Return()

		bi := usecase.NewBostonInteractor(gameMock, pMock)
		bi.Reset()
		gameMock.AssertNumberOfCalls(t, "CpuPlay", 3)
	})

	t.Run("at the settlement", func(t *testing.T) {
		pMock := new(presenter.MockBostonPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(bostonMockOutput)
		gameMock := new(interfaces.MockBostonGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.BostonPhaseHandEnd)

		bi := usecase.NewBostonInteractor(gameMock, pMock)
		bi.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})
}

func TestBostonInteractor_GetConfigAndActionLog(t *testing.T) {
	cfg := domain.DefaultBostonConfig()
	pMock := new(presenter.MockBostonPresenter)
	pMock.On("ActionLogOutput", mock.Anything).Return(`[]`)
	gameMock := new(interfaces.MockBostonGame)
	gameMock.On("GetConfig").Return(cfg)

	bi := usecase.NewBostonInteractor(gameMock, pMock)
	assert.Equal(t, cfg, bi.GetConfig())
	assert.Equal(t, `[]`, bi.ActionLog())
}

func TestBostonInteractor_SnapshotAndRestore(t *testing.T) {
	pMock := new(presenter.MockBostonPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(bostonMockOutput)

	g := domain.NewDefaultBoston()
	g.Reset()
	bi := usecase.NewBostonInteractor(g, pMock)
	data, err := bi.Snapshot()
	assert.NoError(t, err)

	restored, err := usecase.RestoreBostonInteractor(data, pMock)
	assert.NoError(t, err)
	assert.Equal(t, g.GetConfig(), restored.GetConfig())

	_, err = usecase.RestoreBostonInteractor([]byte(`{`), pMock)
	assert.Error(t, err)
}
