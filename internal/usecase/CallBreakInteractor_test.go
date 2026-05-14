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

func TestNewCallBreakInteractor_NilGuards(t *testing.T) {
	spMock := new(presenter.MockCallBreakPresenter)

	t.Run("panics when cb is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "CallBreakInteractor: cb must not be nil", func() {
			usecase.NewCallBreakInteractor(nil, spMock)
		})
	})

	t.Run("panics when sp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockCallBreakGame)
		assert.PanicsWithValue(t, "CallBreakInteractor: sp must not be nil", func() {
			usecase.NewCallBreakInteractor(gameMock, nil)
		})
	})
}

func TestCallBreakInteractor_Reset(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("stays in bid phase", func(t *testing.T) {
		spMock := new(presenter.MockCallBreakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCallBreakGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.CallBreakPhaseBid)
		gameMock.On("IsHumanBidTurn").Return(true)

		ci := usecase.NewCallBreakInteractor(gameMock, spMock)
		assert.Equal(t, mockOutput, ci.Reset())
		gameMock.AssertCalled(t, "Reset")
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("transitions to play phase and runs CPU turns", func(t *testing.T) {
		spMock := new(presenter.MockCallBreakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCallBreakGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.CallBreakPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewCallBreakInteractor(gameMock, spMock)
		assert.Equal(t, mockOutput, ci.Reset())
		gameMock.AssertCalled(t, "Reset")
	})
}

func TestCallBreakInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("sets config then resets", func(t *testing.T) {
		spMock := new(presenter.MockCallBreakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCallBreakGame)
		cfg := domain.CallBreakConfig{CpuDifficulty: domain.CallBreakCpuDifficultyHard, MaxRounds: 3}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.CallBreakPhaseBid)
		gameMock.On("IsHumanBidTurn").Return(true)

		ci := usecase.NewCallBreakInteractor(gameMock, spMock)
		assert.Equal(t, mockOutput, ci.ResetWithConfig(cfg))
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})
}

func TestCallBreakInteractor_ResetWithConfig_ValidationError(t *testing.T) {
	spMock := new(presenter.MockCallBreakPresenter)
	gameMock := new(interfaces.MockCallBreakGame)
	spMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

	ci := usecase.NewCallBreakInteractor(gameMock, spMock)
	cfg := domain.CallBreakConfig{CpuDifficulty: domain.CallBreakCpuDifficulty(-1), MaxRounds: 5}
	assert.Equal(t, "validation error", ci.ResetWithConfig(cfg))
	gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
}

func TestCallBreakInteractor_Bid(t *testing.T) {
	mockOutput := `{"phase":1}`

	t.Run("game ended returns output without bidding", func(t *testing.T) {
		spMock := new(presenter.MockCallBreakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCallBreakGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewCallBreakInteractor(gameMock, spMock)
		assert.Equal(t, mockOutput, ci.Bid(3))
		gameMock.AssertNotCalled(t, "PlayerBid")
	})

	t.Run("bid error returned to presenter", func(t *testing.T) {
		bidErr := errors.New("invalid bid")
		spMock := new(presenter.MockCallBreakPresenter)
		spMock.On("Output", mock.Anything, bidErr).Return(mockOutput)
		gameMock := new(interfaces.MockCallBreakGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerBid", 3).Return(bidErr)

		ci := usecase.NewCallBreakInteractor(gameMock, spMock)
		assert.Equal(t, mockOutput, ci.Bid(3))
		gameMock.AssertNotCalled(t, "CpuBid")
	})

	t.Run("valid bid runs CPU bids then stays in bid phase", func(t *testing.T) {
		spMock := new(presenter.MockCallBreakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCallBreakGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerBid", 3).Return(nil)
		gameMock.On("GetPhase").Return(domain.CallBreakPhaseBid)
		gameMock.On("IsHumanBidTurn").Return(false).Once()
		gameMock.On("CpuBid").Return()
		gameMock.On("IsHumanBidTurn").Return(true)

		ci := usecase.NewCallBreakInteractor(gameMock, spMock)
		assert.Equal(t, mockOutput, ci.Bid(3))
		gameMock.AssertCalled(t, "CpuBid")
	})

	t.Run("valid bid transitions to play phase and runs CPU turns", func(t *testing.T) {
		spMock := new(presenter.MockCallBreakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCallBreakGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerBid", 5).Return(nil)
		gameMock.On("GetPhase").Return(domain.CallBreakPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewCallBreakInteractor(gameMock, spMock)
		assert.Equal(t, mockOutput, ci.Bid(5))
	})
}

func TestCallBreakInteractor_Play(t *testing.T) {
	mockOutput := `{"phase":1}`

	t.Run("game ended returns output without playing", func(t *testing.T) {
		spMock := new(presenter.MockCallBreakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCallBreakGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewCallBreakInteractor(gameMock, spMock)
		assert.Equal(t, mockOutput, ci.Play(0))
		gameMock.AssertNotCalled(t, "PlayerPlay")
	})

	t.Run("not human turn returns output without playing", func(t *testing.T) {
		spMock := new(presenter.MockCallBreakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCallBreakGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		ci := usecase.NewCallBreakInteractor(gameMock, spMock)
		assert.Equal(t, mockOutput, ci.Play(0))
		gameMock.AssertNotCalled(t, "PlayerPlay")
	})

	t.Run("play error returned to presenter", func(t *testing.T) {
		playErr := errors.New("invalid play")
		spMock := new(presenter.MockCallBreakPresenter)
		spMock.On("Output", mock.Anything, playErr).Return(mockOutput)
		gameMock := new(interfaces.MockCallBreakGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerPlay", 0).Return(playErr)

		ci := usecase.NewCallBreakInteractor(gameMock, spMock)
		assert.Equal(t, mockOutput, ci.Play(0))
	})

	t.Run("valid play runs CPU turns", func(t *testing.T) {
		spMock := new(presenter.MockCallBreakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCallBreakGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerPlay", 2).Return(nil)
		gameMock.On("GetPhase").Return(domain.CallBreakPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewCallBreakInteractor(gameMock, spMock)
		assert.Equal(t, mockOutput, ci.Play(2))
		gameMock.AssertCalled(t, "PlayerPlay", 2)
	})
}

func TestCallBreakInteractor_NextTrick(t *testing.T) {
	mockOutput := `{"phase":1}`
	spMock := new(presenter.MockCallBreakPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockCallBreakGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.CallBreakPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewCallBreakInteractor(gameMock, spMock)
	assert.Equal(t, mockOutput, ci.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestCallBreakInteractor_NextRound(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ends after scoring", func(t *testing.T) {
		spMock := new(presenter.MockCallBreakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCallBreakGame)
		gameMock.On("ScoreRound").Return()
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewCallBreakInteractor(gameMock, spMock)
		assert.Equal(t, mockOutput, ci.NextRound())
		gameMock.AssertCalled(t, "ScoreRound")
		gameMock.AssertNotCalled(t, "NextRound")
	})

	t.Run("game continues to next round", func(t *testing.T) {
		spMock := new(presenter.MockCallBreakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCallBreakGame)
		gameMock.On("ScoreRound").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextRound").Return()
		gameMock.On("GetPhase").Return(domain.CallBreakPhaseBid)
		gameMock.On("IsHumanBidTurn").Return(true)

		ci := usecase.NewCallBreakInteractor(gameMock, spMock)
		assert.Equal(t, mockOutput, ci.NextRound())
		gameMock.AssertCalled(t, "NextRound")
	})
}

func TestCallBreakInteractor_GetConfig(t *testing.T) {
	expected := domain.CallBreakConfig{CpuDifficulty: domain.CallBreakCpuDifficultyHard, MaxRounds: 7}
	spMock := new(presenter.MockCallBreakPresenter)
	gameMock := new(interfaces.MockCallBreakGame)
	gameMock.On("GetConfig").Return(expected)

	ci := usecase.NewCallBreakInteractor(gameMock, spMock)
	assert.Equal(t, expected, ci.GetConfig())
}

func TestCallBreakInteractor_Hint(t *testing.T) {
	spMock := new(presenter.MockCallBreakPresenter)
	gameMock := new(interfaces.MockCallBreakGame)
	spMock.On("HintOutput", gameMock).Return(`{"hint":{"cardIndex":0}}`)

	ci := usecase.NewCallBreakInteractor(gameMock, spMock)
	assert.Equal(t, `{"hint":{"cardIndex":0}}`, ci.Hint())
	spMock.AssertExpectations(t)
}

func TestCallBreakInteractor_ActionLog(t *testing.T) {
	spMock := new(presenter.MockCallBreakPresenter)
	gameMock := new(interfaces.MockCallBreakGame)
	spMock.On("ActionLogOutput", gameMock).Return(`{"entries":[]}`)

	ci := usecase.NewCallBreakInteractor(gameMock, spMock)
	assert.Equal(t, `{"entries":[]}`, ci.ActionLog())
}

func TestCallBreakInteractor_RunCpuTurns_AllStopPaths(t *testing.T) {
	mockOutput := `{"phase":1}`

	t.Run("stops on game end", func(t *testing.T) {
		spMock := new(presenter.MockCallBreakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCallBreakGame)
		gameMock.On("NextTrick").Return()
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewCallBreakInteractor(gameMock, spMock)
		ci.NextTrick()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("CpuPlay then ResolveTrick continues with next trick", func(t *testing.T) {
		spMock := new(presenter.MockCallBreakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCallBreakGame)
		gameMock.On("NextTrick").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.CallBreakPhasePlay).Once()
		gameMock.On("IsHumanTurn").Return(false).Once()
		gameMock.On("CpuPlay").Return()
		gameMock.On("GetPhase").Return(domain.CallBreakPhaseTrickEnd).Once()
		gameMock.On("ResolveTrick").Return()
		gameMock.On("GetPhase").Return(domain.CallBreakPhaseRoundEnd)

		ci := usecase.NewCallBreakInteractor(gameMock, spMock)
		ci.NextTrick()
		gameMock.AssertCalled(t, "ResolveTrick")
	})

	t.Run("ends on GameEnd phase after CpuPlay", func(t *testing.T) {
		spMock := new(presenter.MockCallBreakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCallBreakGame)
		gameMock.On("NextTrick").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.CallBreakPhaseGameEnd)

		ci := usecase.NewCallBreakInteractor(gameMock, spMock)
		ci.NextTrick()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})
}

func TestCallBreakInteractor_RunCpuBids_StopPaths(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("stops when not bid phase", func(t *testing.T) {
		spMock := new(presenter.MockCallBreakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCallBreakGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.CallBreakPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewCallBreakInteractor(gameMock, spMock)
		ci.Reset()
		gameMock.AssertNotCalled(t, "CpuBid")
	})

	t.Run("stops when game ended", func(t *testing.T) {
		spMock := new(presenter.MockCallBreakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCallBreakGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(true)
		gameMock.On("GetPhase").Return(domain.CallBreakPhaseBid)

		ci := usecase.NewCallBreakInteractor(gameMock, spMock)
		ci.Reset()
		gameMock.AssertNotCalled(t, "CpuBid")
	})

	t.Run("CPU bids until human bid turn", func(t *testing.T) {
		spMock := new(presenter.MockCallBreakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCallBreakGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.CallBreakPhaseBid)
		gameMock.On("IsHumanBidTurn").Return(false).Once()
		gameMock.On("CpuBid").Return()
		gameMock.On("IsHumanBidTurn").Return(true)

		ci := usecase.NewCallBreakInteractor(gameMock, spMock)
		ci.Reset()
		gameMock.AssertCalled(t, "CpuBid")
	})
}

func TestRestoreCallBreakInteractor_InvalidJSON(t *testing.T) {
	spMock := new(presenter.MockCallBreakPresenter)
	_, err := usecase.RestoreCallBreakInteractor([]byte("not json"), spMock)
	assert.Error(t, err)
}

func TestRestoreCallBreakInteractor_Valid(t *testing.T) {
	cb := domain.NewDefaultCallBreak()
	cb.Reset()
	data, err := cb.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	spMock := new(presenter.MockCallBreakPresenter)
	ci, err := usecase.RestoreCallBreakInteractor(data, spMock)
	assert.NoError(t, err)
	assert.NotNil(t, ci)
}
