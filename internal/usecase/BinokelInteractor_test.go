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

func TestNewBinokelInteractor_NilGuards(t *testing.T) {
	ppMock := new(presenter.MockBinokelPresenter)

	t.Run("panics when p is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "BinokelInteractor: p must not be nil", func() {
			usecase.NewBinokelInteractor(nil, ppMock)
		})
	})

	t.Run("panics when pp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockBinokelGame)
		assert.PanicsWithValue(t, "BinokelInteractor: pp must not be nil", func() {
			usecase.NewBinokelInteractor(gameMock, nil)
		})
	})
}

func setupBinokelInteractorMocks(phase domain.BinokelPhase) (*interfaces.MockBinokelGame, *presenter.MockBinokelPresenter) {
	ppMock := new(presenter.MockBinokelPresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return(`{"phase":0}`)
	gameMock := new(interfaces.MockBinokelGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(phase)
	gameMock.On("IsHumanBidTurn").Return(true)
	gameMock.On("IsHumanDabbTurn").Return(true)
	gameMock.On("IsHumanTurn").Return(true)
	return gameMock, ppMock
}

func TestBinokelInteractor_Reset(t *testing.T) {
	gameMock, ppMock := setupBinokelInteractorMocks(domain.BinokelPhaseBid)
	gameMock.On("Reset").Return()

	pi := usecase.NewBinokelInteractor(gameMock, ppMock)
	result := pi.Reset()
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "Reset")
}

func TestBinokelInteractor_ResetWithConfig(t *testing.T) {
	t.Run("sets config then resets", func(t *testing.T) {
		gameMock, ppMock := setupBinokelInteractorMocks(domain.BinokelPhaseBid)
		cfg := domain.BinokelConfig{CpuDifficulty: domain.BinokelCpuDifficultyHard, PointLimit: 500}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()

		pi := usecase.NewBinokelInteractor(gameMock, ppMock)
		result := pi.ResetWithConfig(cfg)
		assert.Contains(t, result, "phase")
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config returns error", func(t *testing.T) {
		gameMock, ppMock := setupBinokelInteractorMocks(domain.BinokelPhaseBid)
		invalidCfg := domain.BinokelConfig{CpuDifficulty: 99, PointLimit: 10}

		pi := usecase.NewBinokelInteractor(gameMock, ppMock)
		result := pi.ResetWithConfig(invalidCfg)
		assert.Contains(t, result, "phase")
	})
}

func TestBinokelInteractor_Bid(t *testing.T) {
	gameMock, ppMock := setupBinokelInteractorMocks(domain.BinokelPhaseBid)
	gameMock.On("PlayerBid", 150).Return(nil)

	pi := usecase.NewBinokelInteractor(gameMock, ppMock)
	result := pi.Bid(150)
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "PlayerBid", 150)
}

func TestBinokelInteractor_Bid_Error(t *testing.T) {
	gameMock, ppMock := setupBinokelInteractorMocks(domain.BinokelPhaseBid)
	gameMock.On("PlayerBid", 50).Return(domain.NewDomainError(domain.ErrInvalidAmount, "too low"))

	pi := usecase.NewBinokelInteractor(gameMock, ppMock)
	result := pi.Bid(50)
	assert.Contains(t, result, "phase")
}

func TestBinokelInteractor_DiscardToDabb(t *testing.T) {
	gameMock, ppMock := setupBinokelInteractorMocks(domain.BinokelPhaseDabb)
	gameMock.On("PlayerDiscardToDabb", []int{0, 1, 2}).Return(nil)

	pi := usecase.NewBinokelInteractor(gameMock, ppMock)
	result := pi.DiscardToDabb([]int{0, 1, 2})
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "PlayerDiscardToDabb", []int{0, 1, 2})
}

