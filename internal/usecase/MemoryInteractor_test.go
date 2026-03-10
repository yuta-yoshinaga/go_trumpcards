package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockMemoryGame() *interfaces.MockMemoryGame {
	return new(interfaces.MockMemoryGame)
}

func newMockMemoryPresenter() *presenter.MockMemoryPresenter {
	return new(presenter.MockMemoryPresenter)
}

func TestNewMemoryInteractor(t *testing.T) {
	mg := newMockMemoryGame()
	mp := newMockMemoryPresenter()
	mi := NewMemoryInteractor(mg, mp)
	assert.NotNil(t, mi)
}

func TestNewMemoryInteractorPanicsOnNil(t *testing.T) {
	mp := newMockMemoryPresenter()
	assert.Panics(t, func() { NewMemoryInteractor(nil, mp) })
	mg := newMockMemoryGame()
	assert.Panics(t, func() { NewMemoryInteractor(mg, nil) })
}

func TestMemoryInteractorReset(t *testing.T) {
	mg := newMockMemoryGame()
	mp := newMockMemoryPresenter()
	mi := NewMemoryInteractor(mg, mp)

	mg.On("Reset").Return()
	mp.On("Output", mg, nil).Return("reset_output")

	result := mi.Reset()
	assert.Equal(t, "reset_output", result)
	mg.AssertCalled(t, "Reset")
}

func TestMemoryInteractorResetWithConfig(t *testing.T) {
	mg := newMockMemoryGame()
	mp := newMockMemoryPresenter()
	mi := NewMemoryInteractor(mg, mp)

	cfg := domain.MemoryConfig{CpuDifficulty: domain.MemoryCpuDifficultyHard}
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return()
	mp.On("Output", mg, nil).Return("config_output")

	result := mi.ResetWithConfig(cfg)
	assert.Equal(t, "config_output", result)
	mg.AssertCalled(t, "SetConfig", cfg)
}

func TestMemoryInteractorFlip(t *testing.T) {
	t.Run("successful flip", func(t *testing.T) {
		mg := newMockMemoryGame()
		mp := newMockMemoryPresenter()
		mi := NewMemoryInteractor(mg, mp)

		mg.On("GetGameEndFlag").Return(false)
		mg.On("IsHumanTurn").Return(true)
		mg.On("PlayerFlip", 5).Return(nil)
		mp.On("Output", mg, nil).Return("flip_output")

		result := mi.Flip(5)
		assert.Equal(t, "flip_output", result)
	})

	t.Run("flip error", func(t *testing.T) {
		mg := newMockMemoryGame()
		mp := newMockMemoryPresenter()
		mi := NewMemoryInteractor(mg, mp)

		err := errors.New("invalid position")
		mg.On("GetGameEndFlag").Return(false)
		mg.On("IsHumanTurn").Return(true)
		mg.On("PlayerFlip", -1).Return(err)
		mp.On("Output", mg, err).Return("error_output")

		result := mi.Flip(-1)
		assert.Equal(t, "error_output", result)
	})

	t.Run("flip when game over", func(t *testing.T) {
		mg := newMockMemoryGame()
		mp := newMockMemoryPresenter()
		mi := NewMemoryInteractor(mg, mp)

		mg.On("GetGameEndFlag").Return(true)
		mp.On("Output", mg, nil).Return("game_over")

		result := mi.Flip(0)
		assert.Equal(t, "game_over", result)
	})

	t.Run("flip when not human turn", func(t *testing.T) {
		mg := newMockMemoryGame()
		mp := newMockMemoryPresenter()
		mi := NewMemoryInteractor(mg, mp)

		mg.On("GetGameEndFlag").Return(false)
		mg.On("IsHumanTurn").Return(false)
		mp.On("Output", mg, nil).Return("not_your_turn")

		result := mi.Flip(0)
		assert.Equal(t, "not_your_turn", result)
	})
}

func TestMemoryInteractorNext(t *testing.T) {
	t.Run("resolve and CPU plays", func(t *testing.T) {
		mg := newMockMemoryGame()
		mp := newMockMemoryPresenter()
		mi := NewMemoryInteractor(mg, mp)

		mg.On("GetGameEndFlag").Return(false).Times(3)
		mg.On("ResolveFlip").Return()
		// After resolve: CPU turn
		mg.On("GetPhase").Return(domain.MemoryPhaseFlip1).Once()
		mg.On("IsHumanTurn").Return(false).Once()
		mg.On("CpuFlip").Return().Once()

		// After CPU resolves: back to human
		mg.On("GetGameEndFlag").Return(false).Once()
		mg.On("GetPhase").Return(domain.MemoryPhaseFlip1).Once()
		mg.On("IsHumanTurn").Return(true).Once()

		mp.On("Output", mg, nil).Return("next_output")

		result := mi.Next()
		assert.Equal(t, "next_output", result)
		mg.AssertCalled(t, "CpuFlip")
		mg.AssertCalled(t, "ResolveFlip")
	})

	t.Run("next when game over", func(t *testing.T) {
		mg := newMockMemoryGame()
		mp := newMockMemoryPresenter()
		mi := NewMemoryInteractor(mg, mp)

		mg.On("GetGameEndFlag").Return(true)
		mp.On("Output", mg, nil).Return("game_over")

		result := mi.Next()
		assert.Equal(t, "game_over", result)
		mg.AssertNotCalled(t, "ResolveFlip")
	})

	t.Run("CPU continues on match then game ends", func(t *testing.T) {
		mg := newMockMemoryGame()
		mp := newMockMemoryPresenter()
		mi := NewMemoryInteractor(mg, mp)

		// First call to Next
		mg.On("GetGameEndFlag").Return(false).Once()
		mg.On("ResolveFlip").Return()

		// runCpuTurns: loop iteration 1
		mg.On("GetGameEndFlag").Return(false).Once()
		mg.On("GetPhase").Return(domain.MemoryPhaseFlip1).Once()
		mg.On("IsHumanTurn").Return(false).Once()
		mg.On("CpuFlip").Return().Once()

		// After CpuFlip + ResolveFlip: game ends
		mg.On("GetGameEndFlag").Return(true).Once()

		mp.On("Output", mg, nil).Return("cpu_turn_output")

		result := mi.Next()
		assert.Equal(t, "cpu_turn_output", result)
	})

	t.Run("stops on non-flip1 phase", func(t *testing.T) {
		mg := newMockMemoryGame()
		mp := newMockMemoryPresenter()
		mi := NewMemoryInteractor(mg, mp)

		mg.On("GetGameEndFlag").Return(false).Times(2)
		mg.On("ResolveFlip").Return()
		mg.On("GetPhase").Return(domain.MemoryPhaseGameEnd).Once()
		mp.On("Output", mg, nil).Return("phase_stop")

		result := mi.Next()
		assert.Equal(t, "phase_stop", result)
	})
}

func TestMemoryInteractorGetConfig(t *testing.T) {
	mg := newMockMemoryGame()
	mp := newMockMemoryPresenter()
	mi := NewMemoryInteractor(mg, mp)

	cfg := domain.DefaultMemoryConfig()
	mg.On("GetConfig").Return(cfg)

	result := mi.GetConfig()
	assert.Equal(t, cfg, result)
}

func TestMemoryInteractorActionLog(t *testing.T) {
	mg := newMockMemoryGame()
	mp := newMockMemoryPresenter()
	mi := NewMemoryInteractor(mg, mp)

	mp.On("ActionLogOutput", mock.Anything).Return("log_output")

	result := mi.ActionLog()
	assert.Equal(t, "log_output", result)
}
