package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestGetCountingBetAmount(t *testing.T) {
	// Balanced system (Hi-Lo): uses trueCount
	t.Run("balanced count <= 1 returns 1x base", func(t *testing.T) {
		bet := domain.GetCountingBetAmount(1.0, 5, domain.BJCountingHiLo, 10000)
		assert.Equal(t, 50, bet)
	})
	t.Run("balanced count <= 1 negative", func(t *testing.T) {
		bet := domain.GetCountingBetAmount(-2.0, 5, domain.BJCountingHiLo, 10000)
		assert.Equal(t, 50, bet)
	})
	t.Run("balanced count 2 returns 2x base", func(t *testing.T) {
		bet := domain.GetCountingBetAmount(2.0, 0, domain.BJCountingHiLo, 10000)
		assert.Equal(t, 100, bet)
	})
	t.Run("balanced count 2.5 returns 2x base", func(t *testing.T) {
		bet := domain.GetCountingBetAmount(2.5, 0, domain.BJCountingHiLo, 10000)
		assert.Equal(t, 100, bet)
	})
	t.Run("balanced count 3 returns 4x base", func(t *testing.T) {
		bet := domain.GetCountingBetAmount(3.0, 0, domain.BJCountingHiLo, 10000)
		assert.Equal(t, 200, bet)
	})
	t.Run("balanced count 4 returns 8x base", func(t *testing.T) {
		bet := domain.GetCountingBetAmount(4.0, 0, domain.BJCountingHiLo, 10000)
		assert.Equal(t, 400, bet)
	})
	t.Run("balanced count 5 returns 16x base", func(t *testing.T) {
		bet := domain.GetCountingBetAmount(5.0, 0, domain.BJCountingHiLo, 10000)
		assert.Equal(t, 800, bet)
	})
	t.Run("balanced count 10 returns 16x base (>=5 cap)", func(t *testing.T) {
		bet := domain.GetCountingBetAmount(10.0, 0, domain.BJCountingHiLo, 10000)
		assert.Equal(t, 800, bet)
	})

	// KO (unbalanced): uses runningCount
	t.Run("KO count <= 1 returns 1x base", func(t *testing.T) {
		bet := domain.GetCountingBetAmount(0.0, 1, domain.BJCountingKO, 10000)
		assert.Equal(t, 50, bet)
	})
	t.Run("KO count negative returns 1x base", func(t *testing.T) {
		bet := domain.GetCountingBetAmount(0.0, -3, domain.BJCountingKO, 10000)
		assert.Equal(t, 50, bet)
	})
	t.Run("KO count 2 returns 2x base", func(t *testing.T) {
		bet := domain.GetCountingBetAmount(0.0, 2, domain.BJCountingKO, 10000)
		assert.Equal(t, 100, bet)
	})
	t.Run("KO count 3 returns 4x base", func(t *testing.T) {
		bet := domain.GetCountingBetAmount(0.0, 3, domain.BJCountingKO, 10000)
		assert.Equal(t, 200, bet)
	})
	t.Run("KO count 4 returns 8x base", func(t *testing.T) {
		bet := domain.GetCountingBetAmount(0.0, 4, domain.BJCountingKO, 10000)
		assert.Equal(t, 400, bet)
	})
	t.Run("KO count >= 5 returns 16x base", func(t *testing.T) {
		bet := domain.GetCountingBetAmount(0.0, 5, domain.BJCountingKO, 10000)
		assert.Equal(t, 800, bet)
	})

	// Chip clamping
	t.Run("bet clamped to available chips", func(t *testing.T) {
		bet := domain.GetCountingBetAmount(5.0, 0, domain.BJCountingHiLo, 150)
		// 800 clamped to 150, rounded to 150
		assert.Equal(t, 150, bet)
	})
	t.Run("bet clamped and rounded down to BJMinBet multiple", func(t *testing.T) {
		bet := domain.GetCountingBetAmount(5.0, 0, domain.BJCountingHiLo, 155)
		// 800 clamped to 155, rounded to 150
		assert.Equal(t, 150, bet)
	})
	t.Run("available chips less than BJMinBet returns 0", func(t *testing.T) {
		bet := domain.GetCountingBetAmount(5.0, 0, domain.BJCountingHiLo, 5)
		assert.Equal(t, 0, bet)
	})

	// Zen Count (balanced) uses trueCount
	t.Run("Zen count 3 returns 4x base", func(t *testing.T) {
		bet := domain.GetCountingBetAmount(3.0, 10, domain.BJCountingZen, 10000)
		assert.Equal(t, 200, bet)
	})

	// Omega II (balanced) uses trueCount
	t.Run("Omega II count 4 returns 8x base", func(t *testing.T) {
		bet := domain.GetCountingBetAmount(4.0, 10, domain.BJCountingOmegaII, 10000)
		assert.Equal(t, 400, bet)
	})
}

