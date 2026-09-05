//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestNewBatakPlayer(t *testing.T) {
	t.Run("human", func(t *testing.T) {
		p := domain.NewBatakPlayer(true)
		assert.True(t, p.GetIsHuman())
		assert.Equal(t, -1, p.GetBid())
		assert.Equal(t, 0, p.GetTrickCount())
		assert.Equal(t, 0, p.GetRoundScore())
		assert.Equal(t, 0, p.GetCumulativeScore())
	})

	t.Run("cpu", func(t *testing.T) {
		p := domain.NewBatakPlayer(false)
		assert.False(t, p.GetIsHuman())
	})
}

func TestBatakPlayer_BidGetterSetter(t *testing.T) {
	p := domain.NewBatakPlayer(false)
	p.SetBid(5)
	assert.Equal(t, 5, p.GetBid())
	p.SetBid(0)
	assert.Equal(t, 0, p.GetBid())
}

func TestBatakPlayer_CommitRoundScore(t *testing.T) {
	p := domain.NewBatakPlayer(false)
	p.SetRoundScore(5)
	p.CommitRoundScore()
	assert.Equal(t, 5, p.GetCumulativeScore())

	p.SetRoundScore(-5)
	p.CommitRoundScore()
	assert.Equal(t, 0, p.GetCumulativeScore())
}

func TestBatakPlayer_ResetRound(t *testing.T) {
	p := domain.NewBatakPlayer(true)
	p.SetBid(5)
	p.SetRoundScore(5)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	p.AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 2, false)})
	p.SetIsFinished(true)

	p.ResetRound()

	assert.Equal(t, -1, p.GetBid())
	assert.Equal(t, 0, p.GetRoundScore())
	assert.Equal(t, 0, p.GetCardsSize())
	assert.Equal(t, 0, p.GetTrickCount())
	assert.False(t, p.GetIsFinished())
}

func TestBatakPlayer_JSONRoundTrip(t *testing.T) {
	p := domain.NewBatakPlayer(true)
	p.SetBid(5)
	p.SetRoundScore(5)
	p.SetCumulativeScore(10)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var p2 domain.BatakPlayer
	require.NoError(t, json.Unmarshal(data, &p2))

	assert.Equal(t, p.GetBid(), p2.GetBid())
	assert.Equal(t, p.GetRoundScore(), p2.GetRoundScore())
	assert.Equal(t, p.GetCumulativeScore(), p2.GetCumulativeScore())
	assert.Equal(t, p.GetIsHuman(), p2.GetIsHuman())
	assert.Equal(t, p.GetCardsSize(), p2.GetCardsSize())
}

func TestBatakPlayer_UnmarshalJSON_Invalid(t *testing.T) {
	var p domain.BatakPlayer
	err := json.Unmarshal([]byte("not json"), &p)
	assert.Error(t, err)
}

func TestBatakPlayer_UnmarshalJSON_Empty(t *testing.T) {
	var p domain.BatakPlayer
	err := json.Unmarshal([]byte("{}"), &p)
	require.NoError(t, err)
	// empty JSON should produce defaulted player (not panic on subsequent calls)
	assert.False(t, p.GetIsHuman())
}
