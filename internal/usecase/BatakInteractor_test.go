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

func TestNewBatakInteractor_NilGuards(t *testing.T) {
	spMock := new(presenter.MockBatakPresenter)

	t.Run("panics when cb is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "BatakInteractor: cb must not be nil", func() {
			usecase.NewBatakInteractor(nil, spMock)
		})
	})

	t.Run("panics when sp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockBatakGame)
		assert.PanicsWithValue(t, "BatakInteractor: sp must not be nil", func() {
			usecase.NewBatakInteractor(gameMock, nil)
		})
	})
}

func TestBatakInteractor_Reset(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("stays in bid phase", func(t *testing.T) {
		spMock := new(presenter.MockBatakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBatakGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.BatakPhaseBid)
		gameMock.On("IsHumanBidTurn").Return(true)

		ci := usecase.NewBatakInteractor(gameMock, spMock)
		assert.Equal(t, mockOutput, ci.Reset())
		gameMock.AssertCalled(t, "Reset")
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("transitions to play phase and runs CPU turns", func(t *testing.T) {
		spMock := new(presenter.MockBatakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBatakGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.BatakPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewBatakInteractor(gameMock, spMock)
		assert.Equal(t, mockOutput, ci.Reset())
		gameMock.AssertCalled(t, "Reset")
	})
}

func TestBatakInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("sets config then resets", func(t *testing.T) {
		spMock := new(presenter.MockBatakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBatakGame)
		cfg := domain.BatakConfig{CpuDifficulty: domain.BatakCpuDifficultyHard, MaxRounds: 3}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.BatakPhaseBid)
		gameMock.On("IsHumanBidTurn").Return(true)

		ci := usecase.NewBatakInteractor(gameMock, spMock)
		assert.Equal(t, mockOutput, ci.ResetWithConfig(cfg))
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})
}

func TestBatakInteractor_ResetWithConfig_ValidationError(t *testing.T) {
	spMock := new(presenter.MockBatakPresenter)
	gameMock := new(interfaces.MockBatakGame)
	spMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

	ci := usecase.NewBatakInteractor(gameMock, spMock)
	cfg := domain.BatakConfig{CpuDifficulty: domain.BatakCpuDifficulty(-1), MaxRounds: 5}
	assert.Equal(t, "validation error", ci.ResetWithConfig(cfg))
	gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
}

func TestBatakInteractor_Bid(t *testing.T) {
	mockOutput := `{"phase":1}`

	t.Run("game ended returns output without bidding", func(t *testing.T) {
		spMock := new(presenter.MockBatakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBatakGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewBatakInteractor(gameMock, spMock)
		assert.Equal(t, mockOutput, ci.Bid(3))
		gameMock.AssertNotCalled(t, "PlayerBid")
	})

	t.Run("bid error returned to presenter", func(t *testing.T) {
		bidErr := errors.New("invalid bid")
		spMock := new(presenter.MockBatakPresenter)
		spMock.On("Output", mock.Anything, bidErr).Return(mockOutput)
		gameMock := new(interfaces.MockBatakGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerBid", 3).Return(bidErr)

		ci := usecase.NewBatakInteractor(gameMock, spMock)
		assert.Equal(t, mockOutput, ci.Bid(3))
		gameMock.AssertNotCalled(t, "CpuBid")
	})

	t.Run("valid bid runs CPU bids then stays in bid phase", func(t *testing.T) {
		spMock := new(presenter.MockBatakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBatakGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerBid", 3).Return(nil)
		gameMock.On("GetPhase").Return(domain.BatakPhaseBid)
		gameMock.On("IsHumanBidTurn").Return(false).Once()
		gameMock.On("CpuBid").Return()
		gameMock.On("IsHumanBidTurn").Return(true)

		ci := usecase.NewBatakInteractor(gameMock, spMock)
		assert.Equal(t, mockOutput, ci.Bid(3))
		gameMock.AssertCalled(t, "CpuBid")
	})

	t.Run("valid bid transitions to play phase and runs CPU turns", func(t *testing.T) {
		spMock := new(presenter.MockBatakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBatakGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerBid", 5).Return(nil)
		gameMock.On("GetPhase").Return(domain.BatakPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewBatakInteractor(gameMock, spMock)
		assert.Equal(t, mockOutput, ci.Bid(5))
	})
}

func TestBatakInteractor_Play(t *testing.T) {
	mockOutput := `{"phase":1}`

	t.Run("game ended returns output without playing", func(t *testing.T) {
		spMock := new(presenter.MockBatakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBatakGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewBatakInteractor(gameMock, spMock)
		assert.Equal(t, mockOutput, ci.Play(0))
		gameMock.AssertNotCalled(t, "PlayerPlay")
	})

	t.Run("not human turn returns output without playing", func(t *testing.T) {
		spMock := new(presenter.MockBatakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBatakGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		ci := usecase.NewBatakInteractor(gameMock, spMock)
		assert.Equal(t, mockOutput, ci.Play(0))
		gameMock.AssertNotCalled(t, "PlayerPlay")
	})

	t.Run("play error returned to presenter", func(t *testing.T) {
		playErr := errors.New("invalid play")
		spMock := new(presenter.MockBatakPresenter)
		spMock.On("Output", mock.Anything, playErr).Return(mockOutput)
		gameMock := new(interfaces.MockBatakGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerPlay", 0).Return(playErr)

		ci := usecase.NewBatakInteractor(gameMock, spMock)
		assert.Equal(t, mockOutput, ci.Play(0))
	})

	t.Run("valid play runs CPU turns", func(t *testing.T) {
		spMock := new(presenter.MockBatakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBatakGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerPlay", 2).Return(nil)
		gameMock.On("GetPhase").Return(domain.BatakPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewBatakInteractor(gameMock, spMock)
		assert.Equal(t, mockOutput, ci.Play(2))
		gameMock.AssertCalled(t, "PlayerPlay", 2)
	})

	t.Run("human completes trick calls ResolveTrick", func(t *testing.T) {
		spMock := new(presenter.MockBatakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBatakGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerPlay", 2).Return(nil)
		gameMock.On("GetPhase").Return(domain.BatakPhaseTrickEnd).Once()
		gameMock.On("ResolveTrick").Return()
		gameMock.On("GetPhase").Return(domain.BatakPhaseTrickEnd)

		ci := usecase.NewBatakInteractor(gameMock, spMock)
		assert.Equal(t, mockOutput, ci.Play(2))
		gameMock.AssertCalled(t, "ResolveTrick")
	})
}

func TestBatakInteractor_NextTrick(t *testing.T) {
	mockOutput := `{"phase":1}`
	spMock := new(presenter.MockBatakPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockBatakGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.BatakPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewBatakInteractor(gameMock, spMock)
	assert.Equal(t, mockOutput, ci.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestBatakInteractor_NextRound(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ends after scoring", func(t *testing.T) {
		spMock := new(presenter.MockBatakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBatakGame)
		gameMock.On("ScoreRound").Return()
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewBatakInteractor(gameMock, spMock)
		assert.Equal(t, mockOutput, ci.NextRound())
		gameMock.AssertCalled(t, "ScoreRound")
		gameMock.AssertNotCalled(t, "NextRound")
	})

	t.Run("game continues to next round", func(t *testing.T) {
		spMock := new(presenter.MockBatakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBatakGame)
		gameMock.On("ScoreRound").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextRound").Return()
		gameMock.On("GetPhase").Return(domain.BatakPhaseBid)
		gameMock.On("IsHumanBidTurn").Return(true)

		ci := usecase.NewBatakInteractor(gameMock, spMock)
		assert.Equal(t, mockOutput, ci.NextRound())
		gameMock.AssertCalled(t, "NextRound")
	})
}

func TestBatakInteractor_GetConfig(t *testing.T) {
	expected := domain.BatakConfig{CpuDifficulty: domain.BatakCpuDifficultyHard, MaxRounds: 7}
	spMock := new(presenter.MockBatakPresenter)
	gameMock := new(interfaces.MockBatakGame)
	gameMock.On("GetConfig").Return(expected)

	ci := usecase.NewBatakInteractor(gameMock, spMock)
	assert.Equal(t, expected, ci.GetConfig())
}

func TestBatakInteractor_Hint(t *testing.T) {
	spMock := new(presenter.MockBatakPresenter)
	gameMock := new(interfaces.MockBatakGame)
	spMock.On("HintOutput", gameMock).Return(`{"hint":{"cardIndex":0}}`)

	ci := usecase.NewBatakInteractor(gameMock, spMock)
	assert.Equal(t, `{"hint":{"cardIndex":0}}`, ci.Hint())
	spMock.AssertExpectations(t)
}

func TestBatakInteractor_ActionLog(t *testing.T) {
	spMock := new(presenter.MockBatakPresenter)
	gameMock := new(interfaces.MockBatakGame)
	spMock.On("ActionLogOutput", gameMock).Return(`{"entries":[]}`)

	ci := usecase.NewBatakInteractor(gameMock, spMock)
	assert.Equal(t, `{"entries":[]}`, ci.ActionLog())
}

func TestBatakInteractor_RunCpuTurns_AllStopPaths(t *testing.T) {
	mockOutput := `{"phase":1}`

	t.Run("stops on game end", func(t *testing.T) {
		spMock := new(presenter.MockBatakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBatakGame)
		gameMock.On("NextTrick").Return()
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewBatakInteractor(gameMock, spMock)
		ci.NextTrick()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("CpuPlay then ResolveTrick continues with next trick", func(t *testing.T) {
		spMock := new(presenter.MockBatakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBatakGame)
		gameMock.On("NextTrick").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.BatakPhasePlay).Once()
		gameMock.On("IsHumanTurn").Return(false).Once()
		gameMock.On("CpuPlay").Return()
		gameMock.On("GetPhase").Return(domain.BatakPhaseTrickEnd).Once()
		gameMock.On("ResolveTrick").Return()
		gameMock.On("GetPhase").Return(domain.BatakPhaseRoundEnd)

		ci := usecase.NewBatakInteractor(gameMock, spMock)
		ci.NextTrick()
		gameMock.AssertCalled(t, "ResolveTrick")
	})

	t.Run("ends on GameEnd phase after CpuPlay", func(t *testing.T) {
		spMock := new(presenter.MockBatakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBatakGame)
		gameMock.On("NextTrick").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.BatakPhaseGameEnd)

		ci := usecase.NewBatakInteractor(gameMock, spMock)
		ci.NextTrick()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})
}

func TestBatakInteractor_RunCpuBids_StopPaths(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("stops when not bid phase", func(t *testing.T) {
		spMock := new(presenter.MockBatakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBatakGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.BatakPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewBatakInteractor(gameMock, spMock)
		ci.Reset()
		gameMock.AssertNotCalled(t, "CpuBid")
	})

	t.Run("stops when game ended", func(t *testing.T) {
		spMock := new(presenter.MockBatakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBatakGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(true)
		gameMock.On("GetPhase").Return(domain.BatakPhaseBid)

		ci := usecase.NewBatakInteractor(gameMock, spMock)
		ci.Reset()
		gameMock.AssertNotCalled(t, "CpuBid")
	})

	t.Run("CPU bids until human bid turn", func(t *testing.T) {
		spMock := new(presenter.MockBatakPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBatakGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.BatakPhaseBid)
		gameMock.On("IsHumanBidTurn").Return(false).Once()
		gameMock.On("CpuBid").Return()
		gameMock.On("IsHumanBidTurn").Return(true)

		ci := usecase.NewBatakInteractor(gameMock, spMock)
		ci.Reset()
		gameMock.AssertCalled(t, "CpuBid")
	})
}

func TestRestoreBatakInteractor_InvalidJSON(t *testing.T) {
	spMock := new(presenter.MockBatakPresenter)
	_, err := usecase.RestoreBatakInteractor([]byte("not json"), spMock)
	assert.Error(t, err)
}

func TestRestoreBatakInteractor_Valid(t *testing.T) {
	cb := domain.NewDefaultBatak()
	cb.Reset()
	data, err := cb.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	spMock := new(presenter.MockBatakPresenter)
	ci, err := usecase.RestoreBatakInteractor(data, spMock)
	assert.NoError(t, err)
	assert.NotNil(t, ci)
}
