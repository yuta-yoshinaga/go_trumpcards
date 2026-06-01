//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestNewMacauPlayer(t *testing.T) {
	t.Run("human player", func(t *testing.T) {
		p := domain.NewMacauPlayer(true)
		assert.True(t, p.GetIsHuman())
		assert.Equal(t, 0, p.GetRoundScore())
		assert.Equal(t, 0, p.GetCumulativeScore())
		assert.Equal(t, 0, p.GetCardsSize())
		assert.False(t, p.GetHasDeclared())
	})

	t.Run("CPU player", func(t *testing.T) {
		p := domain.NewMacauPlayer(false)
		assert.False(t, p.GetIsHuman())
		assert.Equal(t, 0, p.GetRoundScore())
		assert.Equal(t, 0, p.GetCumulativeScore())
	})
}

func TestMacauPlayer_HasDeclaredGetterSetter(t *testing.T) {
	p := domain.NewMacauPlayer(false)
	assert.False(t, p.GetHasDeclared())
	p.SetHasDeclared(true)
	assert.True(t, p.GetHasDeclared())
	p.SetHasDeclared(false)
	assert.False(t, p.GetHasDeclared())
}

func TestMacauPlayer_CommitRoundScore(t *testing.T) {
	p := domain.NewMacauPlayer(false)
	p.SetRoundScore(30)
	p.CommitRoundScore()
	assert.Equal(t, 30, p.GetCumulativeScore())

	p.SetRoundScore(20)
	p.CommitRoundScore()
	assert.Equal(t, 50, p.GetCumulativeScore())
}

func TestMacauPlayer_ResetRound(t *testing.T) {
	p := domain.NewMacauPlayer(true)
	p.SetRoundScore(30)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	p.SetIsFinished(true)
	p.SetHasDeclared(true)

	p.ResetRound()

	assert.Equal(t, 0, p.GetRoundScore())
	assert.Equal(t, 0, p.GetCardsSize())
	assert.False(t, p.GetIsFinished())
	assert.False(t, p.GetHasDeclared())
}