func TestBinokelInteractor_DiscardToDabb_Error(t *testing.T) {
	gameMock, ppMock := setupBinokelInteractorMocks(domain.BinokelPhaseDabb)
	gameMock.On("PlayerDiscardToDabb", []int{0}).Return(domain.NewDomainError(domain.ErrInvalidCard, "invalid"))

	pi := usecase.NewBinokelInteractor(gameMock, ppMock)
	result := pi.DiscardToDabb([]int{0})
	assert.Contains(t, result, "phase")
}

func TestBinokelInteractor_Pass(t *testing.T) {
	gameMock, ppMock := setupBinokelInteractorMocks(domain.BinokelPhaseBid)
	gameMock.On("PlayerPass").Return(nil)

	pi := usecase.NewBinokelInteractor(gameMock, ppMock)
	result := pi.Pass()
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "PlayerPass")
}

func TestBinokelInteractor_CallTrump(t *testing.T) {
	gameMock, ppMock := setupBinokelInteractorMocks(domain.BinokelPhaseTrump)
	gameMock.On("PlayerCallTrump", 3).Return(nil)

	pi := usecase.NewBinokelInteractor(gameMock, ppMock)
	result := pi.CallTrump(3)
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "PlayerCallTrump", 3)
}

func TestBinokelInteractor_ConfirmMelds(t *testing.T) {
	gameMock, ppMock := setupBinokelInteractorMocks(domain.BinokelPhaseMeld)
	gameMock.On("ConfirmMelds").Return()

	pi := usecase.NewBinokelInteractor(gameMock, ppMock)
	result := pi.ConfirmMelds()
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "ConfirmMelds")
}

func TestBinokelInteractor_Play(t *testing.T) {
	gameMock, ppMock := setupBinokelInteractorMocks(domain.BinokelPhasePlay)
	gameMock.On("PlayerPlay", 2).Return(nil)

	pi := usecase.NewBinokelInteractor(gameMock, ppMock)
	result := pi.Play(2)
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "PlayerPlay", 2)
}

func TestBinokelInteractor_NextTrick(t *testing.T) {
	gameMock, ppMock := setupBinokelInteractorMocks(domain.BinokelPhaseTrickEnd)
	gameMock.On("ResolveTrick").Return()
	gameMock.On("NextTrick").Return()

	pi := usecase.NewBinokelInteractor(gameMock, ppMock)
	result := pi.NextTrick()
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "ResolveTrick")
	gameMock.AssertCalled(t, "NextTrick")
}

func TestBinokelInteractor_NextRound(t *testing.T) {
	gameMock, ppMock := setupBinokelInteractorMocks(domain.BinokelPhaseRoundEnd)
	gameMock.On("NextRound").Return()

	pi := usecase.NewBinokelInteractor(gameMock, ppMock)
	result := pi.NextRound()
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "NextRound")
}

func TestBinokelInteractor_Hint(t *testing.T) {
	gameMock, ppMock := setupBinokelInteractorMocks(domain.BinokelPhaseBid)
	ppMock.On("HintOutput", mock.Anything).Return(`{"hint":"bid"}`)

	pi := usecase.NewBinokelInteractor(gameMock, ppMock)
	result := pi.Hint()
	assert.Contains(t, result, "hint")
}

func TestBinokelInteractor_ActionLog(t *testing.T) {
	gameMock, ppMock := setupBinokelInteractorMocks(domain.BinokelPhaseBid)
	ppMock.On("ActionLogOutput", mock.Anything).Return(`[]`)

	pi := usecase.NewBinokelInteractor(gameMock, ppMock)
	result := pi.ActionLog()
	assert.Equal(t, "[]", result)
}

func TestBinokelInteractor_GetConfig(t *testing.T) {
	gameMock, ppMock := setupBinokelInteractorMocks(domain.BinokelPhaseBid)
	cfg := domain.DefaultBinokelConfig()
	gameMock.On("GetConfig").Return(cfg)

	pi := usecase.NewBinokelInteractor(gameMock, ppMock)
	result := pi.GetConfig()
	assert.Equal(t, cfg, result)
}

