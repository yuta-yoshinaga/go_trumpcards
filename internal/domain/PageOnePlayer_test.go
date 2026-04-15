//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestNewPageOnePlayer(t *testing.T) {
	t.Run("human player", func(t *testing.T) {
		p := domain.NewPageOnePlayer(true)
		assert.True(t, p.GetIsHuman())
		assert.Equal(t, 0, p.GetRoundScore())
		assert.Equal(t, 0, p.GetCumulativeScore())
		assert.Equal(t, 0, p.GetCardsSize())
		assert.False(t, p.GetHasDeclared())
	})

	t.Run("CPU player", func(t *testing.T) {
		p := domain.NewPageOnePlayer(false)
		assert.False(t, p.GetIsHuman())
	})
}

func TestPageOnePlayer_RoundScore(t *testing.T) {
	p := domain.NewPageOnePlayer(false)
	p.SetRoundScore(42)
	assert.Equal(t, 42, p.GetRoundScore())
}

func TestPageOnePlayer_CommitRoundScore(t *testing.T) {
	p := domain.NewPageOnePlayer(false)
	p.SetRoundScore(30)
	p.CommitRoundScore()
	assert.Equal(t, 30, p.GetCumulativeScore())

	p.SetRoundScore(20)
	p.CommitRoundScore()
	assert.Equal(t, 50, p.GetCumulativeScore())
}

func TestPageOnePlayer_ResetRound(t *testing.T) {
	p := domain.NewPageOnePlayer(true)
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

func TestPageOnePlayer_HasDeclared(t *testing.T) {
	p := domain.NewPageOnePlayer(true)
	assert.False(t, p.GetHasDeclared())
	p.SetHasDeclared(true)
	assert.True(t, p.GetHasDeclared())
}

func TestPageOnePlayer_JSONRoundTrip(t *testing.T) {
	p := domain.NewPageOnePlayer(true)
	p.SetRoundScore(10)
	p.SetCumulativeScore(25)
	p.SetHasDeclared(true)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))

	data, err := json.Marshal(p)
	require.NoError(t, err)

	out := domain.NewPageOnePlayer(false)
	require.NoError(t, json.Unmarshal(data, out))

	assert.True(t, out.GetIsHuman())
	assert.Equal(t, 10, out.GetRoundScore())
	assert.Equal(t, 25, out.GetCumulativeScore())
	assert.True(t, out.GetHasDeclared())
	assert.Equal(t, 1, out.GetCardsSize())
}
