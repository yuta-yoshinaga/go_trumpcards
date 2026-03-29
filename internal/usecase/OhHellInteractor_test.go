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

func TestNewOhHellInteractor_NilGuards(t *testing.T) {
	opMock := new(presenter.MockOhHellPresenter)

	t.Run("panics when o is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "OhHellInteractor: o must not be nil", func() {
			usecase.NewOhHellInteractor(nil, opMock)
		})
	})

	t.Run("panics when op is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockOhHellGame)
		assert.PanicsWithValue(t, "OhHellInteractor: op must not be nil", func() {
			usecase.NewOhHellInteractor(gameMock, nil)
		})
	})
}

func TestOhHellInteractor_Reset(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("reset stays in bid phase", func(t *testing.T) {
		opMock := new(presenter.MockOhHellPresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockOhHellGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.OhHellPhaseBid)
		gameMock.On("IsHumanBidTurn").Return(true)

		oi := usecase.NewOhHellInteractor(gameMock, opMock)
		result := oi.Reset()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "Reset")
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("reset transitions to play phase and runs CPU turns", func(t *testing.T) {
		opMock := new(presenter.MockOhHellPresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockOhHellGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.OhHellPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		oi := usecase.NewOhHellInteractor(gameMock, opMock)
		result := oi.Reset()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "Reset")
	})
}

func TestOhHellInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("sets config then resets", func(t *testing.T) {
		opMock := new(presenter.MockOhHellPresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockOhHellGame)
		cfg := domain.OhHellConfig{CpuDifficulty: domain.OhHellCpuDifficultyHard, MaxHandSize: 10}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.OhHellPhaseBid)
		gameMock.On("IsHumanBidTurn").Return(true)

		oi := usecase.NewOhHellInteractor(gameMock, opMock)
		result := oi.ResetWithConfig(cfg)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})
}

func TestOhHellInteractor_ResetWithConfig_ValidationError(t *testing.T) {
	opMock := new(presenter.MockOhHellPresenter)
	gameMock := new(interfaces.MockOhHellGame)
	opMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

	oi := usecase.NewOhHellInteractor(gameMock, opMock)
	cfg := domain.OhHellConfig{CpuDifficulty: domain.OhHellCpuDifficulty(-1), MaxHandSize: 10}
	result := oi.ResetWithConfig(cfg)
	assert.Equal(t, "validation error", result)
	gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
}