func TestShouldTakeInsurance(t *testing.T) {
	// Balanced system: uses trueCount
	t.Run("balanced TC >= 3 takes insurance", func(t *testing.T) {
		assert.True(t, domain.ShouldTakeInsurance(3.0, 0, domain.BJCountingHiLo))
	})
	t.Run("balanced TC > 3 takes insurance", func(t *testing.T) {
		assert.True(t, domain.ShouldTakeInsurance(5.0, 0, domain.BJCountingHiLo))
	})
	t.Run("balanced TC < 3 declines insurance", func(t *testing.T) {
		assert.False(t, domain.ShouldTakeInsurance(2.9, 0, domain.BJCountingHiLo))
	})
	t.Run("balanced TC = 0 declines insurance", func(t *testing.T) {
		assert.False(t, domain.ShouldTakeInsurance(0.0, 0, domain.BJCountingHiLo))
	})
	t.Run("balanced TC negative declines insurance", func(t *testing.T) {
		assert.False(t, domain.ShouldTakeInsurance(-2.0, 0, domain.BJCountingHiLo))
	})

	// KO: uses runningCount
	t.Run("KO RC >= 3 takes insurance", func(t *testing.T) {
		assert.True(t, domain.ShouldTakeInsurance(0.0, 3, domain.BJCountingKO))
	})
	t.Run("KO RC > 3 takes insurance", func(t *testing.T) {
		assert.True(t, domain.ShouldTakeInsurance(0.0, 5, domain.BJCountingKO))
	})
	t.Run("KO RC < 3 declines insurance", func(t *testing.T) {
		assert.False(t, domain.ShouldTakeInsurance(0.0, 2, domain.BJCountingKO))
	})
	t.Run("KO RC = 0 declines insurance", func(t *testing.T) {
		assert.False(t, domain.ShouldTakeInsurance(0.0, 0, domain.BJCountingKO))
	})
	t.Run("KO RC negative declines insurance", func(t *testing.T) {
		assert.False(t, domain.ShouldTakeInsurance(0.0, -1, domain.BJCountingKO))
	})

	// Zen Count (balanced)
	t.Run("Zen TC >= 3 takes insurance", func(t *testing.T) {
		assert.True(t, domain.ShouldTakeInsurance(3.0, 10, domain.BJCountingZen))
	})
	t.Run("Zen TC < 3 declines insurance", func(t *testing.T) {
		assert.False(t, domain.ShouldTakeInsurance(2.0, 10, domain.BJCountingZen))
	})

	// Omega II (balanced)
	t.Run("Omega II TC >= 3 takes insurance", func(t *testing.T) {
		assert.True(t, domain.ShouldTakeInsurance(4.0, 10, domain.BJCountingOmegaII))
	})
	t.Run("Omega II TC < 3 declines insurance", func(t *testing.T) {
		assert.False(t, domain.ShouldTakeInsurance(1.0, 10, domain.BJCountingOmegaII))
	})
}
