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

func TestNewSpadesInteractor_NilGuards(t *testing.T) {
	spMock := new(presenter.MockSpadesPresenter)

	t.Run("panics when s is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "SpadesInteractor: s must not be nil", func() {
			usecase.NewSpadesInteractor(nil, spMock)
		})
	})

	t.Run("panics when sp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockSpadesGame)
		assert.PanicsWithValue(t, "SpadesInteractor: sp must not be nil", func() {
			usecase.NewSpadesInteractor(gameMock, nil)
		})
	})
}

func TestSpadesInteractor_Reset(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("reset stays in bid phase (does not run CPU turns)", func(t *testing.T) {
		spMock := new(presenter.MockSpadesPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockSpadesGame)
		gameMock.On("Reset").Return()
		// runCpuBids: not ended, phase=Bid, human bid turn → break
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.SpadesPhaseBid)
		gameMock.On("IsHumanBidTurn").Return(true)

		si := usecase.NewSpadesInteractor(gameMock, spMock)
		result := si.Reset()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "Reset")
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("reset transitions to play phase and runs CPU turns", func(t *testing.T) {
		spMock := new(presenter.MockSpadesPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockSpadesGame)
		gameMock.On("Reset").Return()
		// runCpuBids: not ended, phase != Bid → break (already in Play)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.SpadesPhasePlay)
		// runCpuTurns: human turn → break immediately
		gameMock.On("IsHumanTurn").Return(true)

		si := usecase.NewSpadesInteractor(gameMock, spMock)
		result := si.Reset()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "Reset")
	})
}

func TestSpadesInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("sets config then resets", func(t *testing.T) {
		spMock := new(presenter.MockSpadesPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockSpadesGame)
		cfg := domain.SpadesConfig{CpuDifficulty: domain.SpadesCpuDifficultyHard, PointLimit: 300, NilBonus: 100, BagPenaltyThreshold: 10}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.SpadesPhaseBid)
		gameMock.On("IsHumanBidTurn").Return(true)

		si := usecase.NewSpadesInteractor(gameMock, spMock)
		result := si.ResetWithConfig(cfg)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "SetConfig", cfg)
		gameMock.AssertCalled(t, "Reset")
	})
}

func TestSpadesInteractor_ResetWithConfig_ValidationError(t *testing.T) {
	spMock := new(presenter.MockSpadesPresenter)
	gameMock := new(interfaces.MockSpadesGame)
	spMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

	si := usecase.NewSpadesInteractor(gameMock, spMock)
	cfg := domain.SpadesConfig{CpuDifficulty: domain.SpadesCpuDifficulty(-1), PointLimit: 100, NilBonus: 100, BagPenaltyThreshold: 10}
	result := si.ResetWithConfig(cfg)
	assert.Equal(t, "validation error", result)
	gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
}

