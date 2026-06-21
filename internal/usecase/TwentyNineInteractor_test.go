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

const twentyNineMockOutput = `{"phase":0}`

func newTwentyNineMocks() (*presenter.MockTwentyNinePresenter, *interfaces.MockTwentyNineGame) {
	sp := new(presenter.MockTwentyNinePresenter)
	sp.On("Output", mock.Anything, mock.Anything).Return(twentyNineMockOutput)
	sp.On("HintOutput", mock.Anything).Return("hint")
	sp.On("ActionLogOutput", mock.Anything).Return("log")
	return sp, new(interfaces.MockTwentyNineGame)
}

func TestNewTwentyNineInteractor_NilGuards(t *testing.T) {
	sp := new(presenter.MockTwentyNinePresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "TwentyNineInteractor: g must not be nil", func() {
			usecase.NewTwentyNineInteractor(nil, sp)
		})
	})
	t.Run("panics when sp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockTwentyNineGame)
		assert.PanicsWithValue(t, "TwentyNineInteractor: sp must not be nil", func() {
			usecase.NewTwentyNineInteractor(gameMock, nil)
		})
	})
}

func TestTwentyNineInteractor_Reset(t *testing.T) {
	sp, gameMock := newTwentyNineMocks()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TwentyNinePhaseBid)
	gameMock.On("IsHumanBidTurn").Return(true)

	ti := usecase.NewTwentyNineInteractor(gameMock, sp)
	assert.Equal(t, twentyNineMockOutput, ti.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestTwentyNineInteractor_ResetWithConfig(t *testing.T) {
	sp, gameMock := newTwentyNineMocks()
	cfg := domain.TwentyNineConfig{CpuDifficulty: domain.TwentyNineCpuDifficultyHard, TargetPoints: 6}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TwentyNinePhaseBid)
	gameMock.On("IsHumanBidTurn").Return(true)

	ti := usecase.NewTwentyNineInteractor(gameMock, sp)
	assert.Equal(t, twentyNineMockOutput, ti.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestTwentyNineInteractor_ResetWithConfigInvalid(t *testing.T) {
	sp, gameMock := newTwentyNineMocks()

	ti := usecase.NewTwentyNineInteractor(gameMock, sp)
	bad := domain.TwentyNineConfig{CpuDifficulty: domain.TwentyNineCpuDifficultyNormal, TargetPoints: 0}
	assert.Equal(t, twentyNineMockOutput, ti.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestTwentyNineInteractor_Bid(t *testing.T) {
	sp, gameMock := newTwentyNineMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerBid", domain.TwentyNineBidTwenty).Return(nil)
	gameMock.On("GetPhase").Return(domain.TwentyNinePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewTwentyNineInteractor(gameMock, sp)
	assert.Equal(t, twentyNineMockOutput, ti.Bid(int(domain.TwentyNineBidTwenty)))
	gameMock.AssertCalled(t, "PlayerBid", domain.TwentyNineBidTwenty)
}

func TestTwentyNineInteractor_BidError(t *testing.T) {
	sp, gameMock := newTwentyNineMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerBid", domain.TwentyNineBidTwenty).Return(errors.New("must beat current bid"))

	ti := usecase.NewTwentyNineInteractor(gameMock, sp)
	assert.Equal(t, twentyNineMockOutput, ti.Bid(int(domain.TwentyNineBidTwenty)))
	gameMock.AssertNotCalled(t, "CpuBid")
}

func TestTwentyNineInteractor_BidGameEnded(t *testing.T) {
	sp, gameMock := newTwentyNineMocks()
	gameMock.On("GetGameEndFlag").Return(true)

	ti := usecase.NewTwentyNineInteractor(gameMock, sp)
	assert.Equal(t, twentyNineMockOutput, ti.Bid(0))
	gameMock.AssertNotCalled(t, "PlayerBid", mock.Anything)
}

func TestTwentyNineInteractor_PlayResolvesTrick(t *testing.T) {
	sp, gameMock := newTwentyNineMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TwentyNinePhaseTrickEnd).Once()
	gameMock.On("GetPhase").Return(domain.TwentyNinePhaseRoundEnd)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 2).Return(nil)
	gameMock.On("ResolveTrick").Return()

	ti := usecase.NewTwentyNineInteractor(gameMock, sp)
	assert.Equal(t, twentyNineMockOutput, ti.Play(2))
	gameMock.AssertCalled(t, "PlayerPlay", 2)
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestTwentyNineInteractor_PlayNoResolveWhenNotTrickEnd(t *testing.T) {
	sp, gameMock := newTwentyNineMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TwentyNinePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 1).Return(nil)

	ti := usecase.NewTwentyNineInteractor(gameMock, sp)
	assert.Equal(t, twentyNineMockOutput, ti.Play(1))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestTwentyNineInteractor_PlayError(t *testing.T) {
	sp, gameMock := newTwentyNineMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TwentyNinePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 9).Return(errors.New("invalid card"))

	ti := usecase.NewTwentyNineInteractor(gameMock, sp)
	assert.Equal(t, twentyNineMockOutput, ti.Play(9))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestTwentyNineInteractor_PlayNotHumanTurn(t *testing.T) {
	sp, gameMock := newTwentyNineMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	ti := usecase.NewTwentyNineInteractor(gameMock, sp)
	assert.Equal(t, twentyNineMockOutput, ti.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestTwentyNineInteractor_NextTrick(t *testing.T) {
	sp, gameMock := newTwentyNineMocks()
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TwentyNinePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewTwentyNineInteractor(gameMock, sp)
	assert.Equal(t, twentyNineMockOutput, ti.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestTwentyNineInteractor_NextRound(t *testing.T) {
	sp, gameMock := newTwentyNineMocks()
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("NextRound").Return()
	gameMock.On("GetPhase").Return(domain.TwentyNinePhaseBid)
	gameMock.On("IsHumanBidTurn").Return(true)

	ti := usecase.NewTwentyNineInteractor(gameMock, sp)
	assert.Equal(t, twentyNineMockOutput, ti.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestTwentyNineInteractor_NextRoundGameEnded(t *testing.T) {
	sp, gameMock := newTwentyNineMocks()
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	ti := usecase.NewTwentyNineInteractor(gameMock, sp)
	assert.Equal(t, twentyNineMockOutput, ti.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestTwentyNineInteractor_GetConfigHintActionLog(t *testing.T) {
	sp, gameMock := newTwentyNineMocks()
	cfg := domain.DefaultTwentyNineConfig()
	gameMock.On("GetConfig").Return(cfg)

	ti := usecase.NewTwentyNineInteractor(gameMock, sp)
	assert.Equal(t, cfg, ti.GetConfig())
	assert.Equal(t, "hint", ti.Hint())
	assert.Equal(t, "log", ti.ActionLog())
}

func TestTwentyNineInteractor_AdvanceCpuBids(t *testing.T) {
	sp, gameMock := newTwentyNineMocks()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	// Two CPU bids, then the human's bid turn.
	gameMock.On("GetPhase").Return(domain.TwentyNinePhaseBid)
	gameMock.On("IsHumanBidTurn").Return(false).Twice()
	gameMock.On("IsHumanBidTurn").Return(true)
	gameMock.On("CpuBid").Return()

	ti := usecase.NewTwentyNineInteractor(gameMock, sp)
	assert.Equal(t, twentyNineMockOutput, ti.Reset())
	gameMock.AssertNumberOfCalls(t, "CpuBid", 2)
}

func TestRestoreTwentyNineInteractor(t *testing.T) {
	sp := new(presenter.MockTwentyNinePresenter)
	src := domain.NewDefaultTwentyNine()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	ti, err := usecase.RestoreTwentyNineInteractor(data, sp)
	assert.NoError(t, err)
	assert.NotNil(t, ti)

	_, err = usecase.RestoreTwentyNineInteractor([]byte(`{`), sp)
	assert.Error(t, err)
}
