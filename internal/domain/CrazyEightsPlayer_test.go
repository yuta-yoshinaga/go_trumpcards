//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestNewCrazyEightsPlayer(t *testing.T) {
	t.Run("human player", func(t *testing.T) {
		p := domain.NewCrazyEightsPlayer(true)
		assert.True(t, p.GetIsHuman())
		assert.Equal(t, 0, p.GetRoundScore())
		assert.Equal(t, 0, p.GetCumulativeScore())
		assert.Equal(t, 0, p.GetCardsSize())
	})

	t.Run("CPU player", func(t *testing.T) {
		p := domain.NewCrazyEightsPlayer(false)
		assert.False(t, p.GetIsHuman())
		assert.Equal(t, 0, p.GetRoundScore())
		assert.Equal(t, 0, p.GetCumulativeScore())
	})
}

func TestCrazyEightsPlayer_RoundScoreGetterSetter(t *testing.T) {
	p := domain.NewCrazyEightsPlayer(false)
	p.SetRoundScore(42)
	assert.Equal(t, 42, p.GetRoundScore())
	p.SetRoundScore(0)
	assert.Equal(t, 0, p.GetRoundScore())
}

func TestCrazyEightsPlayer_CumulativeScoreGetterSetter(t *testing.T) {
	p := domain.NewCrazyEightsPlayer(false)
	p.SetCumulativeScore(100)
	assert.Equal(t, 100, p.GetCumulativeScore())
	p.SetCumulativeScore(0)
	assert.Equal(t, 0, p.GetCumulativeScore())
}

func TestCrazyEightsPlayer_CommitRoundScore(t *testing.T) {
	p := domain.NewCrazyEightsPlayer(false)
	p.SetRoundScore(30)
	p.CommitRoundScore()
	assert.Equal(t, 30, p.GetCumulativeScore())

	p.SetRoundScore(20)
	p.CommitRoundScore()
	assert.Equal(t, 50, p.GetCumulativeScore())
}

func TestCrazyEightsPlayer_ResetRound(t *testing.T) {
	p := domain.NewCrazyEightsPlayer(true)
	p.SetRoundScore(30)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	p.SetIsFinished(true)

	p.ResetRound()

	assert.Equal(t, 0, p.GetRoundScore())
	assert.Equal(t, 0, p.GetCardsSize())
	assert.False(t, p.GetIsFinished())
}
