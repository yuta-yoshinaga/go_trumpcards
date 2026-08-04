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

const kaiserMockOutput = `{"phase":0}`

// kzPlayable stubs a game sitting on the human's turn.
func kzPlayable() (*interfaces.MockKaiserGame, *presenter.MockKaiserPresenter) {
	pMock := new(presenter.MockKaiserPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(kaiserMockOutput)
	gameMock := new(interfaces.MockKaiserGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("GetPhase").Return(domain.KaiserPhaseHandEnd)
	return gameMock, pMock
}

func TestNewKaiserInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockKaiserPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "KaiserInteractor: g must not be nil", func() {
			usecase.NewKaiserInteractor(nil, pMock)
		})
	})

	t.Run("panics when gp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockKaiserGame)
		assert.PanicsWithValue(t, "KaiserInteractor: gp must not be nil", func() {
			usecase.NewKaiserInteractor(gameMock, nil)
		})
	})
}

func TestKaiserInteractor_Reset(t *testing.T) {
	pMock := new(presenter.MockKaiserPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(kaiserMockOutput)
	gameMock := new(interfaces.MockKaiserGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.KaiserPhaseBid)
	gameMock.On("IsHumanTurn").Return(true)

	ki := usecase.NewKaiserInteractor(gameMock, pMock)
	assert.Equal(t, kaiserMockOutput, ki.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestKaiserInteractor_ResetWithConfig(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockKaiserPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(kaiserMockOutput)
		gameMock := new(interfaces.MockKaiserGame)
		cfg := domain.KaiserConfig{AllowNoTrump: false}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.KaiserPhaseBid)
		gameMock.On("IsHumanTurn").Return(true)

		ki := usecase.NewKaiserInteractor(gameMock, pMock)
		assert.Equal(t, kaiserMockOutput, ki.ResetWithConfig(cfg))
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config never reaches the game", func(t *testing.T) {
		pMock := new(presenter.MockKaiserPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(kaiserMockOutput)
		gameMock := new(interfaces.MockKaiserGame)

		ki := usecase.NewKaiserInteractor(gameMock, pMock)
		assert.Equal(t, kaiserMockOutput, ki.ResetWithConfig(domain.KaiserConfig{CpuDifficulty: 9}))
		gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
		gameMock.AssertNotCalled(t, "Reset")
	})
}

// **フェーズごとに使う席が違う。**ビッドは bidIdx、切札と捨て札は declarerIdx、
// プレイは currentIdx。取り違えると常に別の席で動こうとして弾かれる。
func TestKaiserInteractor_UsesTheRightSeatPerPhase(t *testing.T) {
	t.Run("bidding uses the bid seat", func(t *testing.T) {
		gameMock, pMock := kzPlayable()
		gameMock.On("GetBidPlayerIdx").Return(2)
		gameMock.On("Bid", 2, 8, domain.KaiserContractTrump).Return(nil)
		gameMock.On("PassBid", 2).Return(nil)

		ki := usecase.NewKaiserInteractor(gameMock, pMock)
		ki.Bid(8, domain.KaiserContractTrump)
		ki.PassBid()

		gameMock.AssertCalled(t, "Bid", 2, 8, domain.KaiserContractTrump)
		gameMock.AssertCalled(t, "PassBid", 2)
	})

	// **切札と捨て札は落札者の席。**ビッド手番はもう進んでいる。
	t.Run("naming trump and discarding use the declarer seat", func(t *testing.T) {
		gameMock, pMock := kzPlayable()
		gameMock.On("GetDeclarerIdx").Return(3)
		gameMock.On("SetTrump", 3, domain.CardDesignHeart).Return(nil)
		gameMock.On("Discard", 3, []int{0, 1}).Return(nil)

		ki := usecase.NewKaiserInteractor(gameMock, pMock)
		ki.SetTrump(domain.CardDesignHeart)
		ki.Discard([]int{0, 1})

		gameMock.AssertCalled(t, "SetTrump", 3, domain.CardDesignHeart)
		gameMock.AssertCalled(t, "Discard", 3, []int{0, 1})
		gameMock.AssertNotCalled(t, "GetBidPlayerIdx")
	})

	t.Run("playing uses the current seat", func(t *testing.T) {
		gameMock, pMock := kzPlayable()
		gameMock.On("GetCurrentPlayerIdx").Return(0)
		gameMock.On("PlayCard", 0, 4).Return(nil)

		ki := usecase.NewKaiserInteractor(gameMock, pMock)
		assert.Equal(t, kaiserMockOutput, ki.PlayCard(4))
		gameMock.AssertCalled(t, "PlayCard", 0, 4)
	})
}

func TestKaiserInteractor_BlockedWhenNotTheHumansTurn(t *testing.T) {
	pMock := new(presenter.MockKaiserPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(kaiserMockOutput)
	gameMock := new(interfaces.MockKaiserGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	ki := usecase.NewKaiserInteractor(gameMock, pMock)
	assert.Equal(t, kaiserMockOutput, ki.PlayCard(0))
	assert.Equal(t, kaiserMockOutput, ki.Bid(8, domain.KaiserContractTrump))
	assert.Equal(t, kaiserMockOutput, ki.Discard([]int{0, 1}))
	gameMock.AssertNotCalled(t, "PlayCard", mock.Anything, mock.Anything)
	gameMock.AssertNotCalled(t, "Bid", mock.Anything, mock.Anything, mock.Anything)
	gameMock.AssertNotCalled(t, "Discard", mock.Anything, mock.Anything)
}

func TestKaiserInteractor_DomainErrorIsPresented(t *testing.T) {
	wantErr := errors.New("the five of hearts and three of spades may not be discarded")
	pMock := new(presenter.MockKaiserPresenter)
	pMock.On("Output", mock.Anything, wantErr).Return(kaiserMockOutput)
	gameMock := new(interfaces.MockKaiserGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("GetDeclarerIdx").Return(0)
	gameMock.On("Discard", 0, []int{0, 1}).Return(wantErr)

	ki := usecase.NewKaiserInteractor(gameMock, pMock)
	assert.Equal(t, kaiserMockOutput, ki.Discard([]int{0, 1}))
	pMock.AssertCalled(t, "Output", mock.Anything, wantErr)
	// エラーで止まったら CPU は動かさない。
	gameMock.AssertNotCalled(t, "CpuPlay")
}

func TestKaiserInteractor_NextHand(t *testing.T) {
	t.Run("deals again", func(t *testing.T) {
		pMock := new(presenter.MockKaiserPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(kaiserMockOutput)
		gameMock := new(interfaces.MockKaiserGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextHand").Return(nil)
		gameMock.On("GetPhase").Return(domain.KaiserPhaseBid)
		gameMock.On("IsHumanTurn").Return(true)

		ki := usecase.NewKaiserInteractor(gameMock, pMock)
		assert.Equal(t, kaiserMockOutput, ki.NextHand())
		gameMock.AssertCalled(t, "NextHand")
	})

	t.Run("blocked after the game ends", func(t *testing.T) {
		pMock := new(presenter.MockKaiserPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(kaiserMockOutput)
		gameMock := new(interfaces.MockKaiserGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ki := usecase.NewKaiserInteractor(gameMock, pMock)
		assert.Equal(t, kaiserMockOutput, ki.NextHand())
		gameMock.AssertNotCalled(t, "NextHand")
	})

	t.Run("an error is presented", func(t *testing.T) {
		wantErr := errors.New("the hand is still in progress")
		pMock := new(presenter.MockKaiserPresenter)
		pMock.On("Output", mock.Anything, wantErr).Return(kaiserMockOutput)
		gameMock := new(interfaces.MockKaiserGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextHand").Return(wantErr)

		ki := usecase.NewKaiserInteractor(gameMock, pMock)
		assert.Equal(t, kaiserMockOutput, ki.NextHand())
		pMock.AssertCalled(t, "Output", mock.Anything, wantErr)
	})
}

// **CPU ループは人間の手番と精算で止まる。**止まらないと操作できない。
func TestKaiserInteractor_RunCpuTurnsStops(t *testing.T) {
	t.Run("at the human's turn", func(t *testing.T) {
		pMock := new(presenter.MockKaiserPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(kaiserMockOutput)
		gameMock := new(interfaces.MockKaiserGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.KaiserPhasePlay)
		gameMock.On("IsHumanTurn").Return(false).Times(3)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("CpuPlay").Return()

		ki := usecase.NewKaiserInteractor(gameMock, pMock)
		ki.Reset()
		gameMock.AssertNumberOfCalls(t, "CpuPlay", 3)
	})

	t.Run("at the settlement", func(t *testing.T) {
		pMock := new(presenter.MockKaiserPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(kaiserMockOutput)
		gameMock := new(interfaces.MockKaiserGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.KaiserPhaseHandEnd)

		ki := usecase.NewKaiserInteractor(gameMock, pMock)
		ki.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})
}

func TestKaiserInteractor_GetConfigAndActionLog(t *testing.T) {
	cfg := domain.DefaultKaiserConfig()
	pMock := new(presenter.MockKaiserPresenter)
	pMock.On("ActionLogOutput", mock.Anything).Return(`[]`)
	gameMock := new(interfaces.MockKaiserGame)
	gameMock.On("GetConfig").Return(cfg)

	ki := usecase.NewKaiserInteractor(gameMock, pMock)
	assert.Equal(t, cfg, ki.GetConfig())
	assert.Equal(t, `[]`, ki.ActionLog())
}

// **CUI のヒントはここを通る。**presenter の HintOutput に素通しする (#4938)。
func TestKaiserInteractor_Hint(t *testing.T) {
	pMock := new(presenter.MockKaiserPresenter)
	pMock.On("HintOutput", mock.Anything).Return("advice")
	gameMock := new(interfaces.MockKaiserGame)

	ki := usecase.NewKaiserInteractor(gameMock, pMock)
	assert.Equal(t, "advice", ki.Hint())
	pMock.AssertCalled(t, "HintOutput", gameMock)
}

func TestKaiserInteractor_SnapshotAndRestore(t *testing.T) {
	pMock := new(presenter.MockKaiserPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(kaiserMockOutput)

	g := domain.NewDefaultKaiser()
	g.Reset()
	ki := usecase.NewKaiserInteractor(g, pMock)
	data, err := ki.Snapshot()
	assert.NoError(t, err)

	restored, err := usecase.RestoreKaiserInteractor(data, pMock)
	assert.NoError(t, err)
	assert.Equal(t, g.GetConfig(), restored.GetConfig())

	_, err = usecase.RestoreKaiserInteractor([]byte(`{`), pMock)
	assert.Error(t, err)
}
