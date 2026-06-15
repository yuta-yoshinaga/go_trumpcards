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

const napMockOutput = `{"phase":0}`

func newNapMocks() (*presenter.MockNapPresenter, *interfaces.MockNapGame) {
	sp := new(presenter.MockNapPresenter)
	sp.On("Output", mock.Anything, mock.Anything).Return(napMockOutput)
	sp.On("HintOutput", mock.Anything).Return("hint")
	sp.On("ActionLogOutput", mock.Anything).Return("log")
	return sp, new(interfaces.MockNapGame)
}

func TestNewNapInteractor_NilGuards(t *testing.T) {
	sp := new(presenter.MockNapPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "NapInteractor: g must not be nil", func() {
			usecase.NewNapInteractor(nil, sp)
		})
	})
	t.Run("panics when sp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockNapGame)
		assert.PanicsWithValue(t, "NapInteractor: sp must not be nil", func() {
			usecase.NewNapInteractor(gameMock, nil)
		})
	})
}

func TestNapInteractor_Reset(t *testing.T) {
	sp, gameMock := newNapMocks()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.NapPhaseBid)
	gameMock.On("IsHumanBidTurn").Return(true)

	ni := usecase.NewNapInteractor(gameMock, sp)
	assert.Equal(t, napMockOutput, ni.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestNapInteractor_ResetWithConfig(t *testing.T) {
	sp, gameMock := newNapMocks()
	cfg := domain.NapConfig{CpuDifficulty: domain.NapCpuDifficultyHard, TargetPoints: 15}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.NapPhaseBid)
	gameMock.On("IsHumanBidTurn").Return(true)

	ni := usecase.NewNapInteractor(gameMock, sp)
	assert.Equal(t, napMockOutput, ni.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestNapInteractor_ResetWithConfigInvalid(t *testing.T) {
	sp, gameMock := newNapMocks()

	ni := usecase.NewNapInteractor(gameMock, sp)
	bad := domain.NapConfig{CpuDifficulty: domain.NapCpuDifficultyNormal, TargetPoints: 0}
	assert.Equal(t, napMockOutput, ni.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestNapInteractor_Bid(t *testing.T) {
	sp, gameMock := newNapMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerBid", domain.NapBidThree).Return(nil)
	gameMock.On("GetPhase").Return(domain.NapPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ni := usecase.NewNapInteractor(gameMock, sp)
	assert.Equal(t, napMockOutput, ni.Bid(int(domain.NapBidThree)))
	gameMock.AssertCalled(t, "PlayerBid", domain.NapBidThree)
}

func TestNapInteractor_BidError(t *testing.T) {
	sp, gameMock := newNapMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerBid", domain.NapBidThree).Return(errors.New("must beat current bid"))

	ni := usecase.NewNapInteractor(gameMock, sp)
	assert.Equal(t, napMockOutput, ni.Bid(int(domain.NapBidThree)))
	gameMock.AssertNotCalled(t, "CpuBid")
}

func TestNapInteractor_BidGameEnded(t *testing.T) {
	sp, gameMock := newNapMocks()
	gameMock.On("GetGameEndFlag").Return(true)

	ni := usecase.NewNapInteractor(gameMock, sp)
	assert.Equal(t, napMockOutput, ni.Bid(0))
	gameMock.AssertNotCalled(t, "PlayerBid", mock.Anything)
}

func TestNapInteractor_PlayResolvesTrick(t *testing.T) {
	sp, gameMock := newNapMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.NapPhaseTrickEnd).Once()
	gameMock.On("GetPhase").Return(domain.NapPhaseRoundEnd)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 2).Return(nil)
	gameMock.On("ResolveTrick").Return()

	ni := usecase.NewNapInteractor(gameMock, sp)
	assert.Equal(t, napMockOutput, ni.Play(2))
	gameMock.AssertCalled(t, "PlayerPlay", 2)
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestNapInteractor_PlayNoResolveWhenNotTrickEnd(t *testing.T) {
	sp, gameMock := newNapMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.NapPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 1).Return(nil)

	ni := usecase.NewNapInteractor(gameMock, sp)
	assert.Equal(t, napMockOutput, ni.Play(1))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestNapInteractor_PlayError(t *testing.T) {
	sp, gameMock := newNapMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.NapPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 9).Return(errors.New("invalid card"))

	ni := usecase.NewNapInteractor(gameMock, sp)
	assert.Equal(t, napMockOutput, ni.Play(9))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestNapInteractor_PlayNotHumanTurn(t *testing.T) {
	sp, gameMock := newNapMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	ni := usecase.NewNapInteractor(gameMock, sp)
	assert.Equal(t, napMockOutput, ni.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestNapInteractor_NextTrick(t *testing.T) {
	sp, gameMock := newNapMocks()
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.NapPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ni := usecase.NewNapInteractor(gameMock, sp)
	assert.Equal(t, napMockOutput, ni.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestNapInteractor_NextRound(t *testing.T) {
	sp, gameMock := newNapMocks()
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("NextRound").Return()
	gameMock.On("GetPhase").Return(domain.NapPhaseBid)
	gameMock.On("IsHumanBidTurn").Return(true)

	ni := usecase.NewNapInteractor(gameMock, sp)
	assert.Equal(t, napMockOutput, ni.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestNapInteractor_NextRoundGameEnded(t *testing.T) {
	sp, gameMock := newNapMocks()
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	ni := usecase.NewNapInteractor(gameMock, sp)
	assert.Equal(t, napMockOutput, ni.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestNapInteractor_GetConfigHintActionLog(t *testing.T) {
	sp, gameMock := newNapMocks()
	cfg := domain.DefaultNapConfig()
	gameMock.On("GetConfig").Return(cfg)

	ni := usecase.NewNapInteractor(gameMock, sp)
	assert.Equal(t, cfg, ni.GetConfig())
	assert.Equal(t, "hint", ni.Hint())
	assert.Equal(t, "log", ni.ActionLog())
}

func TestNapInteractor_AdvanceCpuBids(t *testing.T) {
	sp, gameMock := newNapMocks()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	// Two CPU bids, then the human's bid turn.
	gameMock.On("GetPhase").Return(domain.NapPhaseBid)
	gameMock.On("IsHumanBidTurn").Return(false).Twice()
	gameMock.On("IsHumanBidTurn").Return(true)
	gameMock.On("CpuBid").Return()

	ni := usecase.NewNapInteractor(gameMock, sp)
	assert.Equal(t, napMockOutput, ni.Reset())
	gameMock.AssertNumberOfCalls(t, "CpuBid", 2)
}

func TestRestoreNapInteractor(t *testing.T) {
	sp := new(presenter.MockNapPresenter)
	src := domain.NewDefaultNap()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	ni, err := usecase.RestoreNapInteractor(data, sp)
	assert.NoError(t, err)
	assert.NotNil(t, ni)

	_, err = usecase.RestoreNapInteractor([]byte(`{`), sp)
	assert.Error(t, err)
}
