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

func TestNewNapoleonInteractor_NilGuards(t *testing.T) {
	npMock := new(presenter.MockNapoleonPresenter)

	t.Run("panics when n is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "NapoleonInteractor: n must not be nil", func() {
			usecase.NewNapoleonInteractor(nil, npMock)
		})
	})

	t.Run("panics when np is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockNapoleonGame)
		assert.PanicsWithValue(t, "NapoleonInteractor: np must not be nil", func() {
			usecase.NewNapoleonInteractor(gameMock, nil)
		})
	})
}

func TestNapoleonInteractor_Reset(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("reset stays in bid phase", func(t *testing.T) {
		npMock := new(presenter.MockNapoleonPresenter)
		npMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockNapoleonGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.NapoleonPhaseBid)
		gameMock.On("IsHumanBidTurn").Return(true)

		ni := usecase.NewNapoleonInteractor(gameMock, npMock)
		result := ni.Reset()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "Reset")
	})

	t.Run("reset transitions to play phase and runs CPU turns", func(t *testing.T) {
		npMock := new(presenter.MockNapoleonPresenter)
		npMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockNapoleonGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.NapoleonPhasePlay)
		gameMock.On("IsHumanDeclareTurn").Return(true)
		gameMock.On("IsHumanExchangeTurn").Return(true)
		gameMock.On("IsHumanTurn").Return(true)

		ni := usecase.NewNapoleonInteractor(gameMock, npMock)
		result := ni.Reset()
		assert.Equal(t, mockOutput, result)
	})
}

func TestNapoleonInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("sets config then resets", func(t *testing.T) {
		npMock := new(presenter.MockNapoleonPresenter)
		npMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockNapoleonGame)
		cfg := domain.NapoleonConfig{CpuDifficulty: domain.NapoleonCpuDifficultyHard, MinBid: 12, PointLimit: 200}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.NapoleonPhaseBid)
		gameMock.On("IsHumanBidTurn").Return(true)

		ni := usecase.NewNapoleonInteractor(gameMock, npMock)
		result := ni.ResetWithConfig(cfg)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config returns error", func(t *testing.T) {
		npMock := new(presenter.MockNapoleonPresenter)
		npMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockNapoleonGame)
		cfg := domain.NapoleonConfig{CpuDifficulty: domain.NapoleonCpuDifficulty(99), MinBid: 12, PointLimit: 200}

		ni := usecase.NewNapoleonInteractor(gameMock, npMock)
		ni.ResetWithConfig(cfg)
		npMock.AssertCalled(t, "Output", gameMock, mock.MatchedBy(func(err error) bool {
			return err != nil
		}))
	})
}

func TestNapoleonInteractor_Bid(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("successful bid", func(t *testing.T) {
		npMock := new(presenter.MockNapoleonPresenter)
		npMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockNapoleonGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerBid", 12).Return(nil)
		gameMock.On("GetPhase").Return(domain.NapoleonPhaseBid)
		gameMock.On("IsHumanBidTurn").Return(true)

		ni := usecase.NewNapoleonInteractor(gameMock, npMock)
		result := ni.Bid(12)
		assert.Equal(t, mockOutput, result)
	})

	t.Run("bid error", func(t *testing.T) {
		npMock := new(presenter.MockNapoleonPresenter)
		npMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockNapoleonGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerBid", 5).Return(errors.New("invalid"))

		ni := usecase.NewNapoleonInteractor(gameMock, npMock)
		ni.Bid(5)
		npMock.AssertCalled(t, "Output", gameMock, mock.MatchedBy(func(err error) bool {
			return err != nil
		}))
	})

	t.Run("game ended guard", func(t *testing.T) {
		npMock := new(presenter.MockNapoleonPresenter)
		npMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockNapoleonGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ni := usecase.NewNapoleonInteractor(gameMock, npMock)
		ni.Bid(12)
		gameMock.AssertNotCalled(t, "PlayerBid", mock.Anything)
	})
}

func TestNapoleonInteractor_DeclareTrump(t *testing.T) {
	mockOutput := `{"phase":2}`

	t.Run("successful declare", func(t *testing.T) {
		npMock := new(presenter.MockNapoleonPresenter)
		npMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockNapoleonGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerDeclareTrump", 1, 3, 1).Return(nil)
		gameMock.On("GetPhase").Return(domain.NapoleonPhaseKittyExchange)
		gameMock.On("IsHumanExchangeTurn").Return(true)

		ni := usecase.NewNapoleonInteractor(gameMock, npMock)
		result := ni.DeclareTrump(1, 3, 1)
		assert.Equal(t, mockOutput, result)
	})

	t.Run("declare error", func(t *testing.T) {
		npMock := new(presenter.MockNapoleonPresenter)
		npMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockNapoleonGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerDeclareTrump", 0, 0, 0).Return(errors.New("invalid"))

		ni := usecase.NewNapoleonInteractor(gameMock, npMock)
		ni.DeclareTrump(0, 0, 0)
		npMock.AssertCalled(t, "Output", gameMock, mock.MatchedBy(func(err error) bool {
			return err != nil
		}))
	})

	t.Run("game ended guard", func(t *testing.T) {
		npMock := new(presenter.MockNapoleonPresenter)
		npMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockNapoleonGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ni := usecase.NewNapoleonInteractor(gameMock, npMock)
		ni.DeclareTrump(1, 3, 1)
		gameMock.AssertNotCalled(t, "PlayerDeclareTrump", mock.Anything, mock.Anything, mock.Anything)
	})
}

