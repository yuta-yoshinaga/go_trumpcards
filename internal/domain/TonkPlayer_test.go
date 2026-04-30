//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestNewTonkPlayer(t *testing.T) {
	t.Run("human player", func(t *testing.T) {
		p := domain.NewTonkPlayer(true)
		assert.True(t, p.GetIsHuman())
		assert.Equal(t, 0, p.GetRoundScore())
		assert.Equal(t, 0, p.GetCumulativeScore())
		assert.Equal(t, 0, p.GetCardsSize())
	})

	t.Run("CPU player", func(t *testing.T) {
		p := domain.NewTonkPlayer(false)
		assert.False(t, p.GetIsHuman())
	})
}

func TestTonkPlayer_RoundScore(t *testing.T) {
	p := domain.NewTonkPlayer(false)
	p.SetRoundScore(42)
	assert.Equal(t, 42, p.GetRoundScore())
	p.CommitRoundScore()
	assert.Equal(t, 42, p.GetCumulativeScore())
}

func TestTonkPlayer_ResetRound(t *testing.T) {
	p := domain.NewTonkPlayer(true)
	p.SetRoundScore(30)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	p.SetIsFinished(true)

	p.ResetRound()

	assert.Equal(t, 0, p.GetRoundScore())
	assert.Equal(t, 0, p.GetCardsSize())
	assert.False(t, p.GetIsFinished())
}

func TestTonkPlayer_JSONRoundtrip(t *testing.T) {
	p := domain.NewTonkPlayer(true)
	p.SetRoundScore(15)
	p.SetCumulativeScore(60)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var restored domain.TonkPlayer
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.True(t, restored.GetIsHuman())
	assert.Equal(t, 15, restored.GetRoundScore())
	assert.Equal(t, 60, restored.GetCumulativeScore())
	assert.Equal(t, 1, restored.GetCardsSize())
}

func TestTonkPlayer_UnmarshalNilGamePlayer(t *testing.T) {
	// Empty JSON → fields default
	var p domain.TonkPlayer
	require.NoError(t, json.Unmarshal([]byte(`{}`), &p))
	assert.False(t, p.GetIsHuman())
}
