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

const ganjifaMockOutput = `{"phase":0}`

func newGanjifaMocks() (*presenter.MockGanjifaPresenter, *interfaces.MockGanjifaGame) {
	sp := new(presenter.MockGanjifaPresenter)
	sp.On("Output", mock.Anything, mock.Anything).Return(ganjifaMockOutput)
	sp.On("HintOutput", mock.Anything).Return("hint")
	sp.On("ActionLogOutput", mock.Anything).Return("log")
	return sp, new(interfaces.MockGanjifaGame)
}

func TestNewGanjifaInteractor_NilGuards(t *testing.T) {
	sp := new(presenter.MockGanjifaPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "GanjifaInteractor: g must not be nil", func() {
			usecase.NewGanjifaInteractor(nil, sp)
		})
	})
	t.Run("panics when sp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockGanjifaGame)
		assert.PanicsWithValue(t, "GanjifaInteractor: sp must not be nil", func() {
			usecase.NewGanjifaInteractor(gameMock, nil)
		})
	})
}

func TestGanjifaInteractor_Reset(t *testing.T) {
	sp, gameMock := newGanjifaMocks()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.GanjifaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	pi := usecase.NewGanjifaInteractor(gameMock, sp)
	assert.Equal(t, ganjifaMockOutput, pi.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestGanjifaInteractor_ResetWithConfig(t *testing.T) {
	sp, gameMock := newGanjifaMocks()
	cfg := domain.GanjifaConfig{CpuDifficulty: domain.GanjifaCpuDifficultyHard, TargetRounds: 6}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.GanjifaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	pi := usecase.NewGanjifaInteractor(gameMock, sp)
	assert.Equal(t, ganjifaMockOutput, pi.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestGanjifaInteractor_ResetWithConfigInvalid(t *testing.T) {
	sp, gameMock := newGanjifaMocks()

	pi := usecase.NewGanjifaInteractor(gameMock, sp)
	bad := domain.GanjifaConfig{CpuDifficulty: domain.GanjifaCpuDifficultyNormal, TargetRounds: 0}
	assert.Equal(t, ganjifaMockOutput, pi.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestGanjifaInteractor_PlayResolvesTrick(t *testing.T) {
	sp, gameMock := newGanjifaMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.GanjifaPhaseTrickEnd).Once()
	gameMock.On("GetPhase").Return(domain.GanjifaPhaseRoundEnd)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 2).Return(nil)
	gameMock.On("ResolveTrick").Return()

	pi := usecase.NewGanjifaInteractor(gameMock, sp)
	assert.Equal(t, ganjifaMockOutput, pi.Play(2))
	gameMock.AssertCalled(t, "PlayerPlay", 2)
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestGanjifaInteractor_PlayNoResolveWhenNotTrickEnd(t *testing.T) {
	sp, gameMock := newGanjifaMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.GanjifaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 1).Return(nil)

	pi := usecase.NewGanjifaInteractor(gameMock, sp)
	assert.Equal(t, ganjifaMockOutput, pi.Play(1))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestGanjifaInteractor_PlayError(t *testing.T) {
	sp, gameMock := newGanjifaMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.GanjifaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 9).Return(errors.New("invalid card"))

	pi := usecase.NewGanjifaInteractor(gameMock, sp)
	assert.Equal(t, ganjifaMockOutput, pi.Play(9))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestGanjifaInteractor_PlayNotHumanTurn(t *testing.T) {
	sp, gameMock := newGanjifaMocks()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	pi := usecase.NewGanjifaInteractor(gameMock, sp)
	assert.Equal(t, ganjifaMockOutput, pi.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestGanjifaInteractor_NextTrick(t *testing.T) {
	sp, gameMock := newGanjifaMocks()
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.GanjifaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	pi := usecase.NewGanjifaInteractor(gameMock, sp)
	assert.Equal(t, ganjifaMockOutput, pi.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestGanjifaInteractor_NextRound(t *testing.T) {
	sp, gameMock := newGanjifaMocks()
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("NextRound").Return()
	gameMock.On("GetPhase").Return(domain.GanjifaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	pi := usecase.NewGanjifaInteractor(gameMock, sp)
	assert.Equal(t, ganjifaMockOutput, pi.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestGanjifaInteractor_NextRoundGameEnded(t *testing.T) {
	sp, gameMock := newGanjifaMocks()
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	pi := usecase.NewGanjifaInteractor(gameMock, sp)
	assert.Equal(t, ganjifaMockOutput, pi.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestGanjifaInteractor_GetConfigHintActionLog(t *testing.T) {
	sp, gameMock := newGanjifaMocks()
	cfg := domain.DefaultGanjifaConfig()
	gameMock.On("GetConfig").Return(cfg)

	pi := usecase.NewGanjifaInteractor(gameMock, sp)
	assert.Equal(t, cfg, pi.GetConfig())
	assert.Equal(t, "hint", pi.Hint())
	assert.Equal(t, "log", pi.ActionLog())
}

// Ganjifa has no bidding, so the only CPU auto-advance is trick play: the
// interactor must keep calling CpuPlay until the human is on turn again.
func TestGanjifaInteractor_AdvanceCpuPlays(t *testing.T) {
	sp, gameMock := newGanjifaMocks()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.GanjifaPhasePlay)
	gameMock.On("IsHumanTurn").Return(false).Twice()
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("CpuPlay").Return()

	pi := usecase.NewGanjifaInteractor(gameMock, sp)
	assert.Equal(t, ganjifaMockOutput, pi.Reset())
	gameMock.AssertNumberOfCalls(t, "CpuPlay", 2)
}

func TestRestoreGanjifaInteractor(t *testing.T) {
	sp := new(presenter.MockGanjifaPresenter)
	src := domain.NewDefaultGanjifa()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	pi, err := usecase.RestoreGanjifaInteractor(data, sp)
	assert.NoError(t, err)
	assert.NotNil(t, pi)

	_, err = usecase.RestoreGanjifaInteractor([]byte(`{`), sp)
	assert.Error(t, err)
}