func TestOhHellInteractor_Bid(t *testing.T) {
	mockOutput := `{"phase":1}`

	t.Run("game ended returns output without bidding", func(t *testing.T) {
		opMock := new(presenter.MockOhHellPresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockOhHellGame)
		gameMock.On("GetGameEndFlag").Return(true)

		oi := usecase.NewOhHellInteractor(gameMock, opMock)
		result := oi.Bid(3)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerBid")
	})

	t.Run("bid error returns error output", func(t *testing.T) {
		opMock := new(presenter.MockOhHellPresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return("bid error")
		gameMock := new(interfaces.MockOhHellGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerBid", 99).Return(errors.New("invalid bid"))

		oi := usecase.NewOhHellInteractor(gameMock, opMock)
		result := oi.Bid(99)
		assert.Equal(t, "bid error", result)
	})

	t.Run("successful bid runs CPU bids", func(t *testing.T) {
		opMock := new(presenter.MockOhHellPresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockOhHellGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerBid", 3).Return(nil)
		gameMock.On("GetPhase").Return(domain.OhHellPhaseBid)
		gameMock.On("IsHumanBidTurn").Return(true)

		oi := usecase.NewOhHellInteractor(gameMock, opMock)
		result := oi.Bid(3)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerBid", 3)
	})
}

func TestOhHellInteractor_Play(t *testing.T) {
	mockOutput := `{"phase":1}`

	t.Run("game ended returns output without playing", func(t *testing.T) {
		opMock := new(presenter.MockOhHellPresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockOhHellGame)
		gameMock.On("GetGameEndFlag").Return(true)

		oi := usecase.NewOhHellInteractor(gameMock, opMock)
		result := oi.Play(0)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerPlay")
	})

	t.Run("not human turn returns output", func(t *testing.T) {
		opMock := new(presenter.MockOhHellPresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockOhHellGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		oi := usecase.NewOhHellInteractor(gameMock, opMock)
		result := oi.Play(0)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerPlay")
	})

	t.Run("play error returns error output", func(t *testing.T) {
		opMock := new(presenter.MockOhHellPresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return("play error")
		gameMock := new(interfaces.MockOhHellGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerPlay", 99).Return(errors.New("invalid"))

		oi := usecase.NewOhHellInteractor(gameMock, opMock)
		result := oi.Play(99)
		assert.Equal(t, "play error", result)
	})

	t.Run("successful play runs CPU turns", func(t *testing.T) {
		opMock := new(presenter.MockOhHellPresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockOhHellGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerPlay", 0).Return(nil)
		// runCpuTurns: phase=TrickEnd → break
		gameMock.On("GetPhase").Return(domain.OhHellPhaseTrickEnd)

		oi := usecase.NewOhHellInteractor(gameMock, opMock)
		result := oi.Play(0)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerPlay", 0)
	})
}

func TestOhHellInteractor_NextTrick(t *testing.T) {
	mockOutput := `{"phase":1}`

	opMock := new(presenter.MockOhHellPresenter)
	opMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockOhHellGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.OhHellPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	oi := usecase.NewOhHellInteractor(gameMock, opMock)
	result := oi.NextTrick()
	assert.Equal(t, mockOutput, result)
	gameMock.AssertCalled(t, "NextTrick")
}

func TestOhHellInteractor_NextRound(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("scores round, then starts next", func(t *testing.T) {
		opMock := new(presenter.MockOhHellPresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockOhHellGame)
		gameMock.On("ScoreRound").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextRound").Return()
		gameMock.On("GetPhase").Return(domain.OhHellPhaseBid)
		gameMock.On("IsHumanBidTurn").Return(true)

		oi := usecase.NewOhHellInteractor(gameMock, opMock)
		result := oi.NextRound()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "ScoreRound")
		gameMock.AssertCalled(t, "NextRound")
	})

	t.Run("game ends after score", func(t *testing.T) {
		opMock := new(presenter.MockOhHellPresenter)
		opMock.On("Output", mock.Anything, mock.Anything).Return(`{"gameEndFlag":true}`)
		gameMock := new(interfaces.MockOhHellGame)
		gameMock.On("ScoreRound").Return()
		gameMock.On("GetGameEndFlag").Return(true)

		oi := usecase.NewOhHellInteractor(gameMock, opMock)
		result := oi.NextRound()
		assert.Equal(t, `{"gameEndFlag":true}`, result)
		gameMock.AssertNotCalled(t, "NextRound")
	})
}

func TestOhHellInteractor_GetConfig(t *testing.T) {
	opMock := new(presenter.MockOhHellPresenter)
	gameMock := new(interfaces.MockOhHellGame)
	cfg := domain.DefaultOhHellConfig()
	gameMock.On("GetConfig").Return(cfg)

	oi := usecase.NewOhHellInteractor(gameMock, opMock)
	assert.Equal(t, cfg, oi.GetConfig())
}

func TestOhHellInteractor_Hint(t *testing.T) {
	opMock := new(presenter.MockOhHellPresenter)
	gameMock := new(interfaces.MockOhHellGame)
	opMock.On("HintOutput", gameMock).Return("hint result")

	oi := usecase.NewOhHellInteractor(gameMock, opMock)
	assert.Equal(t, "hint result", oi.Hint())
}

func TestOhHellInteractor_ActionLog(t *testing.T) {
	opMock := new(presenter.MockOhHellPresenter)
	gameMock := new(interfaces.MockOhHellGame)
	opMock.On("ActionLogOutput", gameMock).Return("log result")

	oi := usecase.NewOhHellInteractor(gameMock, opMock)
	assert.Equal(t, "log result", oi.ActionLog())
}