func TestSpadesInteractor_Bid(t *testing.T) {
	mockOutput := `{"phase":1}`

	t.Run("game ended returns output without bidding", func(t *testing.T) {
		spMock := new(presenter.MockSpadesPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockSpadesGame)
		gameMock.On("GetGameEndFlag").Return(true)

		si := usecase.NewSpadesInteractor(gameMock, spMock)
		result := si.Bid(3)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerBid")
	})

	t.Run("bid error is returned to presenter", func(t *testing.T) {
		bidErr := errors.New("invalid bid")
		spMock := new(presenter.MockSpadesPresenter)
		spMock.On("Output", mock.Anything, bidErr).Return(mockOutput)
		gameMock := new(interfaces.MockSpadesGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerBid", 3).Return(bidErr)

		si := usecase.NewSpadesInteractor(gameMock, spMock)
		result := si.Bid(3)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "CpuBid")
	})

	t.Run("valid bid runs CPU bids then stays in bid phase", func(t *testing.T) {
		spMock := new(presenter.MockSpadesPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockSpadesGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerBid", 3).Return(nil)
		// runCpuBids: phase=Bid, not human bid turn → CpuBid
		gameMock.On("GetPhase").Return(domain.SpadesPhaseBid)
		gameMock.On("IsHumanBidTurn").Return(false).Once()
		gameMock.On("CpuBid").Return()
		// after CpuBid: human bid turn → break
		gameMock.On("IsHumanBidTurn").Return(true)

		si := usecase.NewSpadesInteractor(gameMock, spMock)
		result := si.Bid(3)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "CpuBid")
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("valid bid transitions to play phase and runs CPU turns", func(t *testing.T) {
		spMock := new(presenter.MockSpadesPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockSpadesGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerBid", 5).Return(nil)
		// runCpuBids: phase != Bid (already Play) → break
		gameMock.On("GetPhase").Return(domain.SpadesPhasePlay)
		// runCpuTurns: human turn → break
		gameMock.On("IsHumanTurn").Return(true)

		si := usecase.NewSpadesInteractor(gameMock, spMock)
		result := si.Bid(5)
		assert.Equal(t, mockOutput, result)
	})
}

func TestSpadesInteractor_Play(t *testing.T) {
	mockOutput := `{"phase":1}`

	t.Run("game ended returns output without playing", func(t *testing.T) {
		spMock := new(presenter.MockSpadesPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockSpadesGame)
		gameMock.On("GetGameEndFlag").Return(true)

		si := usecase.NewSpadesInteractor(gameMock, spMock)
		result := si.Play(0)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerPlay")
	})

	t.Run("not human turn returns output without playing", func(t *testing.T) {
		spMock := new(presenter.MockSpadesPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockSpadesGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		si := usecase.NewSpadesInteractor(gameMock, spMock)
		result := si.Play(0)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerPlay")
	})

	t.Run("play error is returned to presenter", func(t *testing.T) {
		playErr := errors.New("invalid play")
		spMock := new(presenter.MockSpadesPresenter)
		spMock.On("Output", mock.Anything, playErr).Return(mockOutput)
		gameMock := new(interfaces.MockSpadesGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerPlay", 0).Return(playErr)

		si := usecase.NewSpadesInteractor(gameMock, spMock)
		result := si.Play(0)
		assert.Equal(t, mockOutput, result)
	})

	t.Run("valid play runs CPU turns", func(t *testing.T) {
		spMock := new(presenter.MockSpadesPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockSpadesGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerPlay", 2).Return(nil)
		// runCpuTurns: human turn → break
		gameMock.On("GetPhase").Return(domain.SpadesPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		si := usecase.NewSpadesInteractor(gameMock, spMock)
		result := si.Play(2)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerPlay", 2)
	})
}

func TestSpadesInteractor_NextTrick(t *testing.T) {
	mockOutput := `{"phase":1}`

	t.Run("calls NextTrick and runs CPU turns", func(t *testing.T) {
		spMock := new(presenter.MockSpadesPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockSpadesGame)
		gameMock.On("NextTrick").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.SpadesPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		si := usecase.NewSpadesInteractor(gameMock, spMock)
		result := si.NextTrick()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "NextTrick")
	})
}

func TestSpadesInteractor_NextRound(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ends after scoring", func(t *testing.T) {
		spMock := new(presenter.MockSpadesPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockSpadesGame)
		gameMock.On("ScoreRound").Return()
		gameMock.On("GetGameEndFlag").Return(true)

		si := usecase.NewSpadesInteractor(gameMock, spMock)
		result := si.NextRound()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "ScoreRound")
		gameMock.AssertNotCalled(t, "NextRound")
	})

	t.Run("game continues to next round", func(t *testing.T) {
		spMock := new(presenter.MockSpadesPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockSpadesGame)
		gameMock.On("ScoreRound").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextRound").Return()
		// runCpuBids: phase=Bid, human bid turn → break
		gameMock.On("GetPhase").Return(domain.SpadesPhaseBid)
		gameMock.On("IsHumanBidTurn").Return(true)

		si := usecase.NewSpadesInteractor(gameMock, spMock)
		result := si.NextRound()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "ScoreRound")
		gameMock.AssertCalled(t, "NextRound")
	})
}

func TestSpadesInteractor_GetConfig(t *testing.T) {
	t.Run("returns config from game", func(t *testing.T) {
		spMock := new(presenter.MockSpadesPresenter)
		gameMock := new(interfaces.MockSpadesGame)
		expected := domain.SpadesConfig{CpuDifficulty: domain.SpadesCpuDifficultyHard, PointLimit: 300, NilBonus: 100, BagPenaltyThreshold: 10}
		gameMock.On("GetConfig").Return(expected)

		si := usecase.NewSpadesInteractor(gameMock, spMock)
		result := si.GetConfig()
		assert.Equal(t, expected, result)
	})
}

func TestSpadesInteractor_ActionLog(t *testing.T) {
	spMock := new(presenter.MockSpadesPresenter)
	gameMock := new(interfaces.MockSpadesGame)
	spMock.On("ActionLogOutput", gameMock).Return(`{"entries":[]}`)

	si := usecase.NewSpadesInteractor(gameMock, spMock)
	result := si.ActionLog()
	assert.Equal(t, `{"entries":[]}`, result)
	spMock.AssertExpectations(t)
}

func TestSpadesInteractor_RunCpuTurns(t *testing.T) {
	mockOutput := `{"phase":1}`

	t.Run("stops when game ended", func(t *testing.T) {
		spMock := new(presenter.MockSpadesPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockSpadesGame)
		gameMock.On("NextTrick").Return()
		gameMock.On("GetGameEndFlag").Return(true)

		si := usecase.NewSpadesInteractor(gameMock, spMock)
		si.NextTrick()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("stops when phase is TrickEnd", func(t *testing.T) {
		spMock := new(presenter.MockSpadesPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockSpadesGame)
		gameMock.On("NextTrick").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.SpadesPhaseTrickEnd)

		si := usecase.NewSpadesInteractor(gameMock, spMock)
		si.NextTrick()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("stops when phase is RoundEnd", func(t *testing.T) {
		spMock := new(presenter.MockSpadesPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockSpadesGame)
		gameMock.On("NextTrick").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.SpadesPhaseRoundEnd)

		si := usecase.NewSpadesInteractor(gameMock, spMock)
		si.NextTrick()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("stops when phase is GameEnd", func(t *testing.T) {
		spMock := new(presenter.MockSpadesPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockSpadesGame)
		gameMock.On("NextTrick").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.SpadesPhaseGameEnd)

		si := usecase.NewSpadesInteractor(gameMock, spMock)
		si.NextTrick()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("stops when phase is not Play (e.g. Bid)", func(t *testing.T) {
		spMock := new(presenter.MockSpadesPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockSpadesGame)
		gameMock.On("NextTrick").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.SpadesPhaseBid)

		si := usecase.NewSpadesInteractor(gameMock, spMock)
		si.NextTrick()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("CPU plays then trick ends with ResolveTrick leading to RoundEnd", func(t *testing.T) {
		spMock := new(presenter.MockSpadesPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockSpadesGame)
		gameMock.On("NextTrick").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		// First iteration: Play phase, not human → CpuPlay
		gameMock.On("GetPhase").Return(domain.SpadesPhasePlay).Once()
		gameMock.On("IsHumanTurn").Return(false).Once()
		gameMock.On("CpuPlay").Return().Once()
		// After CpuPlay: phase becomes TrickEnd → ResolveTrick
		gameMock.On("GetPhase").Return(domain.SpadesPhaseTrickEnd).Once()
		gameMock.On("ResolveTrick").Return()
		// After ResolveTrick: phase becomes RoundEnd → break
		gameMock.On("GetPhase").Return(domain.SpadesPhaseRoundEnd)

		si := usecase.NewSpadesInteractor(gameMock, spMock)
		si.NextTrick()
		gameMock.AssertCalled(t, "CpuPlay")
		gameMock.AssertCalled(t, "ResolveTrick")
		gameMock.AssertNumberOfCalls(t, "NextTrick", 1)
	})

	t.Run("CPU plays then trick ends with ResolveTrick then continues", func(t *testing.T) {
		spMock := new(presenter.MockSpadesPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockSpadesGame)
		gameMock.On("NextTrick").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		// First iteration: Play phase, not human → CpuPlay
		gameMock.On("GetPhase").Return(domain.SpadesPhasePlay).Once()
		gameMock.On("IsHumanTurn").Return(false).Once()
		gameMock.On("CpuPlay").Return()
		// After CpuPlay: phase becomes TrickEnd → ResolveTrick
		gameMock.On("GetPhase").Return(domain.SpadesPhaseTrickEnd).Once()
		gameMock.On("ResolveTrick").Return()
		// After ResolveTrick: phase is Play → NextTrick inside runCpuTurns
		gameMock.On("GetPhase").Return(domain.SpadesPhasePlay).Once()
		// Next iteration: phase=Play, human turn → break
		gameMock.On("GetPhase").Return(domain.SpadesPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		si := usecase.NewSpadesInteractor(gameMock, spMock)
		si.NextTrick()
		gameMock.AssertCalled(t, "CpuPlay")
		gameMock.AssertCalled(t, "ResolveTrick")
	})

	t.Run("CPU play does not trigger trick end", func(t *testing.T) {
		spMock := new(presenter.MockSpadesPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockSpadesGame)
		gameMock.On("NextTrick").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		// First iteration: Play phase, not human → CpuPlay
		gameMock.On("GetPhase").Return(domain.SpadesPhasePlay).Once()
		gameMock.On("IsHumanTurn").Return(false).Once()
		gameMock.On("CpuPlay").Return()
		// After CpuPlay: phase stays Play (not TrickEnd)
		gameMock.On("GetPhase").Return(domain.SpadesPhasePlay)
		// Next iteration: human turn → break
		gameMock.On("IsHumanTurn").Return(true)

		si := usecase.NewSpadesInteractor(gameMock, spMock)
		si.NextTrick()
		gameMock.AssertCalled(t, "CpuPlay")
		gameMock.AssertNotCalled(t, "ResolveTrick")
	})
}

func TestSpadesInteractor_RunCpuBids(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("stops when not bid phase", func(t *testing.T) {
		spMock := new(presenter.MockSpadesPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockSpadesGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.SpadesPhasePlay)
		// runCpuTurns: human turn → break
		gameMock.On("IsHumanTurn").Return(true)

		si := usecase.NewSpadesInteractor(gameMock, spMock)
		si.Reset()
		gameMock.AssertNotCalled(t, "CpuBid")
	})

	t.Run("stops when human bid turn", func(t *testing.T) {
		spMock := new(presenter.MockSpadesPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockSpadesGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.SpadesPhaseBid)
		gameMock.On("IsHumanBidTurn").Return(true)

		si := usecase.NewSpadesInteractor(gameMock, spMock)
		si.Reset()
		gameMock.AssertNotCalled(t, "CpuBid")
	})

	t.Run("stops when game ended during bids", func(t *testing.T) {
		spMock := new(presenter.MockSpadesPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockSpadesGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(true)
		// After runCpuBids exits, Reset still checks GetPhase
		gameMock.On("GetPhase").Return(domain.SpadesPhaseBid)

		si := usecase.NewSpadesInteractor(gameMock, spMock)
		si.Reset()
		gameMock.AssertNotCalled(t, "CpuBid")
	})
}
