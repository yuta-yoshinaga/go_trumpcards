//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestNewTarneebPlayer(t *testing.T) {
	p := domain.NewTarneebPlayer(true, 0)
	assert.True(t, p.GetIsHuman())
	assert.Equal(t, 0, p.GetTeam())
	assert.Equal(t, -1, p.GetBid())
}

func TestTarneebPlayer_SetBid(t *testing.T) {
	p := domain.NewTarneebPlayer(false, 1)
	p.SetBid(7)
	assert.Equal(t, 7, p.GetBid())
}

func TestTarneebPlayer_ResetRound(t *testing.T) {
	p := domain.NewTarneebPlayer(true, 0)
	p.SetBid(9)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	p.SetRoundScore(42)
	p.AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 5, false)})

	p.ResetRound()
	assert.Equal(t, -1, p.GetBid())
	assert.Equal(t, 0, p.GetRoundScore())
	assert.Equal(t, 0, p.GetCardsSize())
	assert.Equal(t, 0, p.GetTrickCount())
}

func TestTarneebPlayer_JSONRoundTrip(t *testing.T) {
	p := domain.NewTarneebPlayer(true, 1)
	p.SetBid(8)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
	p.SetRoundScore(5)
	p.CommitRoundScore()

	data, err := json.Marshal(p)
	require.NoError(t, err)
	var got domain.TarneebPlayer
	require.NoError(t, json.Unmarshal(data, &got))
	assert.True(t, got.GetIsHuman())
	assert.Equal(t, 1, got.GetTeam())
	assert.Equal(t, 8, got.GetBid())
	assert.Equal(t, 5, got.GetCumulativeScore())
	assert.Equal(t, 1, got.GetCardsSize())
}