func TestBinokelInteractor_GameEndGuard(t *testing.T) {
	ppMock := new(presenter.MockBinokelPresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return(`{"gameEndFlag":true}`)
	gameMock := new(interfaces.MockBinokelGame)
	gameMock.On("GetGameEndFlag").Return(true)

	pi := usecase.NewBinokelInteractor(gameMock, ppMock)

	assert.Contains(t, pi.Bid(150), "gameEndFlag")
	assert.Contains(t, pi.Pass(), "gameEndFlag")
	assert.Contains(t, pi.DiscardToDabb([]int{0, 1, 2}), "gameEndFlag")
	assert.Contains(t, pi.CallTrump(1), "gameEndFlag")
	assert.Contains(t, pi.NextRound(), "gameEndFlag")
	assert.Contains(t, pi.ConfirmMelds(), "gameEndFlag")
}

func TestBinokelInteractor_RunCpuBids_CpuDabb(t *testing.T) {
	ppMock := new(presenter.MockBinokelPresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return(`{"phase":2}`)
	gameMock := new(interfaces.MockBinokelGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("Reset").Return()
	gameMock.On("GetPhase").Return(domain.BinokelPhaseDabb).Once()
	gameMock.On("IsHumanDabbTurn").Return(false).Once()
	gameMock.On("CpuDiscardToDabb").Return().Once()
	gameMock.On("GetPhase").Return(domain.BinokelPhaseTrump).Once()
	gameMock.On("IsHumanTurn").Return(true).Once()

	pi := usecase.NewBinokelInteractor(gameMock, ppMock)
	_ = pi.Reset()
	gameMock.AssertCalled(t, "CpuDiscardToDabb")
}

func TestBinokelInteractor_Play_Error(t *testing.T) {
	gameMock, ppMock := setupBinokelInteractorMocks(domain.BinokelPhasePlay)
	gameMock.On("PlayerPlay", 99).Return(domain.NewDomainError(domain.ErrInvalidCard, "invalid"))

	pi := usecase.NewBinokelInteractor(gameMock, ppMock)
	result := pi.Play(99)
	assert.Contains(t, result, "phase")
}

func TestBinokelInteractor_Play_NotHumanTurn(t *testing.T) {
	ppMock := new(presenter.MockBinokelPresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return(`{"notHuman":true}`)
	gameMock := new(interfaces.MockBinokelGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	pi := usecase.NewBinokelInteractor(gameMock, ppMock)
	result := pi.Play(0)
	assert.Contains(t, result, "notHuman")
}

func TestBinokelInteractor_NextTrick_GameEnd(t *testing.T) {
	ppMock := new(presenter.MockBinokelPresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return(`{"gameEnd":true}`)
	gameMock := new(interfaces.MockBinokelGame)
	gameMock.On("ResolveTrick").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	pi := usecase.NewBinokelInteractor(gameMock, ppMock)
	result := pi.NextTrick()
	assert.Contains(t, result, "gameEnd")
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestRestoreBinokelInteractor(t *testing.T) {
	players := []*domain.BinokelPlayer{
		domain.NewBinokelPlayer(true),
		domain.NewBinokelPlayer(false),
		domain.NewBinokelPlayer(false),
	}
	p := domain.NewBinokel(domain.NewTrumpCardsBinokel(), players, domain.DefaultBinokelConfig())
	ppMock := new(presenter.MockBinokelPresenter)
	pi := usecase.NewBinokelInteractor(p, ppMock)

	data, err := pi.Snapshot()
	assert.NoError(t, err)

	restored, err := usecase.RestoreBinokelInteractor(data, ppMock)
	assert.NoError(t, err)
	assert.NotNil(t, restored)
}

func TestRestoreBinokelInteractor_InvalidJSON(t *testing.T) {
	ppMock := new(presenter.MockBinokelPresenter)
	_, err := usecase.RestoreBinokelInteractor([]byte("invalid"), ppMock)
	assert.Error(t, err)
}
