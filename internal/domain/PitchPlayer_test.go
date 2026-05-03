//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestNewPitchPlayer(t *testing.T) {
	p := domain.NewPitchPlayer(true)
	assert.True(t, p.GetIsHuman())
	assert.Equal(t, -1, p.GetBid())
	assert.Equal(t, 0, p.GetCardsSize())
}

func TestPitchPlayer_BidGetSet(t *testing.T) {
	p := domain.NewPitchPlayer(false)
	p.SetBid(3)
	assert.Equal(t, 3, p.GetBid())
	p.SetBid(0)
	assert.Equal(t, 0, p.GetBid())
}

func TestPitchPlayer_ResetRound(t *testing.T) {
	p := domain.NewPitchPlayer(false)
	p.SetBid(4)
	p.SetRoundScore(5)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	p.AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)})
	p.SetIsFinished(true)

	p.ResetRound()

	assert.Equal(t, -1, p.GetBid())
	assert.Equal(t, 0, p.GetRoundScore())
	assert.Equal(t, 0, p.GetCardsSize())
	assert.Equal(t, 0, p.GetTrickCount())
	assert.False(t, p.GetIsFinished())
}

func TestPitchPlayer_JSONRoundTrip(t *testing.T) {
	p := domain.NewPitchPlayer(true)
	p.SetBid(3)
	p.SetRoundScore(2)
	p.SetCumulativeScore(5)
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))
	p.AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignClover, 9, false)})

	data, err := json.Marshal(p)
	assert.NoError(t, err)

	var clone domain.PitchPlayer
	assert.NoError(t, json.Unmarshal(data, &clone))
	assert.True(t, clone.GetIsHuman())
	assert.Equal(t, 3, clone.GetBid())
	assert.Equal(t, 2, clone.GetRoundScore())
	assert.Equal(t, 5, clone.GetCumulativeScore())
	assert.Equal(t, 1, clone.GetCardsSize())
	assert.Equal(t, 1, clone.GetTrickCount())
}
