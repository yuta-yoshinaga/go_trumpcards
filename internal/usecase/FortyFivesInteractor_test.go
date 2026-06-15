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

const fortyFivesMockOutput = `{"phase":0}`

func newFortyFivesMocks() (*presenter.MockFortyFivesPresenter, *interfaces.MockFortyFivesGame) {
	sp := new(presenter.MockFortyFivesPresenter)
	sp.On("Output", mock.Anything, mock.Anything).Return(fortyFivesMockOutput)
	sp.On("HintOutput", mock.Anything).Return("hint")
	sp.On("ActionLogOutput", mock.Anything).Return("log")
	return sp, new(interfaces.MockFortyFivesGame)
}

func TestNewFortyFivesInteractor_NilGuards(t *testing.T) {
	sp := new(presenter.MockFortyFivesPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "FortyFivesInteractor: g must not be nil", func() {
			usecase.NewFortyFivesInteractor(nil, sp)
		})
	})
	t.Run("panics when sp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockFortyFivesGame)
		assert.PanicsWithValue(t, "FortyFivesInteractor: sp must not be nil", func() {
			usecase.NewFortyFivesInteractor(gameMock, nil)
		})
	})
}

func TestFortyFivesInteractor_Reset(t *testing.T) {
	sp, gameMock := newFortyFivesMocks()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.FortyFivesPhaseBid)
	gameMock.On("IsHumanBidTurn").Return(true)

	fi := usecase.NewFortyFivesInteractor(gameMock, sp)
	assert.Equal(t, fortyFivesMockOutput, fi.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestFortyFivesInteractor_ResetWithConfig(t *testing.T) {
	sp, gameMock := newFortyFivesMocks()
	cfg := domain.FortyFivesConfig{CpuDifficulty: domain.FortyFivesCpuDifficultyHard, TargetPoints: 45}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.FortyFivesPhaseBid)
	gameMock.On("IsHumanBidTurn").Return(true)

	fi := usecase.NewFortyFivesInteractor(gameMock, sp)
	assert.Equal(t, fortyFivesMockOutput, fi.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestFortyFivesInteractor_ResetWithConfigInvalid(t *testing.T) {
	sp, gameMock := newFortyFivesMocks()

	fi := usecase.NewFortyFivesInteractor(gameMock, sp)
	bad := domain.FortyFivesConfig{CpuDifficulty: domain.FortyFivesCpuDifficultyNormal, TargetPoints: 0}
	assert.Equal(t, fortyFivesMockOutput, fi.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestFortyFivesInteractor_Bid(t *testing.T) {
	sp, gameMock := newFortyFivesMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerBid", domain.FortyFivesBidTwenty).Return(nil)
	gameMock.On("GetPhase").Return(domain.FortyFivesPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	fi := usecase.NewFortyFivesInteractor(gameMock, sp)
	assert.Equal(t, fortyFivesMockOutput, fi.Bid(int(domain.FortyFivesBidTwenty)))
	gameMock.AssertCalled(t, "PlayerBid", domain.FortyFivesBidTwenty)
}

func TestFortyFivesInteractor_BidError(t *testing.T) {
	sp, gameMock := newFortyFivesMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerBid", domain.FortyFivesBidTwenty).Return(errors.New("must beat current bid"))

	fi := usecase.NewFortyFivesInteractor(gameMock, sp)
	assert.Equal(t, fortyFivesMockOutput, fi.Bid(int(domain.FortyFivesBidTwenty)))
	gameMock.AssertNotCalled(t, "CpuBid")
}

func TestFortyFivesInteractor_BidGameEnded(t *testing.T) {
	sp, gameMock := newFortyFivesMocks()
	gameMock.On("GetGameEndFlag").Return(true)

	fi := usecase.NewFortyFivesInteractor(gameMock, sp)
	assert.Equal(t, fortyFivesMockOutput, fi.Bid(0))
	gameMock.AssertNotCalled(t, "PlayerBid", mock.Anything)
}

func TestFortyFivesInteractor_PlayResolvesTrick(t *testing.T) {
	sp, gameMock := newFortyFivesMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.FortyFivesPhaseTrickEnd).Once()
	gameMock.On("GetPhase").Return(domain.FortyFivesPhaseRoundEnd)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 2).Return(nil)
	gameMock.On("ResolveTrick").Return()

	fi := usecase.NewFortyFivesInteractor(gameMock, sp)
	assert.Equal(t, fortyFivesMockOutput, fi.Play(2))
	gameMock.AssertCalled(t, "PlayerPlay", 2)
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestFortyFivesInteractor_PlayNoResolveWhenNotTrickEnd(t *testing.T) {
	sp, gameMock := newFortyFivesMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.FortyFivesPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 1).Return(nil)

	fi := usecase.NewFortyFivesInteractor(gameMock, sp)
	assert.Equal(t, fortyFivesMockOutput, fi.Play(1))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestFortyFivesInteractor_PlayError(t *testing.T) {
	sp, gameMock := newFortyFivesMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.FortyFivesPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 9).Return(errors.New("invalid card"))

	fi := usecase.NewFortyFivesInteractor(gameMock, sp)
	assert.Equal(t, fortyFivesMockOutput, fi.Play(9))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestFortyFivesInteractor_PlayNotHumanTurn(t *testing.T) {
	sp, gameMock := newFortyFivesMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	fi := usecase.NewFortyFivesInteractor(gameMock, sp)
	assert.Equal(t, fortyFivesMockOutput, fi.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestFortyFivesInteractor_NextTrick(t *testing.T) {
	sp, gameMock := newFortyFivesMocks()
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.FortyFivesPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	fi := usecase.NewFortyFivesInteractor(gameMock, sp)
	assert.Equal(t, fortyFivesMockOutput, fi.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestFortyFivesInteractor_NextRound(t *testing.T) {
	sp, gameMock := newFortyFivesMocks()
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("NextRound").Return()
	gameMock.On("GetPhase").Return(domain.FortyFivesPhaseBid)
	gameMock.On("IsHumanBidTurn").Return(true)

	fi := usecase.NewFortyFivesInteractor(gameMock, sp)
	assert.Equal(t, fortyFivesMockOutput, fi.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestFortyFivesInteractor_NextRoundGameEnded(t *testing.T) {
	sp, gameMock := newFortyFivesMocks()
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	fi := usecase.NewFortyFivesInteractor(gameMock, sp)
	assert.Equal(t, fortyFivesMockOutput, fi.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestFortyFivesInteractor_GetConfigHintActionLog(t *testing.T) {
	sp, gameMock := newFortyFivesMocks()
	cfg := domain.DefaultFortyFivesConfig()
	gameMock.On("GetConfig").Return(cfg)

	fi := usecase.NewFortyFivesInteractor(gameMock, sp)
	assert.Equal(t, cfg, fi.GetConfig())
	assert.Equal(t, "hint", fi.Hint())
	assert.Equal(t, "log", fi.ActionLog())
}

func TestFortyFivesInteractor_AdvanceCpuBids(t *testing.T) {
	sp, gameMock := newFortyFivesMocks()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	// Two CPU bids, then the human's bid turn.
	gameMock.On("GetPhase").Return(domain.FortyFivesPhaseBid)
	gameMock.On("IsHumanBidTurn").Return(false).Twice()
	gameMock.On("IsHumanBidTurn").Return(true)
	gameMock.On("CpuBid").Return()

	fi := usecase.NewFortyFivesInteractor(gameMock, sp)
	assert.Equal(t, fortyFivesMockOutput, fi.Reset())
	gameMock.AssertNumberOfCalls(t, "CpuBid", 2)
}

func TestRestoreFortyFivesInteractor(t *testing.T) {
	sp := new(presenter.MockFortyFivesPresenter)
	src := domain.NewDefaultFortyFives()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	fi, err := usecase.RestoreFortyFivesInteractor(data, sp)
	assert.NoError(t, err)
	assert.NotNil(t, fi)

	_, err = usecase.RestoreFortyFivesInteractor([]byte(`{`), sp)
	assert.Error(t, err)
}