func TestNapoleonInteractor_ExchangeKitty(t *testing.T) {
	mockOutput := `{"phase":3}`

	t.Run("successful exchange", func(t *testing.T) {
		npMock := new(presenter.MockNapoleonPresenter)
		npMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockNapoleonGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerExchangeKitty", 5).Return(nil)
		gameMock.On("GetPhase").Return(domain.NapoleonPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		ni := usecase.NewNapoleonInteractor(gameMock, npMock)
		result := ni.ExchangeKitty(5)
		assert.Equal(t, mockOutput, result)
	})

	t.Run("exchange error", func(t *testing.T) {
		npMock := new(presenter.MockNapoleonPresenter)
		npMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockNapoleonGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerExchangeKitty", -1).Return(errors.New("invalid"))

		ni := usecase.NewNapoleonInteractor(gameMock, npMock)
		ni.ExchangeKitty(-1)
	})

	t.Run("game ended guard", func(t *testing.T) {
		npMock := new(presenter.MockNapoleonPresenter)
		npMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockNapoleonGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ni := usecase.NewNapoleonInteractor(gameMock, npMock)
		ni.ExchangeKitty(5)
		gameMock.AssertNotCalled(t, "PlayerExchangeKitty", mock.Anything)
	})
}

func TestNapoleonInteractor_Play(t *testing.T) {
	mockOutput := `{"phase":3}`

	t.Run("successful play", func(t *testing.T) {
		npMock := new(presenter.MockNapoleonPresenter)
		npMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockNapoleonGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerPlay", 3).Return(nil)
		gameMock.On("GetPhase").Return(domain.NapoleonPhasePlay)

		ni := usecase.NewNapoleonInteractor(gameMock, npMock)
		result := ni.Play(3)
		assert.Equal(t, mockOutput, result)
	})

	t.Run("play error", func(t *testing.T) {
		npMock := new(presenter.MockNapoleonPresenter)
		npMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockNapoleonGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerPlay", 99).Return(errors.New("invalid"))

		ni := usecase.NewNapoleonInteractor(gameMock, npMock)
		ni.Play(99)
	})

	t.Run("not human turn guard", func(t *testing.T) {
		npMock := new(presenter.MockNapoleonPresenter)
		npMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockNapoleonGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		ni := usecase.NewNapoleonInteractor(gameMock, npMock)
		ni.Play(0)
		gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
	})
}

func TestNapoleonInteractor_NextTrick(t *testing.T) {
	mockOutput := `{"phase":3}`
	npMock := new(presenter.MockNapoleonPresenter)
	npMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockNapoleonGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.NapoleonPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ni := usecase.NewNapoleonInteractor(gameMock, npMock)
	result := ni.NextTrick()
	assert.Equal(t, mockOutput, result)
	gameMock.AssertCalled(t, "NextTrick")
}

func TestNapoleonInteractor_NextRound(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("scores round and starts next round", func(t *testing.T) {
		npMock := new(presenter.MockNapoleonPresenter)
		npMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockNapoleonGame)
		gameMock.On("ScoreRound").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextRound").Return()
		gameMock.On("GetPhase").Return(domain.NapoleonPhaseBid)
		gameMock.On("IsHumanBidTurn").Return(true)

		ni := usecase.NewNapoleonInteractor(gameMock, npMock)
		result := ni.NextRound()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "ScoreRound")
		gameMock.AssertCalled(t, "NextRound")
	})

	t.Run("game ended after scoring", func(t *testing.T) {
		npMock := new(presenter.MockNapoleonPresenter)
		npMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockNapoleonGame)
		gameMock.On("ScoreRound").Return()
		gameMock.On("GetGameEndFlag").Return(true)

		ni := usecase.NewNapoleonInteractor(gameMock, npMock)
		ni.NextRound()
		gameMock.AssertNotCalled(t, "NextRound")
	})
}

func TestNapoleonInteractor_GetConfig(t *testing.T) {
	npMock := new(presenter.MockNapoleonPresenter)
	gameMock := new(interfaces.MockNapoleonGame)
	cfg := domain.DefaultNapoleonConfig()
	gameMock.On("GetConfig").Return(cfg)

	ni := usecase.NewNapoleonInteractor(gameMock, npMock)
	assert.Equal(t, cfg, ni.GetConfig())
}

func TestNapoleonInteractor_Hint(t *testing.T) {
	npMock := new(presenter.MockNapoleonPresenter)
	gameMock := new(interfaces.MockNapoleonGame)
	npMock.On("HintOutput", gameMock).Return("hint")

	ni := usecase.NewNapoleonInteractor(gameMock, npMock)
	assert.Equal(t, "hint", ni.Hint())
}

func TestNapoleonInteractor_ActionLog(t *testing.T) {
	npMock := new(presenter.MockNapoleonPresenter)
	gameMock := new(interfaces.MockNapoleonGame)
	npMock.On("ActionLogOutput", gameMock).Return("log")

	ni := usecase.NewNapoleonInteractor(gameMock, npMock)
	assert.Equal(t, "log", ni.ActionLog())
}
