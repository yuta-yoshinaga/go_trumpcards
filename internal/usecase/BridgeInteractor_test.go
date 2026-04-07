//go:build test

package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func TestNewBridgeInteractor_NilGuards(t *testing.T) {
	bpMock := new(presenter.MockBridgePresenter)

	t.Run("panics when b is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "BridgeInteractor: b must not be nil", func() {
			usecase.NewBridgeInteractor(nil, bpMock)
		})
	})

	t.Run("panics when bp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockBridgeGame)
		assert.PanicsWithValue(t, "BridgeInteractor: bp must not be nil", func() {
			usecase.NewBridgeInteractor(gameMock, nil)
		})
	})
}

func setupBridgeInteractorMocks(phase domain.BridgePhase) (*interfaces.MockBridgeGame, *presenter.MockBridgePresenter) {
	bpMock := new(presenter.MockBridgePresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(`{"phase":0}`)
	gameMock := new(interfaces.MockBridgeGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(phase)
	gameMock.On("IsHumanBidTurn").Return(true)
	gameMock.On("IsHumanTurn").Return(true)
	return gameMock, bpMock
}

func TestBridgeInteractor_Reset(t *testing.T) {
	t.Run("resets and stays in bid phase", func(t *testing.T) {
		gameMock, bpMock := setupBridgeInteractorMocks(domain.BridgePhaseBid)
		gameMock.On("Reset").Return()

		bi := usecase.NewBridgeInteractor(gameMock, bpMock)
		result := bi.Reset()
		assert.Contains(t, result, "phase")
		gameMock.AssertCalled(t, "Reset")
	})
}

func TestBridgeInteractor_ResetWithConfig(t *testing.T) {
	t.Run("sets config then resets", func(t *testing.T) {
		gameMock, bpMock := setupBridgeInteractorMocks(domain.BridgePhaseBid)
		cfg := domain.BridgeConfig{CpuDifficulty: domain.BridgeCpuDifficultyHard}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()

		bi := usecase.NewBridgeInteractor(gameMock, bpMock)
		result := bi.ResetWithConfig(cfg)
		assert.Contains(t, result, "phase")
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config returns error", func(t *testing.T) {
		gameMock, bpMock := setupBridgeInteractorMocks(domain.BridgePhaseBid)
		invalidCfg := domain.BridgeConfig{CpuDifficulty: 99}

		bi := usecase.NewBridgeInteractor(gameMock, bpMock)
		result := bi.ResetWithConfig(invalidCfg)
		assert.Contains(t, result, "phase")
	})
}

func TestBridgeInteractor_Bid(t *testing.T) {
	t.Run("successful bid", func(t *testing.T) {
		gameMock, bpMock := setupBridgeInteractorMocks(domain.BridgePhaseBid)
		gameMock.On("PlayerBid", 1, 1, 1).Return(nil)

		bi := usecase.NewBridgeInteractor(gameMock, bpMock)
		result := bi.Bid(1, 1, 1)
		assert.Contains(t, result, "phase")
	})

	t.Run("bid error", func(t *testing.T) {
		gameMock, bpMock := setupBridgeInteractorMocks(domain.BridgePhaseBid)
		gameMock.On("PlayerBid", 1, 0, 1).Return(domain.ErrInvalidPlay)

		bi := usecase.NewBridgeInteractor(gameMock, bpMock)
		result := bi.Bid(1, 0, 1)
		assert.Contains(t, result, "phase")
	})

	t.Run("game ended", func(t *testing.T) {
		bpMock := new(presenter.MockBridgePresenter)
		bpMock.On("Output", mock.Anything, mock.Anything).Return(`{"gameEndFlag":true}`)
		gameMock := new(interfaces.MockBridgeGame)
		gameMock.On("GetGameEndFlag").Return(true)

		bi := usecase.NewBridgeInteractor(gameMock, bpMock)
		result := bi.Bid(0, 0, 0)
		assert.Contains(t, result, "gameEndFlag")
	})
}

func TestBridgeInteractor_Play(t *testing.T) {
	t.Run("successful play", func(t *testing.T) {
		gameMock, bpMock := setupBridgeInteractorMocks(domain.BridgePhasePlay)
		gameMock.On("PlayerPlay", 0).Return(nil)

		bi := usecase.NewBridgeInteractor(gameMock, bpMock)
		result := bi.Play(0)
		assert.Contains(t, result, "phase")
	})

	t.Run("play error", func(t *testing.T) {
		gameMock, bpMock := setupBridgeInteractorMocks(domain.BridgePhasePlay)
		gameMock.On("PlayerPlay", -1).Return(domain.ErrInvalidCard)

		bi := usecase.NewBridgeInteractor(gameMock, bpMock)
		result := bi.Play(-1)
		assert.Contains(t, result, "phase")
	})
}

func TestBridgeInteractor_NextTrick(t *testing.T) {
	gameMock, bpMock := setupBridgeInteractorMocks(domain.BridgePhasePlay)
	gameMock.On("ResolveTrick").Return()
	gameMock.On("NextTrick").Return()

	bi := usecase.NewBridgeInteractor(gameMock, bpMock)
	result := bi.NextTrick()
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestBridgeInteractor_NextRound(t *testing.T) {
	gameMock, bpMock := setupBridgeInteractorMocks(domain.BridgePhaseBid)
	gameMock.On("ScoreRound").Return()
	gameMock.On("NextRound").Return()

	bi := usecase.NewBridgeInteractor(gameMock, bpMock)
	result := bi.NextRound()
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "ScoreRound")
}

func TestBridgeInteractor_GetConfig(t *testing.T) {
	gameMock, bpMock := setupBridgeInteractorMocks(domain.BridgePhaseBid)
	gameMock.On("GetConfig").Return(domain.DefaultBridgeConfig())

	bi := usecase.NewBridgeInteractor(gameMock, bpMock)
	cfg := bi.GetConfig()
	assert.Equal(t, domain.BridgeCpuDifficultyNormal, cfg.CpuDifficulty)
}

func TestBridgeInteractor_Hint(t *testing.T) {
	gameMock, bpMock := setupBridgeInteractorMocks(domain.BridgePhaseBid)
	bpMock.On("HintOutput", mock.Anything).Return(`{"hint":{}}`)

	bi := usecase.NewBridgeInteractor(gameMock, bpMock)
	result := bi.Hint()
	assert.Contains(t, result, "hint")
}

func TestBridgeInteractor_ActionLog(t *testing.T) {
	gameMock, bpMock := setupBridgeInteractorMocks(domain.BridgePhaseBid)
	bpMock.On("ActionLogOutput", mock.Anything).Return(`{"log":[]}`)

	bi := usecase.NewBridgeInteractor(gameMock, bpMock)
	result := bi.ActionLog()
	assert.Contains(t, result, "log")
}

func TestBridgeInteractor_Play_NotHumanTurn(t *testing.T) {
	bpMock := new(presenter.MockBridgePresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(`{"notHuman":true}`)
	gameMock := new(interfaces.MockBridgeGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	bi := usecase.NewBridgeInteractor(gameMock, bpMock)
	result := bi.Play(0)
	assert.Contains(t, result, "notHuman")
}

func TestBridgeInteractor_NextTrick_GameEnd(t *testing.T) {
	bpMock := new(presenter.MockBridgePresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(`{"gameEnd":true}`)
	gameMock := new(interfaces.MockBridgeGame)
	gameMock.On("ResolveTrick").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	bi := usecase.NewBridgeInteractor(gameMock, bpMock)
	result := bi.NextTrick()
	assert.Contains(t, result, "gameEnd")
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestBridgeInteractor_NextRound_GameEnd(t *testing.T) {
	bpMock := new(presenter.MockBridgePresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(`{"gameEnd":true}`)
	gameMock := new(interfaces.MockBridgeGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	bi := usecase.NewBridgeInteractor(gameMock, bpMock)
	result := bi.NextRound()
	assert.Contains(t, result, "gameEnd")
	gameMock.AssertCalled(t, "ScoreRound")
}

func TestRestoreBridgeInteractor(t *testing.T) {
	players := []*domain.BridgePlayer{
		domain.NewBridgePlayer(true, 0),
		domain.NewBridgePlayer(false, 1),
		domain.NewBridgePlayer(false, 0),
		domain.NewBridgePlayer(false, 1),
	}
	b := domain.NewBridge(domain.NewTrumpCards(0), players, domain.DefaultBridgeConfig())
	bpMock := new(presenter.MockBridgePresenter)
	bi := usecase.NewBridgeInteractor(b, bpMock)

	data, err := bi.Snapshot()
	assert.NoError(t, err)

	restored, err := usecase.RestoreBridgeInteractor(data, bpMock)
	assert.NoError(t, err)
	assert.NotNil(t, restored)
}

func TestRestoreBridgeInteractor_InvalidJSON(t *testing.T) {
	bpMock := new(presenter.MockBridgePresenter)
	_, err := usecase.RestoreBridgeInteractor([]byte("invalid"), bpMock)
	assert.Error(t, err)
}
