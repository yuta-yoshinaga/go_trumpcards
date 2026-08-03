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

const viraMockOutput = `{"phase":0}`

func newViraMocks() (*presenter.MockViraPresenter, *interfaces.MockViraGame) {
	sp := new(presenter.MockViraPresenter)
	sp.On("Output", mock.Anything, mock.Anything).Return(viraMockOutput)
	sp.On("HintOutput", mock.Anything).Return("hint")
	sp.On("ActionLogOutput", mock.Anything).Return("log")
	return sp, new(interfaces.MockViraGame)
}

func TestNewViraInteractor_NilGuards(t *testing.T) {
	sp := new(presenter.MockViraPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "ViraInteractor: g must not be nil", func() {
			usecase.NewViraInteractor(nil, sp)
		})
	})
	t.Run("panics when sp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockViraGame)
		assert.PanicsWithValue(t, "ViraInteractor: sp must not be nil", func() {
			usecase.NewViraInteractor(gameMock, nil)
		})
	})
}

func TestViraInteractor_Reset(t *testing.T) {
	sp, gameMock := newViraMocks()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.ViraPhaseBid)
	gameMock.On("IsHumanBidTurn").Return(true)

	pi := usecase.NewViraInteractor(gameMock, sp)
	assert.Equal(t, viraMockOutput, pi.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestViraInteractor_ResetWithConfig(t *testing.T) {
	sp, gameMock := newViraMocks()
	cfg := domain.ViraConfig{CpuDifficulty: domain.ViraCpuDifficultyHard, TargetRounds: 9}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.ViraPhaseBid)
	gameMock.On("IsHumanBidTurn").Return(true)

	pi := usecase.NewViraInteractor(gameMock, sp)
	assert.Equal(t, viraMockOutput, pi.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

// Rounds must be a multiple of the player count so every seat deals once; a
// config that breaks that must never reach the game.
func TestViraInteractor_ResetWithConfigInvalid(t *testing.T) {
	for name, cfg := range map[string]domain.ViraConfig{
		"zero rounds": {CpuDifficulty: domain.ViraCpuDifficultyNormal, TargetRounds: 0},
		"not a multiple of the player count": {
			CpuDifficulty: domain.ViraCpuDifficultyNormal, TargetRounds: 7,
		},
		"bad difficulty": {CpuDifficulty: 99, TargetRounds: 6},
	} {
		t.Run(name, func(t *testing.T) {
			sp, gameMock := newViraMocks()
			pi := usecase.NewViraInteractor(gameMock, sp)
			assert.Equal(t, viraMockOutput, pi.ResetWithConfig(cfg))
			gameMock.AssertNotCalled(t, "Reset")
		})
	}
}

func TestViraInteractor_Bid(t *testing.T) {
	sp, gameMock := newViraMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerBid", domain.ViraBidSolo).Return(nil)
	gameMock.On("GetPhase").Return(domain.ViraPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	pi := usecase.NewViraInteractor(gameMock, sp)
	assert.Equal(t, viraMockOutput, pi.Bid(int(domain.ViraBidSolo)))
	gameMock.AssertCalled(t, "PlayerBid", domain.ViraBidSolo)
}

func TestViraInteractor_BidError(t *testing.T) {
	sp, gameMock := newViraMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerBid", domain.ViraBidGask).Return(errors.New("must beat current bid"))

	pi := usecase.NewViraInteractor(gameMock, sp)
	assert.Equal(t, viraMockOutput, pi.Bid(int(domain.ViraBidGask)))
	gameMock.AssertNotCalled(t, "CpuBid")
}

func TestViraInteractor_BidGameEnded(t *testing.T) {
	sp, gameMock := newViraMocks()
	gameMock.On("GetGameEndFlag").Return(true)

	pi := usecase.NewViraInteractor(gameMock, sp)
	assert.Equal(t, viraMockOutput, pi.Bid(0))
	gameMock.AssertNotCalled(t, "PlayerBid", mock.Anything)
}

func TestViraInteractor_PlayResolvesTrick(t *testing.T) {
	sp, gameMock := newViraMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.ViraPhaseTrickEnd).Once()
	gameMock.On("GetPhase").Return(domain.ViraPhaseRoundEnd)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 2).Return(nil)
	gameMock.On("ResolveTrick").Return()

	pi := usecase.NewViraInteractor(gameMock, sp)
	assert.Equal(t, viraMockOutput, pi.Play(2))
	gameMock.AssertCalled(t, "PlayerPlay", 2)
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestViraInteractor_PlayNoResolveWhenNotTrickEnd(t *testing.T) {
	sp, gameMock := newViraMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.ViraPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 1).Return(nil)

	pi := usecase.NewViraInteractor(gameMock, sp)
	assert.Equal(t, viraMockOutput, pi.Play(1))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestViraInteractor_PlayError(t *testing.T) {
	sp, gameMock := newViraMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.ViraPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 9).Return(errors.New("invalid card"))

	pi := usecase.NewViraInteractor(gameMock, sp)
	assert.Equal(t, viraMockOutput, pi.Play(9))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestViraInteractor_PlayNotHumanTurn(t *testing.T) {
	sp, gameMock := newViraMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	pi := usecase.NewViraInteractor(gameMock, sp)
	assert.Equal(t, viraMockOutput, pi.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestViraInteractor_NextTrick(t *testing.T) {
	sp, gameMock := newViraMocks()
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.ViraPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	pi := usecase.NewViraInteractor(gameMock, sp)
	assert.Equal(t, viraMockOutput, pi.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestViraInteractor_NextRound(t *testing.T) {
	sp, gameMock := newViraMocks()
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("NextRound").Return()
	gameMock.On("GetPhase").Return(domain.ViraPhaseBid)
	gameMock.On("IsHumanBidTurn").Return(true)

	pi := usecase.NewViraInteractor(gameMock, sp)
	assert.Equal(t, viraMockOutput, pi.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestViraInteractor_NextRoundGameEnded(t *testing.T) {
	sp, gameMock := newViraMocks()
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	pi := usecase.NewViraInteractor(gameMock, sp)
	assert.Equal(t, viraMockOutput, pi.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestViraInteractor_GetConfigHintActionLog(t *testing.T) {
	sp, gameMock := newViraMocks()
	cfg := domain.DefaultViraConfig()
	gameMock.On("GetConfig").Return(cfg)

	pi := usecase.NewViraInteractor(gameMock, sp)
	assert.Equal(t, cfg, pi.GetConfig())
	assert.Equal(t, "hint", pi.Hint())
	assert.Equal(t, "log", pi.ActionLog())
}

func TestViraInteractor_AdvanceCpuBids(t *testing.T) {
	sp, gameMock := newViraMocks()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.ViraPhaseBid)
	gameMock.On("IsHumanBidTurn").Return(false).Twice()
	gameMock.On("IsHumanBidTurn").Return(true)
	gameMock.On("CpuBid").Return()

	pi := usecase.NewViraInteractor(gameMock, sp)
	assert.Equal(t, viraMockOutput, pi.Reset())
	gameMock.AssertNumberOfCalls(t, "CpuBid", 2)
}

func TestViraInteractor_AdvanceCpuPlays(t *testing.T) {
	sp, gameMock := newViraMocks()
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.ViraPhasePlay)
	gameMock.On("IsHumanTurn").Return(false).Twice()
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("CpuPlay").Return()

	pi := usecase.NewViraInteractor(gameMock, sp)
	assert.Equal(t, viraMockOutput, pi.NextTrick())
	gameMock.AssertNumberOfCalls(t, "CpuPlay", 2)
}

func TestRestoreViraInteractor(t *testing.T) {
	sp := new(presenter.MockViraPresenter)
	src := domain.NewDefaultVira()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	pi, err := usecase.RestoreViraInteractor(data, sp)
	assert.NoError(t, err)
	assert.NotNil(t, pi)

	_, err = usecase.RestoreViraInteractor([]byte(`{`), sp)
	assert.Error(t, err)
}
