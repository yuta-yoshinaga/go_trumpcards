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

const soloWhistMockOutput = `{"phase":0}`

func newSoloWhistMocks() (*presenter.MockSoloWhistPresenter, *interfaces.MockSoloWhistGame) {
	sp := new(presenter.MockSoloWhistPresenter)
	sp.On("Output", mock.Anything, mock.Anything).Return(soloWhistMockOutput)
	sp.On("HintOutput", mock.Anything).Return("hint")
	sp.On("ActionLogOutput", mock.Anything).Return("log")
	return sp, new(interfaces.MockSoloWhistGame)
}

func TestNewSoloWhistInteractor_NilGuards(t *testing.T) {
	sp := new(presenter.MockSoloWhistPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "SoloWhistInteractor: g must not be nil", func() {
			usecase.NewSoloWhistInteractor(nil, sp)
		})
	})
	t.Run("panics when sp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockSoloWhistGame)
		assert.PanicsWithValue(t, "SoloWhistInteractor: sp must not be nil", func() {
			usecase.NewSoloWhistInteractor(gameMock, nil)
		})
	})
}

func TestSoloWhistInteractor_Reset(t *testing.T) {
	sp, gameMock := newSoloWhistMocks()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SoloWhistPhaseBid)
	gameMock.On("IsHumanBidTurn").Return(true)

	si := usecase.NewSoloWhistInteractor(gameMock, sp)
	assert.Equal(t, soloWhistMockOutput, si.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestSoloWhistInteractor_ResetWithConfig(t *testing.T) {
	sp, gameMock := newSoloWhistMocks()
	cfg := domain.SoloWhistConfig{CpuDifficulty: domain.SoloWhistCpuDifficultyHard, TargetPoints: 15}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SoloWhistPhaseBid)
	gameMock.On("IsHumanBidTurn").Return(true)

	si := usecase.NewSoloWhistInteractor(gameMock, sp)
	assert.Equal(t, soloWhistMockOutput, si.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestSoloWhistInteractor_ResetWithConfigInvalid(t *testing.T) {
	sp, gameMock := newSoloWhistMocks()

	si := usecase.NewSoloWhistInteractor(gameMock, sp)
	bad := domain.SoloWhistConfig{CpuDifficulty: domain.SoloWhistCpuDifficultyNormal, TargetPoints: 0}
	assert.Equal(t, soloWhistMockOutput, si.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestSoloWhistInteractor_Bid(t *testing.T) {
	sp, gameMock := newSoloWhistMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerBid", domain.SoloWhistBidSolo).Return(nil)
	gameMock.On("GetPhase").Return(domain.SoloWhistPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	si := usecase.NewSoloWhistInteractor(gameMock, sp)
	assert.Equal(t, soloWhistMockOutput, si.Bid(int(domain.SoloWhistBidSolo)))
	gameMock.AssertCalled(t, "PlayerBid", domain.SoloWhistBidSolo)
}

func TestSoloWhistInteractor_BidError(t *testing.T) {
	sp, gameMock := newSoloWhistMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerBid", domain.SoloWhistBidSolo).Return(errors.New("must beat current bid"))

	si := usecase.NewSoloWhistInteractor(gameMock, sp)
	assert.Equal(t, soloWhistMockOutput, si.Bid(int(domain.SoloWhistBidSolo)))
	gameMock.AssertNotCalled(t, "CpuBid")
}

func TestSoloWhistInteractor_BidGameEnded(t *testing.T) {
	sp, gameMock := newSoloWhistMocks()
	gameMock.On("GetGameEndFlag").Return(true)

	si := usecase.NewSoloWhistInteractor(gameMock, sp)
	assert.Equal(t, soloWhistMockOutput, si.Bid(0))
	gameMock.AssertNotCalled(t, "PlayerBid", mock.Anything)
}

func TestSoloWhistInteractor_PlayResolvesTrick(t *testing.T) {
	sp, gameMock := newSoloWhistMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SoloWhistPhaseTrickEnd).Once()
	gameMock.On("GetPhase").Return(domain.SoloWhistPhaseRoundEnd)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 2).Return(nil)
	gameMock.On("ResolveTrick").Return()

	si := usecase.NewSoloWhistInteractor(gameMock, sp)
	assert.Equal(t, soloWhistMockOutput, si.Play(2))
	gameMock.AssertCalled(t, "PlayerPlay", 2)
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestSoloWhistInteractor_PlayNoResolveWhenNotTrickEnd(t *testing.T) {
	sp, gameMock := newSoloWhistMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SoloWhistPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 1).Return(nil)

	si := usecase.NewSoloWhistInteractor(gameMock, sp)
	assert.Equal(t, soloWhistMockOutput, si.Play(1))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestSoloWhistInteractor_PlayError(t *testing.T) {
	sp, gameMock := newSoloWhistMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SoloWhistPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 9).Return(errors.New("invalid card"))

	si := usecase.NewSoloWhistInteractor(gameMock, sp)
	assert.Equal(t, soloWhistMockOutput, si.Play(9))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestSoloWhistInteractor_PlayNotHumanTurn(t *testing.T) {
	sp, gameMock := newSoloWhistMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	si := usecase.NewSoloWhistInteractor(gameMock, sp)
	assert.Equal(t, soloWhistMockOutput, si.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestSoloWhistInteractor_NextTrick(t *testing.T) {
	sp, gameMock := newSoloWhistMocks()
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SoloWhistPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	si := usecase.NewSoloWhistInteractor(gameMock, sp)
	assert.Equal(t, soloWhistMockOutput, si.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestSoloWhistInteractor_NextRound(t *testing.T) {
	sp, gameMock := newSoloWhistMocks()
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("NextRound").Return()
	gameMock.On("GetPhase").Return(domain.SoloWhistPhaseBid)
	gameMock.On("IsHumanBidTurn").Return(true)

	si := usecase.NewSoloWhistInteractor(gameMock, sp)
	assert.Equal(t, soloWhistMockOutput, si.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestSoloWhistInteractor_NextRoundGameEnded(t *testing.T) {
	sp, gameMock := newSoloWhistMocks()
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	si := usecase.NewSoloWhistInteractor(gameMock, sp)
	assert.Equal(t, soloWhistMockOutput, si.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestSoloWhistInteractor_GetConfigHintActionLog(t *testing.T) {
	sp, gameMock := newSoloWhistMocks()
	cfg := domain.DefaultSoloWhistConfig()
	gameMock.On("GetConfig").Return(cfg)

	si := usecase.NewSoloWhistInteractor(gameMock, sp)
	assert.Equal(t, cfg, si.GetConfig())
	assert.Equal(t, "hint", si.Hint())
	assert.Equal(t, "log", si.ActionLog())
}

func TestSoloWhistInteractor_AdvanceCpuBids(t *testing.T) {
	sp, gameMock := newSoloWhistMocks()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	// Two CPU bids, then the human's bid turn.
	gameMock.On("GetPhase").Return(domain.SoloWhistPhaseBid)
	gameMock.On("IsHumanBidTurn").Return(false).Twice()
	gameMock.On("IsHumanBidTurn").Return(true)
	gameMock.On("CpuBid").Return()

	si := usecase.NewSoloWhistInteractor(gameMock, sp)
	assert.Equal(t, soloWhistMockOutput, si.Reset())
	gameMock.AssertNumberOfCalls(t, "CpuBid", 2)
}

func TestRestoreSoloWhistInteractor(t *testing.T) {
	sp := new(presenter.MockSoloWhistPresenter)
	src := domain.NewDefaultSoloWhist()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	si, err := usecase.RestoreSoloWhistInteractor(data, sp)
	assert.NoError(t, err)
	assert.NotNil(t, si)

	_, err = usecase.RestoreSoloWhistInteractor([]byte(`{`), sp)
	assert.Error(t, err)
}
