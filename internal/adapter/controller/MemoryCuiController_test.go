package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockMemoryInteractor() *mockusecase.MockMemoryInteractor {
	return new(mockusecase.MockMemoryInteractor)
}

func TestMemoryCuiControllerQuit(t *testing.T) {
	mi := newMockMemoryInteractor()
	c := NewMemoryCuiController(mi)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestMemoryCuiControllerReset(t *testing.T) {
	mi := newMockMemoryInteractor()
	c := NewMemoryCuiController(mi)
	cfg := domain.DefaultMemoryConfig()
	mi.On("GetConfig").Return(cfg)
	mi.On("ResetWithConfig", cfg).Return("reset_output")
	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestMemoryCuiControllerFlip(t *testing.T) {
	t.Run("valid flip", func(t *testing.T) {
		mi := newMockMemoryInteractor()
		c := NewMemoryCuiController(mi)
		mi.On("Flip", 5).Return("flip_output")
		assert.Equal(t, "flip_output", c.Exec("f 5"))
	})
	t.Run("flip alias", func(t *testing.T) {
		mi := newMockMemoryInteractor()
		c := NewMemoryCuiController(mi)
		mi.On("Flip", 10).Return("flip_output")
		assert.Equal(t, "flip_output", c.Exec("flip 10"))
	})
	t.Run("flip no args", func(t *testing.T) {
		mi := newMockMemoryInteractor()
		c := NewMemoryCuiController(mi)
		result := c.Exec("f")
		assert.Equal(t, msgKey("positionRequired"), result)
	})
	t.Run("flip invalid position", func(t *testing.T) {
		mi := newMockMemoryInteractor()
		c := NewMemoryCuiController(mi)
		result := c.Exec("f abc")
		assert.Contains(t, result, msgStem("invalidPosition"))
	})
}

func TestMemoryCuiControllerNext(t *testing.T) {
	mi := newMockMemoryInteractor()
	c := NewMemoryCuiController(mi)
	mi.On("Next").Return("next_output")
	assert.Equal(t, "next_output", c.Exec("n"))
	assert.Equal(t, "next_output", c.Exec("next"))
}

func TestMemoryCuiControllerSetDifficulty(t *testing.T) {
	t.Run("valid difficulty", func(t *testing.T) {
		mi := newMockMemoryInteractor()
		c := NewMemoryCuiController(mi)
		cfg := domain.DefaultMemoryConfig()
		mi.On("GetConfig").Return(cfg)
		newCfg := cfg
		newCfg.CpuDifficulty = domain.MemoryCpuDifficultyHard
		mi.On("ResetWithConfig", newCfg).Return("hard_output")
		assert.Equal(t, "hard_output", c.Exec("sd 2"))
	})
	t.Run("setdifficulty alias", func(t *testing.T) {
		mi := newMockMemoryInteractor()
		c := NewMemoryCuiController(mi)
		cfg := domain.DefaultMemoryConfig()
		mi.On("GetConfig").Return(cfg)
		newCfg := cfg
		newCfg.CpuDifficulty = domain.MemoryCpuDifficultyEasy
		mi.On("ResetWithConfig", newCfg).Return("easy_output")
		assert.Equal(t, "easy_output", c.Exec("setdifficulty 0"))
	})
	t.Run("no args", func(t *testing.T) {
		mi := newMockMemoryInteractor()
		c := NewMemoryCuiController(mi)
		result := c.Exec("sd")
		assert.Contains(t, result, msgCpuDifficultyRequired())
	})
	t.Run("invalid difficulty", func(t *testing.T) {
		mi := newMockMemoryInteractor()
		c := NewMemoryCuiController(mi)
		result := c.Exec("sd 5")
		assert.Contains(t, result, msgInvalidCpuDifficultyPrefix())
	})
	t.Run("non-numeric difficulty", func(t *testing.T) {
		mi := newMockMemoryInteractor()
		c := NewMemoryCuiController(mi)
		result := c.Exec("sd abc")
		assert.Contains(t, result, msgInvalidCpuDifficultyPrefix())
	})
}

func TestMemoryCuiControllerLog(t *testing.T) {
	mi := newMockMemoryInteractor()
	c := NewMemoryCuiController(mi)
	mi.On("ActionLog").Return("log_output")
	assert.Equal(t, "log_output", c.Exec("l"))
	assert.Equal(t, "log_output", c.Exec("log"))
}

func TestMemoryCuiControllerUnknown(t *testing.T) {
	mi := newMockMemoryInteractor()
	c := NewMemoryCuiController(mi)
	result := c.Exec("xyz")
	assert.Contains(t, result, "コマンドが不明です")
}

func TestMemoryCuiControllerEmpty(t *testing.T) {
	mi := newMockMemoryInteractor()
	c := NewMemoryCuiController(mi)
	result := c.Exec("")
	assert.Contains(t, result, "'help' でコマンド一覧を表示します。")
}
